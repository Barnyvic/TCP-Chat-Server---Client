package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal("Failed to connect to server:", err)
	}
	defer conn.Close()

	fmt.Println("🔗 Connected to TCP Chat Server!")
	
	setupGracefulShutdown(conn)

	go readFromServer(conn)

	readFromUser(conn)
}

func readFromServer(conn net.Conn) {
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("\n❌ Connection to server lost")
			os.Exit(1)
		}
		
		message := string(buffer[:n])
		if message != "" {
			fmt.Print(message)
		}
	}
}

func readFromUser(conn net.Conn) {
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		if !scanner.Scan() {
			break 	
		}
		
		message := strings.TrimSpace(scanner.Text())
		
		if message == "" {
			continue
		}
		
		_, err := conn.Write([]byte(message + "\n"))
		if err != nil {
			fmt.Println("❌ Failed to send message:", err)
			return
		}
	}
}

func setupGracefulShutdown(conn net.Conn) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		fmt.Println("\n👋 Disconnecting gracefully...")
		conn.Close()
		os.Exit(0)
	}()
} 