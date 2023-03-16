package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"main/src/model"
	"os"

	"github.com/lucas-clemente/quic-go"
)

type Server struct {
	serverURL   string
	serverPort  int
	queuePolicy QueuePolicy
}

func NewServer(serverURL string, serverPort int, queuePolicy string) *Server {
	return &Server{
		serverURL:   serverURL,
		serverPort:  serverPort,
		queuePolicy: QueuePolicy(queuePolicy),
	}
}

func (s *Server) Start() {
	url := fmt.Sprintf("%s:%d", s.serverURL, s.serverPort)

	listener, err := quic.ListenAddr(url, generateTLSConfig(), nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server listening on", url)

	for {
		connection, err := listener.Accept(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		go s.handleStream(connection)
		go s.handleStream(connection)
	}
}

func (s *Server) handleStream(connection quic.Connection) {
	// open stream
	stream, err := connection.AcceptStream(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// receive file request
	req := s.receiveData(stream)

	// read file
	data := s.readFile(req.Bitrate, req.Segment, req.Tile)

	// send file response
	s.sendData(stream, req.Priority, req.Bitrate, req.Segment, req.Tile, data)

	// close stream
	stream.Close()
}

// Read file
func (s *Server) readFile(bitrate model.Bitrate, segment int, tile int) []byte {
	basePath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// TODO check the file name logic
	data, err := os.ReadFile(basePath + fmt.Sprintf("/data/segments/video_tiled_%d_dash_track%d_%d.m4s", bitrate, segment, tile))
	if err != nil {
		log.Fatal(err)
	}
	return data
}

// Receive file request
func (s *Server) receiveData(stream quic.Stream) (req model.VideoPacketRequest) {
	if err := json.NewDecoder(stream).Decode(&req); err != nil {
		log.Fatal(err)
	}
	// streamId := stream.StreamID()
	// fmt.Printf("Server stream %d: Got '%+v'\n", streamId, req)
	return
}

// Send file response
func (s *Server) sendData(stream quic.Stream, priority model.Priority, bitrate model.Bitrate, segment int, tile int, data []byte) {
	// streamId := stream.StreamID()
	// fmt.Printf("Server stream %d: Sending '%+v'\n", streamId, res)
	res := model.VideoPacketResponse{
		Priority: priority,
		Bitrate:  bitrate,
		Segment:  segment,
		Tile:     tile,
		Data:     data,
	}
	if err := json.NewEncoder(stream).Encode(&res); err != nil {
		log.Fatal(err)
	}
}
