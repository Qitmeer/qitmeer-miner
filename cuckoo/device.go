package cuckoo

import (
    "encoding/binary"
    "fmt"
    `github.com/Qitmeer/go-opencl/cl`
    `github.com/Qitmeer/qitmeer/crypto/cuckoo/siphash`
    "golang.org/x/crypto/blake2b"
    "log"
    "os"
    `qitmeer-miner/kernel`
    "sync"
    `time`
    "unsafe"
)
const(
    DUCK_SIZE_A = 129
    DUCK_SIZE_B = 83
    BUFFER_SIZE_A1 = DUCK_SIZE_A * 1024 * (4096 - 128) * 2
    BUFFER_SIZE_A2 = DUCK_SIZE_A * 1024 * 256 * 2
    BUFFER_SIZE_B = DUCK_SIZE_B * 1024 * 4096 * 2
    INDEX_SIZE = 256 * 256 * 4
)
type Device struct{
    DeviceName string
    HasNewWork bool
    AllDiffOneShares uint64
    AverageHashRate    float64
    MinerId          uint32
    Context          *cl.Context
    CommandQueue     *cl.CommandQueue
    LocalItemSize     int
    NonceOut     []byte
    BlockObj     *cl.MemObject
    BufferA1     *cl.MemObject //u32
    BufferA2     *cl.MemObject
    BufferI1     *cl.MemObject
    BufferI2     *cl.MemObject
    BufferB     *cl.MemObject
    BufferR     *cl.MemObject
    BufferNonces     *cl.MemObject
    Indexes     *cl.MemObject
    NonceOutObj     *cl.MemObject
    Kernels     map[string]*cl.Kernel
    Program     	*cl.Program
    ClDevice         *cl.Device
    Started          uint32
    GlobalItemSize int
    CurrentWorkID uint32
    Quit chan os.Signal //must init
    sync.Mutex
    Wg sync.WaitGroup
    EventList []*cl.Event
}

func (this *Device)Mine()  {
    var err error
    this.Context, err = cl.CreateContext([]*cl.Device{this.ClDevice})
    if err != nil {
        log.Println("-1", this.MinerId, err)
        return
    }
    this.CommandQueue, err = this.Context.CreateCommandQueue(this.ClDevice, 0)
    if err != nil {
        log.Println("-2", this.MinerId,  err)
    }
    this.Program, err = this.Context.CreateProgramWithSource([]string{kernel.CuckarooKernelBak})
    if err != nil {
        log.Println("-3", this.MinerId, this.DeviceName, err)
        return
    }

    err = this.Program.BuildProgram([]*cl.Device{this.ClDevice}, "")
    if err != nil {
        log.Println("-build", this.MinerId, err)
        return
    }

    this.Kernels = make(map[string]*cl.Kernel,0)
    this.Kernels["kernel_seed_a"], err = this.Program.CreateKernel("FluffySeed2A")
    this.Kernels["kernel_seed_b1"], err = this.Program.CreateKernel("FluffySeed2B")
    this.Kernels["kernel_seed_b2"], err = this.Program.CreateKernel("FluffySeed2B")
    this.Kernels["kernel_round1"], err = this.Program.CreateKernel("FluffyRound1")
    this.Kernels["kernel_round0"], err = this.Program.CreateKernel("FluffyRoundNO1")
    this.Kernels["kernel_round_na"], err = this.Program.CreateKernel("FluffyRoundNON")
    this.Kernels["kernel_round_nb"], err = this.Program.CreateKernel("FluffyRoundNON")
    this.Kernels["kernel_tail"], err = this.Program.CreateKernel("FluffyTailO")
    this.BufferA1, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, BUFFER_SIZE_A1)
    this.BufferA2, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, BUFFER_SIZE_A2)
    this.BufferB, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, BUFFER_SIZE_B)
    this.BufferI1, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, INDEX_SIZE)
    this.BufferI2, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, INDEX_SIZE)
    this.BufferR, err = this.Context.CreateEmptyBuffer(cl.MemReadOnly, 42*2)
    this.BufferNonces, err = this.Context.CreateEmptyBuffer(cl.MemReadWrite, INDEX_SIZE)

    for i:=0;i<2<<32;i++{
        str := fmt.Sprintf("hello world%d",i)
        hdrkey := blake2b.Sum256([]byte(str))

        sip := siphash.Newsip(hdrkey[:])
        //init kernels
        err = this.Kernels["kernel_seed_a"].SetArg(0,uint64(sip.V[0]))

        err = this.Kernels["kernel_seed_a"].SetArg(1,uint64(sip.V[1]))

        err = this.Kernels["kernel_seed_a"].SetArg(2,uint64(sip.V[2]))

        err = this.Kernels["kernel_seed_a"].SetArg(3,uint64(sip.V[3]))

        err = this.Kernels["kernel_seed_a"].SetArg(4,this.BufferB)
        err = this.Kernels["kernel_seed_a"].SetArg(5,this.BufferA1)
        err = this.Kernels["kernel_seed_a"].SetArg(6,this.BufferI1)
        //init kernels
        err = this.Kernels["kernel_seed_b1"].SetArg(0,this.BufferA1)

        err = this.Kernels["kernel_seed_b1"].SetArg(1,this.BufferA1)

        err = this.Kernels["kernel_seed_b1"].SetArg(2,this.BufferA2)

        err = this.Kernels["kernel_seed_b1"].SetArg(3,this.BufferI1)
        err = this.Kernels["kernel_seed_b1"].SetArg(4,this.BufferI2)

        err = this.Kernels["kernel_seed_b1"].SetArg(5,uint32(32))
        //init kernels
        err = this.Kernels["kernel_seed_b2"].SetArg(0,this.BufferB)

        err = this.Kernels["kernel_seed_b2"].SetArg(1,this.BufferA1)

        err = this.Kernels["kernel_seed_b2"].SetArg(2,this.BufferA2)

        err = this.Kernels["kernel_seed_b2"].SetArg(3,this.BufferI1)
        err = this.Kernels["kernel_seed_b2"].SetArg(4,this.BufferI2)

        err = this.Kernels["kernel_seed_b2"].SetArg(5,uint32(0))
        //init kernels
        err = this.Kernels["kernel_round1"].SetArg(0,this.BufferA1)

        err = this.Kernels["kernel_round1"].SetArg(1,this.BufferA2)

        err = this.Kernels["kernel_round1"].SetArg(2,this.BufferB)

        err = this.Kernels["kernel_round1"].SetArg(3,this.BufferI2)
        err = this.Kernels["kernel_round1"].SetArg(4,this.BufferI1)

        err = this.Kernels["kernel_round1"].SetArg(5,uint32(DUCK_SIZE_A * 1024))
        err = this.Kernels["kernel_round1"].SetArg(6,uint32(DUCK_SIZE_A * 1024))


        //init kernels
        err = this.Kernels["kernel_round0"].SetArg(0,this.BufferB)

        err = this.Kernels["kernel_round0"].SetArg(1,this.BufferA1)

        err = this.Kernels["kernel_round0"].SetArg(2,this.BufferI1)

        err = this.Kernels["kernel_round0"].SetArg(3,this.BufferI2)

        //init kernels
        err = this.Kernels["kernel_round_na"].SetArg(0,this.BufferB)

        err = this.Kernels["kernel_round_na"].SetArg(1,this.BufferA1)

        err = this.Kernels["kernel_round_na"].SetArg(2,this.BufferI1)

        err = this.Kernels["kernel_round_na"].SetArg(3,this.BufferI2)

        //init kernels
        err = this.Kernels["kernel_round_nb"].SetArg(0,this.BufferA1)

        err = this.Kernels["kernel_round_nb"].SetArg(1,this.BufferB)

        err = this.Kernels["kernel_round_nb"].SetArg(2,this.BufferI2)

        err = this.Kernels["kernel_round_nb"].SetArg(3,this.BufferI1)

        //init kernels
        err = this.Kernels["kernel_tail"].SetArg(0,this.BufferB)

        err = this.Kernels["kernel_tail"].SetArg(1,this.BufferA1)

        err = this.Kernels["kernel_tail"].SetArg(2,this.BufferI1)

        err = this.Kernels["kernel_tail"].SetArg(3,this.BufferI2)

        this.ClearBuffer(this.BufferI1)
        this.ClearBuffer(this.BufferI2)

        this.KernelEnq("kernel_seed_a",this.Kernels["kernel_seed_a"],2048,128)
        this.KernelEnq("kernel_seed_b1",this.Kernels["kernel_seed_b1"],1024,128)
        this.KernelEnq("kernel_seed_b2",this.Kernels["kernel_seed_b2"],1024,128)
        this.ClearBuffer(this.BufferI1)
        this.KernelEnq("kernel_round1",this.Kernels["kernel_round1"],4096,1024)
        this.ClearBuffer(this.BufferI2)
        this.KernelEnq("kernel_round0",this.Kernels["kernel_round0"],4096,1024)
        this.ClearBuffer(this.BufferI1)
        this.KernelEnq("kernel_round_nb",this.Kernels["kernel_round_nb"],4096,1024)
        for j:= 0;j<120;j++{
            this.ClearBuffer(this.BufferI2)
            this.KernelEnq("kernel_round_na",this.Kernels["kernel_round_na"],4096,1024)
            this.ClearBuffer(this.BufferI1)
            this.KernelEnq("kernel_round_nb",this.Kernels["kernel_round_nb"],4096,1024)
        }
        this.ClearBuffer(this.BufferI2)
        this.KernelEnq("kernel_tail",this.Kernels["kernel_tail"],4096,1024)
        time.Sleep(1*time.Second)
        edges_bytes := make([]byte,8)
        event,err := this.CommandQueue.EnqueueReadBufferByte(this.BufferI2,true,0,edges_bytes,nil)
        if err != nil {
            fmt.Println("error read edges_bytes:",err)
            this.Release()
            return
        }
        event.Release()
        fmt.Println("edges_bytes:",edges_bytes)
        fmt.Println("left edges:",binary.LittleEndian.Uint32(edges_bytes[0:4])*2)
    }
    this.Release()
}

func (this *Device)Update()  {
    this.CurrentWorkID++
}

func (this *Device)ClearBuffer(buffMem *cl.MemObject)  {
    clearBytes := make([]byte,4)
    event,err := this.CommandQueue.EnqueueFillBuffer(buffMem,unsafe.Pointer(&clearBytes[0]),4,0,INDEX_SIZE,nil)
    if err != nil {
        log.Println("clear kernel error " , this.MinerId,err)
        return
    }
    event.Release()
}

func (this *Device)KernelEnq(name string,k *cl.Kernel,localSize,workSize int)  {
    event , err := this.CommandQueue.EnqueueNDRangeKernel(k, []int{0}, []int{localSize*workSize}, []int{workSize}, nil);
    if err != nil {
        log.Println("exec kernel error ", name , this.MinerId,err)
        return
    }
    this.EventList = append(this.EventList,event)
}

func (d *Device)Release()  {
    d.Context.Release()
    d.Program.Release()
    d.CommandQueue.Release()
    d.BufferA1.Release()
    d.BufferA2.Release()
    d.BufferB.Release()
    d.BufferR.Release()
    d.BufferNonces.Release()
    d.BufferI2.Release()
    d.BufferI1.Release()
    for _,v:=range d.Kernels{
        v.Release()
    }
    d.Kernels = map[string]*cl.Kernel{}
    for _,v:=range d.EventList{
        v.Release()
    }
    d.EventList = []*cl.Event{}
}