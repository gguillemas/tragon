package tragon

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"testing"
)

const (
	testTimeout = 1
	delta       = 1
)

func newConnMock(session []byte) connMock {
	buf := bytes.NewBuffer(session)
	r := bufio.NewReader(buf)
	w := bufio.NewWriter(buf)
	return connMock{r: r, w: w}
}

type connMock struct {
	net.Conn
	r io.Reader
	w io.Writer
}

func (cm connMock) Read(b []byte) (int, error) {
	return cm.r.Read(b)
}

func (cm connMock) Write(p []byte) (int, error) {
	return cm.w.Write(p)
}

func (cm connMock) Close() error {
	return nil
}

var errTimeout = errors.New("timeout")

type testCase struct {
	Name     string
	Session  []byte
	WantData []byte
}

var testSuite = []testCase{
	testCase{
		Name: "email",
		Session: []byte(`HELO tragon
MAIL FROM: tragon@example.com
RCPT TO: tragon@example.com
DATA
This is a test of a simple email.
.
QUIT
`),
		WantData: []byte(`This is a test of a simple email.
`),
	},
	testCase{
		Name: "email-with-headers",
		Session: []byte(`HELO tragon
MAIL FROM: tragon@example.com
RCPT TO: tragon@example.com
DATA
Content-Language: en
Subject: Email with headers
This is a test of an email with headers.
.
QUIT
`),
		WantData: []byte(`Content-Language: en
Subject: Email with headers
This is a test of an email with headers.
`),
	},
	testCase{
		Name: "email-incomplete",
		Session: []byte(`HELO tragon
MAIL FROM: tragon@example.com
RCPT TO: tragon@example.com
DATA
This is the start of an email that will never finish.
`),
		WantData: []byte(`This is the start of an email that will never finish.
`),
	},
}

func TestHandleClient(t *testing.T) {
	logger := log.New(os.Stderr, "tragon: ", log.Llongfile)

	for _, tc := range testSuite {
		log.Printf("running \"%v\" test", tc.Name)

		var data []byte
		var wg sync.WaitGroup

		wg.Add(1)
		s := New(":0", testTimeout, DefaultReplies, logger, func(message []byte) {
			data = message
			wg.Done()
		})

		conn := newConnMock(tc.Session)
		s.handleClient(conn)
		conn.Write(tc.Session)

		wg.Wait()

		if bytes.Compare(data, tc.WantData) != 0 {
			t.Fatalf("expected \"%s\", got \"%s\"", tc.WantData, data)
		}
	}
}
