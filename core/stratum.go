package core

import (
	"fmt"
	"sync"
	"net"
	"bufio"
	"hlc-miner/common"
	"strings"
	"time"
	"errors"
	"encoding/json"
	"hlc-miner/common/socks"
	"log"
)
// ErrJsonType is an error for json that we do not expect.
var ErrJsonType = errors.New("Unexpected type in json.")

type StratumMsg struct {
	Method string `json:"method"`
	// Need to make generic.
	Params []string    `json:"params"`
	ID     interface{} `json:"id"`
}

// ErrStratumStaleWork indicates that the work to send to the pool was stale.
var ErrStratumStaleWork = fmt.Errorf("Stale work, throwing away")

type Stratum struct {
	sync.Mutex
	Cfg       *common.Config
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
func (this *Stratum)StratumConn(cfg *common.Config) error {
	this.Cfg = cfg
	pool := cfg.Pool
	log.Println("【Connect pool】:", pool)
	proto := "stratum+tcp://"
	if strings.HasPrefix(this.Cfg.Pool, proto) {
		pool = strings.Replace(pool, proto, "", 1)
	} else {
		err := errors.New("Only stratum pools supported.")
		return err
	}
	this.Cfg.Pool = pool
	this.ID = 1
	this.Reconnect()
	go func() {
		if uint32(time.Now().Unix()) - this.Timeout > 30{
			log.Println("【timeout】reconnect")
			this.Reconnect()
		}
	}()
	return nil
}

func (this *Stratum)Listen(handle func(data string))  {
	log.Println("Starting Stratum Listener")
	for {
		//s.Conn.SetReadDeadline(time.Now().Add(time.Second * 10))
		data, err := this.Reader.ReadString('\n')
		if err != nil {
			for{
				log.Println("【Connection lost!  Reconnecting...】")
				err = this.Reconnect()
				if err != nil {
					fmt.Println("【Reconnect failed sleep 2s】.",err)
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
	if s.Cfg.Proxy != "" {
		proxy := &socks.Proxy{
			Addr:     s.Cfg.Proxy,
			Username: s.Cfg.ProxyUser,
			Password: s.Cfg.ProxyPass,
		}
		conn, err = proxy.Dial("tcp", s.Cfg.Pool)
	} else {
		conn, err = net.Dial("tcp", s.Cfg.Pool)
	}
	if err != nil {
		log.Println("【init reconnect error】",err)
		return err
	}
	s.Conn = conn
	s.Reader = bufio.NewReader(s.Conn)
	err = s.Subscribe()
	if err != nil {
		log.Println("【subscribe reconnect】",err)
		return nil
	}
	// Should NOT need this.
	time.Sleep(5 * time.Second)
	// XXX Do I really need to re-auth here?
	err = s.Auth()
	if err != nil {
		log.Println("【auth reconnect】",err)
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
		Params: []string{s.Cfg.PoolUser, s.Cfg.PoolPassword},
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
		log.Println("【auth connect】",err)
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
		Params: []string{"halalchainminer/" + s.Cfg.Version},
	}
	s.SubID = msg.ID.(uint64)
	s.ID++
	m, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = s.Conn.Write(m)
	if err != nil {
		log.Println("【subscribe connect】",err)
		return err
	}
	_, err = s.Conn.Write([]byte("\n"))
	if err != nil {
		return err
	}
	return nil
}