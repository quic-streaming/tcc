package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"main/src/model"
	"sync"

	"github.com/lucas-clemente/quic-go"
)

const (
	PORT     = 8000
	BASE_URL = "localhost"
)

func main() {
	url := fmt.Sprintf("%s:%d", BASE_URL, PORT)

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}
	// Create new QUIC connection
	connection, err := quic.DialAddr(url, tlsConf, nil)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	// High priority stream
	wg.Add(1)
	go func() {
		handleStream(connection, model.HIGH_PRIORITY)
		wg.Done()
	}()

	// Low priority stream
	wg.Add(1)
	go func() {
		handleStream(connection, model.LOW_PRIORITY)
		wg.Done()
	}()

	// Consumer
	wg.Add(1)
	go func() {
		consumeBuffer()
		wg.Done()
	}()

	wg.Wait()
}

func handleStream(connection quic.Connection, priority model.Priority) {
	req := model.VideoPacketRequest{
		Priority: priority,
		Bitrate:  model.HIGH_BITRATE,
		Segment:  0,
		Tile:     0,
	}

	// create stream
	stream, err := connection.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	streamId := stream.StreamID()

	// send file request
	fmt.Printf("Client stream %d: Sending '%+v'\n", streamId, req)
	if json.NewEncoder(stream).Encode(&req); err != nil {
		log.Fatal(err)
	}

	// receive file response
	var res model.VideoPacketResponse
	if err = json.NewDecoder(stream).Decode(&res); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Client stream %d: Got '%+v'\n", streamId, res)

	// close stream
	stream.Close()
}

func consumeBuffer() {
}
