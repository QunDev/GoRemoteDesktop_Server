package websocket

type Hub struct {
	clients map[*Client]bool
	hosts   map[*Client]bool

	broadcast chan map[*Client][]byte

	register chan *Client

	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan map[*Client][]byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		hosts:      make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if client.role == "host" {
				h.hosts[client] = true
			} else if client.role == "client" {
				h.clients[client] = true
			}
		case client := <-h.unregister:
			if client.role == "host" {
				if _, ok := h.hosts[client]; ok {
					delete(h.hosts, client)
					close(client.send)
				}
			} else if client.role == "client" {
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
				}
			}
		case data := <-h.broadcast:
			for client, message := range data {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
