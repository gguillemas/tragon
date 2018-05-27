package tragon

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
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

type testCase struct {
	Name        string
	SessionFile string
	MessageFile string
}

var testSuite = []testCase{
	testCase{
		Name:        "basic",
		SessionFile: "test/sessions/basic",
		MessageFile: "test/messages/basic",
	},
	testCase{
		Name:        "headers",
		SessionFile: "test/sessions/headers",
		MessageFile: "test/messages/headers",
	},
	testCase{
		Name:        "incomplete",
		SessionFile: "test/sessions/incomplete",
		MessageFile: "test/messages/incomplete",
	},
}

func TestHandleClient(t *testing.T) {
	logger := log.New(os.Stderr, "tragon: ", log.Llongfile)

	for _, tc := range testSuite {
		log.Printf("running \"%v\" test", tc.Name)

		var message []byte
		var wg sync.WaitGroup

		wg.Add(1)
		s := New(":0", testTimeout, DefaultReplies, logger, func(m []byte) {
			message = m
			wg.Done()
		})

		session, err := ioutil.ReadFile(tc.SessionFile)
		if err != nil {
			t.Fatal(err)
		}

		conn := newConnMock(session)
		s.handleClient(conn)

		wg.Wait()

		wantMessage, err := ioutil.ReadFile(tc.MessageFile)
		if err != nil {
			t.Fatal(err)
		}

		if bytes.Compare(message, wantMessage) != 0 {
			t.Fatalf("expected \"%v\", got \"%v\"", wantMessage, message)
		}
	}
}
