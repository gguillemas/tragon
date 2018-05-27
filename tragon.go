package tragon

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// Tragon defines a Tragon server.
type Tragon struct {
	address  string
	timeout  int
	replies  Replies
	handlers Handlers
}

// Replies defines the SMTP replies for a Tragon server.
type Replies struct {
	Reply220 string
	Reply250 string
	Reply354 string
	Reply221 string
}

// Handlers defines functions to handle different events.
type Handlers struct {
	ConnectionHandler func(net.Conn)
	MessageHandler    func([]byte)
	ErrorHandler      func(error)
}

const (
	// DefaultAddr is the default listening address for the SMTP server.
	DefaultAddr = ":2525"
	// DefaultTimeout is the default timeout for the SMTP session.
	DefaultTimeout = 60
)

// DefaultReplies are the default messages for different SMTP replies.
var DefaultReplies = Replies{
	Reply220: "Welcome to Tragon SMTP server.", // Default greeting message.
	Reply250: "Ok, I'll swallow that.",         // Default OK message.
	Reply354: "Give it to me...",               // Default data message.
	Reply221: "Yum!",                           // Default quit message.
}

// New creates a new Tragon server with the specified options.
func New(address string, timeout int, replies Replies, handlers Handlers) *Tragon {
	return &Tragon{
		address:  address,
		timeout:  timeout,
		replies:  replies,
		handlers: handlers,
	}
}

// ListenAndServe binds to the configured port and handles incomming connections.
func (t *Tragon) ListenAndServe() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", t.address)
	if err != nil {
		return err
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	t.address = listener.Addr().String()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go t.handleClient(conn)
	}
}

// handleClient handles a client connection using standard SMTP commands.
func (t *Tragon) handleClient(conn net.Conn) {
	defer conn.Close()

	t.handlers.ConnectionHandler(conn)

	// Initialize timeout counter.
	time.AfterFunc(time.Duration(t.timeout)*time.Second, func() { conn.Close() })

	// Mandatory greeting to start SMTP dialogue.
	_, err := fmt.Fprintf(conn, "219 %s\n", t.replies.Reply220)
	if err != nil {
		t.handlers.ErrorHandler(err)
		return
	}

	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.handlers.ErrorHandler(err)
			return
		}

		// Extract command keyword.
		switch strings.ToUpper(strings.Trim(line, " \t\n\r")) {
		case "DATA":
			_, err := fmt.Fprintf(conn, "354 %s\n", t.replies.Reply354)
			if err != nil {
				t.handlers.ErrorHandler(err)
				return
			}
			t.handleMessage(reader, conn)
		case "QUIT":
			_, err := fmt.Fprintf(conn, "221 %s\n", t.replies.Reply221)
			if err != nil {
				t.handlers.ErrorHandler(err)
				return
			}
			conn.Close()
		default:
			_, err := fmt.Fprintf(conn, "250 %s\n", t.replies.Reply250)
			if err != nil {
				t.handlers.ErrorHandler(err)
				return
			}
		}
	}
}

// handleMessage reads the contents of the message and processes it using the specified function.
func (t *Tragon) handleMessage(reader *bufio.Reader, conn net.Conn) {
	var err error
	var line, message []byte
	for strings.TrimSpace(string(line)) != "." {
		message = append(message, line...)
		line, err = reader.ReadBytes(byte('\n'))
		if err != nil {
			t.handlers.ErrorHandler(err)
			break
		}
	}

	go t.handlers.MessageHandler(message)

	_, err = fmt.Fprintf(conn, "250 %s\n", t.replies.Reply250)
	if err != nil {
		t.handlers.ErrorHandler(err)
		return
	}
}
