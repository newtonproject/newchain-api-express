package utils

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// account faucet
const (
	AccountFaucetDripped = 1
)

var (
	BigNEW *big.Int = big.NewInt(1e+18) // BigNEW in WEI
)

// SplitBalance convert balance to a decimal number in NEW
// return the integer and fractional parts
func SplitBalance(balance *big.Int) (uint64, uint64) {
	return big.NewInt(0).Div(balance, BigNEW).Uint64(), big.NewInt(0).Mod(balance, BigNEW).Uint64()
}

// MergeBalance merge balance in interger and fractional into big.Int
func MergeBalance(i, d uint64) *big.Int {
	return big.NewInt(0).Add(big.NewInt(0).Mul(big.NewInt(0).SetUint64(i), BigNEW), big.NewInt(0).SetUint64(d))
}

// MergeBalanceNEW MergeBalanceNEW string
func MergeBalanceNEW(i, d uint64) string {
	strD := big.NewInt(0).SetUint64(d).String()

	return big.NewInt(0).SetUint64(i).String() + "." + strings.Repeat("0", 18-len(strD)) + strD
}

func AddBalance(xi, xd, yi, yd uint64) (i, d uint64) {
	i = xi + yi

	d = xd + yd
	if d < xd || d < yd {
		i++
	}

	return i, d
}

var (
	big10        = big.NewInt(10)
	big1NEWInWEI = new(big.Int).Exp(big10, big.NewInt(18), nil)

	// newton
	// 100,000,000,000 NEW
	NewtonMax = big.NewInt(0).Mul(big.NewInt(100000000000), big1NEWInWEI)

	// errors
	errBigSetString  = errors.New("convert string to big error")
	errLessThan0Wei  = errors.New("the transaction amount is less than 0 ISAAC")
	errIllegalAmount = errors.New("illegal amount")
	errIllegalUnit   = errors.New("illegal unit")
)

func GetNewtonAmountISAAC(amountStr, unit string) (*big.Int, error) {
	amount, err := GetAmountISAAC(amountStr, unit)
	if err != nil {
		return nil, err
	}
	if amount.Cmp(big.NewInt(0)) < 0 {
		return nil, errLessThan0Wei
	}
	if amount.Cmp(NewtonMax) > 0 {
		return nil, errors.New("max")
	}

	return amount, nil
}

func GetAmountISAAC(amountStr, unit string) (*big.Int, error) {
	if amountStr == "" || amountStr == "0" {
		return big.NewInt(0), nil
	}
	switch unit {
	case "NEW":
		index := strings.IndexByte(amountStr, '.')
		if index <= 0 {
			amountWei, ok := new(big.Int).SetString(amountStr, 10)
			if !ok {
				return nil, errBigSetString
			}
			return new(big.Int).Mul(amountWei, big1NEWInWEI), nil
		}
		amountStrInt := amountStr[:index]
		amountStrDec := amountStr[index+1:]
		amountStrDecLen := len(amountStrDec)
		if amountStrDecLen > 18 {
			return nil, errIllegalAmount
		}
		amountStrInt = amountStrInt + strings.Repeat("0", 18)
		amountStrDec = amountStrDec + strings.Repeat("0", 18-amountStrDecLen)
		if amountStrInt[0] == '-' {
			amountStrDec = fmt.Sprintf("-%s", amountStrDec)
		}

		amountStrIntBig, ok := new(big.Int).SetString(amountStrInt, 10)
		if !ok {
			return nil, errBigSetString
		}
		amountStrDecBig, ok := new(big.Int).SetString(amountStrDec, 10)
		if !ok {
			return nil, errBigSetString
		}

		return new(big.Int).Add(amountStrIntBig, amountStrDecBig), nil
	case "ISAAC":
		amountWei, ok := new(big.Int).SetString(amountStr, 10)
		if !ok {
			return nil, errBigSetString
		}
		return amountWei, nil
	}

	return nil, errIllegalUnit
}

func GetISAACAmountTextUnitByUnit(amount *big.Int, unit string) string {
	if amount == nil {
		return "0 ISAAC"
	}
	amountStr := amount.String()
	amountStrLen := len(amountStr)
	if unit == "" {
		if amountStrLen <= 18 {
			unit = "ISAAC"
		} else {
			unit = "NEW"
		}
	}

	return fmt.Sprintf("%s %s", GetISAACAmountTextByUnit(amount, unit), unit)
}

// GetISAACAmountTextByUnit convert 1000000000 ISAAC to string of 0.000000001 NEW or 1000000000 ISAAC
func GetISAACAmountTextByUnit(amount *big.Int, unit string) string {
	if amount == nil {
		return "0"
	}
	var (
		negStr string
		amountStr string
	)
	if amount.Cmp(big.NewInt(0)) < 0 {
		negStr = "-"
		amountStr = big.NewInt(0).Neg(amount).String()
	} else {
		amountStr = amount.String()
	}

	amountStrLen := len(amountStr)

	switch unit {
	case "NEW":
		var amountStrDec, amountStrInt string
		if amountStrLen <= 18 {
			amountStrDec = strings.Repeat("0", 18-amountStrLen) + amountStr
			amountStrInt = "0"
		} else {
			amountStrDec = amountStr[amountStrLen-18:]
			amountStrInt = amountStr[:amountStrLen-18]
		}
		amountStrDec = strings.TrimRight(amountStrDec, "0")
		if len(amountStrDec) <= 0 {
			return negStr + amountStrInt
		}
		return negStr + amountStrInt + "." + amountStrDec

	case "ISAAC":
		return negStr + amountStr
	}

	return errIllegalUnit.Error()
}

func GetTodayUnix() (int64) {
	return GetTodayStartUnix()
}

func GetTodayStartUnix() (int64) {
	year, month, day := time.Now().UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix()
}

func GetTodayEndUnix() (int64) {
	return GetTomorrowUnix()
}

func GetYesterdayUnix() (int64) {
	year, month, day := time.Now().UTC().Add(-1 * time.Hour * 24 ).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix()
}

func GetYesterdayStartUnix() (int64) {
	year, month, day := time.Now().UTC().Add(-1 * time.Hour * 24 ).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix()
}

func GetYesterdayEndUnix() (int64) {
	year, month, day := time.Now().UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix()
}

func GetTomorrowUnix() (int64) {
	year, month, day := time.Now().UTC().Add(time.Hour * 24 ).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix()
}