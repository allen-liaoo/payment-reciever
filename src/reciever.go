package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	var usdcAddr = os.Getenv("USDC_ADDRESS")
	var myWalletAddr = os.Getenv("MY_WALLET_ADDRESS")
	if usdcAddr == "" || myWalletAddr == "" {
		panic("USDC_ADDRESS or MY_WALLET_ADDRESS environment variable is not set")
	}
	fee, err := getGasPrice()
	if err != nil {
		panic(err)
	}
	fmt.Println(fee)

	balance, err := getWalletBalance(myWalletAddr)
	if err != nil {
		panic(err)
	}
	fmt.Println(balance)

	tokenBalance, err := getTokenBalance(usdcAddr, myWalletAddr)
	if err != nil {
		panic(err)
	}
	fmt.Println(tokenBalance)
}
