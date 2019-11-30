package main

import (
    `fmt`
    `qitmeer-miner/common`
	`time`
)

func main()  {
	for{
	    fmt.Println(time.Now().Unix())
	    common.Usleep(1000)
    }
}

