package start

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"os"
	"os/signal"
	"regexp"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

func SubscribingToNewBlocks(ctx context.Context, cli *ethclient.Client) error {
	headers := make(chan *types.Header)

	sub, err := cli.SubscribeNewHead(ctx, headers)
	if err != nil {
		return errors.Wrap(err, "cannot subscribe to new heeaders")
	}
	defer sub.Unsubscribe()

	for {
		select {
		case err := <-sub.Err():
			return errors.Wrap(err, "error on from sub")
		case header := <-headers:
			fmt.Println(header.Hash().Hex())

			block, err := cli.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(block.Hash().Hex())        // 0xbc10defa8dda384c96a17640d84de5578804945d347072e091b4e5f390ddea7f
			fmt.Println(block.Number().Uint64())   // 3477413
			fmt.Println(block.Time())              // 1529525947
			fmt.Println(block.Nonce())             // 130524141876765836
			fmt.Println(len(block.Transactions())) // 7
		case <-ctx.Done():
			return nil
		}
	}

	return nil
}

func QueryBlocks(ctx context.Context, client *ethclient.Client) error {
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "error on get header by number")
	}

	log.Println(header.Number.String())

	blockNumber := big.NewInt(15946232)
	block, err := client.BlockByNumber(ctx, blockNumber)
	if err != nil {
		return errors.Wrap(err, "error on get block by number")
	}

	fmt.Println(block.Number().Uint64())     // 5671744
	fmt.Println(block.Time())                // 1527211625
	fmt.Println(block.Difficulty().Uint64()) // 3217000136609065
	fmt.Println(block.Hash().Hex())          // 0x9e8751ebb5069389b855bba72d94902cc385042661498a415979b7b6ee9ba4b9
	fmt.Println(len(block.Transactions()))   // 144

	count, err := client.TransactionCount(ctx, block.Hash())
	if err != nil {
		return errors.Wrap(err, "error on get transaction count")
	}

	fmt.Println("transaction count =", count)

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return errors.Wrap(err, "error on get NetworkID")
	}

	for _, tx := range block.Transactions() {
		fmt.Println(tx.Hash().Hex())        // 0x5d49fcaa394c97ec8a9c3e7bd9e8388d420fb050a52083ca52ff24b3b65bc9c2
		fmt.Println(tx.Gas())               // 105000
		fmt.Println(tx.GasPrice().Uint64()) // 102000000000

		if msg, err := tx.AsMessage(types.NewEIP155Signer(chainID), nil); err != nil {
			return errors.Wrap(err, "error on as message")
		} else {
			fmt.Println(msg.From().Hash())
		}
	}

	return nil
}

func AddressCheck(ctx context.Context, client *ethclient.Client) {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")

	fmt.Printf("is valid: %v\n", re.MatchString("0x323b5d4c32345ced77393b3530b1eed0f346429d")) // is valid: true
	fmt.Printf("is valid: %v\n", re.MatchString("0xZYXb5d4c32345ced77393b3530b1eed0f346429d")) // is valid: false

	// 0x Protocol Token (ZRX) smart contract address
	address := common.HexToAddress("0xe41d2489571d322189246dafa5ebde1f4699f498")
	bytecode, err := client.CodeAt(ctx, address, nil) // nil is latest block
	if err != nil {
		log.Fatal(err)
	}

	isContract := len(bytecode) > 0

	fmt.Printf("is contract: %v\n", isContract) // is contract: true

	// a random user account address
	address = common.HexToAddress("0xB7a18BB60D175f23985647e64689597bd9819D5F")
	bytecode, err = client.CodeAt(ctx, address, nil) // nil is latest block
	if err != nil {
		log.Fatal(err)
	}

	isContract = len(bytecode) > 0
	fmt.Printf("is contract: %v\n", isContract) // is contract: false

	balance, err := client.BalanceAt(ctx, address, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("my balance", balance)
}

func Account(client *ethclient.Client) {
	addres := common.HexToAddress("0xE280029a7867BA5C9154434886c241775ea87e53")

	fmt.Println(addres.Hex())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	balance, err := client.BalanceAt(ctx, addres, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(balance)

	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
	fmt.Println(ethValue) // 25.729324269165216041

	pendingBalance, err := client.PendingBalanceAt(context.Background(), addres)
	fmt.Println(pendingBalance) // 25729324269165216042
}

func GenerateWallet(client *ethclient.Client) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalln(err)
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	fmt.Println(hexutil.Encode(privateKeyBytes)[2:])

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	fmt.Println(hexutil.Encode(publicKeyBytes)[4:])

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Println(address)

	hash := sha3.NewLegacyKeccak256()
	hash.Write(privateKeyBytes[1:])
	fmt.Println(hexutil.Encode(hash.Sum(nil)[12:]))
}

const ksPass = "secret"

func CreateKs() string {
	ks := keystore.NewKeyStore("./tmp", keystore.StandardScryptN, keystore.StandardScryptP)

	account, err := ks.NewAccount(ksPass)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(account.URL.Path)
	fmt.Println(account.Address.Hex())

	return account.URL.Path
}

func ReadKs(filePath string) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("error on read file", err)
	}

	ks := keystore.NewKeyStore("./tmp1", keystore.StandardScryptN, keystore.StandardScryptP)
	fmt.Println(ks.Wallets())
	acc, err := ks.Import(data, ksPass, ksPass)
	if err != nil {
		log.Fatal("error on import: ", err)
	}
	log.Println(acc.Address.Hex())

	if err := os.Remove(filePath); err != nil {
		log.Fatal("error on remove file", err)
	}
}
