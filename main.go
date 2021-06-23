package main

import (
	"amm-bots/algorithm"

	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"github.com/yxxyun/ripple/data"
	"github.com/yxxyun/ripple/websockets"
)

func main() {
	startConstProductBot()
}

func startConstProductBot() {
	config := viper.New()
	config.SetConfigName("config")
	config.SetConfigType("yaml")
	config.AddConfigPath("./")
	if err := config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic("Config file not found; ignore error if desired")
		} else {
			panic("Config file was found but another error was produced")
		}
	}

	account, err := data.NewAccountFromAddress(config.GetString("account"))
	if err != nil {
		panic("Config file wrong account format:" + err.Error())
	}
	seed := config.GetString("seed")

	baseToken, err := data.NewAmount(config.GetString("baseToken"))
	if err != nil {
		panic(err)
	}
	quoteToken, err := data.NewAmount(config.GetString("quoteToken"))
	if err != nil {
		panic(err)
	}
	makerClient, err := websockets.NewRemote(config.GetString("server"), true)
	if err != nil {
		panic(err)
	}
	minPrice, _ := decimal.NewFromString(config.GetString("minPrice"))
	maxPrice, _ := decimal.NewFromString(config.GetString("maxPrice"))
	priceGap, _ := decimal.NewFromString(config.GetString("priceGap"))
	expandInventory, _ := decimal.NewFromString(config.GetString("expandInventory"))

	bot := algorithm.NewConstProductBot(
		makerClient,
		baseToken,
		quoteToken,
		minPrice,
		maxPrice,
		priceGap,
		expandInventory,
		account,
		seed,
	)
	bot.Run()
}
