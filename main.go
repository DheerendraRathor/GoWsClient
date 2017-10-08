package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pongTimeout = 60 * time.Second
)

var addrFlag = flag.String("addr", "ws://echo.websocket.org", "http service address")
var subProtocolFlag = flag.String("protocols", "", "Comma separated list of protocols to use")
var echoDelayFlag = flag.Uint("echoDelay", 0, "Delay before echoing back received message from server")

func main() {
	flag.Parse()

	var address string = *addrFlag
	var subProtocols []string
	var echoDelay = *echoDelayFlag

	if address == "" {
		log.Fatalf("Server address \"%s\" is invalid.", address)
	}

	if *subProtocolFlag != "" {
		subProtocols = strings.Split(*subProtocolFlag, ",")
		log.Printf("Using sub protocols: %s\n", strings.Join(subProtocols, ", "))
	}

	if echoDelay < 0 {
		log.Fatalln("Echo delay cannot be negative")
	}

	var dialer *websocket.Dialer = &websocket.Dialer{
		Subprotocols: subProtocols,
	}

	conn, _, err := dialer.Dial(address, nil)
	if err != nil {
		log.Fatalf("Unable to dial ws connection to agent %s with error %s\n", address, err)
	}

	log.Println("Connection stabilized")
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

	var wg sync.WaitGroup

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

		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(time.Second * time.Duration(echoDelay)))

			err = conn.WriteMessage(msgType, rawMsg)
			if err == nil {
				log.Printf("Sent received message back to server")
			} else {
				log.Printf("Unable to send message to server. Error: %s\n", err)
			}
		}()
	}

	wg.Wait()

	log.Println("Exiting...")
}
