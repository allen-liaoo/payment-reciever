package reciever

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"

	"allen-liaoo/payment-reciever/util"
)

var client *ethclient.Client

var usdcAddress common.Address
var providerWalletAddress common.Address
var providerWalletPK *ecdsa.PrivateKey
var middlewareWalletMneumonic string
var desAddress common.Address

func init() {
	godotenv.Load("../.env")
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
	desAddress = providerWalletAddress
}

type PaymentResult struct {
	ProviderToMiddlewareReceipt *types.Receipt
	MiddlewareToDestinationTx   *types.Transaction
	BaseFee                     *big.Int
	GasTipCap                   *big.Int
	GasFeeCap                   *big.Int
	GasUnit                     uint64
}

// Check if a middleware wallet has enough balance to sweep, then sweep and return the transaction receipt
// from providerWallet to middleware, and the hex of the transaction from middleware to destination wallet
func SweepMiddleware(middlewareWallet *accounts.Account, privateKey *ecdsa.PrivateKey, minBalance *big.Int, gasCostThreshold *big.Int) (*PaymentResult, error) {

	result := &PaymentResult{
		ProviderToMiddlewareReceipt: nil,
		MiddlewareToDestinationTx:   nil,
		BaseFee:                     nil,
		GasTipCap:                   nil,
		GasFeeCap:                   nil,
		GasUnit:                     0,
	}

	// Check Balance
	balance, err := util.GetTokenBalance(client, usdcAddress, middlewareWallet.Address)
	if err != nil {
		return result, err
	}

	// Check if middleware wallet has expected balance to sweep
	if balance.Cmp(minBalance) < 0 {
		return result, fmt.Errorf("middleware wallet does not have enough balance to sweep")
	}

	// Estimate gas fee, which means getting
	// 1. BaseFee from the latest block header
	// 2. PriorityFee/GasTipCap = SuggestGasTipCap
	// 3. GasFeeCap = BaseFee + GasTipCap
	// 4. GasUnit = EstimateGas
	// 5. MiddlewareGasFee = GasFeeCap * GasUnit (amount we send to middleware)
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return result, err
	}
	result.BaseFee = header.BaseFee

	result.GasTipCap, err = client.SuggestGasTipCap(context.Background())
	if err != nil {
		return result, err
	}

	result.GasFeeCap = new(big.Int).Add(result.BaseFee, result.GasTipCap)

	data := util.BuildTokenTxDataField(desAddress, balance) // data field for contract tokens transfer
	msg := ethereum.CallMsg{                                // test transaction
		From:  middlewareWallet.Address,
		To:    &usdcAddress,
		Value: big.NewInt(0), // value
		Data:  data,
	}

	result.GasUnit, err = client.EstimateGas(context.Background(), msg)
	if err != nil {
		log.Printf("EstimateGas failed, using default value 65000: %v", err)
		result.GasUnit = 65000
	}

	middlewareGasFee := new(big.Int).Mul(result.GasFeeCap, big.NewInt(int64(result.GasUnit)))

	if gasCostThreshold.Cmp(result.GasFeeCap) > 0 {
		return result, fmt.Errorf("gas fee cap is too high")
	}

	// sweep transaction
	// 1. Transfer ETH gas fee from provider wallet to middleware wallet
	tx1, err := util.SendTx(&util.TxInput{
		Client:     client,
		From:       providerWalletAddress,
		To:         middlewareWallet.Address,
		Amount:     middlewareGasFee,
		GasTipCap:  result.GasTipCap,
		GasFeeCap:  result.GasFeeCap,
		GasUnit:    21000,
		PrivateKey: providerWalletPK,
	})
	if err != nil {
		return result, err
	}

	result.ProviderToMiddlewareReceipt, err = bind.WaitMined(context.Background(), client, tx1)
	if err != nil {
		return result, err
	} else if result.ProviderToMiddlewareReceipt.Status != 1 {
		return result, fmt.Errorf("provider to middleware transaction failed")
	}

	// 2. Transfer USDC from middleware wallet to destination wallet
	result.MiddlewareToDestinationTx, err = util.SendTokenTx(usdcAddress, &util.TxInput{
		Client:     client,
		From:       middlewareWallet.Address,
		To:         desAddress,
		Amount:     balance,
		GasTipCap:  result.GasTipCap,
		GasFeeCap:  result.GasFeeCap,
		GasUnit:    result.GasUnit,
		PrivateKey: privateKey,
	})
	if err != nil {
		return result, err
	}
	return result, nil
}
