package main

import (
    `fmt`
    `github.com/Qitmeer/qitmeer-miner/common`
	`time`
)

func main()  {
	for{
	    fmt.Println(time.Now().Unix())
	    common.Usleep(1000)
    }
}

