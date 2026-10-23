package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"chat-room/chatroom"
)

const (
    cleanupInterval = 5 * time.Second
    userTimeout     = 15 * time.Second
)


type User struct {
	Messages []chatroom.Message
	LastSeen time.Time
}

type Server struct {
	mu    sync.Mutex
	users map[string]*User
}

func NewServer() *Server {
	return &Server{
		users: make(map[string]*User),
	}
}

func (s *Server) Join(
	args *chatroom.JoinArgs,
	reply *chatroom.JoinReply,
) error {
	username := strings.TrimSpace(args.Username)

	if username == "" {
		reply.Success = false
		reply.Message = "Username cannot be empty"
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[username]; exists {
		reply.Success = false
		reply.Message = "Username already taken"
		return nil
	}

	s.users[username] = &User{
		Messages: make([]chatroom.Message, 0),
		LastSeen: time.Now(),
	}

	s.broadcastSystemLocked(
		username + " joined the chat",
		username,
	)

	reply.Success = true
	reply.Message = "Welcome, " + username
	return nil
}

func (s *Server) SendMessage(
	args *chatroom.SendMessageArgs,
	reply *chatroom.ActionReply,
) error {
	content := strings.TrimSpace(args.Content)

	if content == "" {
		reply.Success = false
		reply.Message = "Message cannot be empty"
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	sender, exists := s.users[args.Username]
	if !exists {
		reply.Success = false
		reply.Message = "User is not connected"
		return nil
	}

	sender.LastSeen = time.Now()

	msg := chatroom.Message{
		From:    args.Username,
		Content: content,
		Kind:    "chat",
		Time:    time.Now().Format("15:04:05"),
	}

	for username, user := range s.users {
		if username == args.Username {
			continue
		}

		user.Messages = append(user.Messages, msg)
	}

	reply.Success = true
	reply.Message = "Message sent"
	return nil
}

func (s *Server) PollMessages(
	args *chatroom.PollArgs,
	reply *chatroom.PollReply,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[args.Username]
	if !exists {
		reply.Success = false
		reply.Message = "User is not connected"
		return nil
	}

	user.LastSeen = time.Now()

	reply.Messages = user.Messages
	user.Messages = make([]chatroom.Message, 0)

	reply.Success = true
	return nil
}

func (s *Server) ListUsers(
	args *chatroom.ListUsersArgs,
	reply *chatroom.ListUsersReply,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	reply.Users = make([]string, 0, len(s.users))

	for username := range s.users {
		reply.Users = append(reply.Users, username)
	}

	sort.Strings(reply.Users)
	return nil
}


func (s *Server) Leave(
	args *chatroom.LeaveArgs,
	reply *chatroom.ActionReply,
) error {
	fmt.Println("DEBUG: Leave called for:", args.Username)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[args.Username]; !exists {
		fmt.Println("DEBUG: User already disconnected:", args.Username)

		reply.Success = true
		reply.Message = "User already disconnected"
		return nil
	}

	delete(s.users, args.Username)

	fmt.Println("DEBUG: User removed:", args.Username)

	s.broadcastSystemLocked(
		args.Username+" left the chat",
		"",
	)

	reply.Success = true
	reply.Message = "Disconnected"
	return nil
}


// Caller must already hold s.mu.
func (s *Server) broadcastSystemLocked(
	content string,
	exclude string,
) {
	msg := chatroom.Message{
		Content: content,
		Kind:    "system",
		Time:    time.Now().Format("15:04:05"),
	}

	for username, user := range s.users {
		if username == exclude {
			continue
		}

		user.Messages = append(user.Messages, msg)
	}
}

// Remove clients that have stopped polling.

func (s *Server) cleanupInactive() {
    ticker := time.NewTicker(cleanupInterval)
    defer ticker.Stop()

    for range ticker.C {
        s.mu.Lock()

        now := time.Now()

        for username, user := range s.users {
            fmt.Println(
                "Checking user:",
                username,
                "Inactive for:",
                time.Since(user.LastSeen),
            )

            if now.Sub(user.LastSeen) > userTimeout {
                fmt.Println("DEBUG: Timeout detected for:", username)

                delete(s.users, username)

                s.broadcastSystemLocked(
                    username+" disconnected unexpectedly",
                    "",
                )

                fmt.Println("Removed inactive user:", username)
            }
        }

        s.mu.Unlock()
    }
}


func main() {
	server := NewServer()

	if err := rpc.RegisterName("Chat", server); err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Chat server listening on :1234")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(
		interrupt,
		os.Interrupt,
		syscall.SIGTERM,
	)

	go server.cleanupInactive()

	go func() {
		<-interrupt
		fmt.Println("\nShutting down server...")
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}

		go rpc.ServeConn(conn)
	}

	fmt.Println("Server stopped")
}