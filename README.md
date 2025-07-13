# TCP Chat Server & Client

A simple TCP-based chat application built in Go to demonstrate:

- TCP server/client networking with `net` package
- Goroutines for concurrent client handling
- Message broadcasting between multiple clients
- CLI-based chat interface

## Project Structure

```
tcp-chat/
├── go.mod              # Go module definition
├── server/
│   └── main.go         # TCP Chat Server
├── client/
│   └── main.go         # TCP Chat Client
└── README.md           # This file
```

## How to Run

### Clone the Repository

```bash
git clone <repository-url>
cd tcp-chat
```

### Start the Server

```bash
go run server/main.go
```

### Connect Clients (in separate terminals)

```bash
go run client/main.go
```

### Build Binaries (Optional)

```bash
go build -o tcp-chat-server server/main.go

go build -o tcp-chat-client client/main.go

./tcp-chat-server
./tcp-chat-client
```

## 🏗️ Architecture

### Server Architecture

```
┌─────────────────────────────────────────┐
│              TCP Server                 │
│                                         │
│  ┌─────────────┐    ┌─────────────────┐ │
│  │ ChatServer  │    │ Message Channel │ │
│  │ - clients   │◄──►│ (buffered)      │ │
│  │ - mutex     │    │                 │ │
│  │ - broadcast │    └─────────────────┘ │
│  └─────────────┘                        │
│         │                               │
│  ┌─────────────┐    ┌─────────────────┐ │
│  │handleClient │    │ handleBroadcast │ │
│  │(goroutine)  │    │ (goroutine)     │ │
│  └─────────────┘    └─────────────────┘ │
└─────────────────────────────────────────┘
```

### Client Architecture

```
┌─────────────────────────────────────────┐
│           TCP Client                    │
│                                         │
│  ┌─────────────┐    ┌─────────────────┐ │
│  │readFromUser │    │ readFromServer  │ │
│  │(main thread)│    │ (goroutine)     │ │
│  └─────────────┘    └─────────────────┘ │
│         │                    │          │
│  ┌─────────────┐    ┌─────────────────┐ │
│  │ bufio.Scanner│    │ conn.Read()     │ │
│  │ (stdin)     │    │ (network)       │ │
│  └─────────────┘    └─────────────────┘ │
└─────────────────────────────────────────┘
```


### Git Workflow

```bash
git add .

git commit -m "Add new feature"

git push origin main
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
