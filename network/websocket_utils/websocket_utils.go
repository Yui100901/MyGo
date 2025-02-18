package websocket_utils

import (
	"github.com/Yui100901/MyGo/log_utils"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

//
// @Author yfy2001
// @Date 2025/1/16 09 56
//

var WSServer = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		//验证
		return true
	},
}

type WebSocket struct {
	Conn *websocket.Conn //底层websocket连接对象
	Done chan struct{}
	mu   sync.Mutex
}

func NewWebSocketByUpgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*WebSocket, error) {
	conn, err := WSServer.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}
	return &WebSocket{
		Conn: conn,
		Done: make(chan struct{}),
		mu:   sync.Mutex{},
	}, nil
}

func NewWebSocketByDail(url string, requestHeader http.Header) (*WebSocket, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, requestHeader)
	if err != nil {
		return nil, err
	}
	return &WebSocket{
		Conn: conn,
		Done: make(chan struct{}),
		mu:   sync.Mutex{},
	}, nil
}

func (ws *WebSocket) OnMessage(handler func([]byte)) {
	defer close(ws.Done)
	for {
		_, message, err := ws.Conn.ReadMessage()
		if err != nil {
			log_utils.Error.Println("Websocket Read ERROR:", err)
			return
		}
		if handler != nil {
			handler(message)
		}
		log_utils.Info.Printf("Websocket Receive: %s", message)
	}
}

func (ws *WebSocket) SendMessage(data func() []byte) error {
	ws.mu.Lock()
	err := ws.Conn.WriteMessage(websocket.TextMessage, data())
	ws.mu.Unlock()
	return err
}

func (ws *WebSocket) Close() {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.Conn.Close()
}
