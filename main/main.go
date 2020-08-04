package main

import (
	"crypto/ecdsa"
	"net"

	"google.golang.org/grpc"
	"istio.io/istio/authzadaptor"
)

func main() {
	listener, err := net.Listen("tcp", ":9070")
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	authzadaptor.RegisterHandleAuthzadaptorServiceServer(server, authzadaptor.AuthZAdaptor{URLToPublicKeyDict: make(map[string]*ecdsa.PublicKey)})
	server.Serve(listener)
}
