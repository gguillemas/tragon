package tragon

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"testing"
)

const (
	delta       = 1
	testTimeout = 1
)

var errTimeout error = errors.New("timeout")

type testCase struct {
	Name     string
	Session  string
	WantData []byte
}

var testSuite []testCase = []testCase{
	testCase{
		Name: "email",
		Session: `HELO tragon
MAIL FROM: tragon@example.com
RCPT TO: tragon@example.com
DATA
This is a test of a simple email.
.
QUIT
`,
		WantData: []byte(`This is a test of a simple email.
`),
	},
	testCase{
		Name: "email-with-headers",
		Session: `HELO tragon
MAIL FROM: tragon@example.com
RCPT TO: tragon@example.com
DATA
Content-Language: en
Subject: Email with headers
This is a test of an email with headers.
.
QUIT
`,
		WantData: []byte(`Content-Language: en
Subject: Email with headers
This is a test of an email with headers.
`),
	},
	testCase{
		Name: "email-incomplete",
		Session: `HELO tragon
MAIL FROM: tragon@example.com
RCPT TO: tragon@example.com
DATA
This is the start of an email that will never finish.
`,
		WantData: []byte(`This is the start of an email that will never finish.
`),
	},
}

func TestHandleMessages(t *testing.T) {
	logger := log.New(os.Stderr, "tragon: ", log.Llongfile)

	for _, tc := range testSuite {
		log.Printf("running \"%v\" test", tc.Name)

		var err error
		var data []byte
		var wg sync.WaitGroup

		wg.Add(1)
		s := New(":0", testTimeout, DefaultReplies, logger, func(message []byte) {
			data = message
			wg.Done()
		})
		s.ListenAndServe()

		conn, err := net.Dial("tcp", s.addr)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Fprintf(conn, tc.Session)

		wg.Wait()
		conn.Close()

		if bytes.Compare(data, tc.WantData) != 0 {
			t.Fatalf("expected \"%s\", got \"%s\"", tc.WantData, data)
		}
	}
}
