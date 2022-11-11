package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/SealTV/soliditylearn/start"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// client, err := ethclient.Dial("http://localhost:8545")
	// client, err := ethclient.Dial("https://cloudflare-eth.com")
	client, err := ethclient.Dial("wss://ropsten.infura.io/ws")
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	start.Account(client)
	start.GenerateWallet(client)
	fp := start.CreateKs()
	start.ReadKs(fp)

	start.AddressCheck(ctx, client)

	if err := start.QueryBlocks(ctx, client); err != nil {
		log.Fatal(err)
	}

	if err := start.SubscribingToNewBlocks(ctx, client); err != nil {
		log.Fatal(err)
	}
}
