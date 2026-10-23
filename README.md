# Go RPC Chat Room

A concurrent client-server chat application built with **Go** and **net/rpc**. The application allows multiple clients to join a shared chat room, exchange messages, view connected users, and handle user disconnections.

## Features

- **RPC Communication:** Uses Go's `net/rpc` package for client-server communication.
- **Multiple Clients:** Supports multiple clients connecting to the same server.
- **Username Validation:** Rejects empty and duplicate usernames.
- **Real-Time Chat:** Clients periodically poll the server for incoming messages.
- **Connected Users:** Lists currently connected users using the `/users` command.
- **System Notifications:** Notifies users when someone joins, leaves, or disconnects unexpectedly.
- **Concurrency Safety:** Uses goroutines and `sync.Mutex` to handle concurrent requests safely.
- **Inactive User Cleanup:** Removes inactive users after a configured timeout.
- **Graceful Client Exit:** Supports leaving the chat using `/quit` and handles Ctrl+C.

## Technologies Used

- Go
- `net/rpc`
- Goroutines
- Channels
- `sync.Mutex`
- TCP Networking

## Project Structure

```text
chat-room/
├── go.mod
├── README.md
├── chatroom/
│   └── types.go
└── cmd/
    ├── server/
    │   └── main.go
    └── client/
        └── main.go
```

## Requirements

- Go installed on your machine.
- A terminal or command-line interface.
- Git (optional, for cloning the repository).

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/Gemmedmohamed/go-rpc-chat-room.git
cd go-rpc-chat-room
```

### 2. Start the Server

Open a terminal in the project root and run:

```bash
go run ./cmd/server
```

The server listens on port `1234` by default.

### 3. Start a Client

Open another terminal in the project root and run:

```bash
go run ./cmd/client
```

Enter a username when prompted.

To connect additional users, run the client command in separate terminals.

## Available Commands

| Command | Description |
|---|---|
| `/users` | Displays the currently connected users. |
| `/quit` | Leaves the chat room and disconnects the client. |
| `Ctrl+C` | Interrupts the client or server process, depending on which terminal receives the signal. |

Any ordinary text entered by a client is sent as a chat message to the other connected users.

## How It Works

1. The server starts a TCP listener and registers the `Chat` RPC service.
2. Each client connects to the server and joins using a unique username.
3. Clients send messages through RPC calls.
4. The server distributes messages to the other connected users.
5. Clients periodically poll the server to retrieve queued messages.
6. The server tracks user activity and removes inactive users after the configured timeout.
7. System messages notify connected users about join, leave, and unexpected-disconnection events.

## Concurrency and Synchronization

The server uses goroutines to handle incoming RPC connections concurrently. A mutex protects shared user data and message queues to prevent concurrent access from corrupting the server's state.

## Configuration

The server's default address is:

```text
:1234
```

The inactive-user cleanup settings are configured in the server code:

- Cleanup interval: 5 seconds
- Inactivity timeout: 15 seconds

## Academic Context

Developed as a Distributed Systems assignment to practice RPC communication, TCP networking, concurrency, synchronization, and client-server architecture using Go.

## License

This project was developed for educational purposes.
