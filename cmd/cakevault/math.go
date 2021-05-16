package main

import "math/big"

func multiplyBy(x, y *big.Float) *big.Float {
	return big.NewFloat(0).Mul(x, y)
}

func divideBy(x, y *big.Float) *big.Float {
	return big.NewFloat(0).Quo(x, y)
}

func Pow(a *big.Int, b *big.Int) *big.Float {
	i := new(big.Int).Exp(a, b, nil)
	return new(big.Float).SetInt(i)
}

func getFullDisplayBalance(balance *big.Float, decimals int) *big.Float {
	pow := Pow(big.NewInt(10), big.NewInt(int64(decimals)))
	return divideBy(balance, pow)
}
