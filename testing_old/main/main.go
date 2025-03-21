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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func privToAddress(pk *ecdsa.PrivateKey) common.Address {
	return crypto.PubkeyToAddress(*pk.Public().(*ecdsa.PublicKey))
}

func createContract(contractBytecode string, creatorPrivateKey *ecdsa.PrivateKey, client *ethclient.Client) error {
	creatorAddr := privToAddress(creatorPrivateKey)
	fmt.Printf("creator address: %v\n", creatorAddr.Hex())
	// get nonce
	nonce, err := client.PendingNonceAt(context.Background(), creatorAddr)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %v", err)
	}

	// gasPrice, err := client.SuggestGasPrice(context.Background())
	// if err != nil {
	// 	return fmt.Errorf("failed to get gas price: %v", err)
	// }
	// fmt.Printf("gas price: %v\n", gasPrice)

	data := common.FromHex(contractBytecode)
	// gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
	// 	From: creatorAddr,
	// 	To:   &creatorAddr, // can't be nil when estimating gas
	// 	Data: data,
	// })
	// if err != nil {
	// 	return fmt.Errorf("failed to estimate gas: %v", err)
	// }
	// gasLimit += 35000 // + 3200 for contract creation
	// fmt.Printf("gas limit: %v\n", gasLimit)

	// deploy contract
	tx := types.NewContractCreation(nonce, big.NewInt(0), 30000000, big.NewInt(2000000000), data)

	chainId, err := client.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %v", err)
	}

	// sign the transaction
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainId), creatorPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	// send transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return fmt.Errorf("failed to send contract deployment transaction: %v", err)
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

	cmd, err := StartChain()
	if err != nil {
		log.Printf("Failed to start chain: %v", err)
		return
	}
	defer func() {
		StopChain(cmd)
		log.Printf("Chain at %d stopped\n", cmd.Process.Pid)
	}()

	const rpcHost = "127.0.0.1"
	const rpcPort = 8545

	var usdcContractBytecode string
	b, err := os.ReadFile("../contracts/build/USDT.bin")
	if err != nil {
		log.Printf("Failed to read USDC contract bytecode: %v", err)
		return
	}
	usdcContractBytecode = string(b)

	conCreatorPKHex := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	conCreatorPK, err := crypto.HexToECDSA(conCreatorPKHex)
	if err != nil {
		log.Printf("failed to load contract creator private key: %v", err)
		return
	}
	rpcURL := "http://" + rpcHost + ":" + fmt.Sprintf("%d/", rpcPort)

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Printf("Failed to connect to the Ethereum client: %v", err)
		return
	}

	id, err := client.ChainID(context.Background())
	if err != nil {
		log.Printf("Failed to get chain ID: %v", err)
		return
	}
	log.Printf("Connected to chain with ID: %v", id)

	err = createContract(usdcContractBytecode, conCreatorPK, client)
	if err != nil {
		log.Printf("Failed to deploy contract: %v", err)
	} else {
		fmt.Printf("Contract deployed!")
	}

}
