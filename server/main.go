package main

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type ChatServer struct {	
	clients   map[net.Conn]string  
	mutex     sync.RWMutex         
	broadcast chan Message         
}

type Message struct {
	sender  string
	content string
	conn    net.Conn 
}

var chatServer = &ChatServer{
	clients:   make(map[net.Conn]string),
	broadcast: make(chan Message, 100), 
}

func main() {
	go chatServer.handleBroadcast()

	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
	defer listener.Close()

	fmt.Println("🚀 TCP Chat Server started on 127.0.0.1:8080")
	fmt.Println("Waiting for clients to connect...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}

		go handleClient(conn)
	}
}

func (cs *ChatServer) handleBroadcast() {
	for message := range cs.broadcast {
		cs.mutex.RLock()
		for conn := range cs.clients {
			if conn == message.conn {
				continue
			}
			
			formattedMsg := fmt.Sprintf("[%s]: %s", message.sender, message.content)
			
			conn.Write([]byte(formattedMsg))
		}
		cs.mutex.RUnlock()
	}
}

func (cs *ChatServer) addClient(conn net.Conn, name string) {
	cs.mutex.Lock()
	cs.clients[conn] = name
	cs.mutex.Unlock()
	
	fmt.Printf("✅ Client connected: %s (%s)\n", conn.RemoteAddr(), name)
	fmt.Printf("📊 Total clients: %d\n", len(cs.clients))
}	

func (cs *ChatServer) removeClient(conn net.Conn) {
	cs.mutex.Lock()
	clientName := cs.clients[conn]
	delete(cs.clients, conn)
	cs.mutex.Unlock()
	
	fmt.Printf("❌ Client disconnected: %s (%s)\n", conn.RemoteAddr(), clientName)
	fmt.Printf("📊 Total clients: %d\n", len(cs.clients))
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	defer chatServer.removeClient(conn)
	
	clientAddr := conn.RemoteAddr().String()
	
	chatServer.addClient(conn, clientAddr)
	
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return
		}
		
		message := string(buffer[:n])
		fmt.Printf("📩 Message from %s: %s", clientAddr, message)
		
		chatServer.broadcast <- Message{
			sender:  clientAddr,
			content: message,
			conn:    conn,
		}
	}
} 