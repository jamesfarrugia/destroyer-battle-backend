package main

import (
	"encoding/binary"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var clients map[int]*Client
var sessions map[string]*Session

// serveWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}

	doProcessConnection(conn)
}

func doProcessConnection(conn *websocket.Conn) {
	/* when client connects, they can either tell us they're new or they are reconnecting:
	 * - new clients send us a 0x0 byte, and we'll respond with an ID;
	 * - reconnecting clients send us 0x1 followed by their id;
	 *
	 * when a client disconnects without informing us, one minute is allowed for reconnection,
	 * otherwise it is removed.  If a client connects with an ID that we no longer know,
	 * we inform it and send a new ID */

	msgType, msgData, err := conn.ReadMessage()
	if err != nil {
		log.Error(err)
	}

	if msgType == websocket.TextMessage {
		err = conn.WriteMessage(websocket.TextMessage, []byte("Invalid header"))
		if err != nil {
			log.Error("Failed to respond with invalid header message")
		}
		err = conn.Close()
		if err != nil {
			log.Error("Failed to close connection")
		}
		return
	}

	if len(msgData) == 0 {
		if err != nil {
			log.Error("Empty header")
		}
		return
	}

	if msgData[0] == 0 {
		doProcessNewClient(conn)
	} else if msgData[0] == 1 {
		doProcessReconnectingClient(conn)
	}
}

func doProcessNewClient(conn *websocket.Conn) {
	client := &Client{id: time.Now().Nanosecond(), conn: conn, send: make(chan []byte, 256)}
	clients[client.id] = client

	log.Info("New client ", client.id)

	bleID := make([]byte, 4)
	binary.LittleEndian.PutUint32(bleID, uint32(client.id))
	log.Debug(bleID)
	err := conn.WriteMessage(websocket.BinaryMessage, bleID)
	if err != nil {
		log.Error("Failed to write client ID")
	}

	client.handle()
}

func doProcessReconnectingClient(conn *websocket.Conn) {
	var rq = []byte{1}

	log.Info("Client claims to be reconnecting, sending ack")
	err := conn.WriteMessage(websocket.BinaryMessage, rq)
	if err != nil {
		log.Error("Failed to write ID expectation ACK")
	}

	_, idData, err := conn.ReadMessage()
	if err != nil {
		log.Error(err)
	}

	id := binary.BigEndian.Uint32(idData)
	log.Info("Claiming ID", id)

	client := clients[int(id)]

	// send 0x1 if we did not find it, and client should reconnect or disconnect
	if client == nil {
		log.Info("Invalid, responding with error")
		err := conn.WriteMessage(websocket.BinaryMessage, rq)
		if err != nil {
			log.Error("Failed to write client not found error")
		}
	} else {
		log.Info("Client reconnected")
		rq[0] = 0
		err := conn.WriteMessage(websocket.BinaryMessage, rq)
		if err != nil {
			log.Error("Failed to write client reconnect ACK")
		}

		client.handle()
	}
}
