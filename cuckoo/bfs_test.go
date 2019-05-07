package cuckoo

import (
	"fmt"
	"testing"
)

const SIZE = 5

type point struct {
	x int64   //顶点1 x坐标
	y int64   //顶点2 y坐标
	union bool //是否联通
} //x,y坐标

var points map[int64]point

func (this *point)status()bool{
	if this.x < 0 || this.y < 0{
		return false
	}
	return true
}

func getPointKey(x,y int64) int64 {
	return x * SIZE + y
}

//上部坐标
func (this *point)up()point{
	if this.x > 0{
		//不在第一行
		k := getPointKey(this.x-1,this.y)
		if _,ok := points[k];ok{
			return points[k]
		}
	}
	return point{-1,-1,false}
}

//下部坐标
func (this *point)down()point{
	if this.x < 4{
		//不在最后行
		k := getPointKey(this.x+1,this.y)
		if _,ok := points[k];ok{
			return points[k]
		}
	}
	return point{-1,-1,false}
}

//左部坐标
func (this *point)left()point{
	if this.y >0{
		//不在第一列
		k := getPointKey(this.x,this.y-1)
		if _,ok := points[k];ok{
			return points[k]
		}
	}
	return point{-1,-1,false}
}

//左部坐标
func (this *point)right()point{
	if this.y <4{
		//不在最后一列
		k := getPointKey(this.x,this.y+1)
		if _,ok := points[k];ok{
			return points[k]
		}
	}
	return point{-1,-1,false}
}

//一个点是否是另一个点的相邻点
func (this *point)hasnext(po point)bool{
	if this.up().x == po.x && this.up().y == po.y{
		return true
	}
	if this.down().x == po.x && this.down().y == po.y{
		return true
	}
	if this.left().x == po.x && this.left().y == po.y{
		return true
	}
	if this.right().x == po.x && this.right().y == po.y{
		return true
	}
	return false
}

type unionpath [][]int64

/**

	1表示走不通 0可通

	0 1 0 0 0
	0 1 0 1 0
	0 0 0 0 0
	0 1 1 1 0
	0 0 0 1 0
 */

var allpath map[int64]unionpath

func TestBFS(t *testing.T){
	//初始化矩阵

	in := [SIZE][SIZE]int64{{0,1,0,0,0},{0,1,0,1,0},{0,0,0,0,0},{0,1,1,1,0},{0,0,0,1,0}}

	//初始化图形点
	points = make(map[int64]point,0)
	for x:=0;x<SIZE;x++{
		for y:=0;y<SIZE;y++{
			p := point{}
			p.x = int64(x)
			p.y = int64(y)
			p.union = false
			if in[x][y] == 0{
				p.union = true
			}
			points[getPointKey(int64(x),int64(y))] = p
		}
	}

	fmt.Println(len(points),points)
	//找出所有路径
	allpath = make(map[int64]unionpath,0)
	findPath(int64(0),[]int64{},points[0])
	//将路径整理出来
	for k:= 0;k<len(allpath);k++{
		for _,v := range allpath[int64(k)]{
			if inarray(0,v) && inarray(24,v){
				fmt.Println("找到完整路径：",v,"长度：",len(v))
			}
		}
		fmt.Println(allpath[int64(k)],"长度：",len(allpath[int64(k)]))
	}

}

func inarray(address int64,m []int64) bool {
	for _,v := range m{
		if v == address{
			return true
		}
	}
	return false
}

func inpath(po point,pa []int64) bool  {
	for _,k := range pa{
		if getPointKey(po.x,po.y) == k{
			return true
		}
	}

	return false
}

func findPath(level int64,parents []int64,po point){

	if inpath(po,parents){
		return
	}

	if po.status() && po.union {
		//连通的点
		parents = append(parents,getPointKey(po.x, po.y))
		allpath[level] = append(allpath[level],parents)
		//上
		findPath(level+1,parents,po.up())
		//下
		findPath(level+1,parents,po.down())
		//左
		findPath(level+1,parents,po.left())
		//下
		findPath(level+1,parents,po.right())
	}
	return
}