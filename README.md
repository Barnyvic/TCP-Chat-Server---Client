# TCP Chat Server & Client

A real-time TCP-based chat application built in Go featuring concurrent client handling, message broadcasting, and interactive CLI interface.

## Features

- **Multi-client Support**: Handle unlimited concurrent connections
- **Real-time Messaging**: Instant message broadcasting to all connected clients
- **Thread-safe Operations**: Concurrent access management with goroutines and mutexes
- **Interactive CLI**: User-friendly command-line interface
- **Graceful Shutdown**: Clean exit handling with quit commands and signal interruption
- **Connection Management**: Automatic client tracking and cleanup

## Prerequisites

- Go 1.21 or higher
- Terminal or command prompt

## Installation

```bash
git clone https://github.com/Barnyvic/TCP-Chat-Server---Client
cd tcp-chat
```

## Usage

### Starting the Server

```bash
go run server/main.go
```

Server will start listening on `127.0.0.1:8080` and display:

```
🚀 TCP Chat Server started on 127.0.0.1:8080
Waiting for clients to connect...
```

### Connecting Clients

Open multiple terminal windows and run:

```bash
go run client/main.go
```

Each client will display:

```
🔗 Connected to TCP Chat Server!
📍 Your address: 127.0.0.1:XXXXX
💬 Type your messages and press Enter to send
📝 Type 'quit' or press Ctrl+C to exit
----------------------------------------
>
```

### Building Executables

```bash
go build -o tcp-chat-server server/main.go
go build -o tcp-chat-client client/main.go

./tcp-chat-server
./tcp-chat-client
```

## Architecture

### Server Components

- **ChatServer**: Main server struct managing client connections
- **Message Broadcasting**: Channel-based message distribution system
- **Client Management**: Thread-safe client tracking with RWMutex
- **Connection Handling**: Individual goroutines for each client

### Client Components

- **Concurrent I/O**: Simultaneous reading from server and user input
- **Message Reception**: Dedicated goroutine for server messages
- **User Input**: Main thread handling keyboard input
- **Signal Handling**: Graceful shutdown on interruption

## Technical Implementation

### Server Architecture

```
TCP Server (127.0.0.1:8080)
├── ChatServer Struct
│   ├── clients map[net.Conn]string
│   ├── mutex sync.RWMutex
│   └── broadcast chan Message
├── handleClient() goroutines
├── handleBroadcast() goroutine
└── Connection Management
```

### Client Architecture

```
TCP Client
├── readFromServer() goroutine
├── readFromUser() main thread
├── Signal Handler goroutine
└── Connection Management
```

### Message Flow

1. Client sends message → Server receives in `handleClient()`
2. Server puts message in broadcast channel
3. `handleBroadcast()` distributes to all other clients
4. Clients receive and display formatted messages

## Key Technologies

- **Go net package**: TCP networking functionality
- **Goroutines**: Lightweight concurrent programming
- **Channels**: Inter-goroutine communication
- **Mutexes**: Thread-safe data access
- **Signal handling**: Graceful shutdown management
- **bufio**: Efficient I/O operations

## Testing

### Basic Test Scenario

1. Start the server
2. Connect 2-3 clients in separate terminals
3. Send messages from different clients
4. Verify messages appear in all other clients
5. Test graceful disconnection with 'quit' command

### Expected Behavior

- Messages broadcast to all clients except sender
- Client connection/disconnection logged on server
- Real-time message delivery
- Stable operation under multiple concurrent clients

## Project Structure

```
tcp-chat/
├── .git/
├── .gitignore
├── go.mod
├── README.md
├── server/
│   └── main.go
└── client/
    └── main.go
```

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/name`)
3. Commit changes (`git commit -m 'Add feature'`)
4. Push to branch (`git push origin feature/name`)
5. Open Pull Request

## License

This project is open source and available under the MIT License.
