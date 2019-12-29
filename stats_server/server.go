package stats_server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/Qitmeer/qitmeer-miner/common"
	"github.com/Qitmeer/qitmeer-miner/core"
	"github.com/Qitmeer/qitmeer-miner/stats_server/websocket"
	"strconv"
	"strings"
)

func HandleRouter(cfg *common.GlobalConfig,devices []core.BaseDevice ) {
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/miner_data", func (w http.ResponseWriter, r *http.Request) {
		MinerData(w,r,devices,cfg)
	})
	http.HandleFunc("/set_devices", func (w http.ResponseWriter, r *http.Request) {
		cfg.OptionConfig.UseDevices = r.FormValue("ids")
		inters := r.FormValue("inters")
		lsizes := r.FormValue("lsizes")
		intersArr := strings.Split(inters,",")
		lsizesArr := strings.Split(lsizes,",")
		idsAttr := []string{}
		if cfg.OptionConfig.UseDevices != ""{
			idsAttr = strings.Split(cfg.OptionConfig.UseDevices,",")
		}
		for i,dev := range devices{
			inter := cfg.OptionConfig.Intensity
			if len(intersArr) > i{
				inter,_ = strconv.Atoi(intersArr[i])
			}
			lsizes := cfg.OptionConfig.WorkSize
			if len(lsizesArr) > i{
				lsizes,_ = strconv.Atoi(lsizesArr[i])
			}
			dev.SetIntensity(inter)
			dev.SetWorkSize(lsizes)
			dev.SetIsValid(true)
			if !common.InArray(strconv.Itoa(dev.GetMinerId()),idsAttr){
				dev.SetIsValid(false)
			}
		}
		_,_ = fmt.Fprintf(w, string("success"))
	})
	http.HandleFunc("/set_params", func (w http.ResponseWriter, r *http.Request) {
		cfg.SoloConfig.MinerAddr = r.FormValue("miner_addr")
		cfg.SoloConfig.RPCServer = r.FormValue("rpc_server")
		cfg.SoloConfig.RPCUser = r.FormValue("rpc_username")
		cfg.SoloConfig.RPCPassword = r.FormValue("rpc_password")
		cfg.PoolConfig.Pool = r.FormValue("stratum_addr")
		cfg.PoolConfig.PoolUser = r.FormValue("stratum_user")
		cfg.PoolConfig.PoolPassword = r.FormValue("stratum_pass")
		cfg.OptionConfig.Restart = 1
		_,_ = fmt.Fprintf(w, string("success"))
	})
	common.MinerLoger.Info("stats server start","server",cfg.OptionConfig.StatsServer)
	go websocket.Manager.Start()
	http.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {
		statsData := &websocket.StatsData{}
		statsData.Cfg = cfg
		statsData.Devices = devices

		websocket.WsPage(writer,request,statsData)
	})
	if err := http.ListenAndServe(cfg.OptionConfig.StatsServer, nil) ;err != nil{
		common.MinerLoger.Error(err.Error())
	}
}


func MinerData(w http.ResponseWriter, r *http.Request,devices []core.BaseDevice,cfg *common.GlobalConfig) {
	var res = map[string]interface{}{}
	var devs = []map[string]interface{}{}
	for _,dev := range devices{
		devs = append(devs, map[string]interface{}{
			"id":dev.GetMinerId(),
			"name":dev.GetName(),
			"hashrate":dev.GetAverageHashRate(),
			"isValid":dev.GetIsValid(),
			"global_size":dev.GetIntensity(),
			"local_size":dev.GetWorkSize(),
		})
	}
	res["devices"] = devs
	res["config"] = cfg
	b,_ := json.Marshal(res)
	_,_ = fmt.Fprintf(w, string(b))
}