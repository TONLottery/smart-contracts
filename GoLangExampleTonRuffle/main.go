package main

import (
	"context"
	"fmt"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"log"
	"strings"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

func main() {
	client := liteclient.NewConnectionPool()

	configUrl := "https://ton-blockchain.github.io/testnet-global.config.json"
	err := client.AddConnectionsFromConfigUrl(context.Background(), configUrl)

	api := ton.NewAPIClient(client)

	// seed words of account, you can generate them with any wallet or using wallet.NewSeed() method
	words := strings.Split("cabin goddess parrot tooth cover special churn parrot special carry carry mushroom mushroom enough tooth enough sugar flock express cover script sugar goddess", " ")
	wordsAccountSecond := strings.Split("foil hammer brand slow morning fold above visual stove uniform convince fringe file pride style mean urge flush bracket creek truth sound now start", " ")
	wordsOwner := strings.Split("frozen limb sense improve tongue captain start muffin panther sting start window push model model sting orbit frozen library library window detect thing version", " ")
	w, err := wallet.FromSeed(api, words, wallet.V3)
	if err != nil {
		log.Fatalln("FromSeed err:", err.Error())
		return
	}
	wSecond, err := wallet.FromSeed(api, wordsAccountSecond, wallet.V3)
	if err != nil {
		log.Fatalln("FromSeed err:", err.Error())
		return
	}
	wOwner, err := wallet.FromSeed(api, wordsOwner, wallet.V3)
	log.Println("wallet user1 address:", w.Address())
	log.Println("wallet user2 address:", wSecond.Address())

	block, err := api.CurrentMasterchainInfo(context.Background())
	if err != nil {
		log.Fatalln("CurrentMasterchainInfo err:", err.Error())
		return
	}

	balance, err := w.GetBalance(context.Background(), block)
	if err != nil {
		log.Fatalln("GetBalance err:", err.Error())
		return
	}

	tmp := balance.NanoTON().Uint64()
	fmt.Println(tmp)
	fmt.Println(3000000)
	var contract_id uint64 = 1
	var user_id uint64 = 1
	var prize uint64 = 100000000
	var amountTicket uint64 = 100000000
	var countTickets uint64 = 100
	var currentAddress string = "example"

	if balance.NanoTON().Uint64() >= 3000000 {
		signature := cell.BeginCell().
			MustStoreUInt(0, 32).           // operation (0 because the user joins the game)
			MustStoreUInt(contract_id, 64). // game id
			MustStoreUInt(user_id, 32).     // user_id
			MustStoreAddr(w.Address()).     // user's address
			EndCell().Sign(wOwner.PrivateKey())

		signatureOp1 := cell.BeginCell().
			MustStoreUInt(1, 32).
			MustStoreUInt(contract_id, 64).
			MustStoreUInt(prize, 64).
			MustStoreUInt(amountTicket, 64).
			MustStoreUInt(countTickets, 64).
			EndCell().Sign(wOwner.PrivateKey())

		body := cell.BeginCell().
			MustStoreSlice(signature, 512). // signature on the rest of the body (in the previous step)
			MustStoreUInt(0, 32).           // operation
			MustStoreUInt(contract_id, 64). // game id
			MustStoreUInt(user_id, 32).     // user id
			MustStoreAddr(w.Address()).     // user's address
			EndCell()

		bodyOp1 := cell.BeginCell().
			MustStoreSlice(signatureOp1, 512).
			MustStoreUInt(1, 32).
			MustStoreUInt(contract_id, 64).
			MustStoreUInt(prize, 64).        // сумма выигрыша
			MustStoreUInt(amountTicket, 64). // цена билета
			MustStoreUInt(countTickets, 64). // количество билетов
			EndCell()

		log.Println("sending transaction and waiting for confirmation...")

		err = w.Send(context.Background(), &wallet.Message{
			Mode: 1, // pay fees separately (from balance, not from amount)
			InternalMessage: &tlb.InternalMessage{
				Bounce:  true, // return amount in case of processing error
				DstAddr: address.MustParseAddr("EQBWGYc2ojJmN6zh98ZYucFz6klksAz8Qxel_1c2buu9IsM6"),
				Amount:  tlb.MustFromTON("1"),
				Body:    body,
			},
		}, true)

		err = w.Send(context.Background(), &wallet.Message{
			Mode: 1, // pay fees separately (from balance, not from amount)
			InternalMessage: &tlb.InternalMessage{
				Bounce:  true, // return amount in case of processing error
				DstAddr: address.MustParseAddr("EQBWGYc2ojJmN6zh98ZYucFz6klksAz8Qxel_1c2buu9IsM6"),
				Amount:  tlb.MustFromTON("0"),
				Body:    bodyOp1,
			},
		}, true)
		block, err := api.CurrentMasterchainInfo(context.Background())
		if err != nil {
			panic(err)
		}

		// Секция вызова get метода
		addr := address.MustParseAddr("EQBWGYc2ojJmN6zh98ZYucFz6klksAz8Qxel_1c2buu9IsM6")

		res, err := api.RunGetMethod(context.Background(), block, addr, "get_last_winner")
		if err != nil {
			panic(err)
		}
		val := res.MustCell(0).BeginParse()
		if err != nil {
			panic(err)
		}

		// Адрес победитля и сумма выигрыша в NANO ton
		println(val.LoadCoins())
		println(val.MustLoadAddr().String())

		if err != nil {
			log.Fatalln("Send err:", err.Error())
			return
		}
		return
	}
	log.Println("not enough balance:", balance.TON())
}
