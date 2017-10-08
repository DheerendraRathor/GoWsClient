SimpleWsClient
==============

A simple websocket client in Golang to debug websocket server. 

### Features:
- Awesome logging. Every action taken and every event registered is logged.
- Takes care of sending pong.
- Send received message back to server
- In case of binary message, it logs base64 encoded message.

### Installation:
- Download latest exe/binary for required platform from Github releases
- **OR** if go is installed then use `go get -u github.com/DheerendraRathor/SimpleWsClient`

### Usages:
```bash
$ ./SimpleWsClient -help
Usage of ./SimpleWsClient:
  -addr string
        http service address (default "ws://echo.websocket.org")
  -echoDelay uint
        Delay before echoing back received message from server
  -protocols string
        Comma separated list of protocols to use

$ ./SimpleWsClient -addr wss://echo.websocket.org -protocols echo,chat
```
