package core

import (
	"bufio"
	"encoding/json"
	"errors"
	"net"
	"qitmeer-miner/common"
	"qitmeer-miner/common/socks"
	"strings"
	"sync"
	"time"
)
// ErrJsonType is an error for json that we do not expect.
var ErrJsonType = errors.New("Unexpected type in json.")

type StratumMsg struct {
	Method string `json:"method"`
	// Need to make generic.
	Params []string    `json:"params"`
	ID     interface{} `json:"id"`
}

type Stratum struct {
	sync.Mutex
	Cfg       *common.GlobalConfig
	Conn      net.Conn
	Reader    *bufio.Reader
	ID        uint64
	Started uint32
	Timeout uint32
	ValidShares uint64
	InvalidShares uint64
	StaleShares uint64
	SubmitIDs	[]uint64
	SubID	uint64
	AuthID uint64
}


// StratumConn starts the initial connection to a stratum pool and sets defaults
// in the pool object.
func (this *Stratum)StratumConn(cfg *common.GlobalConfig) error {
	this.Cfg = cfg
	pool := cfg.PoolConfig.Pool
	common.MinerLoger.Debugf("【Connect pool】:%v", pool)
	proto := "stratum+tcp://"
	if strings.HasPrefix(this.Cfg.PoolConfig.Pool, proto) {
		pool = strings.Replace(pool, proto, "", 1)
	} else {
		err := errors.New("[error] Only stratum pools supported.stratum+tcp://")
		return err
	}
	this.Cfg.PoolConfig.Pool = pool
	this.ID = 1
	_ = this.Reconnect()

	go func() {
		if uint32(time.Now().Unix()) - this.Timeout > 30{
			common.MinerLoger.Debug("【timeout】reconnect")
			_ = this.Reconnect()
		}
	}()
	return nil
}

func (this *Stratum)Listen(handle func(data string))  {
	common.MinerLoger.Debug("Starting Stratum Listener")
	for {
		if this.Cfg.OptionConfig.Restart == 1{
			_ = this.Conn.Close()
			common.MinerLoger.Debug("pool restart")
			return
		}
		//s.Conn.SetReadDeadline(time.Now().Add(time.Second * 10))
		var data string
		var err error
		if this.Reader != nil{
			data, err = this.Reader.ReadString('\n')
		} else{
			err = errors.New("network wrong!")
		}

		if err != nil {
			for{
				common.MinerLoger.Debug("【Connection lost!  Reconnecting...】")
				err = this.Reconnect()
				if err != nil {
					common.MinerLoger.Debugf("【Reconnect failed sleep 2s】.%s",err.Error())
					time.Sleep(2*time.Second)
					continue
				}
				break
			}
		}
		handle(data)
		this.Timeout = uint32(time.Now().Unix())
	}
}

// Reconnect reconnects to a stratum server if the connection has been lost.
func (s *Stratum) Reconnect() error {
	var conn net.Conn
	var err error
	if s.Cfg.OptionConfig.Proxy != "" {
		proxy := &socks.Proxy{
			Addr:     s.Cfg.OptionConfig.Proxy,
			Username: s.Cfg.OptionConfig.ProxyUser,
			Password: s.Cfg.OptionConfig.ProxyPass,
		}
		conn, err = proxy.Dial("tcp", s.Cfg.PoolConfig.Pool)
	} else {
		conn, err = net.Dial("tcp", s.Cfg.PoolConfig.Pool)
	}
	if err != nil {
		common.MinerLoger.Debugf("【init reconnect error】%v",err)
		return err
	}
	s.Conn = conn
	s.Reader = bufio.NewReader(s.Conn)
	err = s.Subscribe()
	if err != nil {
		common.MinerLoger.Debugf("【subscribe reconnect】%v",err)
		return nil
	}
	// Should NOT need this.
	time.Sleep(5 * time.Second)
	// XXX Do I really need to re-auth here?
	err = s.Auth()
	if err != nil {
		common.MinerLoger.Debugf("【auth reconnect】%v",err)
		return nil
	}
	// If we were able to reconnect, restart counter
	s.Started = uint32(time.Now().Unix())
	s.Timeout = uint32(time.Now().Unix())
	return nil
}

// Auth sends a message to the pool to authorize a worker.
func (s *Stratum) Auth() error {
	msg := StratumMsg{
		Method: "mining.authorize",
		ID:     s.ID,
		Params: []string{s.Cfg.PoolConfig.PoolUser, s.Cfg.PoolConfig.PoolPassword},
	}
	// Auth reply has no method so need a way to identify it.
	// Ugly, but not much choice.
	id, ok := msg.ID.(uint64)
	if !ok {
		return ErrJsonType
	}
	s.ID++
	s.AuthID = id
	m, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = s.Conn.Write(m)
	if err != nil {
		common.MinerLoger.Debugf("【auth connect】%v",err)
		return err
	}
	_, err = s.Conn.Write([]byte("\n"))
	if err != nil {
		return err
	}
	return nil
}

// Subscribe sends the subscribe message to get mining info for a worker.
func (s *Stratum) Subscribe() error {
	msg := StratumMsg{
		Method: "mining.subscribe",
		ID:     s.ID,
		Params: []string{"qitmeer-miner/v0.0.1"},
	}
	s.SubID = msg.ID.(uint64)
	s.ID++
	m, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = s.Conn.Write(m)
	if err != nil {
		common.MinerLoger.Debugf("【subscribe connect】%v",err)
		return err
	}
	_, err = s.Conn.Write([]byte("\n"))
	if err != nil {
		return err
	}
	return nil
}