package testing

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
)

// this function creates a goroutine
// shutdown is a two-way channel
// send a nil value to shutdown the chain, and afterwards, receive nil if the chain shutdown successfully, or error(s) if the chain failed to shutdown
func CreateChain(rpcHost string, rpcPort int, genesis *core.Genesis, shutdown chan error) error {
	// init directory to store chain data
	datadir := "./privatechain"
	os.RemoveAll(datadir)
	err := os.MkdirAll(datadir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	// create an eth node with rpc server
	config := &node.Config{
		DataDir:  datadir,
		Name:     "privatechain",
		HTTPHost: rpcHost, // rpc host
		HTTPPort: rpcPort, // rpc port
		// HTTPModules: []string{"eth", "net", "web3"}, // enable modules?
	}

	stack, err := node.New(config)
	if err != nil {
		return fmt.Errorf("failed to create Ethereum node: %v", err)
	}

	ethConfig := &eth.Config{Genesis: genesis}
	_, err = eth.New(stack, ethConfig) // ethBackend
	if err != nil {
		return fmt.Errorf("failed to create Ethereum backend: %v", err)
	}

	// start the node and expose rpc
	err = stack.Start()
	if err != nil {
		return fmt.Errorf("failed to start Ethereum node: %v", err)
	}

	// creates a new goroutine inside this function (not relying on user of the function to create a goroutine)
	// because if the chain wasn't created successfully, we want to return the error to the user immediatel
	go waitForShutdown(shutdown, stack)
	return nil
}

func waitForShutdown(shutdown chan error, stack *node.Node) {
	<-shutdown // wait for shutdown signal

	err := stack.Close()
	if err != nil {
		shutdown <- fmt.Errorf("error during shutdown: %v", err)
		return
	}
	shutdown <- nil // signal successful shutdown
}
