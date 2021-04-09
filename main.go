package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type message struct {
	msg    string
	sender string
}

var whoisListenAddr = flag.String("whois.listen", ":8843", "")
var logFile = flag.String("log.file", "./chat.log", "")

var messages map[string][]message

func main() {
	flag.Parse()

	messages = make(map[string][]message)

	l, err := net.Listen("tcp", *whoisListenAddr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Print(err)
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	in, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Print(err)
		return
	}

	senderHost, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		log.Print(err)
		fmt.Fprintf(conn, "error :(\r\n")
		return
	}
	senderID := senderHost

	fmt.Fprintf(conn, "hai, %s!\n", senderID)

	senderMessages, ok := messages[senderID]
	if !ok || len(senderMessages) == 0 {
		fmt.Fprintf(conn, "no messages :(\r\n")
	}

	senderMsgCount := len(senderMessages)
	for i, m := range senderMessages {
		fmt.Fprintf(conn, "%d/%d: msg from %s: %s\r\n", i+1, senderMsgCount, m.sender, m.msg)
	}

	in = strings.ReplaceAll(in, "\r\n", "")

	parts := strings.SplitN(in, " ", 2)
	if len(parts) != 2 {
		fmt.Fprintf(conn, "oki bai!\r\n")
		return
	}

	rcpt := strings.ToLower(parts[0])
	msg := parts[1]

	msg = strings.Map(func(r rune) rune {
		if r > 126 || r < 32 {
			return -1
		}
		return r
	}, msg)

	rcptIP := net.ParseIP(rcpt)
	if rcptIP == nil {
		fmt.Fprintf(conn, "invalid rcpt :(\r\n")
		return
	}

	logMessage(conn.RemoteAddr().String(), rcpt, msg)

	if _, ok := messages[rcpt]; !ok {
		messages[rcpt] = make([]message, 0)
	}
	messages[rcpt] = append(messages[rcpt], message{msg: msg, sender: senderID})
	fmt.Fprintf(conn, "oki! :3\r\n")

	fmt.Fprintf(conn, "oki bai!\r\n")
}

func logMessage(sender string, rcpt string, message string) {
	log.Printf("msg from: %s, to: %s, with content: %s", sender, rcpt, message)
	f, err := os.OpenFile(*logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		log.Print(err)
	}
	defer f.Close()
	fmt.Fprintf(f, "msg from: %s, to: %s, with content: %s\n", sender, rcpt, message)
}
