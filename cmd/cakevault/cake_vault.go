package main

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/cryptoitona/tools/internal/bsc"
	"github.com/cryptoitona/tools/internal/cakevault"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

//go:generate abigen --abi=./internal/cakevault/cakeVault.abi --pkg=cakevault --out=./internal/cakevault/cake_vault.go

type CakeVault struct {
	cli *bsc.Client
	cv  *cakevault.Cakevault
}

const cakeVaultAddress = "0xa80240Eb5d7E05d3F250cF000eEc0891d00b51CC"

func NewCakeVault(ctx context.Context, walletKey *ecdsa.PrivateKey) (*CakeVault, error) {
	cli, err := bsc.NewClient(ctx, bsc.GetURL(), walletKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	cv, err := cakevault.NewCakevault(
		common.HexToAddress(cakeVaultAddress),
		cli.RawClient(),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &CakeVault{
		cli: cli,
		cv:  cv,
	}, nil
}

func (c *CakeVault) Claim(ctx context.Context) (*types.Transaction, error) {
	txopts, err := c.cli.BaseTransactOpts(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	tx, err := c.cv.Harvest(txopts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return tx, nil
}

func (c *CakeVault) Calc(ctx context.Context, cakePrice float64) (*CalcResult, error) {
	eg, ctx := errgroup.WithContext(ctx)

	var (
		estimatedRewards        *big.Int
		totalPendingCakeRewards *big.Int

		// related fee
		contractCallFee *big.Int
	)

	opt := &bind.CallOpts{
		From:    c.cli.WalletAddress(),
		Context: ctx,
	}

	eg.Go(func() (err error) {
		estimatedRewards, err = c.cv.CalculateHarvestCakeRewards(opt)
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	})
	eg.Go(func() (err error) {
		totalPendingCakeRewards, err = c.cv.CalculateTotalPendingCakeRewards(opt)
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	})
	eg.Go(func() (err error) {
		contractCallFee, err = c.cv.CallFee(opt)
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	estimatedCallBountyReward := new(big.Float).SetInt(estimatedRewards)

	return &CalcResult{
		totalPendingCakeRewards:   totalPendingCakeRewards,
		estimatedCallBountyReward: estimatedCallBountyReward,
		dollarValueOfReward:       multiplyBy(estimatedCallBountyReward, big.NewFloat(cakePrice)),
		callFee:                   contractCallFee,
	}, nil
}

type CalcResult struct {
	totalPendingCakeRewards   *big.Int
	estimatedCallBountyReward *big.Float
	dollarValueOfReward       *big.Float
	callFee                   *big.Int
}

func (c *CalcResult) IsValid() bool {
	zeroFloat := new(big.Float).SetFloat64(0)
	return c.dollarValueOfReward.Cmp(zeroFloat) > 0 &&
		c.estimatedCallBountyReward.Cmp(zeroFloat) > 0
}

func (c *CalcResult) DollarCallBountyToDisplay() *big.Float {
	return getFullDisplayBalance(c.dollarValueOfReward, 18)
}

func (c *CalcResult) CakeCallBountyToDisplay() *big.Float {
	return getFullDisplayBalance(c.estimatedCallBountyReward, 18)
}

func (c *CalcResult) TotalYieldToDisplay() *big.Float {
	tmp := new(big.Float).SetInt(c.totalPendingCakeRewards)
	return getFullDisplayBalance(tmp, 18)
}

func (c *CalcResult) CallFee() float64 {
	return float64(c.callFee.Int64()) / 100.0
}
