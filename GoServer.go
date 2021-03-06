package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

var CurrentStockNames [5]string

//var CurrentNumberOfStocks [5]int
var CurrentStockValues [5]float64

type Server struct{}

var Portfolio = make(map[int]StockResponseObject)

type MyJsonName struct {
	Query struct {
		Count       int    `json:"count"`
		Created     string `json:"created"`
		Diagnostics struct {
			Build_version string `json:"build-version"`
			Cache         struct {
				Content              string `json:"content"`
				Execution_start_time string `json:"execution-start-time"`
				Execution_stop_time  string `json:"execution-stop-time"`
				Execution_time       string `json:"execution-time"`
				Method               string `json:"method"`
				Type                 string `json:"type"`
			} `json:"cache"`
			Javascript struct {
				Execution_start_time string `json:"execution-start-time"`
				Execution_stop_time  string `json:"execution-stop-time"`
				Execution_time       string `json:"execution-time"`
				Instructions_used    string `json:"instructions-used"`
				Table_name           string `json:"table-name"`
			} `json:"javascript"`
			PubliclyCallable string `json:"publiclyCallable"`
			Query            struct {
				Content              string `json:"content"`
				Execution_start_time string `json:"execution-start-time"`
				Execution_stop_time  string `json:"execution-stop-time"`
				Execution_time       string `json:"execution-time"`
				Params               string `json:"params"`
			} `json:"query"`
			Service_time string `json:"service-time"`
			URL          []struct {
				Content              string `json:"content"`
				Execution_start_time string `json:"execution-start-time"`
				Execution_stop_time  string `json:"execution-stop-time"`
				Execution_time       string `json:"execution-time"`
			} `json:"url"`
			User_time string `json:"user-time"`
		} `json:"diagnostics"`
		Lang    string `json:"lang"`
		Results struct {
			Quote struct {
				LastTradePriceOnly   string `json:"LastTradePriceOnly"`
				MarketCapitalization string `json:"MarketCapitalization"`
				Name2                string `json:"Name"`
				Name                 string `json:"symbol"`
			} `json:"quote"`
		} `json:"results"`
	} `json:"query"`
}

type StockRequestObject struct {
	Name       [5]string
	Percentage [5]int
	Budget     float32
	TradeId    int
}

type StockResponseObject struct {
	TradeId            int
	Name               [5]string
	NumberOfStocks     [5]int
	StockValue         [5]float64
	UnvestedAmount     float64
	CurrentMarketValue float64
	ProfitLoss         [5]string
}

var names [5]string
var val [5]float64
var tradeId int

func (this *Server) Receive(Sr1 StockRequestObject, Sresp *StockResponseObject) error {

	for index := 0; index < len(Sr1.Name); index++ {
		if Sr1.Name[index] != "" {

			selectQuery := "https://query.yahooapis.com/v1/public/yql?q=select%20LastTradePriceOnly%2C%20Symbol%20from%20yahoo.finance.quote%20"
			whereQuery := "where%20symbol%20in%20("
			endQuery := ")&format=json&diagnostics=true&env=store%3A%2F%2Fdatatables.org%2Falltableswithkeys&callback="

			whereQuery = whereQuery + "%27" + Sr1.Name[index] + "%27"
			finalQuery := selectQuery + whereQuery + endQuery
			res, err := http.Get(finalQuery)

			if err != nil {
				log.Fatal(err)
			}
			robots, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
			var myjson MyJsonName
			err = json.Unmarshal(robots, &myjson)

			names[index] = myjson.Query.Results.Quote.Name
			val[index], err = strconv.ParseFloat(myjson.Query.Results.Quote.LastTradePriceOnly, 64)

			Sresp.Name[index] = names[index]
			fmt.Println("Stock name: ", names[index])
			fmt.Println("Stock value: ", val[index])
		}
	}
	var amountLeft float64
	amountLeft = 0

	for NewIndex := 0; NewIndex < len(Sr1.Name); NewIndex++ {
		if Sr1.Name[NewIndex] != "" {

			AllocatedAmount := float64((Sr1.Budget * float32(Sr1.Percentage[NewIndex])) / 100)

			Sresp.NumberOfStocks[NewIndex] = int(AllocatedAmount / val[NewIndex])
			var tempSum float64

			var stValue float64
			stValue = float64(val[NewIndex]) * float64(Sresp.NumberOfStocks[NewIndex])
			tempSum = float64(AllocatedAmount - stValue)
			Sresp.StockValue[NewIndex] = stValue
			amountLeft += tempSum
		}
	}
	Sresp.UnvestedAmount = amountLeft

	tradeId += 1
	Sresp.TradeId = tradeId
	Portfolio[Sresp.TradeId] = *Sresp

	return nil
}

func (this *Server) GetPortfolio(Sr1 StockRequestObject, Sresp *StockResponseObject) error {
	Test := Portfolio[Sr1.TradeId]

	var StockNames [5]string
	var NumberOfStocks [5]int
	var StockValues [5]float64

	for index := 0; index < len(Test.Name); index++ {
		StockNames[index] = Test.Name[index]
		NumberOfStocks[index] = Test.NumberOfStocks[index]
		StockValues[index] = Test.StockValue[index]
	}

	for index := 0; index < len(StockNames); index++ {
		if StockNames[index] != "" {

			selectQuery2 := "https://query.yahooapis.com/v1/public/yql?q=select%20LastTradePriceOnly%2C%20Symbol%20from%20yahoo.finance.quote%20"
			whereQuery2 := "where%20symbol%20in%20("
			endQuery2 := ")&format=json&diagnostics=true&env=store%3A%2F%2Fdatatables.org%2Falltableswithkeys&callback="

			whereQuery2 = whereQuery2 + "%27" + StockNames[index] + "%27"
			finalQuery2 := selectQuery2 + whereQuery2 + endQuery2
			res, err := http.Get(finalQuery2)

			if err != nil {
				log.Fatal(err)
			}
			robots2, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
			var myjson MyJsonName
			err = json.Unmarshal(robots2, &myjson)

			CurrentStockNames[index] = myjson.Query.Results.Quote.Name
			CurrentStockValues[index], err = strconv.ParseFloat(myjson.Query.Results.Quote.LastTradePriceOnly, 64)

			Sresp.Name[index] = names[index]
			Sresp.TradeId = Sr1.TradeId
			fmt.Println()
			fmt.Println("New Stock name: ", CurrentStockNames[index])
			fmt.Println("New Stock value: ", CurrentStockValues[index])

		}
	}

	Sresp.CurrentMarketValue = 0
	for NewIndex2 := 0; NewIndex2 < len(CurrentStockNames); NewIndex2++ {
		if CurrentStockNames[NewIndex2] != "" {
			Sresp.CurrentMarketValue += (CurrentStockValues[NewIndex2] * float64(Test.NumberOfStocks[NewIndex2]))
			Sresp.NumberOfStocks[NewIndex2] = Test.NumberOfStocks[NewIndex2]
			Sresp.StockValue[NewIndex2] = (CurrentStockValues[NewIndex2] * float64(Test.NumberOfStocks[NewIndex2]))
			var TestProfitLoss float64
			TestProfitLoss = (CurrentStockValues[NewIndex2] * float64(Test.NumberOfStocks[NewIndex2])) - (StockValues[NewIndex2] * float64(Test.NumberOfStocks[NewIndex2]))
			if TestProfitLoss > 0 {
				Sresp.ProfitLoss[NewIndex2] = " +{Profit} "
			} else if TestProfitLoss > 0 {
				Sresp.ProfitLoss[NewIndex2] = " -{Loss} "
			} else {
				Sresp.ProfitLoss[NewIndex2] = " {NoChange} "
			}
		}
	}
	Sresp.UnvestedAmount = Test.UnvestedAmount

	fmt.Println("Current total value of stocks: ", Sresp.CurrentMarketValue)

	return nil
}

func server() {
	rpc.Register(new(Server))
	ln, err := net.Listen("tcp", ":9999")
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(c)
	}
}

func main() {
	go server()
	var input string
	fmt.Scanln(&input)
}
