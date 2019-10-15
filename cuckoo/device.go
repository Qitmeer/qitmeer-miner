package cuckoo

import (
    "encoding/binary"
    "fmt"
    "github.com/Qitmeer/go-opencl/cl"
    "github.com/Qitmeer/qitmeer/common/hash"
    Cuckoo "github.com/Qitmeer/qitmeer/crypto/cuckoo"
    "github.com/Qitmeer/qitmeer/crypto/cuckoo/siphash"
    `log`
    "qitmeer-miner/common"
    "qitmeer-miner/core"
    `qitmeer-miner/kernel`
    "sort"
    `strings`
    "sync/atomic"
    "unsafe"
)

type Device struct {
    core.Device
    ClearBytes	[]byte
    EdgesObj              *cl.MemObject
    EdgesBytes            []byte
    DestinationEdgesCountObj              *cl.MemObject
    DestinationEdgesCountBytes            []byte
    EdgesIndexObj         *cl.MemObject
    EdgesIndexBytes       []byte
    DestinationEdgesObj   *cl.MemObject
    DestinationEdgesBytes []byte
    NoncesObj             *cl.MemObject
    NoncesBytes           []byte
    Nonces           []uint32
    NodesObj              *cl.MemObject
    NodesBytes            []byte
    Edges                 []uint32
    CreateEdgeKernel      *cl.Kernel
    Trimmer01Kernel       *cl.Kernel
    Trimmer02Kernel       *cl.Kernel
    RecoveryKernel        *cl.Kernel
}

var (
    EdgeBits = uint8(24)
    Step = 1 << (uint(EdgeBits)-20)
    workgroup = 1 << 8
    localsize = 1 << 12
    Nedge    = 1 << EdgeBits    //number of edges：
    Edgemask  = uint64(Nedge - 1)
)
func (this *Device) InitDevice() {
    this.Device.InitDevice()
    if !this.IsValid {
        return
    }
    var err error
    kernelStr := strings.ReplaceAll(kernel.CuckarooKernelNew,"{{edge_bits}}",fmt.Sprintf("%d",EdgeBits))
    kernelStr = strings.Replace(kernelStr,"{{step}}",fmt.Sprintf("%d",Step),1)
    kernelStr = strings.ReplaceAll(kernelStr,"{{group}}",fmt.Sprintf("%d",workgroup))
    this.Program, err = this.Context.CreateProgramWithSource([]string{kernelStr})
    if err != nil {
        common.MinerLoger.Infof("-%d,%s%v", this.MinerId, this.DeviceName, err)
        this.IsValid = false
        return
    }
    err = this.Program.BuildProgram([]*cl.Device{this.ClDevice}, "")
    if err != nil {
        common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        this.IsValid = false
        return
    }

    this.InitKernelAndParam()

}

func (this *Device) Mine() {

    defer this.Release()
    var err error
    nonce := uint64(13)
    for {
        if this.HasNewWork {
            break
        }
        str := fmt.Sprintf("helloworld%d",nonce)
        fmt.Printf("第 %d 次找环 %s \n",(nonce+1),str)
        nonce++
        hdrkey := hash.HashH([]byte(str))
        sip := siphash.Newsip(hdrkey[:])
        this.InitParamData()
        err = this.CreateEdgeKernel.SetArg(0,uint64(sip.V[0]))
        if err != nil {
            common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
            this.IsValid = false
            return
        }
        err = this.CreateEdgeKernel.SetArg(1,uint64(sip.V[1]))
        if err != nil {
            common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
            this.IsValid = false
            return
        }
        err = this.CreateEdgeKernel.SetArg(2,uint64(sip.V[2]))
        if err != nil {
            common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
            this.IsValid = false
            return
        }
        err = this.CreateEdgeKernel.SetArg(3,uint64(sip.V[3]))
        if err != nil {
            common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
            this.IsValid = false
            return
        }
        err = this.Trimmer01Kernel.SetArg(0,uint64(sip.V[0]))
        if err != nil {
            common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
            this.IsValid = false
            return
        }
        err = this.Trimmer01Kernel.SetArg(1,uint64(sip.V[1]))
        if err != nil {
            common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
            this.IsValid = false
            return
        }
        err = this.Trimmer01Kernel.SetArg(2,uint64(sip.V[2]))
        if err != nil {
            common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
            this.IsValid = false
            return
        }
        err = this.Trimmer01Kernel.SetArg(3,uint64(sip.V[3]))
        if err != nil {
            common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
            this.IsValid = false
            return
        }
        err = this.Trimmer02Kernel.SetArg(0,uint64(sip.V[0]))
        if err != nil {
            common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
            this.IsValid = false
            return
        }
        err = this.Trimmer02Kernel.SetArg(1,uint64(sip.V[1]))
        if err != nil {
            common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
            this.IsValid = false
            return
        }
        err = this.Trimmer02Kernel.SetArg(2,uint64(sip.V[2]))
        if err != nil {
            common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
            this.IsValid = false
            return
        }
        err = this.Trimmer02Kernel.SetArg(3,uint64(sip.V[3]))
        if err != nil {
            common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
            this.IsValid = false
            return
        }
        // 2 ^ 24 2 ^ 11 * 2 ^ 8 * 2 * 2 ^ 4 11+8+1+4=24  12 + 8 + 4
        if this.Event, err = this.CommandQueue.EnqueueNDRangeKernel(this.CreateEdgeKernel, []int{0}, []int{localsize*workgroup}, []int{workgroup}, nil); err != nil {
            common.MinerLoger.Infof("CreateEdgeKernel-1058%d,%v", this.MinerId,err)
            return
        }
        this.Event.Release()
        for i:= 0;i<120;i++{
            if this.Event, err = this.CommandQueue.EnqueueNDRangeKernel(this.Trimmer01Kernel, []int{0}, []int{localsize*workgroup}, []int{workgroup}, nil); err != nil {
                common.MinerLoger.Infof("Trimmer01Kernel-1058%d,%v", this.MinerId,err)
                return
            }
            this.Event.Release()
            _ = this.CommandQueue.Finish()
        }
        if this.Event, err = this.CommandQueue.EnqueueNDRangeKernel(this.Trimmer02Kernel, []int{0}, []int{localsize*workgroup}, []int{workgroup}, nil); err != nil {
            common.MinerLoger.Infof("Trimmer02Kernel-1058%d,%v", this.MinerId,err)
            return
        }
        this.Event.Release()
        this.DestinationEdgesCountBytes = make([]byte,8)
        this.Event,err = this.CommandQueue.EnqueueReadBufferByte(this.DestinationEdgesCountObj,true,0,this.DestinationEdgesCountBytes,nil)
        if err != nil {
            common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
            this.IsValid = false
            return
        }
        this.Event.Release()
        count := binary.LittleEndian.Uint32(this.DestinationEdgesCountBytes[4:8])
        if count < Cuckoo.ProofSize*2 {
            continue
        }
        this.DestinationEdgesBytes = make([]byte,count*4)
        this.Event,err = this.CommandQueue.EnqueueReadBufferByte(this.DestinationEdgesObj,true,0,this.DestinationEdgesBytes,nil)
        if err != nil {
            common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
            this.IsValid = false
            return
        }
        this.Event.Release()
        this.Edges = make([]uint32,0)
        for j:=0;j<len(this.DestinationEdgesBytes);j+=4{
            blockNonce := binary.LittleEndian.Uint32(this.DestinationEdgesBytes[j:j+4])
            u00 := siphash.SiphashPRF(&sip.V, uint64(blockNonce<<1))
            v00 := siphash.SiphashPRF(&sip.V, (uint64(blockNonce)<<1)|1)
            u := uint32(u00&Edgemask) << 1
            v := (uint32(v00&Edgemask) << 1) | 1
            this.Edges = append(this.Edges,u)
            this.Edges = append(this.Edges,v)
        }
        cg := CGraph{}
        cg.SetEdges(this.Edges,int(count))
        atomic.AddUint64(&this.AllDiffOneShares, 1)
        if !cg.FindSolutions(){
            continue
        }
        //fmt.Println(cg.GetNonceEdgesBytes(),"cg.GetNonceEdgesBytes()======")
        edges := cg.CycleEdges.GetData()
        this.Nonces = make([]uint32,0)
        //for j := 0;j<Cuckoo.ProofSize*4;j+=4{
        //    this.Nonces = append(this.Nonces,binary.LittleEndian.Uint32(this.NoncesBytes[j:j+4]))
        //}
        for j:=0;j<len(this.DestinationEdgesBytes);j+=4{
            blockNonce := binary.LittleEndian.Uint32(this.DestinationEdgesBytes[j:j+4])
            u00 := siphash.SiphashPRF(&sip.V, uint64(blockNonce<<1))
            v00 := siphash.SiphashPRF(&sip.V, (uint64(blockNonce)<<1)|1)
            u := uint32(u00&Edgemask) << 1
            v := (uint32(v00&Edgemask) << 1) | 1
            for i := 0;i < 42;i++{
                if (u== uint32(edges[i].Item1) && v == uint32(edges[i].Item2)) || (u== uint32(edges[i].Item2) && v == uint32(edges[i].Item1)) {
                    this.Nonces = append(this.Nonces,blockNonce)
                }
            }
        }
        //if cg.FindCycle(){
        //this.Event,err = this.CommandQueue.EnqueueWriteBufferByte(this.NodesObj,true,0,cg.GetNonceEdgesBytes(),nil)
        //if err != nil {
        //    common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        //    this.IsValid = false
        //    return
        //}
        //this.Event.Release()
        //if this.Event, err = this.CommandQueue.EnqueueNDRangeKernel(this.RecoveryKernel, []int{0}, []int{localsize*workgroup}, []int{workgroup}, nil); err != nil {
        //    common.MinerLoger.Infof("RecoveryKernel-1058%d,%v", this.MinerId,err)
        //    return
        //}
        //this.Event.Release()
        //this.NoncesBytes = make([]byte,4*Cuckoo.ProofSize)
        //this.Event,err = this.CommandQueue.EnqueueReadBufferByte(this.NoncesObj,true,0,this.NoncesBytes,nil)
        //if err != nil {
        //    common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        //    this.IsValid = false
        //    return
        //}
        //this.Event.Release()


        sort.Slice(this.Nonces, func(i, j int) bool {
            return this.Nonces[i] < this.Nonces[j]
        })
        this.AllDiffOneShares += 1
        err := Cuckoo.VerifyCuckaroo(hdrkey[:],this.Nonces[:],uint(EdgeBits))
        if err != nil{
            common.MinerLoger.Errorf("[error]Verify Error!%v",err)
            continue
        }

        log.Println("found 42 circle nonces",this.Nonces)
        break
    }
}


func (this *Device) Release() {
    this.Context.Release()
    this.Program.Release()
    this.CreateEdgeKernel.Release()
    this.Trimmer01Kernel.Release()
    this.Trimmer02Kernel.Release()
    this.EdgesObj.Release()
    this.EdgesIndexObj.Release()
    this.DestinationEdgesObj.Release()
    this.NoncesObj.Release()
    this.NodesObj.Release()
}

func (this *Device) InitParamData() {
    var err error
    this.ClearBytes = make([]byte,4)
    this.Event,err = this.CommandQueue.EnqueueFillBuffer(this.EdgesIndexObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,Nedge*4*2,nil)
    if err != nil {
        common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        this.IsValid = false
        return
    }
    this.Event.Release()
    this.Event,err = this.CommandQueue.EnqueueFillBuffer(this.EdgesObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,Nedge*2,nil)
    if err != nil {
        common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        this.IsValid = false
        return
    }
    this.Event.Release()
    this.Event,err = this.CommandQueue.EnqueueFillBuffer(this.DestinationEdgesObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,Nedge*2,nil)
    if err != nil {
        common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        this.IsValid = false
        return
    }
    this.Event.Release()
    this.Event,err = this.CommandQueue.EnqueueFillBuffer(this.NodesObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,Cuckoo.ProofSize*8,nil)
    if err != nil {
        common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        this.IsValid = false
        return
    }
    this.Event.Release()
    this.Event,err = this.CommandQueue.EnqueueFillBuffer(this.DestinationEdgesCountObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,8,nil)
    if err != nil {
        common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        this.IsValid = false
        return
    }
    this.Event.Release()
    this.Event,err = this.CommandQueue.EnqueueFillBuffer(this.NoncesObj,unsafe.Pointer(&this.ClearBytes[0]),4,0,Cuckoo.ProofSize*4,nil)
    if err != nil {
        common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        this.IsValid = false
        return
    }
    this.Event.Release()
    err = this.CreateEdgeKernel.SetArgBuffer(4,this.EdgesObj)
    err = this.CreateEdgeKernel.SetArgBuffer(5,this.EdgesIndexObj)

    err = this.Trimmer01Kernel.SetArgBuffer(4,this.EdgesObj)
    err = this.Trimmer01Kernel.SetArgBuffer(5,this.EdgesIndexObj)

    err = this.Trimmer02Kernel.SetArgBuffer(4,this.EdgesObj)
    err = this.Trimmer02Kernel.SetArgBuffer(5,this.EdgesIndexObj)
    err = this.Trimmer02Kernel.SetArgBuffer(6,this.DestinationEdgesObj)
    err = this.Trimmer02Kernel.SetArgBuffer(7,this.DestinationEdgesCountObj)

}

func (this *Device) InitKernelAndParam() {
    var err error
    this.CreateEdgeKernel, err = this.Program.CreateKernel("CreateEdges")
    if err != nil {
        common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        this.IsValid = false
        return
    }

    this.Trimmer01Kernel, err = this.Program.CreateKernel("Trimmer01")
    if err != nil {
        common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        this.IsValid = false
        return
    }

    this.Trimmer02Kernel, err = this.Program.CreateKernel("Trimmer02")
    if err != nil {
        common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        this.IsValid = false
        return
    }

    this.EdgesObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, Nedge*2)
    if err != nil {
        common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        this.IsValid = false
        return
    }
    this.DestinationEdgesObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, Nedge*2)
    if err != nil {
        common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        this.IsValid = false
        return
    }
    this.NodesObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, Cuckoo.ProofSize*4*2)
    if err != nil {
        common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        this.IsValid = false
        return
    }
    this.EdgesIndexObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, Nedge*4*2)
    if err != nil {
        common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        this.IsValid = false
        return
    }
    this.DestinationEdgesCountObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, 8)
    if err != nil {
        common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        this.IsValid = false
        return
    }
    this.NoncesObj, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, Cuckoo.ProofSize*4)
    if err != nil {
        common.MinerLoger.Infof("-%d,%v", this.MinerId, err)
        this.IsValid = false
        return
    }

}