package realtime

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn    *websocket.Conn
	StoreID int
	writeMu sync.Mutex
}

func NewClient(conn *websocket.Conn, storeID int) *Client {
	return &Client{
		Conn:    conn,
		StoreID: storeID,
	}
}

func (c *Client) WriteJSON(payload interface{}) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.Conn.WriteJSON(payload)
}

func (c *Client) Close() {
	_ = c.Conn.Close()
}

type Hub struct {
	mu      sync.RWMutex
	clients map[int]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int]map[*Client]struct{}),
	}
}

func (h *Hub) Register(client *Client) {
	if client == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[client.StoreID]; !ok {
		h.clients[client.StoreID] = make(map[*Client]struct{})
	}
	h.clients[client.StoreID][client] = struct{}{}
}

func (h *Hub) Unregister(client *Client) {
	if client == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if storeClients, ok := h.clients[client.StoreID]; ok {
		delete(storeClients, client)
		if len(storeClients) == 0 {
			delete(h.clients, client.StoreID)
		}
	}
	client.Close()
}

func (h *Hub) Broadcast(storeID int, payload interface{}) {
	h.mu.RLock()
	storeClients := h.clients[storeID]
	clients := make([]*Client, 0, len(storeClients))
	for client := range storeClients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		if err := client.WriteJSON(payload); err != nil {
			h.Unregister(client)
		}
	}
}

var QueueHub = NewHub()
