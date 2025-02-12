package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto/sha3"
)

// MockServices is a that tolds different mock services to be passed
// around easily for testing setup

// UintToPaddedString converts an int to string of length 19 by padding with 0
func UintToPaddedString(num int64) string {
	return fmt.Sprintf("%019d", num)
}

// GetTickChannelID is used to get the channel id for OHLCV data streaming
// it takes pairname, duration and units of data streaming
func GetTickChannelID(bt, qt common.Address, unit string, duration int64) string {
	pair := GetPairKey(bt, qt)
	return fmt.Sprintf("%s::%d::%s", pair, duration, unit)
}

// GetPairKey return the pair key identifier corresponding to two
func GetPairKey(bt, qt common.Address) string {
	return strings.ToLower(fmt.Sprintf("%s::%s", bt.Hex(), qt.Hex()))
}

func GetTradeChannelID(bt, qt common.Address) string {
	return strings.ToLower(fmt.Sprintf("%s::%s", bt.Hex(), qt.Hex()))
}

func GetOHLCVChannelID(bt, qt common.Address, unit string, duration int64) string {
	pair := GetPairKey(bt, qt)
	return fmt.Sprintf("%s::%d::%s", pair, duration, unit)
}

func GetOrderBookChannelID(bt, qt common.Address) string {
	return strings.ToLower(fmt.Sprintf("%s::%s", bt.Hex(), qt.Hex()))
}

func GetPriceBoardChannelID(bt, qt common.Address) string {
	return strings.ToLower(fmt.Sprintf("%s::%s", bt.Hex(), qt.Hex()))
}

func GetMarketsChannelID(channel string) string {
	return strings.ToLower(channel)
}

func Retry(retries int, fn func() error) error {
	if err := fn(); err != nil {
		retries--
		if retries <= 0 {
			return err
		}

		// preventing thundering herd problem (https://en.wikipedia.org/wiki/Thundering_herd_problem)
		time.Sleep(time.Second)

		return Retry(retries, fn)
	}

	return nil
}

func PrintJSON(x interface{}) {
	b, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		fmt.Println("Error: ", err)
	}

	fmt.Print(string(b), "\n")
}

func JSON(x interface{}) string {
	b, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		fmt.Println("Error: ", err)
	}

	return fmt.Sprint(string(b), "\n")
}

func PrintError(msg string, err error) {
	log.Printf("\n%v: %v\n", msg, err)
}

// Util function to handle unused variables while testing
func Use(...interface{}) {

}

// GetAddressFromPublicKey get derived address from public key
func GetAddressFromPublicKey(pk []byte) common.Address {
	var buf []byte
	hash := sha3.NewKeccak256()
	hash.Write(pk[1:]) // remove EC prefix 04
	buf = hash.Sum(nil)
	publicAddr := hexutil.Encode(buf[12:])

	publicAddress := common.HexToAddress(publicAddr)

	return publicAddress
}

// Union get set of two array
func Union(a, b []common.Address) []common.Address {
	m := make(map[common.Address]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; !ok {
			a = append(a, item)
		}
	}
	return a
}

func IsNativeTokenByAddress(address common.Address) bool {
	return (address.Hex() == "0x0000000000000000000000000000000000000001")
}
