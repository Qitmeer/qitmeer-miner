package cuckoo

import (
	"encoding/binary"
	"fmt"
	"log"
	cuckaroo "github.com/HalalChain/qitmeer-lib/crypto/cuckoo"
)

type paths struct {
	values []int
}
func (this *paths)Add(key int)  {
	this.values = append(this.values,key)
}

func (this *paths)Contains(val int) bool {
	for _,v := range this.values{
		if v == val{
			return true
		}
	}
	return false
}

func (this *paths)IndexOf(val int) int {
	for index,v := range this.values{
		if v == val{
			return index
		}
	}
	return -1
}

func (this *paths)String() string {
	str := ""
	for k,v:=range this.values{
		str += fmt.Sprintf("num:%d ,value:%d",k,v)
	}
	return str
}
func (this *paths)Count() int {
	return len(this.values)
}
func (this *paths)Take(length int) {
	this.values = this.values[:length]
}
func (this *paths)Skip(offset int) {
	this.values = this.values[offset:]
}
type Dictionary map[int]int
func (this Dictionary)TryGetValue(key int) int {
	if _,ok:=this[key];ok{
		return this[key]
	}
	return -1
}
func (this Dictionary)Remove(key int) {
	if _,ok := this[key];ok{
		delete(this,key)
	}
}


type  CGraph struct {
	U Dictionary
	V Dictionary
	Edges []int
	EdgesCount int
	Dupes int
	Maxlength int
	CycleEdges Edges
	IsFind bool
	AllMaps map[int][]int
}
func (this *CGraph)GetNonceEdges() []uint32{
	result := make([]uint32,0)
	for i:=0;i<len(this.CycleEdges.data);i++{
		if i==0{
			result = append(result,uint32(this.CycleEdges.data[i].Item1))
			result = append(result,uint32(this.CycleEdges.data[i].Item2))
		} else{
			if !InArrayInterface(uint32(this.CycleEdges.data[i].Item1),result){
				result = append(result,uint32(this.CycleEdges.data[i].Item1))
			}
			if !InArrayInterface(uint32(this.CycleEdges.data[i].Item2),result){
				result = append(result,uint32(this.CycleEdges.data[i].Item2))
			}
		}
	}
	return result
}
func (this *CGraph)GetNonceEdgesBytes() (res []byte){
	res = make([]byte,0)
	for _,e := range this.CycleEdges.data{
		b := make([]byte,4)
		binary.LittleEndian.PutUint32(b,uint32(e.Item1))
		res = append(res,b...)
		b = make([]byte,4)
		binary.LittleEndian.PutUint32(b,uint32(e.Item2))
		res = append(res,b...)
	}
	return
}

type Edges struct {
	data []Edge1
	nodeMap map[int]int
}

func (this *Edges)AddEdges(e Edge1){
	if !this.HasEdges(e){
		this.data = append(this.data,e)
		this.nodeMap[e.Item1] += 1
		this.nodeMap[e.Item2] += 1
	}
}

func (this *Edges)Check() bool{
	if len(this.data) != cuckaroo.ProofSize{
		return false
	}
	//log.Println(this.nodeMap)
	//every node will display 2 times in a cycle
	for _,count := range this.nodeMap{
		if count != 2 {
			return false
		}
	}
	return true
}

func (this *CGraph)FindCycle()bool{
	for k:=0;k<this.EdgesCount*2 ;k+=2{
		if this.IsFind{
			return true
		}
		this.CycleEdges = Edges{
			nodeMap: map[int]int{},
			data:make([]Edge1,0),
		}
		this.Find(Edges{data:make([]Edge1,0),nodeMap: map[int]int{}},this.Edges[k])
	}
	return false
}

func (this *CGraph)Find(parents Edges,k int){
	if this.IsFind{
		return
	}
	l := len(parents.data)
	if l >= 4{
		start := parents.data[0]
		end := parents.data[l-1]
		if start.Item1 == end.Item2{
			for _,e := range parents.data{
				this.CycleEdges.AddEdges(e)
			}
			if this.CycleEdges.Check(){
				this.IsFind = true
				log.Println("Find 42 - Cycles")
				return
			}
		}
	}
	for _,v := range this.AllMaps[k]{
		edge := Edge1{}
		edge.Item1 = k
		edge.Item2 = v
		if parents.HasEdges(edge) || k == v{
			continue
		}

		newparents := Edges{data:make([]Edge1,0),nodeMap: map[int]int{}}
		newparents.data = append(newparents.data,parents.data...)
		newparents.data = append(newparents.data,edge)

		this.Find(newparents,v)
	}
	return
}


func (this *Edges)HasEdges(e Edge1) bool{
	for _,ed := range this.data{
		if ed.Item1 == e.Item1 && ed.Item2 == e.Item2{
			return true
		}
		if ed.Item1 == e.Item2 && ed.Item2 == e.Item1{
			return true
		}
	}
	return false
}

func (this *CGraph)SetEdges(edges []uint32,count int ) {
	this.Edges = make([]int,len(edges))
	for k,v := range edges {
		this.Edges[k] = int(v)
	}
	this.CycleEdges = Edges{
		nodeMap: map[int]int{},
		data:make([]Edge1,0),
	}
	this.U = Dictionary{}
	this.V = Dictionary{}
	this.EdgesCount = count
	this.Maxlength = 8192
	this.Dupes = 0
	this.AllMaps = map[int][]int{}
	//init allmap
	for i:=0;i<this.EdgesCount*2;i+=2{
		e := Edge1{Item1:this.Edges[i],Item2:this.Edges[i+1]}
		if _,ok := this.AllMaps[e.Item1];!ok{
			this.AllMaps[e.Item1] = make([]int,0)
		}
		if _,ok := this.AllMaps[e.Item2];!ok{
			this.AllMaps[e.Item2] = make([]int,0)
		}
		if !InArrayInt(e.Item2,this.AllMaps[e.Item1]) && e.Item2 != e.Item1{
			this.AllMaps[e.Item1] = append(this.AllMaps[e.Item1],e.Item2)
		}
		if !InArrayInt(e.Item1,this.AllMaps[e.Item2]) && e.Item2 != e.Item1{
			this.AllMaps[e.Item2] = append(this.AllMaps[e.Item2],e.Item1)
		}

	}
}

type Edge1 struct {
	Item1 int
	Item2 int
}

func (this *CGraph)FindSolutions() bool {
	//log.Println("【Search】In Edge Count ",this.EdgesCount)
	for ee:=0; ee < this.EdgesCount; ee++{
		e := Edge1{Item1:this.Edges[ee*2+0],Item2:this.Edges[ee*2+1]}
		if I1 := this.U.TryGetValue(e.Item1) ;I1 != -1 && int(I1) == e.Item2{
			this.Dupes++
			continue
		}
		if I2 := this.V.TryGetValue(e.Item2) ;I2 != -1 && int(I2) == e.Item1{
			this.Dupes++
			continue
		}
		path1 := this.path(true,uint(e.Item1))
		path2 := this.path(false,uint(e.Item2))
		joinA := -1
		joinB := -1
		cycle := 0
		this.CycleEdges = Edges{
			nodeMap: map[int]int{},
			data:make([]Edge1,0),
		}
		this.CycleEdges.AddEdges(e)
		for i:=0;i<path1.Count();i++{
			ival := path1.values[i]
			if i > 0 && i < path1.Count() {
				if i & 1 == 0{
					this.CycleEdges.AddEdges(Edge1{Item1:ival,Item2:path1.values[i-1]})
				} else{
					this.CycleEdges.AddEdges(Edge1{Item2:ival,Item1:path1.values[i-1]})
				}
			}
			if path2.Contains(ival){
				path2Idx := path2.IndexOf(ival)
				joinA = i
				joinB = path2Idx
				cycle = joinB+joinA + 1
				if cycle == cuckaroo.ProofSize{
					//log.Println(e)
					//log.Println(i)
					//log.Println(path1)
					//log.Println(path2)
					//log.Println(path2Idx)
					for k:=path2Idx;k>0;k--{
						if path2.values[k] & 1 == 0{
							this.CycleEdges.AddEdges(Edge1{Item1:path2.values[k],Item2:path2.values[k-1]})
						} else{
							this.CycleEdges.AddEdges(Edge1{Item2:path2.values[k],Item1:path2.values[k-1]})
						}
					}
					if this.CycleEdges.Check(){
						break
					}
					//log.Println(len(this.CycleEdges.data),this.CycleEdges)
					cycle = 5
				}
				break
			}
		}

		if cycle >= 4 && cycle != cuckaroo.ProofSize{
			//log.Println(fmt.Sprintf("%d-cycle found!",cycle))
		} else if cycle == cuckaroo.ProofSize{
			//log.Println(fmt.Sprintf("%d-cycle found!",PROOF_SIZE),this.CycleEdges)
			return true
		} else{
			if path1.Count() > path2.Count(){
				this.Reverse(path2,false)
				this.V[e.Item2] = e.Item1
			} else{
				this.Reverse(path1,true)
				this.U[e.Item1] = e.Item2
			}
		}

	}
	return false
}

func (this *CGraph)Reverse(p paths,startsInU bool) {
	for i := p.Count() - 2; i >= 0; i--{
		A := p.values[i]
		B := p.values[i+1]
		if startsInU{
			if i & 1 == 0{
				this.U.Remove(A)
				this.V[B] = A
			} else{
				this.V.Remove(A)
				this.U[B] = A
			}
		} else{
			if (i & 1) == 0{
				this.V.Remove(A)
				this.U[B] = A
			} else{
				this.U.Remove(A)
				this.V[B] = A
			}
		}
	}
}

func (this *CGraph) path(_startInGraphU bool,_key uint) paths{
	p := paths{}
	key := _key
	startInGraphU := _startInGraphU
	g := this.V
	if startInGraphU{
		g = this.U
	}
	p.Add(int(key))
	for {
		v := g.TryGetValue(int(key))
		if v == -1{
			break
		}
		if p.Count() >= this.Maxlength{
			break
		}
		p.Add(int(v))
		startInGraphU = !startInGraphU
		g = this.V
		if startInGraphU{
			g = this.U
		}
		key = uint(v)
	}
	return p
}

func InArrayInterface(e uint32,list []uint32)  bool {
	for _,i := range list{
		if i == e {
			return true
		}
	}
	return false
}

func InArrayInt(e int,list []int)  bool {
	for _, i := range list {
		if i == e {
			return true
		}
	}
	return false
}