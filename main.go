package main

import (
	"allen-liaoo/payment-reciever/reciever"
	"allen-liaoo/payment-reciever/util"
	"fmt"
	"math/big"
	"os"

	"github.com/joho/godotenv"
)

var middlewareWalletMneumonic string

func init() {
	godotenv.Load()
	middlewareWalletMneumonic = os.Getenv("MIDDLEWARE_MNEUMONIC")
	if middlewareWalletMneumonic == "" {
		panic("USDC_ADDRESS or PROVIDER_WALLET_ADDRESS or PROVIDER_WALLET_PK or MIDDLEWARE_MNEUMONIC environment variable is not set")
	}
}

func main() {
	// Local testing: Derive necessary middleware wallet
	var testDerivationPath = "m/44'/60'/0'/0/42"
	middlewareWallet, privateKey, err := util.DeriveWallet(middlewareWalletMneumonic, testDerivationPath)
	if err != nil {
		panic(err)
	}
	fmt.Println("middleware wallet address: ", middlewareWallet.Address.Hex())

	result, error := reciever.SweepMiddleware(middlewareWallet, privateKey, big.NewInt(0), big.NewInt(0))
	fmt.Printf("%#v\n", result)
	if error != nil {
		panic(error)
	}
}
