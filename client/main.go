package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// Create a proper 32-byte key for AES-256 (must match server)
var sharedKey = []byte{
	0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
	0xfe, 0xdc, 0xba, 0x98, 0x76, 0x54, 0x32, 0x10,
	0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88,
	0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00,
}

type Encryption struct {
	key []byte
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

	// GCM can handle messages up to 2^39 - 256 bits, which is much larger than we need
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

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal("Failed to connect to server:", err)
	}
	defer conn.Close()

	fmt.Println("Connected to TCP Chat Server!")
	
	setupGracefulShutdown(conn)

	go readFromServer(conn)

	readFromUser(conn)
}

func readFromServer(conn net.Conn) {
	encryption := &Encryption{key: sharedKey}
	buffer := make([]byte, 1024)
	
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("\nConnection to server lost")
			os.Exit(1)
		}
		

		decryptedData, err := encryption.Decrypt(buffer[:n])
		if err != nil {
			fmt.Printf("Failed to decrypt message: %v\n", err)
			continue
		}
		
		message := string(decryptedData)
		if message != "" {
			fmt.Print(message)
		}
	}
}

func readFromUser(conn net.Conn) {
	encryption := &Encryption{key: sharedKey}
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		if !scanner.Scan() {
			break 	
		}
		
		message := strings.TrimSpace(scanner.Text())
		
		if message == "" {
			continue
		}
		
	
		encryptedMsg, err := encryption.Encrypt([]byte(message + "\n"))
		if err != nil {
			fmt.Println("Failed to encrypt message:", err)
			continue
		}
		
		_, err = conn.Write(encryptedMsg)
		if err != nil {
			fmt.Println("Failed to send message:", err)
			return
		}
	}
}

func setupGracefulShutdown(conn net.Conn) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		fmt.Println("\nDisconnecting gracefully...")
		conn.Close()
		os.Exit(0)
	}()
} 