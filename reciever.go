package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"

	wallet "allen-liaoo/payment-reciever/wallet"
)

var client *ethclient.Client

var usdcAddress common.Address
var providerWalletAddress common.Address
var providerWalletPK *ecdsa.PrivateKey
var middlewareWalletMneumonic string
var desAddress common.Address

func init() {
	godotenv.Load()
	var rpcUrl = os.Getenv("RPC_URL")
	var infuraKey = os.Getenv("INFURA_KEY")
	if rpcUrl == "" || infuraKey == "" {
		panic("RPC_URL or INFURA_KEY environment variable is not set")
	}

	// init client
	var err error
	client, err = ethclient.Dial(rpcUrl + infuraKey)
	if err != nil {
		panic(err)
	}

	var usdcAddr = os.Getenv("USDC_ADDRESS")
	var providerWalletAddr = os.Getenv("PROVIDER_WALLET_ADDRESS")
	var providerWalletPrivateKey = os.Getenv("PROVIDER_WALLET_PK")
	middlewareWalletMneumonic = os.Getenv("MIDDLEWARE_MNEUMONIC")
	if usdcAddr == "" || providerWalletAddr == "" || providerWalletPrivateKey == "" || middlewareWalletMneumonic == "" {
		panic("USDC_ADDRESS or PROVIDER_WALLET_ADDRESS or PROVIDER_WALLET_PK or MIDDLEWARE_MNEUMONIC environment variable is not set")
	}
	usdcAddress = common.HexToAddress(usdcAddr)
	providerWalletAddress = common.HexToAddress(providerWalletAddr)
	providerWalletPK, err = crypto.HexToECDSA(providerWalletPrivateKey)
	if err != nil {
		panic(err)
	}
}

func main() {
	// TODO: Derive necessary middleware wallets
	var testDerivationPath = "m/44'/60'/0'/0/5"
	middlewareWallet, privateKey, err := wallet.DeriveWallet(middlewareWalletMneumonic, testDerivationPath)
	if err != nil {
		panic(err)
	}
	fmt.Println("middleware wallet address: ", middlewareWallet.Address.Hex())

	balance, err := wallet.GetTokenBalance(client, usdcAddress, middlewareWallet.Address)
	if err != nil {
		panic(err)
	}
	decimals, err := wallet.GetContractDecimals(client, usdcAddress) // get contract decimals
	if err != nil {
		panic(err)
	}
	fmt.Println("middleware balance (usdc): ", balance, " (", wallet.FromRawUnit(balance, decimals), ")")

	// TODO: Check if middleware wallet has expected balance to sweep
	if balance.Cmp(big.NewInt(0)) < 0 {
		panic("middleware wallet has not enough balance to sweep")
	}

	// Estimate cost to sweep, including
	// fee (base + priority) of middleware wallet to destination wallet
	// and fee of provider wallet to middleware wallet
	// gas1, data1, err := wallet.EstimateGas(client, middlewareWallet.Address, desAddress, balance, true)
	// if err != nil {
	// 	panic(err)
	// }
	// gas1BI := new(big.Int).SetUint64(gas1)
	// fmt.Println("gas from middleware to destination: ", gas1)
	// gas2, data2, err := wallet.EstimateGas(client, providerWalletAddress, middlewareWallet.Address, gas1, false)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("gas from provider to middleware: ", gas2)

	gastipcap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("gas tip cap: ", gastipcap)

	// TODO: Check if gas cost is affordable
	threshold := big.NewInt(0)
	if threshold.Cmp(gastipcap) > 0 {
		panic("gas tip cap is too high")
	}

	// sweep transaction
	// 1. Transfer ETH gas fee from provider wallet to middleware wallet
	tx1, err := wallet.SendTx(client, providerWalletAddress, middlewareWallet.Address, gastipcap, nil, gastipcap, nil, providerWalletPK)
	if err != nil {
		panic(err)
	}
	fmt.Println("tx1: ", tx1)
	fmt.Println("tx1 hash: ", tx1.Hash().Hex())
	fmt.Println("waiting for tx1 to be mined...")
	// _, err = bind.WaitMined(context.Background(), client, tx1)
	// if err != nil {
	// 	panic(err)
	// }

	// 2. Transfer USDC from middleware wallet to destination wallet
	tx2, err := wallet.SendTokenTx(client, middlewareWallet.Address, desAddress, usdcAddress, balance, nil, gastipcap, privateKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("tx2: ", tx2)
	fmt.Println("tx2 hash: ", tx2.Hash().Hex())
}

// func test() {

// 	// get eth balance
// 	balance, err := client.BalanceAt(context.Background(), providerWalletAddress, nil)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("balance (eth): ", balance, " (", wallet.FromRawUnit(balance, 18), ")")

// 	// Get USDC token balance
// 	decimals, err := wallet.GetContractDecimals(client, usdcAddress) // get contract decimals
// 	if err != nil {
// 		panic(err)
// 	}
// 	tokenBalance, err := wallet.GetTokenBalance(client, usdcAddress, providerWalletAddress)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("balance (asdc):", tokenBalance, " (", wallet.FromRawUnit(tokenBalance, decimals), ")")

// 	// General gas price
// 	gasPrice, err := client.SuggestGasPrice(context.Background())
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("gas (general): ", gasPrice, " (", wallet.FromRawUnit(gasPrice, 18), ")")

// 	// Estimate gas for token transfer (From my wallet address to desAddress)
// 	desAddress := common.HexToAddress("0x4d79b07f793FB42e7870c568cB374DBDc3BeBA51") // random address I grabbed from etherscan
// 	preAmount := big.NewInt(1)
// 	amount := wallet.ToRawUnit(preAmount, decimals)                                 // amount with decimals
// 	data := wallet.BuildTokenTxDataField(providerWalletAddress, desAddress, amount) // data field for contract tokens transfer
// 	msg := ethereum.CallMsg{                                                        // test transaction
// 		From:  providerWalletAddress,
// 		To:    &desAddress,
// 		Value: big.NewInt(0),
// 		Data:  data,
// 	}
// 	gas, err := client.EstimateGas(context.Background(), msg)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("gas to", desAddress.Hex(), ":", gas)

// }
