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
// shutdown is a two-way channel; the buffer size should be 1
// send a nil value to shutdown the chain, and afterwards, receive nil if the chain shutdown successfully, or an error if the chain failed to shutdown
func CreateChain(rpcHost string, rpcPort int, shutdown chan error) error {
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
	close(shutdown) // signal successful shutdown
}
