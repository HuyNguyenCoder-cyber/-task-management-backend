package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	gorillaws "github.com/gorilla/websocket"
)

var BroadcastChannel = make(chan any, 100)

var upgrader = gorillaws.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	activeClients = make(map[*gorillaws.Conn]*wsClient)
	clientsMu     sync.RWMutex
	startOnce     sync.Once
)

const broadcasterWorkerCount = 3

type wsClient struct {
	conn      *gorillaws.Conn
	writeMu   sync.Mutex
	closeOnce sync.Once
}

func (c *wsClient) writeJSON(event any) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteJSON(event)
}

func (c *wsClient) close() {
	c.closeOnce.Do(func() {
		_ = c.conn.Close()
	})
}

func StartEventBroadcaster() {
	startOnce.Do(func() {
		for i := 0; i < broadcasterWorkerCount; i++ {
			go startEventBroadcasterWorker(i + 1)
		}
	})
}

func startEventBroadcasterWorker(workerID int) {
	log.Printf("[WS] broadcaster worker started: %d", workerID)
	for event := range BroadcastChannel {
		for _, client := range snapshotClients() {
			if err := client.writeJSON(event); err != nil {
				log.Printf("[WS] Broadcast failed (worker=%d): %v", workerID, err)
				removeClient(client.conn)
			}
		}
	}
}

func HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS] Upgrade failed: %v", err)
		return
	}

	addClient(conn)

	log.Println("[WS] Client connected thanh cong!")

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			removeClient(conn)
			log.Println("[WS] Client disconnected!")
			break
		}
	}
}

func addClient(conn *gorillaws.Conn) {
	clientsMu.Lock()
	activeClients[conn] = &wsClient{conn: conn}
	clientsMu.Unlock()
}

func removeClient(conn *gorillaws.Conn) {
	var client *wsClient
	clientsMu.Lock()
	client = activeClients[conn]
	delete(activeClients, conn)
	clientsMu.Unlock()
	if client != nil {
		client.close()
		return
	}
	_ = conn.Close()
}

func snapshotClients() []*wsClient {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	clients := make([]*wsClient, 0, len(activeClients))
	for _, client := range activeClients {
		clients = append(clients, client)
	}
	return clients
}

func PublishEvent(event any) {
	select {
	case BroadcastChannel <- event:
	default:
		log.Printf("[WS] Broadcast channel is full, dropping event: %T", event)
	}
}
