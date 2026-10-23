package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"chat-room/chatroom"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter username: ")
	if !scanner.Scan() {
		return
	}

	username := strings.TrimSpace(scanner.Text())
	if username == "" {
		fmt.Println("Username cannot be empty")
		return
	}

	conn, err := net.DialTimeout(
		"tcp",
		"localhost:1234",
		5*time.Second,
	)
	if err != nil {
		fmt.Println("Cannot connect to server:", err)
		return
	}

	client := rpc.NewClient(conn)
	defer client.Close()

	var joinReply chatroom.JoinReply

	err = client.Call(
		"Chat.Join",
		&chatroom.JoinArgs{Username: username},
		&joinReply,
	)
	if err != nil {
		fmt.Println("Join failed:", err)
		return
	}

	if !joinReply.Success {
		fmt.Println(joinReply.Message)
		return
	}

	fmt.Println(joinReply.Message)
	fmt.Println("Commands: /users, /quit")
	fmt.Println("Type a message and press Enter.")

	ctx, stopSignal := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stopSignal()

	runCtx, cancel := context.WithCancel(ctx)

	// Receive messages independently of terminal input.
	go pollMessages(runCtx, client, username)

	input := make(chan string)

	go func() {
		for scanner.Scan() {
			input <- scanner.Text()
		}
		close(input)
	}()

	for {
		select {
		case <-ctx.Done():
			goto shutdown

		case line, ok := <-input:
			if !ok {
				goto shutdown
			}

			line = strings.TrimSpace(line)

			if line == "" {
				continue
			}

			switch line {
			case "/quit":
				goto shutdown

			case "/users":
				var reply chatroom.ListUsersReply

				err := client.Call(
					"Chat.ListUsers",
					&chatroom.ListUsersArgs{},
					&reply,
				)
				if err != nil {
					fmt.Println("List users failed:", err)
					continue
				}

				fmt.Println("Connected users:", strings.Join(reply.Users, ", "))

			default:
				var reply chatroom.ActionReply

				err := client.Call(
					"Chat.SendMessage",
					&chatroom.SendMessageArgs{
						Username: username,
						Content:  line,
					},
					&reply,
				)
				if err != nil {
					fmt.Println("Send failed:", err)
					goto shutdown
				}

				if !reply.Success {
					fmt.Println(reply.Message)
				}
			}
		}
	}

shutdown:
	cancel()

	var leaveReply chatroom.ActionReply

	err = client.Call(
		"Chat.Leave",
		&chatroom.LeaveArgs{Username: username},
		&leaveReply,
	)
	if err != nil {
		fmt.Println("Could not notify server about leaving:", err)
	}

	fmt.Println("Disconnected.")
}

func pollMessages(
	ctx context.Context,
	client *rpc.Client,
	username string,
) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			var reply chatroom.PollReply

			err := client.Call(
				"Chat.PollMessages",
				&chatroom.PollArgs{Username: username},
				&reply,
			)

			if err != nil {
				fmt.Println("\nPolling failed:", err)
				return
			}

			if !reply.Success {
				fmt.Println("\nPolling stopped:", reply.Message)
				return
			}

			for _, msg := range reply.Messages {
				if msg.Kind == "system" {
					fmt.Printf(
						"\n[%s] * %s\n",
						msg.Time,
						msg.Content,
					)
				} else {
					fmt.Printf(
						"\n[%s] %s: %s\n",
						msg.Time,
						msg.From,
						msg.Content,
					)
				}
			}
		}
	}
}