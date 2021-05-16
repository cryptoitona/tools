package bsc

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

// Client is a wrapper of go-ethereum/ethclient to use as bsc client
type Client struct {
	rawClient  *ethclient.Client
	privateKey *ecdsa.PrivateKey
}

func NewClient(ctx context.Context, url string, walletKey *ecdsa.PrivateKey) (*Client, error) {
	cli, err := ethclient.DialContext(ctx, url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &Client{
		rawClient:  cli,
		privateKey: walletKey,
	}, nil
}

func (c *Client) RawClient() *ethclient.Client {
	return c.rawClient
}

func (c *Client) BaseTransactOpts(ctx context.Context) (*bind.TransactOpts, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(c.privateKey, big.NewInt(MainnetChainID))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	nonce, err := c.rawClient.PendingNonceAt(ctx, c.WalletAddress())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	gasPrice, err := c.rawClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	opts.Nonce = big.NewInt(int64(nonce))
	opts.GasPrice = gasPrice
	// opts.GasLimit = 0   // Gas limit to set for the transaction execution (0 = estimate)
	opts.Context = ctx

	return opts, nil
}

func (c *Client) WalletAddress() common.Address {
	return crypto.PubkeyToAddress(c.privateKey.PublicKey)
}
