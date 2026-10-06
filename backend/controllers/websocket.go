package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"mcloud/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSProbeRequest struct {
	IP    string `json:"ip"`
	Color string `json:"color"`
}

type WSProbeResponse struct {
	IP    string `json:"ip"`
	Color string `json:"color"`
}

func HandleProbeWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket升级失败:", err)
		return
	}
	defer conn.Close()

	var mu sync.Mutex
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			go handleProbeMessage(conn, message, &mu)
		}
	}()

	<-done
}

func handleProbeMessage(conn *websocket.Conn, message []byte, mu *sync.Mutex) {
	var req WSProbeRequest
	if err := json.Unmarshal(message, &req); err != nil {
		return
	}
	if req.IP == "" || req.Color == "" {
		return
	}

	svc := services.NewStatsService()
	resp, err := svc.Probe(services.ProbeRequest{
		IP:    req.IP,
		Color: req.Color,
	})
	if err != nil {
		return
	}

	data := WSProbeResponse{
		IP:    req.IP,
		Color: resp.Color,
	}
	respBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("探测响应序列化失败(%s): %v", req.IP, err)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	if err := conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
		log.Printf("探测响应推送失败(%s): %v", req.IP, err)
	}
}
