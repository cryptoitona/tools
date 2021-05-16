package bsc

import "math/rand"

var rpcList = []string{
	// 10+ nodes balanced, US/EU
	"https://bsc-dataseed1.ninicoin.io",
	// 10+ nodes balanced, US/EU
	"https://bsc-dataseed1.defibit.io",
	// 10+ nodes balanced in each region, global
	"https://bsc-dataseed.binance.org",
}

func GetURL() string {
	return rpcList[rand.Intn(len(rpcList))]
}

// https://github.com/pancakeswap/pancake-frontend/blob/3a10e0bb5352fc01c876efa17dfa844249a17e06/src/utils/addressHelpers.ts#L5
const MainnetChainID = 56
