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
	"unsafe"

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
	stream, err := connection.AcceptStream(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	var req model.VideoPacketRequest
	reqBytes := make([]byte, unsafe.Sizeof(req))
	if _, err = stream.Read(reqBytes); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Server: Got '%+v'\n", req)

	res := model.VideoPacketResponse{
		Priority: req.Priority,
		Bitrate:  req.Bitrate,
		Segment:  req.Segment,
		Tile:     req.Tile,
		Data:     [1024]byte{},
	}

	fmt.Printf("Server: Sending '%+v'\n", res)
	resBytes, err := json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}
	if _, err = stream.Write(resBytes); err != nil {
		log.Fatal(err)
	}

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
