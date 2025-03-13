package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"

	"allen-liaoo/payment-reciever/testing"
)

func privToAddress(pk *ecdsa.PrivateKey) common.Address {
	return crypto.PubkeyToAddress(*pk.Public().(*ecdsa.PublicKey))
}

func createContract(contractBytecode string, creatorPrivateKey *ecdsa.PrivateKey, client *ethclient.Client) error {
	// get nonce
	nonce, err := client.PendingNonceAt(context.Background(), privToAddress(creatorPrivateKey))
	if err != nil {
		return fmt.Errorf("failed to get nonce: %v", err)
	}

	// get gas price
	// gasPrice, err := client.SuggestGasPrice(context.Background())
	// if err != nil {
	// 	return fmt.Errorf("failed to get gas price: %v", err)
	// }

	// deploy contract
	input := common.FromHex(contractBytecode)
	tx := types.NewContractCreation(nonce, big.NewInt(0), uint64(95420), big.NewInt(0), input)

	// sign the transaction
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(big.NewInt(1234)), creatorPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	// send transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return fmt.Errorf("failed to deploy contract: %v", err)
	}

	// wait for transaction to be mined
	fmt.Printf("contract pending: 0x%x\n", signedTx.Hash())
	recp, err := bind.WaitMined(context.Background(), client, signedTx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction to be mined: %v", err)
	} else if recp.Status != 1 {
		return fmt.Errorf("contract creation transaction failed: %v", recp)
	}

	return nil
}

func main() {
	const rpcHost = "127.0.0.1"
	const rpcPort = 8545

	var usdcContractBytecode string
	b, err := os.ReadFile("../USDC.bin")
	if err != nil {
		log.Fatalf("Failed to read USDC contract bytecode: %v", err)
	}
	usdcContractBytecode = string(b)

	shutdown := make(chan error, 1)

	chainConfig := params.TestChainConfig
	chainConfig.ChainID = big.NewInt(1234)

	conCreatorPKHex := "f37f8d03685b160b6b8d57ca59d13a08c83d642077ed88b30963b9ae931a6e0d"
	conCreatorPK, err := crypto.HexToECDSA(conCreatorPKHex)
	if err != nil {
		log.Fatalf("failed to load contract creator private key: %v", err)
	}

	senderPKBex := "bd07fa0f3ae335fa5dfc5ccf7e1bd2432ba8884caf41d2a6d6ddc6e9b715f369"
	senderPK, err := crypto.HexToECDSA(senderPKBex)
	if err != nil {
		log.Fatalf("failed to load sender private key: %v", err)
	}

	alloc := make(types.GenesisAlloc)
	// contract creation address
	alloc[privToAddress(conCreatorPK)] = types.Account{
		Balance:    big.NewInt(1000000000000000000),
		PrivateKey: []byte(conCreatorPKHex),
	}
	// sender address
	alloc[privToAddress(senderPK)] = types.Account{
		Balance:    big.NewInt(1000000000000000000),
		PrivateKey: []byte(senderPKBex),
	}

	genesis := &core.Genesis{
		Config:     chainConfig,
		Difficulty: big.NewInt(0x1),
		GasLimit:   0x8000000,
		Alloc:      make(types.GenesisAlloc),
	}

	err = testing.CreateChain(rpcHost, rpcPort, genesis, shutdown) // wait for chain to be created
	if err != nil {
		log.Fatalf("Failed to create chain: %v", err)
	}
	rpcURL := "http://" + rpcHost + ":" + fmt.Sprintf("%d", rpcPort)
	log.Printf("Chain created at %s", rpcURL)

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	err = createContract(usdcContractBytecode, conCreatorPK, client)
	if err != nil {
		log.Fatalf("Failed to deploy contract: %v", err)
	} else {
		fmt.Printf("Contract deployed!")
	}

	shutdown <- nil   // signal chain to shutdown
	res := <-shutdown // wait for shutdown
	if res != nil {
		log.Fatalf("%v", res)
	} else {
		log.Printf("Chain shutdown successfully")
	}
}
