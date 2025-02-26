package main

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"

	wallet "allen-liaoo/payment-reciever/wallet"
)

func main() {

	// grab env varaibles
	godotenv.Load()
	var rpcUrl = os.Getenv("RPC_URL")
	var infuraKey = os.Getenv("INFURA_KEY")
	if rpcUrl == "" || infuraKey == "" {
		panic("RPC_URL or INFURA_KEY environment variable is not set")
	}
	var usdcAddr = os.Getenv("USDC_ADDRESS")
	var myWalletAddr = os.Getenv("MY_WALLET_ADDRESS")
	if usdcAddr == "" || myWalletAddr == "" {
		panic("USDC_ADDRESS or MY_WALLET_ADDRESS environment variable is not set")
	}
	var usdcAddress = common.HexToAddress(usdcAddr)
	var myWalletAddress = common.HexToAddress(myWalletAddr)

	// init client
	var client *ethclient.Client
	client, err := ethclient.Dial(rpcUrl + infuraKey)
	if err != nil {
		panic(err)
	}

	// get eth balance
	balance, err := client.BalanceAt(context.Background(), myWalletAddress, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("balance (eth): ", balance, " (", wallet.FromRawUnit(balance, 18), ")")

	// Get USDC token balance
	decimals, err := wallet.GetContractDecimals(client, usdcAddress) // get contract decimals
	if err != nil {
		panic(err)
	}
	tokenBalance, err := wallet.GetTokenBalance(client, usdcAddress, myWalletAddress)
	if err != nil {
		panic(err)
	}
	fmt.Println("balance (asdc):", tokenBalance, " (", wallet.FromRawUnit(tokenBalance, decimals), ")")

	// General gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("gas (general): ", gasPrice, " (", wallet.FromRawUnit(gasPrice, 18), ")")

	// Estimate gas for token transfer (From my wallet address to desAddress)
	desAddress := common.HexToAddress("0x4d79b07f793FB42e7870c568cB374DBDc3BeBA51") // random address I grabbed from etherscan
	preAmount := big.NewInt(1)
	amount := wallet.ToRawUnit(preAmount, decimals)                           // amount with decimals
	data := wallet.BuildTokenTxDataField(myWalletAddress, desAddress, amount) // data field for contract tokens transfer
	msg := ethereum.CallMsg{                                                  // test transaction
		From:  myWalletAddress,
		To:    &desAddress,
		Value: big.NewInt(0),
		Data:  data,
	}
	gas, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		panic(err)
	}
	fmt.Println("gas to", desAddress.Hex(), ":", gas)

}
