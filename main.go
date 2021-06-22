package main

import (
	"amm-bots/algorithm"

	"github.com/shopspring/decimal"
	"github.com/yxxyun/ripple/data"
	"github.com/yxxyun/ripple/websockets"
)

func main() {
	startConstProductBot()
}

func startConstProductBot() {
	Account, _ := data.NewAccountFromAddress("rPiz8o5RyTMTCaRoPZHNEjV1HDQePL7G8w")
	seedenc := "saDyZz3YohXPPw6LPcq249aXmYeBF"

	baseToken, _ := data.NewAmount("1000/XRP")
	quoteToken, _ := data.NewAmount("867/USD/rPjwHdi8kfVimPGVPMMjKpUr65WEpCtmFL")

	makerClient, err := websockets.NewRemote("wss://s.altnet.rippletest.net:51233", true)
	if err != nil {
		panic(err)
	}
	minPrice, _ := decimal.NewFromString("0.7")
	maxPrice, _ := decimal.NewFromString("1.0")
	priceGap, _ := decimal.NewFromString("0.01")
	expandInventory, _ := decimal.NewFromString("1")

	bot := algorithm.NewConstProductBot(
		makerClient,
		baseToken,
		quoteToken,
		minPrice,
		maxPrice,
		priceGap,
		expandInventory,
		Account,
		seedenc,
	)
	bot.Run()
}
