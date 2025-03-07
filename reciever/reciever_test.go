package reciever

import (
	"allen-liaoo/payment-reciever/util"
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

// Performance testing
// Simulates a number of transactions of the form:
// 1. Provider sends USDC to middleware
// 2. Initiate process of checking and sweeping middleware funds
// 3. Check if middleware funds are swept to destination
// 4. Print gas usage for each transaction
func TestHandleMiddleware(t *testing.T) {
	// Define test parameters
	numWallets := 3
	startWalletPath := 50
	USDCAmount := big.NewInt(20)

	fmt.Println("Starting middleware wallet test")
	fmt.Println("Provider wallet:", providerWalletAddress.Hex())

	// Validate provider has enough USDC for tests
	providerBalance, err := util.GetTokenBalance(client, usdcAddress, providerWalletAddress)
	assert.NoError(t, err)
	decimals, err := util.GetContractDecimals(client, usdcAddress)
	assert.NoError(t, err)
	fmt.Printf("Provider USDC balance: %s (%s)\n",
		providerBalance.String(),
		util.FromSmallestUnit(providerBalance, decimals))

	// Check if provider has enough USDC for tests
	minRequired := new(big.Int).Mul(USDCAmount, big.NewInt(int64(numWallets)))
	assert.True(t, providerBalance.Cmp(minRequired) >= 0,
		"Provider needs at least %s USDC for tests", minRequired)

	// Test with multiple wallets
	for i := 0; i < numWallets; i++ {
		t.Run(fmt.Sprintf("Middleware%d", i), func(t *testing.T) {
			// Derive a middleware wallet with a unique derivation path
			derivationPath := fmt.Sprintf("m/44'/60'/0'/0/%d", startWalletPath+i)
			middlewareWallet, privateKey, err := util.DeriveWallet(middlewareWalletMneumonic, derivationPath)
			assert.NoError(t, err)

			fmt.Printf("\nTest %d: Using middleware wallet: %s\n", i+1, middlewareWallet.Address.Hex())
			fmt.Printf("\tFrom path: %s\n", derivationPath)
			fmt.Printf("\tPrivate key: %x \n", crypto.FromECDSA(privateKey))

			// Fund the middleware wallet
			err = sendUSDCToMiddleware(middlewareWallet.Address, USDCAmount)
			assert.NoError(t, err)

			// Verify middleware received the funds
			balance, err := util.GetTokenBalance(client, usdcAddress, middlewareWallet.Address)
			balance = util.ToSmallestUnit(balance, decimals)
			assert.NoError(t, err)
			fmt.Printf("Middleware received %s USDC\n",
				balance.String())
			assert.True(t, balance.Cmp(big.NewInt(0)) > 0, "Middleware should have positive balance")

			// Now handle the middleware wallet (sweep funds)
			startTime := time.Now()
			result, err := SweepMiddleware(middlewareWallet, privateKey, USDCAmount, big.NewInt(0))
			elapsedTime := time.Since(startTime)

			assert.NoError(t, err)
			assert.Equal(t, uint64(1), result.ProviderToMiddlewareReceipt.Status, "1st Transaction should be successful")
			assert.NotNil(t, result.MiddlewareToDestinationTx)

			fmt.Printf("Sweep completed in %v\n", elapsedTime)
			fmt.Printf("\tSweep transaction initiated, tx: %s\n", result.MiddlewareToDestinationTx.Hash().Hex())

			// Wait for the sweep transaction to be mined
			fmt.Println("\tWaiting for sweep transaction to be confirmed...")
			sweepReceipt, err := bind.WaitMined(context.Background(), client, result.MiddlewareToDestinationTx)
			assert.NoError(t, err)
			fmt.Printf("\tSweep transaction confirmed in block %d\n", sweepReceipt.BlockNumber.Uint64())
			assert.Equal(t, uint64(1), sweepReceipt.Status, "2nd Transaction should be successful")

			// Verify middleware wallet balance is now 0
			balanceAfter, err := util.GetTokenBalance(client, usdcAddress, middlewareWallet.Address)
			assert.NoError(t, err)
			assert.Equal(t, 0, balanceAfter.Cmp(big.NewInt(0)), "Middleware wallet should be empty after sweep")

			// Check gas usage for sweep transaction
			sweepEfficiency := float64(sweepReceipt.GasUsed) / float64(result.GasUnit) * 100
			fmt.Printf("\tSweep gas: used %d / %d (%.2f%%)\n",
				sweepReceipt.GasUsed, result.GasUnit, sweepEfficiency)

			// Check leftover ETH balance at middleware wallet
			ethBalance, err := client.BalanceAt(context.Background(), middlewareWallet.Address, nil)
			assert.NoError(t, err)
			ethBalanceF, _ := ethBalance.Float64()
			middlewareGasFee, _ := new(big.Int).Mul(result.GasFeeCap, big.NewInt(int64(result.GasUnit))).Float64()
			fmt.Printf("\tLeftover balance percentage: %.2f%%\n", ethBalanceF/middlewareGasFee)

			fmt.Printf("\t%#v\n", result)
		})
	}
}

// Helper function to send USDC from provider to middleware
func sendUSDCToMiddleware(middlewareAddr common.Address, amount *big.Int) error {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return err
	}

	gasTipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return err
	}

	gasFeeCap := new(big.Int).Add(header.BaseFee, gasTipCap)

	fmt.Printf("Sending %s USDC to middleware\n",
		util.ToSmallestUnit(amount, 6))

	tx, err := util.SendTokenTx(usdcAddress, &util.TxInput{
		Client:     client,
		From:       providerWalletAddress,
		To:         middlewareAddr,
		Amount:     amount,
		GasTipCap:  gasTipCap,
		GasFeeCap:  gasFeeCap,
		GasUnit:    65000,
		PrivateKey: providerWalletPK,
	})

	if err != nil {
		return err
	}

	fmt.Printf("\tFund transfer initiated, tx: %s\n", tx.Hash().Hex())
	// Wait for the transaction to be mined
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return fmt.Errorf("error waiting for transaction to be mined: %w", err)
	}

	// Log transaction confirmed
	fmt.Printf("\tTransaction confirmed in block %d\n", receipt.BlockNumber.Uint64())
	return nil
}
