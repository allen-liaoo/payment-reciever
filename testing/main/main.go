package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"

	"allen-liaoo/payment-reciever/testing"
)

func main() {
	shutdown := make(chan any, 1)
	shutdownRes := make(chan error, 1)
	rpcHost := "127.0.0.1"
	rpcPort := 8545
	rpcURL := "http://" + rpcHost + ":" + fmt.Sprintf("%d", rpcPort)
	err := testing.CreateChain(rpcHost, rpcPort, shutdown, shutdownRes) // wait for chain to be created
	if err != nil {
		log.Fatalf("Failed to create chain: %v", err)
	}
	log.Printf("Chain created at %s", rpcURL)

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	id, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get network ID: %v", err)
	}
	log.Print(id)

	shutdown <- nil // signal chain to shutdown
	<-shutdownRes   // wait for shutdown

}
