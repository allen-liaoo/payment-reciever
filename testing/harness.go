package harness

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// "MiddlewarePKs": [
//     "0xdc0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
//   ],

//   "SweeperPK": "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",

//   "RecipientPK": "0xbc0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

type ChainInfo struct {
	USDTCreatorAddress common.Address
	USDTAddress        common.Address
}

var chainInfo ChainInfo

func init() {
	// Get the absolute path of the current file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("Error retrieving caller information")
		return
	}

	// Resolve the relative path to the target file
	relPath := "chain_info.json"
	absPath := filepath.Join(filepath.Dir(filename), relPath)

	jsonFile, err := os.Open(absPath)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}

	type ChainInfoInternal struct {
		USDTCreatorPK string `json:"USDTCreatorPK"`
		USDTAddress   string `json:"USDTAddress"`
	}

	var chainInfoIn ChainInfoInternal

	if err := json.Unmarshal(byteValue, &chainInfoIn); err != nil {
		panic(err)
	}

	pk, err := crypto.HexToECDSA(chainInfoIn.USDTCreatorPK)
	if err != nil {
		panic(err)
	}

	chainInfo.USDTAddress = common.HexToAddress(chainInfoIn.USDTAddress)
	chainInfo.USDTCreatorAddress = crypto.PubkeyToAddress(*pk.Public().(*ecdsa.PublicKey))
}
