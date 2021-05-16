package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		os.Exit(1)
	}
}

const targetCake = 0.001

func run(ctx context.Context) error {
	privKey := os.Getenv("WALLET_PRIVATE_KEY")
	privateKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		log.Fatal(err)
	}

	cv, err := NewCakeVault(ctx, privateKey)
	if err != nil {
		return err
	}

	getCakePrice := getCakePriceFunc(ctx)

	for i := 0; ; i++ {
		cakePrice, err := getCakePrice()
		if err != nil {
			log.Println("failed to get cake price:", err)
			continue
		}
		result, err := cv.Calc(ctx, cakePrice)
		if err != nil {
			return errors.WithStack(err)
		}
		if !result.IsValid() {
			continue
		}

		cakeBounty := result.CakeCallBountyToDisplay()
		fmt.Printf(
			"Fee: %.2f%%, Doller: $%.3f, Cake: %.5f, Pool total pending yield: %.3f CAKE (per cake; $%.5f), valid: %v\r",
			result.CallFee(),
			result.DollarCallBountyToDisplay(),
			cakeBounty,
			result.TotalYieldToDisplay(),
			cakePrice,
			result.IsValid(),
		)
		if cakeBounty.Cmp(big.NewFloat(targetCake)) < 0 {
			continue
		}
		fmt.Println()

		tx, err := cv.Claim(ctx)
		if err != nil {
			return err
		}
		log.Println("tx hash:", tx.Hash())
		log.Println("bscscan: ", "https://bscscan.com/tx/"+tx.Hash().String())
		return nil
	}
}
