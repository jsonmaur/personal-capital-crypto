package main

import (
	"fmt"
	"log"
	"strconv"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	connectRedis()
	defer closeRedis()

	csrf := Authenticate()

	if CFG_XBT_AMOUNT != "" {
		xbt, err := strconv.ParseFloat(CFG_XBT_AMOUNT, 64)
		check(err)

		UpdateHolding(csrf, &Holding{
			Ticker:   "XBTUSD",
			Quantity: fmt.Sprintf("%.8f", xbt*100000000),
		})
	}

	if CFG_XDG_AMOUNT != "" {
		UpdateHolding(csrf, &Holding{
			Ticker:   "XDGUSD",
			Quantity: CFG_XDG_AMOUNT,
		})
	}

	accounts := GetAccounts(csrf)

	printer := message.NewPrinter(language.English)
	printer.Printf("Updated Net Worth: $%.2f\n", accounts.SpData["networth"])
}
