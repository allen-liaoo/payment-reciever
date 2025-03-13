package testing

import (
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
)

// this function creates a goroutine
// shutdown is a channel to signal the chain to shutdown
// shutdownResult is the channel to signal that the chain has been shutdown (or errored)
func CreateChain(rpcHost string, rpcPort int, shutdown <-chan any, shutdownResult chan<- error) error {
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

	chainConfig := params.TestChainConfig
	chainConfig.ChainID = big.NewInt(1234)
	genesis := &core.Genesis{
		Config:     chainConfig,
		Difficulty: big.NewInt(0x1),
		GasLimit:   0x8000000,
		Alloc:      make(types.GenesisAlloc),
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
	go waitForShutdown(shutdown, shutdownResult, stack)
	return nil
}

func waitForShutdown(shutdown <-chan any, shutdownResult chan<- error, stack *node.Node) {
	<-shutdown // wait for shutdown signal

	err := stack.Close()
	if err != nil {
		shutdownResult <- fmt.Errorf("error during shutdown: %v", err)
		return
	}
	close(shutdownResult) // signal successful shutdown
}
