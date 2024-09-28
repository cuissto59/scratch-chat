package main

import (
	"log"
	"net"
)

type MessageType int

const (
	ClientConnected MessageType = iota + 1
	NewMessage
	ClientDisconnected
)

type Message struct {
	Type MessageType
	Author net.Conn
	Text string
}

const Port = "6969"
const safeMode = false

func safeRemoteAddr(conn net.Conn) string {
	if safeMode {
		return "[REDACTED]"
	} else {
		return conn.RemoteAddr().String()
	}
}

func client(conn net.Conn, messages chan *Message) {
	buffer := make([]byte, 512)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
      log.Printf("Could not read from %s: %s\n", safeRemoteAddr(conn),err)
			conn.Close()
			messages <- &Message{
				Type: ClientDisconnected,
				Author: conn,
			}
			return
		}
		messages <- &Message{
			Type: NewMessage,
			Author: conn,
			Text: string(buffer[0:n]),
		}
	}
}

func server(messages chan *Message) {
	conns := map[string]net.Conn{}
	for {
		msg := <-messages
		switch msg.Type {
		case ClientConnected:
			log.Printf("Client %s connected\n", safeRemoteAddr(msg.Author))
			conns[msg.Author.RemoteAddr().String()] = msg.Author
		case ClientDisconnected:
			log.Printf("Client %s disconnected\n", safeRemoteAddr(msg.Author))
			delete(conns, msg.Author.RemoteAddr().String())
		case NewMessage:
			log.Printf("Client %s sent message %s\n", safeRemoteAddr(msg.Author), msg.Text)
			for _, conn := range conns {
				if conn.RemoteAddr().String() != msg.Author.RemoteAddr().String() {
					_, err := conn.Write([]byte(msg.Text))
					if err != nil {
						log.Printf("Could not send data to %s : %s\n", safeRemoteAddr(conn), err)
					}

				}
			}
		}

	}

}

func main() {

	ln, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		log.Fatalf("ERROR: could not listen to epic port %s: %v\n", Port, err)

	}
	log.Printf("Listning to TCP connections on port %s ...\n", Port)

	messages := make(chan *Message)
	go server(messages)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("ERROR: could not accept a coonection: %s\n", err)
		}
		log.Printf("Accepted connection from %s\n", conn.RemoteAddr())

		messages <- &Message{
			Type: ClientConnected,
			Author: conn,
		}

		go client(conn, messages)
	}
}
