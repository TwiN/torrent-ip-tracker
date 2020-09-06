package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

const (
	ProtocolID     = 0x41727101980
	MaxRequestSize = 100
)

type Action int32

const (
	Connect Action = iota
	Announce
	Scrape
	Error
)

type RequestHeader struct {
	ConnectionID  int64
	Action        int32
	TransactionID int32
}

type ErrorResponse struct {
	Action        int32
	TransactionID int32
	Message       [20]byte
}

type ConnectResponse struct {
	Action        int32
	TransactionID int32
	ConnectionID  int64
}

type ResponseWriter struct {
	Conn *net.UDPConn
	Addr *net.UDPAddr
}

func (w ResponseWriter) Write(p []byte) (int, error) {
	w.Conn.WriteToUDP(p, w.Addr)
	return len(p), nil
}

func main() {
	addr, err := net.ResolveUDPAddr("udp4", ":12345")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listening to UDP on", addr.String())

	for {
		message := make([]byte, MaxRequestSize)
		_, clientAddr, err := conn.ReadFromUDP(message)
		if err != nil {
			log.Println("Failed to read from UDP:", err.Error())
			continue
		}
		writer := ResponseWriter{Conn: conn, Addr: clientAddr}

		var header RequestHeader
		reader := bytes.NewReader(message)
		if err := binary.Read(reader, binary.BigEndian, &header); err != nil {
			log.Println("Unable to read message:", err.Error())
		}

		switch Action(header.Action) {
		case Connect:
			if header.ConnectionID != ProtocolID {
				log.Println("Invalid protocol ID:", header.ConnectionID)
				continue
			}
			response := ConnectResponse{
				Action:        header.Action,
				TransactionID: header.TransactionID,
				ConnectionID:  header.ConnectionID,
			}
			if err := binary.Write(writer, binary.BigEndian, response); err != nil {
				log.Println("Unable to handle announce request:", err.Error())
			}
		case Announce:
			response := ErrorResponse{
				Action:        int32(Error),
				TransactionID: header.TransactionID,
				Message:       [20]byte{},
			}
			copy(response.Message[:], fmt.Sprintf("IP: %s", clientAddr.IP.To4().String()))
			if err := binary.Write(writer, binary.BigEndian, response); err != nil {
				log.Println("Unable to handle announce request:", err.Error())
			}
		}
	}
}
