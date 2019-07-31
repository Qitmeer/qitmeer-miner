// Copyright (c) 2019 The halalchain developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qitmeer

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"qitmeer-miner/core"
	"log"
	"strconv"
	"strings"
	"time"
)
type getResponseJson struct {
	Result BlockHeader
	Id int `json:"id"`
	Error string `json:"error"`
	JsonRpc string `json:"jsonrpc"`
}
var ErrSameWork = fmt.Errorf("Same work, Had Submitted!")
type getSubmitResponseJson struct {
	Result string `json:"result"`
	Id int `json:"id"`
	Error string `json:"error"`
	JsonRpc string `json:"jsonrpc"`
}
type QitmeerWork struct {
	core.Work
	Block BlockHeader
	PoolWork NotifyWork
	stra *QitmeerStratum
	StartWork bool
}

//GetBlockTemplate
func (this *QitmeerWork) Get () bool {
	body := this.Rpc.RpcResult("getBlockTemplate",[]interface{}{})
	if body == nil{
		log.Println("network failed")
		return false
	}
	var blockTemplate getResponseJson
	err := json.Unmarshal(body,&blockTemplate)
	if err != nil {
		var r map[string]interface{}
		_ = json.Unmarshal(body,&r)
		if _,ok := r["error"];ok{
			log.Println("[node reply]",r["error"])
			return false
		}
		log.Println("[node reply]",string(body))
		return false
	}
	if this.Block.Height > 0 && this.Block.Height == blockTemplate.Result.Height{
		//not has new work
		return false
	}
	diff, _ := strconv.ParseUint(blockTemplate.Result.Bits, 16, 32)
	diffi := make([]byte,4)
	binary.LittleEndian.PutUint32(diffi, uint32(diff))
	blockTemplate.Result.Difficulty = binary.LittleEndian.Uint32(diffi)
	blockTemplate.Result.HasCoinbasePack = false
	_ = blockTemplate.Result.CalcCoinBase(this.Cfg.SoloConfig.RandStr,this.Cfg.SoloConfig.MinerAddr)
	this.Block = blockTemplate.Result
	this.Started = uint32(time.Now().Unix())
	return true
}

//Submit
func (this *QitmeerWork) Submit (subm string) error {
	this.Lock()
	defer this.Unlock()
	if this.LastSub == subm{
		return ErrSameWork
	}
	this.LastSub = subm
	body := this.Rpc.RpcResult("submitBlock",[]interface{}{subm})
	var res getSubmitResponseJson
	err := json.Unmarshal(body, &res)
	if err != nil {
		fmt.Println("【submit error】",string(body))
		return err
	}
	if !strings.Contains(res.Result,"Block submitted accepted") {
		if strings.Contains(res.Result,"The tips of block is expired"){
			return ErrSameWork
		}
		return errors.New("【submit data failed】"+res.Result)
	}
	return nil
}

// pool get work
func (this *QitmeerWork) PoolGet () bool {
	if !this.stra.PoolWork.NewWork {
		return false
	}
	err := this.stra.PoolWork.PrepWork()
	if err != nil {
		log.Println(err)
		return false
	}

	if (this.stra.PoolWork.JobID != "" && !this.stra.PoolWork.Clean) || this.PoolWork.JobID == this.stra.PoolWork.JobID{
		return false
	}

	this.PoolWork = this.stra.PoolWork
	return true
}

//pool submit work
func (this *QitmeerWork) PoolSubmit (subm string) error {
	if this.LastSub == subm{
		return ErrSameWork
	}
	this.LastSub = subm
	arr := strings.Split(subm,"-")
	data,err := hex.DecodeString(arr[0])
	if err != nil {
		return err
	}
	sub, err := this.stra.PrepSubmit(data,arr[1],arr[2])
	if err != nil {
		return err
	}
	m, err := json.Marshal(sub)
	if err != nil {
		return err
	}
	_, err = this.stra.Conn.Write(m)
	if err != nil {
		log.Println("【submit error】【pool connect error】",err)
		return err
	}
	_, err = this.stra.Conn.Write([]byte("\n"))
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}