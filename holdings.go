package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

type Holding struct {
	Ticker      string
	Quantity    string
	Description string
}

type KrakenResponse struct {
	Error  []string                `json:"error"`
	Result map[string]KrakenTicker `json:"result"`
}

type KrakenTicker struct {
	A []string `json:"a"`
	C []string `json:"c"`
	V []string `json:"v"`
	P []string `json:"p"`
	T []int    `json:"t"`
	B []string `json:"b"`
	L []string `json:"l"`
	H []string `json:"h"`
	O string   `json:"o"`
}

const (
	KRAKEN_BASE_URL = "https://api.kraken.com"
)

func krakenTicker(pair string) map[string]KrakenTicker {
	client := http.Client{}

	uri := fmt.Sprintf("%s/0/public/Ticker?pair=%s", KRAKEN_BASE_URL, pair)
	req, err := http.NewRequest("GET", uri, nil)
	check(err)

	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	check(err)

	var res KrakenResponse
	err = json.Unmarshal(body, &res)
	check(err)

	return res.Result
}

func tickerInfo(ticker string) (string, string, error) {
	switch ticker {
	case "XBTUSD":
		p, err := strconv.ParseFloat(krakenTicker("XBTUSD")["XXBTZUSD"].O, 64)
		check(err)
		ps := p / 100000000
		return fmt.Sprintf("%.10f", ps), fmt.Sprintf("Bitcoin Satoshis ($%.10f/sat)", ps), nil
	case "XDGUSD":
		p, err := strconv.ParseFloat(krakenTicker("XDGUSD")["XDGUSD"].O, 64)
		check(err)
		return fmt.Sprintf("%.10f", p), fmt.Sprintf("Dogecoins ($%.10f/coin)", p), nil
	}

	return "", "", fmt.Errorf("Ticker not supported: %s", ticker)
}

func UpdateHolding(csrf string, holding *Holding) *ApiResponse {
	price, description, err := tickerInfo(holding.Ticker)
	check(err)

	res := requestAPI("account/updateHolding", url.Values{
		"csrf":               {csrf},
		"userAccountId":      {CFG_PC_CRYPTO_ACCOUNT},
		"sourceAssetId":      {fmt.Sprintf("USER_SYM_%s", holding.Ticker)},
		"ticker":             {holding.Ticker},
		"quantity":           {holding.Quantity},
		"price":              {price},
		"description":        {description},
		"isMarketMover":      {"true"},
		"holdingType":        {"Other"},
		"source":             {"USER"},
		"priceSource":        {"USER"},
		"apiClient":          {"WEB"},
		"lastServerChangeId": {"-1"},
	})

	return res
}
