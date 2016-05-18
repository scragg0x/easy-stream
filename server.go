package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/googollee/go-socket.io"
	zmq "github.com/pebbe/zmq3"
)

const zmqPort = "5555"
const ioPort = "3000"

type ZmqPacket struct {
	Room string `json:"room"`
	Msg  map[string]interface{} `json:"msg"`
}

func main() {

	debug := false
	if os.Getenv("VAGRANT") == "1" || os.Getenv("DEBUG") == "1" {
		debug = true
	}

	if debug {
		log.Println("Debug enabled")
	}

	pull, err := zmq.NewSocket(zmq.PULL)
	if err != nil {
		log.Fatal(err)
	}
	defer pull.Close()

	io, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	io.On("connection", func(so socketio.Socket) {
		so.On("subscribe", func(msg string) {
			so.Join(msg)
		})
		so.On("disconnect", func() {})
	})

	log.Println("Binding zmq to port", zmqPort)
	pull.Bind("tcp://*:" + zmqPort)

	http.Handle("/socket.io/", io)
	go http.ListenAndServe(":"+ioPort, nil)

	for {
		msg, _ := pull.RecvMessage(0)

		for _, val := range msg {
			var packet ZmqPacket
			err := json.Unmarshal([]byte(val), &packet)
			if err != nil || packet.Msg == nil || packet.Room == "" {
				log.Println("Invalid Packet", packet)
				continue
			}

			if debug {
				log.Println(packet)
			}

			io.BroadcastTo(packet.Room, "message", packet.Msg, func(so socketio.Socket, data string) {
				// Do something with client ACK
			})
		}

	}
}
