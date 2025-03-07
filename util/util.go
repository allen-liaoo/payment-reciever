package util

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"golang.org/x/crypto/sha3"

	"allen-liaoo/payment-reciever/erc20"
)

type Contract struct {
	Name     string
	Decimals uint8
}

var knownContracts = map[string]Contract{
	"0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238": {
		Name:     "USDC?",
		Decimals: 6,
	},
}

func GetContractDecimals(client *ethclient.Client, contractAddress common.Address) (uint8, error) {

	con, ok := knownContracts[contractAddress.Hex()]
	if ok {
		return con.Decimals, nil
	}

	// otherwise, lookup contract decimals
	contract, err := erc20.NewErc20(contractAddress, client)
	if err != nil {
		return 0, err
	}

	decimals, err := contract.Decimals(nil)
	if err != nil {
		return 0, err
	}
	fmt.Println("New decimals ", decimals, " for contract address ", contractAddress.Hex())

	return decimals, nil

}

// return balance, decimals, error
func GetTokenBalance(client *ethclient.Client, contractAddress common.Address, walletAddress common.Address) (*big.Int, error) {
	contract, err := erc20.NewErc20(contractAddress, client)
	if err != nil {
		return big.NewInt(0), err
	}

	balance, err := contract.BalanceOf(nil, walletAddress)
	if err != nil {
		return big.NewInt(0), err
	}

	return balance, nil
}

// from/to smallest unit as defined by decimals
func FromSmallestUnit(amount *big.Int, decimals uint8) *big.Int {
	base := big.NewInt(10)
	var newAmount big.Int
	newAmount.Div(amount, base.Exp(base, new(big.Int).SetInt64(int64(decimals)), nil))
	return &newAmount
}

func ToSmallestUnit(amount *big.Int, decimals uint8) *big.Int {
	base := big.NewInt(10)
	var newAmount big.Int
	newAmount.Mul(amount, base.Exp(base, big.NewInt(int64(decimals)), nil))
	return &newAmount
}

// returns the hash of the transaction (in hex)
type TxInput struct {
	Client     *ethclient.Client
	From       common.Address
	To         common.Address
	Amount     *big.Int
	GasFeeCap  *big.Int
	GasTipCap  *big.Int
	GasUnit    uint64
	Data       []byte
	PrivateKey *ecdsa.PrivateKey
}

func sendTx(input *TxInput) (*types.Transaction, error) {
	nonce, err := input.Client.PendingNonceAt(context.Background(), input.From)
	if err != nil {
		return nil, err
	}
	chainID, err := input.Client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	// EIP1559 transaction
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        &input.To,
		Value:     input.Amount,
		GasFeeCap: input.GasFeeCap,
		GasTipCap: input.GasTipCap,
		Gas:       input.GasUnit,
		Data:      input.Data,
	})
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), input.PrivateKey)
	if err != nil {
		return nil, err
	}
	err = input.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

func SendTx(input *TxInput) (*types.Transaction, error) {
	return sendTx(input)
}

// automatically builds data field
func SendTokenTx(contractAddress common.Address, input *TxInput) (*types.Transaction, error) {
	input.Data = BuildTokenTxDataField(input.To, input.Amount)
	input.To = contractAddress
	input.Amount = big.NewInt(0)
	return SendTx(input)
}

func BuildTokenTxDataField(to common.Address, amount *big.Int) []byte {
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4] // method ID is first four bytes

	paddedAddress := common.LeftPadBytes(to.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)
	return data
}

// DeriveWallet derive wallet from mnemonic and path. It returns the account and private key.
func DeriveWallet(mnemonic string, path string) (*accounts.Account, *ecdsa.PrivateKey, error) {
	derPath, err := hdwallet.ParseDerivationPath(path)
	if err != nil {
		return nil, nil, err
	}
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, nil, err
	}
	account, err := wallet.Derive(derPath, false)
	if err != nil {
		return nil, nil, err
	}
	privateKey, err := wallet.PrivateKey(account)
	if err != nil {
		return nil, nil, err
	}
	return &account, privateKey, nil
}
