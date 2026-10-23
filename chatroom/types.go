package chatroom

type Message struct {
	From    string
	Content string
	Kind    string
	Time    string
}

type JoinArgs struct {
	Username string
}

type JoinReply struct {
	Success bool
	Message string
}

type SendMessageArgs struct {
	Username string
	Content  string
}

type ActionReply struct {
	Success bool
	Message string
}

type PollArgs struct {
	Username string
}

type PollReply struct {
	Messages []Message
	Success  bool
	Message  string
}

type ListUsersArgs struct{}

type ListUsersReply struct {
	Users []string
}

type LeaveArgs struct {
	Username string
}