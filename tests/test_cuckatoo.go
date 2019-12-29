package main
/*
#cgo LDFLAGS: -L../lib/cuckoo/target/release -lcuckoo
#include "../lib/cuckoo.h"
#include <stdio.h>
#include <stdlib.h>
*/
import "C"
import (
    "encoding/binary"
    "fmt"
    "github.com/Qitmeer/go-opencl/cl"
    "github.com/Qitmeer/qitmeer/common/hash"
    "github.com/Qitmeer/qitmeer/crypto/cuckoo"
    "github.com/Qitmeer/qitmeer/crypto/cuckoo/siphash"
    "log"
    "os"
    "github.com/Qitmeer/qitmeer-miner/common"
    "github.com/Qitmeer/qitmeer-miner/core"
    "github.com/Qitmeer/qitmeer-miner/kernel"
    "github.com/Qitmeer/qitmeer-miner/symbols/qitmeer"
    "sort"
    "unsafe"
)
const RES_BUFFER_SIZE = 4000000
const LOCAL_WORK_SIZE = 256
const GLOBAL_WORK_SIZE = 1024 * LOCAL_WORK_SIZE
const SetCnt = 1
const Trim = 2
const Extract = 3
//init the config file
func main()  {
    var typ = common.DevicesTypesForGPUMining
    clDevices := common.GetDevices(typ,"")
    deviceMiner := Cuckatoo{}
    for i, device := range clDevices {
        q := make(chan os.Signal)
        deviceMiner.Init(i,device,false,q,&common.GlobalConfig{})
    }
    deviceMiner.InitDevice()
    deviceMiner.Mine()
}

type Cuckatoo struct {
    core.Device
    ClearBytes                 []byte
    EdgesObj                   *cl.MemObject
    EdgesBytes                 []byte
    CountersObj                *cl.MemObject
    EdgesIndexBytes            []byte
    ResultObj                  *cl.MemObject
    ResultBytes                []byte
    NoncesObj                  *cl.MemObject
    NoncesBytes                []byte
    Nonces                     []uint32
    NodesObj                   *cl.MemObject
    NodesBytes                 []byte
    Edges                      []uint32
    CreateEdgeKernel           *cl.Kernel
    Trimmer01Kernel            *cl.Kernel
    Trimmer02Kernel            *cl.Kernel
    RecoveryKernel             *cl.Kernel
    Work                       qitmeer.QitmeerWork
    Transactions               map[int][]qitmeer.Transactions
    header                     qitmeer.MinerBlockData
}
const edges_bits = 29
var el_count = (1024 * 1024 * 512 / 32) << (edges_bits - 29)
var current_mode = SetCnt
var current_uorv = 0
var trims = 128 << (edges_bits - 29)
func (this *Cuckatoo) InitDevice() {
    this.Device.InitDevice()
    if !this.IsValid {
        return
    }
    var err error
    this.Program, err = this.Context.CreateProgramWithSource([]string{kernel.CuckatooKernel})
    if err != nil {
        log.Println("-", this.MinerId, this.DeviceName, err)
        this.IsValid = false
        return
    }

    err = this.Program.BuildProgram([]*cl.Device{this.ClDevice}, "")
    if err != nil {
        log.Println("-", this.MinerId, err)
        this.IsValid = false
        return
    }

    this.InitKernelAndParam()

}

func (this *Cuckatoo) Mine() {

    defer this.Release()


    for {
        var err error
        text := "helloworld"


        for {

            for nonce := 0;nonce <= 0 ;nonce++{
                if this.HasNewWork {
                    break
                }
                text = fmt.Sprintf("%s%d",text,nonce)
                hdrkey := hash.DoubleHashH([]byte(text))
                sip := siphash.Newsip(hdrkey[:16])
                if nonce == 0{
                    sip.V = [4]uint64{
                        0x27580576fe290177,
                        0xf9ea9b2031f4e76e,
                        0x1663308c8607868f,
                        0xb88839b0fa180d0e,
                    }
                } else{
                    sip.V = [4]uint64{
                        0x5c0348cfc71b5ce6,
                        0xbf4141b92a45e49,
                        0x7282d7893f658b88,
                        0x61525294db9b617f,
                    }
                }
                this.InitParamData()
                err = this.CreateEdgeKernel.SetArg(0,uint64(sip.V[0]))
                if err != nil {
                    log.Println("-", this.MinerId, err)
                    this.IsValid = false
                    return
                }
                err = this.CreateEdgeKernel.SetArg(1,uint64(sip.V[1]))
                if err != nil {
                    log.Println("-", this.MinerId, err)
                    this.IsValid = false
                    return
                }
                err = this.CreateEdgeKernel.SetArg(2,uint64(sip.V[2]))
                if err != nil {
                    log.Println("-", this.MinerId, err)
                    this.IsValid = false
                    return
                }
                err = this.CreateEdgeKernel.SetArg(3,uint64(sip.V[3]))
                if err != nil {
                    log.Println("-", this.MinerId, err)
                    this.IsValid = false
                    return
                }
                for l:=uint32(0) ;l<uint32(trims);l++{
                    current_uorv = int(l & 1)
                    current_mode = SetCnt
                    err = this.CreateEdgeKernel.SetArg(7,uint32(current_mode))
                    err = this.CreateEdgeKernel.SetArg(8,uint32(current_uorv))
                    this.Enq(8)
                    current_mode = Trim
                    if int(l) == (trims - 1) {
                        current_mode = Extract
                    }
                    err = this.CreateEdgeKernel.SetArg(7,uint32(current_mode))
                    this.Enq(8)
                    _,err = this.CommandQueue.EnqueueFillBuffer(this.CountersObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,el_count*4,nil)

                }
                this.ResultBytes = make([]byte,RES_BUFFER_SIZE*4)
                _,_ = this.CommandQueue.EnqueueReadBufferByte(this.ResultObj,true,0,this.ResultBytes,nil)
                leftEdges := binary.LittleEndian.Uint32(this.ResultBytes[4:8])
                log.Println(fmt.Sprintf("Trimmed to %d edges",leftEdges))

                p := C.malloc(C.size_t(len(this.ResultBytes)))
                defer C.free(p)
                // copy the data into the buffer, by converting it to a Go array
                cBuf := (*[1 << 30]byte)(p)
                copy(cBuf[:], this.ResultBytes)
                noncesBytes := make([]byte,42*4)
                C.search_circle((*C.uint)(p),C.size_t(len(this.ResultBytes)),(*C.uint)(unsafe.Pointer(&noncesBytes[0])))
                //log.Println(noncesBytes)
                nonces := make([]uint32,0)
                for jj := 0;jj < len(noncesBytes);jj+=4{
                    nonces = append(nonces,binary.LittleEndian.Uint32(noncesBytes[jj:jj+4]))
                }
                sort.Slice(nonces, func(i, j int) bool {
                    return nonces[i]<nonces[j]
                })
                log.Println(nonces)
                if err := cuckoo.VerifyCuckatoo(hdrkey[:16],nonces,edges_bits);err == nil{
                    log.Println("[Found Circle Success]")
                } else{
                    log.Println(err)
                }

            }
            break

        }
        break
    }
}

func (this *Cuckatoo) Enq(num int) {
    offset := 0
    for j:=0;j<num;j++{
        offset = j * GLOBAL_WORK_SIZE
        //log.Println(j,offset)
        // 2 ^ 24 2 ^ 11 * 2 ^ 8 * 2 * 2 ^ 4 11+8+1+4=24
        if _, err := this.CommandQueue.EnqueueNDRangeKernel(this.CreateEdgeKernel, []int{offset}, []int{GLOBAL_WORK_SIZE}, []int{LOCAL_WORK_SIZE}, nil); err != nil {
            log.Println("CreateEdgeKernel-1058", this.MinerId,err)
            return
        }
    }
}


func (this *Cuckatoo) SubmitShare(substr chan string) {
    this.Device.SubmitShare(substr)
}

func (this *Cuckatoo) Release() {
    this.Context.Release()
    this.Program.Release()
    this.CreateEdgeKernel.Release()
    this.EdgesObj.Release()
    this.CountersObj.Release()
    this.ResultObj.Release()
}

func (this *Cuckatoo) InitParamData() {
    var err error
    this.ClearBytes = make([]byte,4)
    allBytes := []byte{255,255,255,255}
    _,err = this.CommandQueue.EnqueueFillBuffer(this.CountersObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,el_count*4,nil)
    if err != nil {
        log.Println("-", this.MinerId, err)
        this.IsValid = false
        return
    }
    _,err = this.CommandQueue.EnqueueFillBuffer(this.EdgesObj,unsafe.Pointer(&allBytes[0]),4,0,el_count*4*8,nil)
    if err != nil {
        log.Println("-", this.MinerId, err)
        this.IsValid = false
        return
    }
    _,err = this.CommandQueue.EnqueueFillBuffer(this.ResultObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,RES_BUFFER_SIZE*4,nil)
    if err != nil {
        log.Println("-", this.MinerId, err)
        this.IsValid = false
        return
    }

    err = this.CreateEdgeKernel.SetArgBuffer(4,this.EdgesObj)
    err = this.CreateEdgeKernel.SetArgBuffer(5,this.CountersObj)
    err = this.CreateEdgeKernel.SetArgBuffer(6,this.ResultObj)
    err = this.CreateEdgeKernel.SetArg(7,uint32(current_mode))
    err = this.CreateEdgeKernel.SetArg(8,uint32(current_uorv))

    if err != nil {
        log.Println("-", this.MinerId, err)
        this.IsValid = false
        return
    }
}

func (this *Cuckatoo) InitKernelAndParam() {
    var err error
    this.CreateEdgeKernel, err = this.Program.CreateKernel("LeanRound")
    if err != nil {
        log.Println("-", this.MinerId, err)
        this.IsValid = false
        return
    }

    this.EdgesObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, el_count*4*8)
    if err != nil {
        log.Println("-", this.MinerId, err)
        this.IsValid = false
        return
    }
    this.CountersObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, el_count*4)
    if err != nil {
        log.Println("-", this.MinerId, err)
        this.IsValid = false
        return
    }
    this.ResultObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, RES_BUFFER_SIZE*4)
    if err != nil {
        log.Println("-", this.MinerId, err)
        this.IsValid = false
        return
    }

}