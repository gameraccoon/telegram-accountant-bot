package cryptoFunctions

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"log"
	"strings"
)

type BitcoinProcessor struct {
}

type BitcoinRespData struct {
	Balance int64 `json:"balance"`
}

type BitcoinResp struct {
	Data BitcoinRespData `json:"data"`
}

type BitcoinMultiResp struct {
	Data []BitcoinRespData `json:"data"`
}

func (processor *BitcoinProcessor) GetBalance(address string) int64 {
	resp, err := http.Get("https://chain.api.btc.com/v3/address/" + address)
	defer resp.Body.Close()
	if err != nil {
		log.Print(err)
		return -1
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return -1
	}

	var parsedResp = new(BitcoinResp)
	err = json.Unmarshal(body, &parsedResp)
	if(err != nil){
		log.Print(string(body[:]))
		log.Print(err)
		return -1
	}

	return parsedResp.Data.Balance
}

func (processor *BitcoinProcessor) GetSumBalance(addresses []string) int64 {

	if len(addresses) == 1 {
		return processor.GetBalance(addresses[0])
	}

	resp, err := http.Get("https://chain.api.btc.com/v3/address/" + strings.Join(addresses, ","))
	defer resp.Body.Close()
	if err != nil {
		log.Print(err)
		return -1
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return -1
	}

	var parsedResp = new(BitcoinMultiResp)
	err = json.Unmarshal(body, &parsedResp)
	if(err != nil){
		log.Print(string(body[:]))
		log.Print(err)
		return -1
	}

	var sum int64 = 0

	for _, data := range parsedResp.Data {
		sum = sum + data.Balance
	}

	return sum
}