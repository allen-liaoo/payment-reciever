package test

import (
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
)

const rpcURL = "https://127.0.0.1:8554/"

var client *ethclient.Client

func init() {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		panic(err)
	}

	// allocate tokens to middlewares

}

func TestHandleMiddleware(t *testing.T) {

}
