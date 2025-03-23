package harness

import (
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
)

const rpcURL = "https://127.0.0.1:8545/"

var client *ethclient.Client

func init() {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		panic(err)
	}

	// allocate tokens to middlewares

	// Process the middleware private keys
	for _, pk := range chainInfo.MiddlewarePKs {
		// Use the private keys as needed
		// Example: allocate tokens to each middleware
		// You'll need to implement the actual token allocation logic
	}
}

func TestHandleMiddleware(t *testing.T) {

}
