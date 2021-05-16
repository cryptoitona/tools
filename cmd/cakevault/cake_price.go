package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cryptoitona/tools/internal/bsc"
	"github.com/pkg/errors"
)

type Address struct {
	Symbol      string
	Addresses   map[int]string
	Decimals    int
	ProjectLink string
}

var tokens = map[string]*Address{
	"cake": {
		Symbol: "CAKE",
		Addresses: map[int]string{
			bsc.MainnetChainID: "0x0e09fabb73bd3ade0a17ecc321fd13a19e81ce82",
			97:                 "0xa35062141Fa33BCA92Ce69FeD37D0E8908868AAe",
		},
		Decimals:    18,
		ProjectLink: "https://pancakeswap.finance/",
	},
}

func getCakeAddress() string {
	return tokens["cake"].Addresses[bsc.MainnetChainID]
}

const tokenAPIURL = "https://api.pancakeswap.info/api/v2/tokens"

func fetchPrices(ctx context.Context) (*PriceAPIResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", tokenAPIURL, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var result PriceAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.WithStack(err)
	}
	return &result, nil
}

var _ json.Unmarshaler = (*PriceAPIResponse)(nil)

type PriceAPIResponse struct {
	UpdatedAt string
	Data      map[string]*PriceAPI
}

func (p *PriceAPIResponse) UnmarshalJSON(b []byte) error {
	type PriceAPIResponseTmp struct {
		UpdatedAt string
		Data      map[string]*PriceAPI
	}
	var v PriceAPIResponseTmp
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	p.UpdatedAt = v.UpdatedAt
	p.Data = make(map[string]*PriceAPI, len(v.Data))
	for k, v := range v.Data {
		p.Data[strings.ToLower(k)] = v
	}
	return nil
}

type PriceAPI struct {
	Name     string      `json:"name"`
	Symbol   string      `json:"symbol"`
	Price    json.Number `json:"price"`
	PriceBNB json.Number `json:"price_BNB"`
}

var priceAPICache atomic.Value

func getCakePriceFunc(ctx context.Context) func() (float64, error) {
	initialize := make(chan struct{}, 1)
	initialize <- struct{}{}
	waitCh := make(chan struct{})
	ticker := time.NewTicker(time.Minute)
	go func() {
		var once sync.Once
		for {
			select {
			case <-initialize:
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
			resp, err := fetchPrices(ctx)
			if err != nil {
				log.Println("error fetchPrices:", err)
				continue
			}
			cake := resp.Data[getCakeAddress()]
			priceAPICache.Store(cake)

			once.Do(func() {
				close(waitCh)
			})
		}
	}()
	return func() (float64, error) {
		<-waitCh
		cake := priceAPICache.Load().(*PriceAPI)
		return cake.Price.Float64()
	}
}
