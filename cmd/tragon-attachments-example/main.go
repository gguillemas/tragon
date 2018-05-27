package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"path"
	"strings"

	"github.com/gguillemas/tragon"
)

var attachmentsDir string

func main() {
	if len(os.Args) < 1 {
		fmt.Println("Usage: tragon-attachments [directory]")
		os.Exit(1)
	}

	// Attachments will be saved to the specified directory.
	attachmentsDir = os.Args[1]

	t := tragon.New(
		tragon.DefaultAddr,
		tragon.DefaultTimeout,
		tragon.DefaultReplies,
		log.New(os.Stderr, "tragon", log.LstdFlags),
		// The attachments of each message will be analyzed.
		analyzeAttachments,
	)

	err := t.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func analyzeAttachments(message []byte) {
	r := bytes.NewReader(message)
	m, err := mail.ReadMessage(r)
	if err != nil {
		log.Println(err)
		return
	}

	// Log basic message metadata.
	log.Printf("To: %v", m.Header.Get("To"))
	log.Printf("From: %v", m.Header.Get("From"))
	log.Printf("Subject: %v", m.Header.Get("Subject"))

	// If the message has no particular content type, return.
	contentType := m.Header.Get("Content-Type")
	if contentType == "" {
		log.Println("message has no attachments")
		return
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.Println(err)
		return
	}

	// If the message is not multipart, return.
	if strings.Split(mediaType, "/")[0] != "multipart" {
		log.Println("message has no attachments")
		return
	}

	// Read each part (or attachment) as identified by its boundary.
	partReader := multipart.NewReader(m.Body, params["boundary"])
	for {
		part, err := partReader.NextPart()
		if err == io.EOF {
			return
		}

		// Retrieve the name of the attachment file.
		log.Printf("Filename: %v", part.FileName())

		// Read the data of the attachment.
		var partData []byte
		n, err := part.Read(partData)

		log.Printf("Size: %v", n)

		// Some MIME word decoding might be necessary here.

		// Compute the SHA-256 hash of the attachment.
		sha256Sum := sha256.Sum256(partData)
		sha256String := hex.EncodeToString(sha256Sum[:32])

		log.Printf("SHA-256: %v", sha256String)

		// Save file as its SHA-256 hash in the specified directory.
		err = ioutil.WriteFile(path.Join(attachmentsDir, sha256String), partData, 0600)
		if err != nil {
			log.Println("could not save attachment: %v", err)
			return
		}
	}
}
