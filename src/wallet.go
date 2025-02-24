package main

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"

	contracts "allen-liaoo/payment-reciever/contracts"
)

var client *ethclient.Client

func init() {
	godotenv.Load()
	var rpcUrl = os.Getenv("RPC_URL")
	var infuraKey = os.Getenv("INFURA_KEY")
	if rpcUrl == "" || infuraKey == "" {
		panic("RPC_URL or INFURA_KEY environment variable is not set")
	}

	cclient, err := ethclient.Dial(rpcUrl + infuraKey)
	if err != nil {
		panic(err)
	}
	client = cclient
}

func getMaxPriorityFee() (*big.Int, error) {
	return client.SuggestGasTipCap(context.Background())
}

func getWalletBalance(walletAddress string) (*big.Int, error) {
	return client.BalanceAt(context.Background(), common.HexToAddress(walletAddress), nil)
}

// return balance, decimals, error
func getTokenBalance(contractAddress string, walletAddress string) (*big.Int, error) {
	contract, err := contracts.NewErc20(common.HexToAddress(contractAddress), client)
	if err != nil {
		return big.NewInt(0), err
	}

	balance, err := contract.BalanceOf(nil, common.HexToAddress(walletAddress))
	if err != nil {
		return big.NewInt(0), err
	}

	decimals, err := contract.Decimals(nil)
	if err != nil {
		return balance, err
	}

	return calculcateBalance(balance, decimals), nil
}

func calculcateBalance(balance *big.Int, decimals uint8) *big.Int {
	base := big.NewInt(10)
	balance.Div(balance, base.Exp(base, big.NewInt(int64(decimals)), nil))
	return balance
}

// returns the hash of the transaction (in hex)
func sendTransaction(from string, to string, amount *big.Int, gasLimit uint64, gasPrice *big.Int, privateKey *ecdsa.PrivateKey) (string, error) {
	nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(from))
	if err != nil {
		return "", err
	}

	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(to),
		amount,
		gasLimit,
		gasPrice,
		[]byte{},
	)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return "", err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	return signedTx.Hash().Hex(), nil
}
