package main

import (
	"flag"
	"log"
	"time"

	"net"

	"encoding/base64"
	"fmt"

	"github.com/gorilla/websocket"
)

const (
	pongTimeout = 60 * time.Second
)

var addr = flag.String("addr", "", "http service address")

func main() {
	flag.Parse()

	if *addr == "" {
		flag.Usage()
		log.Fatalf("Server address \"%s\" is invalid.", *addr)
	}

	var dialer *websocket.Dialer

	conn, _, err := dialer.Dial(*addr, nil)
	if err != nil {
		log.Fatalf("Unable to dial ws connection to agent %s with error %s\n", *addr, err)
	}

	log.Println("Connection stablized")
	defer conn.Close()

	conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("Received close message. Code: %d, Text: %s\n", code, text)
		message := []byte{}
		if code != websocket.CloseNoStatusReceived {
			message = websocket.FormatCloseMessage(code, "")
		}
		conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(pongTimeout))
		return nil
	})

	conn.SetPingHandler(
		func(appData string) error {
			log.Printf("Received ping message: '%s'\n", appData)
			err := conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(pongTimeout))
			if e, ok := err.(net.Error); ok && e.Temporary() {
				log.Printf("Received temporary error while sending pong. %v\n", e)
				return nil
			}
			if err == nil {
				log.Println("Sent pong")
			} else {
				log.Println("Failed to send pong")
			}
			return err
		},
	)

	for {
		msgType, rawMsg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error in reading message:", err)
			break
		}

		switch msgType {
		case websocket.TextMessage:
			log.Printf("Received text message: %s\n", rawMsg)
		case websocket.BinaryMessage:
			log.Printf("Received binary message. Base64: %s\n", base64.StdEncoding.EncodeToString(rawMsg))
		default:
			log.Printf("Received unknown message type: %d.\n Base64 of message: %s\n", msgType, base64.StdEncoding.EncodeToString(rawMsg))
		}

		err = conn.WriteMessage(msgType, rawMsg)
		if err == nil {
			log.Printf("Sent received message back to server")
		} else {
			log.Printf("Unable to send message to server. Error: %s\n", err)
		}
	}

	log.Println("Exiting...")
}
