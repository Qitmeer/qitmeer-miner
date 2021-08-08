package websocket

import (
	"encoding/json"
	"github.com/Qitmeer/qitmeer-miner/common"
	"github.com/Qitmeer/qitmeer-miner/core"
	"github.com/gorilla/websocket"
	"github.com/twinj/uuid"
	"math/big"
	"net/http"
)

type StatsData struct {
	Cfg     *common.GlobalConfig
	Devices []core.BaseDevice
}

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

type Client struct {
	id     string
	socket *websocket.Conn
	send   chan []byte
}

type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

var Manager = ClientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

func (manager *ClientManager) Start() {
	for {
		select {
		case conn := <-manager.register:
			manager.clients[conn] = true
			jsonMessage, _ := json.Marshal(&Message{Content: "/A new socket has connected."})
			manager.send(jsonMessage, conn)
		case conn := <-manager.unregister:
			if _, ok := manager.clients[conn]; ok {
				close(conn.send)
				delete(manager.clients, conn)
				jsonMessage, _ := json.Marshal(&Message{Content: "/A socket has disconnected."})
				manager.send(jsonMessage, conn)
			}
		case message := <-manager.broadcast:
			for conn := range manager.clients {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(manager.clients, conn)
				}
			}
		}
	}
}

func (manager *ClientManager) send(message []byte, ignore *Client) {
	for conn := range manager.clients {
		if conn != ignore {
			conn.send <- message
		}
	}
}

func (c *Client) read() {
	defer func() {
		Manager.unregister <- c
		_ = c.socket.Close()
	}()

	for {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			Manager.unregister <- c
			_ = c.socket.Close()
			break
		}
		jsonMessage, _ := json.Marshal(&Message{Sender: c.id, Content: string(message)})
		Manager.broadcast <- jsonMessage
	}
}

func (c *Client) write(data *StatsData) {
	defer func() {
		_ = c.socket.Close()
	}()
	configD := map[string]interface{}{}
	devStats := map[int]interface{}{}
	allHashrate := 0.00
	var needCalcTimes, canCalcTimes *big.Float
	var bj []byte
	var dev core.BaseDevice
	for {
		select {
		default:
			common.Usleep(5)
			allHashrate = 0.00
			configD["config"] = *data.Cfg
			needCalcTimes = new(big.Float).SetInt(common.GetNeedHashTimesByTarget(data.Cfg.OptionConfig.Target))
			for _, dev = range data.Devices {
				devStats[dev.GetMinerId()] = map[string]interface{}{
					"hashrate": dev.GetAverageHashRate(),
					"id":       dev.GetMinerId(),
					"name":     dev.GetName(),
				}
				allHashrate += dev.GetAverageHashRate()
			}
			configD["needSec"] = 0
			configD["blockTime"] = data.Cfg.NecessaryConfig.Param.TargetTimePerBlock
			canCalcTimes = big.NewFloat(allHashrate)
			if allHashrate > 0 && needCalcTimes.Cmp(big.NewFloat(0)) > 0 {
				needCalcTimes.Quo(needCalcTimes, canCalcTimes) //need seconds
				configD["needSec"] = needCalcTimes
			}
			configD["devices"] = devStats
			bj, _ = json.Marshal(configD)
			_ = c.socket.WriteMessage(websocket.TextMessage, bj)
		case message, ok := <-c.send:
			if !ok {
				_ = c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			_ = c.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func WsPage(res http.ResponseWriter, req *http.Request, statsData *StatsData) {
	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
	if err != nil {
		http.NotFound(res, req)
		return
	}
	client := &Client{id: uuid.NewV4().String(), socket: conn, send: make(chan []byte)}

	Manager.register <- client

	go client.read()
	go client.write(statsData)
}
