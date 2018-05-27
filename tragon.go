package tragon

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// Tragon defines a Tragon server.
type Tragon struct {
	addr       string
	timeout    int
	replies    Replies
	handleFunc func([]byte)
	log        *log.Logger
}

// Replies defines the SMTP replies for a Tragon server.
type Replies struct {
	Reply220 string
	Reply250 string
	Reply354 string
	Reply221 string
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
func New(addr string, timeout int, replies Replies, logger *log.Logger, handleFunc func([]byte)) *Tragon {
	return &Tragon{
		addr:       addr,
		timeout:    timeout,
		replies:    replies,
		log:        logger,
		handleFunc: handleFunc,
	}
}

// ListenAndServe binds to the configured port and handles incomming connections.
func (t *Tragon) ListenAndServe() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", t.addr)
	if err != nil {
		return err
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	t.addr = listener.Addr().String()

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

	// Initialize timeout counter.
	time.AfterFunc(time.Duration(t.timeout)*time.Second, func() { conn.Close() })

	// Mandatory greeting to start SMTP dialogue.
	_, err := fmt.Fprintf(conn, "219 %s\n", t.replies.Reply220)
	if err != nil {
		t.log.Println(err)
		return
	}

	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.log.Printf("error reading from closed connection: %v", err)
			return
		}

		// Extract command keyword.
		switch strings.ToUpper(strings.Trim(line, " \t\n\r")) {
		case "DATA":
			_, err := fmt.Fprintf(conn, "354 %s\n", t.replies.Reply354)
			if err != nil {
				t.log.Println(err)
				return
			}
			t.handleMessage(reader, conn)
		case "QUIT":
			_, err := fmt.Fprintf(conn, "221 %s\n", t.replies.Reply221)
			if err != nil {
				t.log.Println(err)
				return
			}
			conn.Close()
		default:
			_, err := fmt.Fprintf(conn, "250 %s\n", t.replies.Reply250)
			if err != nil {
				t.log.Println(err)
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
			t.log.Printf("error reading from closed connection: %v", err)
			break
		}
	}

	go t.handleFunc(message)

	_, err = fmt.Fprintf(conn, "250 %s\n", t.replies.Reply250)
	if err != nil {
		t.log.Println(err)
		return
	}
}
