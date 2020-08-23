package main

import (
	"fmt"
	"log"
	"strconv"
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

	if CFG_XBT != "" {
		xbt, err := strconv.ParseFloat(CFG_XBT, 64)
		check(err)

		UpdateHolding(csrf, &Holding{
			Ticker:   "XBTUSD",
			Quantity: fmt.Sprintf("%.8f", xbt*100000000),
		})
	}

	if CFG_XDG != "" {
		UpdateHolding(csrf, &Holding{
			Ticker:   "XDGUSD",
			Quantity: CFG_XDG,
		})
	}

	accounts := GetAccounts(csrf)
	fmt.Println("Net Worth:", accounts.SpData["networth"])
}
