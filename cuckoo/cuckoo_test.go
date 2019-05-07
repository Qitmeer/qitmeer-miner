package cuckoo

import (
	"testing"
	"fmt"
	"sort"
	"golang.org/x/crypto/blake2b"
	"github.com/AidosKuneen/numcpu"
	"runtime"
)

const (
	EdgeBits = 12 //边缘指数
	NEdge = 1 << EdgeBits
	NNodes = NEdge << 1
	EdgeMask  = NEdge - 1
	ProofSize = 8 //环长度 10
)
func u8to64(p [32]byte, i uint) uint {
	return ((uint)(p[i]) & 0xff) |
		((uint)(p[i+1])&0xff)<<8 |
		((uint)(p[i+2])&0xff)<<16 |
		((uint)(p[i+3])&0xff)<<24 |
		((uint)(p[i+4])&0xff)<<32 |
		((uint)(p[i+5])&0xff)<<40 |
		((uint)(p[i+6])&0xff)<<48 |
		((uint)(p[i+7])&0xff)<<56
}

func Sipnode(nonce uint, uOrV uint) uint {
	return Siphash24(2*nonce+uOrV) & EdgeMask
}
/**
	1 2 3

	4 5 6

	(1,4) (1,5) (1,6) (2,4) (2,6) (3,4) (3,5)

	1 1 1
	1 0 1
	1 1 0


 */
type Edge struct {
	from int64
	to   int64
}

var allKey map[int64][]int64
var allPath [][]int64

func TestCuckoo(t *testing.T){
	n := numcpu.NumCPU()
	p := runtime.GOMAXPROCS(n)
	for i := 0;i<10000;i++{
		if IsFind{
			break
		}
		Headerbytes := []byte(fmt.Sprintf("helloworld%d",i))
		hdrkey := blake2b.Sum256(Headerbytes)
		//fmt.Printf("hdrkey: %x\n", hdrkey)

		V[0] = u8to64(hdrkey, 0)
		V[1] = u8to64(hdrkey, 8)
		V[2] = u8to64(hdrkey, 16)
		V[3] = u8to64(hdrkey, 24)
		//hash table 1
		//hashTable1 := make(map[int64]int64,NEdge)
		var edges [NEdge][]int64
		//hash table 2
		//hashTable2 := []int64{4,5,6}
		//初始化两个顶点集合
		for nonce := 0;nonce < NEdge;nonce++{
			u := Sipnode(uint(nonce), 0)
			v := EdgeMask+Sipnode(uint(nonce), 1)
			edges[int64(nonce)] = append(edges[int64(nonce)],int64(u),int64(v))
		}

		//所有边
		//allkey每个key对应的所有关系
		allKey = make(map[int64][]int64,0)
		for _,edge := range edges{
			if !inNode(edge[1],allKey[edge[0]]){
				allKey[edge[0]] = append(allKey[edge[0]],edge[1])
			}
			if !inNode(edge[0],allKey[edge[1]]){
				allKey[edge[1]] = append(allKey[edge[1]],edge[0])
			}
		}
		//fmt.Println(allKey)
		allPath = make([][]int64,0)
		//寻找路线 寻找环
		for k := 0;k < NEdge;k++{
			if IsFind{
				break
			}
			findCircle([]Edge{},int64(k))
		}
		for _,p := range allPath{
			if len(p)-1 == ProofSize{
				fmt.Println("环长度：",len(p)-1,p)
			}
		}
	}
	runtime.GOMAXPROCS(p)
}

func inNode(n int64,nodes []int64) bool {
	for _,v := range nodes{
		if v == n{
			return true
		}
	}
	return false
}

func edgeExist(ed Edge,m []Edge) bool {
	for _,v := range m{
		if (v.from == ed.from && v.to == ed.to) ||
			(v.to == ed.from && v.from == ed.to) {
			return true
		}
	}
	return false
}
var IsFind = false
func findCircle(parents []Edge,k int64) {
	if IsFind{
		return
	}
	l := len(parents)
	//是否已经找到 环
	if l >= 4{
		//二分图 最小环路 4条边
		start := parents[0]
		end := parents[l-1]
		if start.from == end.to{
			keys := make([]int,0)
			for k,_ := range parents{
				keys = append(keys,k)
			}
			sort.Ints(keys)
			cpath := make([]int64,0)
			for _,index := range keys {
				if index == 0{
					cpath = append(cpath,parents[index].from,parents[index].to)
				} else{
					cpath = append(cpath,parents[index].to)
				}
				//cpath = append(cpath,parents[index].from,parents[index].to)
			}
			allPath = append(allPath,cpath)
			if len(cpath) - 1 == ProofSize{
				IsFind = true
				return
			}
		}
	}
	//是否有线路 边要大于1 因为回路
	if _,ok := allKey[k];ok && len(allKey[k]) > 1{
		for _,v := range allKey[k]{
			edge := Edge{}
			edge.from = k
			edge.to = v
			if edgeExist(edge,parents){
				continue
			}
			newparents := make([]Edge,0)
			newparents = append(newparents,parents...)
			newparents = append(newparents,edge)
			findCircle(newparents,v)
		}
	}
	return
}
