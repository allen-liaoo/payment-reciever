package util

import (
	"encoding/hex"
	"math/big"
	"slices"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// unit test: BuildTokenTxDataField
func TestBuildTokenTxDataField(t *testing.T) {
	transferMethodID, err := hex.DecodeString("a9059cbb")
	if err != nil {
		t.Fatal(err)
	}
	to := common.HexToAddress("0xf3cE9fE9aD09d5540a4aa07367ebA056bEd45bd0")
	amount := big.NewInt(4000000)
	expected := slices.Concat(transferMethodID, common.LeftPadBytes(to.Bytes(), 32), common.LeftPadBytes(amount.Bytes(), 32))

	t.Run(to.Hex(), func(t *testing.T) {
		result := BuildTokenTxDataField(to, amount)
		if !equal(result, expected) {
			t.Errorf("BuildTokenTxDataField(%v, %v) = %v; want %v", to, amount, result, expected)
		}
	})
}

func equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
