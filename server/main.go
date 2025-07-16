package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
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
	blockedUsers map[string]map[string]bool
	privateHistory map[string][]Message
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
	recipient string
	isPrivate bool
}


var sharedKey = []byte{
	0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
	0xfe, 0xdc, 0xba, 0x98, 0x76, 0x54, 0x32, 0x10,
	0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88,
	0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00,
}

var chatServer = &ChatServer{
	clients:   make(map[net.Conn]*Client),
	username:  make(map[string]net.Conn),
	broadcast: make(chan Message, 100), 
	blockedUsers: make(map[string]map[string]bool),
	privateHistory: make(map[string][]Message),
}

type Encryption struct {
	key []byte
}

func main() {
	go chatServer.handleBroadcast()

	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
	defer listener.Close()

	fmt.Println("TCP Chat Server started on 127.0.0.1:8080")
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
	encryption := &Encryption{key: sharedKey}
	
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
			
			encryptedMsg, err := encryption.Encrypt([]byte(formattedMsg))
			if err != nil {
				log.Printf("Encryption error: %v", err)
				continue
			}
			
			conn.Write(encryptedMsg)
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
	
	fmt.Printf("Client connected: %s (%s)\n", name, client.address)
	fmt.Printf("Total clients: %d\n", len(cs.clients))
}	

func (cs *ChatServer) removeClient(conn net.Conn) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	
	if client, exists := cs.clients[conn]; exists {
		delete(cs.clients, conn)
		delete(cs.username, client.username)
		fmt.Printf("Client disconnected: %s (%s)\n", client.username, client.address)
		fmt.Printf("Total clients: %d\n", len(cs.clients))
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
		content: fmt.Sprintf("Username changed: %s -> %s", oldName, newName),
		conn: conn,
		messageType: "username_change",
	}

	fmt.Printf("Username changed: %s -> %s\n", oldName, newName)
	return true
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	defer chatServer.removeClient(conn)

	encryption := &Encryption{key: sharedKey}

	welcomeMsg := "Welcome to TCP Chat!\nPlease enter your username (3-20 characters, letters, numbers, underscore only): "
	encryptedWelcome, err := encryption.Encrypt([]byte(welcomeMsg))
	if err != nil {
		log.Printf("Failed to encrypt welcome message: %v", err)
		return
	}
	conn.Write(encryptedWelcome)
	
	
	buffer := make([]byte, 1024)
	var username string

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return
		}

	
		decryptedData, err := encryption.Decrypt(buffer[:n])
		if err != nil {
			log.Printf("Failed to decrypt message: %v", err)
			continue
		}

		proposedUsername := strings.TrimSpace(string(decryptedData))

		
		if !chatServer.isUserNameValid(proposedUsername) {
			errorMsg := "Invalid username. Please use 3-20 characters, letters, numbers, underscore only.\n"
			encryptedError, err := encryption.Encrypt([]byte(errorMsg))
			if err != nil {
				log.Printf("Failed to encrypt error message: %v", err)
				continue
			}
			conn.Write(encryptedError)
			continue
		}

		if !chatServer.isUsernameAvailable(proposedUsername) {
			errorMsg := "Username already taken. Please choose another.\n"
			encryptedError, err := encryption.Encrypt([]byte(errorMsg))
			if err != nil {
				log.Printf("Failed to encrypt error message: %v", err)
				continue
			}
			conn.Write(encryptedError)
			continue
		}
            
		username = proposedUsername
		break
	}

	chatServer.addClient(conn, username)
	defer chatServer.removeClient(conn)

	welcomeMsg = fmt.Sprintf("Welcome %s! You can now start chatting.\nCommands: /nick <newname> to change username, /quit to exit\n", username)
	encryptedWelcome, err = encryption.Encrypt([]byte(welcomeMsg))
	if err != nil {
		log.Printf("Failed to encrypt welcome message: %v", err)
		return
	}
	conn.Write(encryptedWelcome)

	chatServer.broadcast <- Message{
		sender: "System",
		content: fmt.Sprintf("%s joined the chat", username),
		conn: nil,
		messageType: "system",
	}

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			chatServer.broadcast <- Message{
				sender: "System",
				content: fmt.Sprintf("%s left the chat", username),
				conn: nil,
				messageType: "system",
			}
			return
		}

		decryptedData, err := encryption.Decrypt(buffer[:n])
		if err != nil {
			log.Printf("Failed to decrypt message: %v", err)
			continue
		}

		message := strings.TrimSpace(string(decryptedData))
		if message == "" {
			continue
		}

		if strings.HasPrefix(message, "/") {
				handleCommand(conn, message, username)
				continue
		}

		fmt.Printf("Message from %s: %s", username, message)

		chatServer.broadcast <- Message{
			sender: username,
			content: message,
			conn: conn,
			messageType: "chat",
		}
	}
} 

func handleCommand(conn net.Conn, command, currentUsername string) {
	encryption := &Encryption{key: sharedKey}
	
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}
	
	switch parts[0] {
	case "/nick":
		if len(parts) != 2 {
			errorMsg := "Usage: /nick <newusername>\n"
			encryptedError, err := encryption.Encrypt([]byte(errorMsg))
			if err != nil {
				log.Printf("Failed to encrypt error message: %v", err)
				return
			}
			conn.Write(encryptedError)
			return
		}
		
		newUsername := parts[1]
		
		if !chatServer.isUserNameValid(newUsername) {
			errorMsg := "Invalid username. Please use 3-20 characters (letters, numbers, underscore only)\n"
			encryptedError, err := encryption.Encrypt([]byte(errorMsg))
			if err != nil {
				log.Printf("Failed to encrypt error message: %v", err)
				return
			}
			conn.Write(encryptedError)
			return
		}
		
		if !chatServer.isUsernameAvailable(newUsername) {
			errorMsg := "Username already taken. Please choose another.\n"
			encryptedError, err := encryption.Encrypt([]byte(errorMsg))
			if err != nil {
				log.Printf("Failed to encrypt error message: %v", err)
				return
			}
			conn.Write(encryptedError)
			return
		}
		
		if chatServer.changeUsername(conn, newUsername) {
			successMsg := fmt.Sprintf("Username changed to: %s\n", newUsername)
			encryptedSuccess, err := encryption.Encrypt([]byte(successMsg))
			if err != nil {
				log.Printf("Failed to encrypt success message: %v", err)
				return
			}
			conn.Write(encryptedSuccess)
		} else {
			errorMsg := "Failed to change username\n"
			encryptedError, err := encryption.Encrypt([]byte(errorMsg))
			if err != nil {
				log.Printf("Failed to encrypt error message: %v", err)
				return
			}
			conn.Write(encryptedError)
		}
		
	case "/quit":
		goodbyeMsg := "Goodbye!\n"
		encryptedGoodbye, err := encryption.Encrypt([]byte(goodbyeMsg))
		if err != nil {
			log.Printf("Failed to encrypt goodbye message: %v", err)
		} else {
			conn.Write(encryptedGoodbye)
		}
		conn.Close()
		
	default:
		errorMsg := "Unknown command. Available commands: /nick <newname>, /quit\n"
		encryptedError, err := encryption.Encrypt([]byte(errorMsg))
		if err != nil {
			log.Printf("Failed to encrypt error message: %v", err)
			return
		}
		conn.Write(encryptedError)
	}
}



func NewEncryption() (*Encryption, error) {
	key := make([]byte, 32)
	 if _, err := io.ReadFull(rand.Reader, key); err != nil {
        return nil, err
    }

	return &Encryption{key: key}, nil
}


func (e *Encryption) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func (e *Encryption) Decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		 return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

