package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"time"

	"github.com/SealTV/soliditylearn/lock"
	"github.com/SealTV/soliditylearn/store"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

func main() {
	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatal(err)
	}

	wsClient, err := ethclient.Dial("ws://127.0.0.1:8545")
	// client, err := ethclient.Dial("https://cloudflare-eth.com")
	// client, err := ethclient.Dial("wss://ropsten.infura.io/ws")
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// storeSmartContract(ctx, client, deployStore)

	if err := subscribeForEvents(ctx, wsClient); err != nil {
		log.Fatal(err)
	}

	if err := loadingSmartContract(ctx, client); err != nil {
		log.Fatal(err)
	}
}

type contractDelpoy func(ctx context.Context, client *ethclient.Client, auth *bind.TransactOpts) error

func deployLock(ctx context.Context, client *ethclient.Client, auth *bind.TransactOpts) error {
	// input := "1.0"
	unlockTime := big.NewInt(time.Now().Add(1 * time.Hour * 24_360).Unix())
	address, tx, instance, err := lock.DeployLock(auth, client, unlockTime)
	if err != nil {
		return err
	}

	fmt.Println(address.Hex())   // 0xacC09486C9e34aa1Dff13b2E64d5482A3648D018
	fmt.Println(tx.Hash().Hex()) // 0x23069a5ff569d3c3f71a4ab12e4ae4077139b55bd45aef3272aa39275db2bdad

	_ = instance
	return nil
}

func checkAccount(ctx context.Context, client *ethclient.Client) {
	account := common.HexToAddress("0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199")
	balance, err := client.BalanceAt(ctx, account, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("account balance:", balance)

}

func deployStore(ctx context.Context, client *ethclient.Client, auth *bind.TransactOpts) error {
	address, tx, instance, err := store.DeployStore(auth, client, "v1.0.0")
	if err != nil {
		return err
	}

	fmt.Println(address.Hex())   // 0xacC09486C9e34aa1Dff13b2E64d5482A3648D018
	fmt.Println(tx.Hash().Hex()) // 0x23069a5ff569d3c3f71a4ab12e4ae4077139b55bd45aef3272aa39275db2bdad

	_ = instance

	return nil
}

func storeSmartContract(ctx context.Context, client *ethclient.Client, deploy contractDelpoy) error {
	privateKey, err := crypto.HexToECDSA("df57089febbacf7ba0bc227dafbffa9fc08a93fdc68e1e42411a14efcf23656e")
	if err != nil {
		return err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)       // in wei
	auth.GasLimit = uint64(30000000) // in units
	auth.GasPrice = gasPrice

	if err := deploy(ctx, client, auth); err != nil {
		return errors.Wrap(err, "cannot deploy contract")
	}
	return nil
}

func loadingSmartContract(ctx context.Context, client *ethclient.Client) error {
	privateKey, err := crypto.HexToECDSA("de9be858da4a475276426320d5e9262ecfc3ba460bfac56360bfa6c4c28b4ee0")
	if err != nil {
		return errors.Wrap(err, "cannot get private key from hex string")
	}

	publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return errors.New("error casting public key to ECDSA")
	}

	nonce, err := client.PendingNonceAt(ctx, crypto.PubkeyToAddress(*publicKeyECDSA))
	if err != nil {
		return errors.Wrap(err, "cannot pending nonce at")
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot suggest gas proice")
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot get chain_id")
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return errors.Wrap(err, "cannon get new keyed transactor")
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	instance, err := getStoreInstance(ctx, client)
	if err != nil {
		return errors.Wrap(err, "cannot connect new smart contract instance")
	}

	version, err := instance.Version(nil)
	if err != nil {
		return errors.Wrap(err, "cannot get version")
	}

	log.Println("Store version:", version)

	log.Println("try to store data")
	key := [32]byte{}
	value := [32]byte{}

	copy(key[:], "key3")
	copy(value[:], "Hello world")

	tx, err := instance.SetItem(auth, key, value)
	if err != nil {
		return errors.Wrap(err, "cannot set item")
	}

	log.Println("transaction const", tx.Cost().Int64())
	log.Printf("tx sent: %s\n", tx.Hash().Hex()) // tx sent: 0x8d490e535678e9a24360e955d75b27ad307bdfb97a1dca51d0f3035dcee3e870

	result, err := instance.Items(nil, key)
	if err != nil {
		log.Fatal(err)
		return errors.Wrap(err, "cannot get items from smartcontract")
	}

	fmt.Println(string(result[:])) // "bar"
	return nil
}

func subscribeForEvents(ctx context.Context, client *ethclient.Client) error {
	instance, err := getStoreInstance(ctx, client)
	if err != nil {
		return errors.Wrap(err, "cannot connect new smart contract instance")
	}

	c := make(chan *store.StoreItemSet)
	sub, err := instance.WatchItemSet(nil, c)
	if err != nil {
		return errors.Wrap(err, "cannot get watcher")
	}

	go func() {
		log.Println("START listening")
		defer log.Println("STOP listening")
		defer sub.Unsubscribe()

		for {
			select {
			case itemSet := <-c:
				log.Println("new data", string(itemSet.Key[:]), string(itemSet.Value[:]))
			case err := <-sub.Err():
				log.Println("error:", err)
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func getStoreInstance(ctx context.Context, client *ethclient.Client) (*store.Store, error) {
	address := common.HexToAddress("0x73511669fd4de447fed18bb79bafeac93ab7f31f")
	instance, err := store.NewStore(address, client)
	if err != nil {
		return nil, errors.Wrap(err, "cannot connect new smart contract instance")
	}

	return instance, nil
}
