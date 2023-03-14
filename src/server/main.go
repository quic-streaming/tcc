package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"main/src/model"
	"math/big"
	"os"

	"github.com/lucas-clemente/quic-go"
)

const (
	PORT     = 8000
	BASE_URL = "localhost"
)

func main() {
	url := fmt.Sprintf("%s:%d", BASE_URL, PORT)

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

		go handleStream(connection)
		go handleStream(connection)
	}
}

func handleStream(connection quic.Connection) {
	// open stream
	stream, err := connection.AcceptStream(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	streamId := stream.StreamID()

	// read file
	basePath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	data, err := os.ReadFile(basePath + "/data/segments/video_tiled_7_dash_track126_3.m4s")
	if err != nil {
		log.Fatal(err)
	}
	_ = data

	// receive file request
	var req model.VideoPacketRequest
	if err = json.NewDecoder(stream).Decode(&req); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Server stream %d: Got '%+v'\n", streamId, req)

	// send file response
	res := model.VideoPacketResponse{
		Priority: req.Priority,
		Bitrate:  req.Bitrate,
		Segment:  req.Segment,
		Tile:     req.Tile,
		// Data:     data,
	}

	fmt.Printf("Server stream %d: Sending '%+v'\n", streamId, res)
	if json.NewEncoder(stream).Encode(&res); err != nil {
		log.Fatal(err)
	}

	// close stream
	stream.Close()
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}
}
