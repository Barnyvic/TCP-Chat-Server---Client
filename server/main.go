package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type ChatServer struct {	
	clients   map[net.Conn]*Client  
	username  map[string]net.Conn
	mutex     sync.RWMutex         
	broadcast chan Message         
}

type Client struct {
	conn net.Conn
	username string
	address string
}

type Message struct {
	sender  string
	content string
	conn    net.Conn 
	messageType string
}

var chatServer = &ChatServer{
	clients:   make(map[net.Conn]*Client),
	username:  make(map[string]net.Conn),
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
			if message.messageType != "system" && conn == message.conn {
				continue
			}

			var formattedMsg string

			switch message.messageType {
				case "username_change":
					formattedMsg = fmt.Sprintf("*** %s ***\n", message.content)
			     case "system":
					formattedMsg = fmt.Sprintf("*** %s ***\n", message.content)
				default:
					formattedMsg = fmt.Sprintf("[%s]: %s", message.sender, message.content)
			}
			
			
			
			conn.Write([]byte(formattedMsg))
		}
		cs.mutex.RUnlock()
	}
}

func (cs * ChatServer) isUserNameValid(name string) bool {
	if len(name) < 3 || len(name) > 20 {
		return false
	}

	for _, char := range name{
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}

func (cs *ChatServer) isUsernameAvailable(name string) bool {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()
	_,exists := cs.username[name]
	return !exists
}

func (cs *ChatServer) addClient(conn net.Conn, name string) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	client := &Client {
		conn: conn,
		username: name,
		address: conn.RemoteAddr().String(),
	}
	
	cs.clients[conn] = client
	cs.username[name] = conn
	
	fmt.Printf("✅ Client connected: %s (%s)\n", name, client.address)
	fmt.Printf("📊 Total clients: %d\n", len(cs.clients))
}	

func (cs *ChatServer) removeClient(conn net.Conn) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	
	if client, exists := cs.clients[conn]; exists {
		delete(cs.clients, conn)
		delete(cs.username, client.username)
		fmt.Printf("❌ Client disconnected: %s (%s)\n", client.username, client.address)
		fmt.Printf("📊 Total clients: %d\n", len(cs.clients))
	}
	
}


func (cs *ChatServer) changeUsername(conn net.Conn, newName string) bool {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

     client, exists := cs.clients[conn]
	 if !exists {
		return false
	 }

	delete(cs.username, client.username)

	oldName := client.username
	client.username = newName
	cs.username[newName] = conn

	cs.broadcast <- Message{
		sender: client.username,
		content: fmt.Sprintf("✅ Username changed: %s -> %s", oldName, newName),
		conn: conn,
		messageType: "username_change",
	}

	fmt.Printf("✅ Username changed: %s -> %s\n", oldName, newName)
	return true
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	defer chatServer.removeClient(conn)

	conn.Write([]byte("Welcome to TCP Chat!\nPlease enter your username (3-20 characters, letters, numbers, underscore only): "))
	
	
	buffer := make([]byte, 1024)
	var username string

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return
		}

		proposedUsername := strings.TrimSpace(string(buffer[:n]))

		
		if !chatServer.isUserNameValid(proposedUsername) {
			conn.Write([]byte("Invalid username. Please use 3-20 characters, letters, numbers, underscore only.\n"))
			continue
		}

		if !chatServer.isUsernameAvailable(proposedUsername) {
			conn.Write([]byte("Username already taken. Please choose another.\n"))
			continue
		}
            
		username = proposedUsername
		break
	}

	chatServer.addClient(conn, username)
	defer chatServer.removeClient(conn)

	conn.Write([]byte(fmt.Sprintf("Welcome %s! You can now start chatting.\nCommands: /nick <newname> to change username, /quit to exit\n", username)))

	chatServer.broadcast <- Message{
		sender: "System",
		content: fmt.Sprintf("✅ %s joined the chat", username),
		conn: nil,
		messageType: "system",
	}

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			chatServer.broadcast <- Message{
				sender: "System",
				content: fmt.Sprintf("❌ %s left the chat", username),
				conn: nil,
				messageType: "system",
			}
			return
		}

		message := strings.TrimSpace(string(buffer[:n]))
		if message == "" {
			continue
		}

		if strings.HasPrefix(message, "/") {
				handleCommand(conn, message, username)
				continue
		}

		fmt.Printf("📩 Message from %s: %s", username, message)

		chatServer.broadcast <- Message{
			sender: username,
			content: message,
			conn: conn,
			messageType: "chat",
		}
	}
} 

func handleCommand(conn net.Conn, command, currentUsername string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}
	
	switch parts[0] {
	case "/nick":
		if len(parts) != 2 {
			conn.Write([]byte("Usage: /nick <newusername>\n"))
			return
		}
		
		newUsername := parts[1]
		
		if !chatServer.isUserNameValid(newUsername) {
			conn.Write([]byte("Invalid username. Please use 3-20 characters (letters, numbers, underscore only)\n"))
			return
		}
		
		if !chatServer.isUsernameAvailable(newUsername) {
			conn.Write([]byte("Username already taken. Please choose another.\n"))
			return
		}
		
		if chatServer.changeUsername(conn, newUsername) {
			conn.Write([]byte(fmt.Sprintf("Username changed to: %s\n", newUsername)))
		} else {
			conn.Write([]byte("Failed to change username\n"))
		}
		
	case "/quit":
		conn.Write([]byte("Goodbye!\n"))
		conn.Close()
		
	default:
		conn.Write([]byte("Unknown command. Available commands: /nick <newname>, /quit\n"))
	}
}