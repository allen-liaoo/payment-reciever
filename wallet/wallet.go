package wallet

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/crypto/sha3"

	contracts "allen-liaoo/payment-reciever/contracts"
)

type Contract struct {
	name     string
	decimals uint8
}

var knownContracts map[string]Contract

func init() {

	file, err := os.Open("wallet/decimals.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	byteValue, _ := io.ReadAll(file)
	json.Unmarshal([]byte(byteValue), &knownContracts)

}

func GetContractDecimals(client *ethclient.Client, contractAddress common.Address) (uint8, error) {

	con, ok := knownContracts[contractAddress.Hex()]
	if ok {
		return con.decimals, nil
	}

	// otherwise, lookup contract decimals
	contract, err := contracts.NewErc20(contractAddress, client)
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
	contract, err := contracts.NewErc20(contractAddress, client)
	if err != nil {
		return big.NewInt(0), err
	}

	balance, err := contract.BalanceOf(nil, walletAddress)
	if err != nil {
		return big.NewInt(0), err
	}

	return balance, nil
}

// display unit to raw unit
func FromRawUnit(amount *big.Int, decimals uint8) *big.Int {
	base := big.NewInt(10)
	var newAmount big.Int
	newAmount.Div(amount, base.Exp(base, big.NewInt(int64(decimals)), nil))
	return &newAmount
}

func ToRawUnit(amount *big.Int, decimals uint8) *big.Int {
	base := big.NewInt(10)
	var newAmount big.Int
	newAmount.Mul(amount, base.Exp(base, big.NewInt(int64(decimals)), nil))
	return &newAmount
}

// returns the hash of the transaction (in hex)
func SendTx(client *ethclient.Client, from common.Address, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, privateKey *ecdsa.PrivateKey) (string, error) {
	nonce, err := client.PendingNonceAt(context.Background(), from)
	if err != nil {
		return "", err
	}

	tx := types.NewTransaction(
		nonce,
		to,
		amount,
		gasLimit,
		gasPrice,
		data,
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

func SendTokenTx(client *ethclient.Client, from common.Address, contractAddress common.Address, gasLimit uint64, gasPrice *big.Int, data []byte, privateKey *ecdsa.PrivateKey) (string, error) {
	return SendTx(client, from, contractAddress, big.NewInt(0), gasLimit, gasPrice, data, privateKey)
}

func BuildTokenTxDataField(contractAddress common.Address, to common.Address, amount *big.Int) []byte {
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4] // method ID is first four bytes
	// fmt.Println("method id:", hexutil.Encode(methodID))

	paddedAddress := common.LeftPadBytes(to.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)
	return data
}
