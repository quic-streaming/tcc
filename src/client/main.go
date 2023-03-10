package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"main/src/model"
	"sync"
	"unsafe"

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

	stream, err := connection.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	streamId := stream.StreamID()

	fmt.Printf("Client %d: Sending '%+v'\n", streamId, req)
	reqBytes, err := json.Marshal(req)
	if err != nil {
		log.Fatal(err)
	}
	if _, err = stream.Write(reqBytes); err != nil {
		log.Fatal(err)
	}

	var res model.VideoPacketResponse
	resBytes := make([]byte, unsafe.Sizeof(res))

	if _, err = io.ReadFull(stream, resBytes); err != nil {
		log.Fatal(err)
	}
	if err = json.Unmarshal(resBytes, &res); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Client %d: Got '%+v'\n", streamId, res)

	stream.Close()
}

func consumeBuffer() {
}
