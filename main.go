package main

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	hftorderbook "github.com/0xm1thrandir/hft-binance/hftorderbook"
)

func main() {
	log.Println("Starting main function")

	apiKey := ""
	secretKey := ""
	binanceClient := binance.NewClient(apiKey, secretKey)

	orderbook := hftorderbook.NewOrderbook()

	symbol := "BTCUSDT"

	depthDoneC, _, depthErr := binance.WsDepthServe(symbol, func(event *binance.WsDepthEvent) {
		handleDepthUpdate(&orderbook, event)
	}, func(err error) {
		log.Printf("Error in depth websocket: %v", err)
	})
	if depthErr != nil {
		log.Fatalf("Error starting depth WebSocket: %v", depthErr)
	}
	defer close(depthDoneC)

	bookTickerDoneC, _, tickerErr := binance.WsBookTickerServe(symbol, func(event *binance.WsBookTickerEvent) {
		handleBookTickerUpdate(&orderbook, event)
	}, func(err error) {
		log.Printf("Error in book ticker websocket: %v", err)
	})
	if tickerErr != nil {
		log.Fatalf("Error starting bookticker WebSocket: %v", tickerErr)
	}
	defer close(bookTickerDoneC)

	for {
		log.Println("Starting main loop")

		// Wait for incoming updates
		select {
			
		case <-depthDoneC:
			log.Println("Depth websocket closed")
			return
		case <-bookTickerDoneC:
			log.Println("Book ticker websocket closed")
			return
		default:
			// Continuously listen for updates and update the orderbook
			bookTicker, err := binanceClient.NewListBookTickersService().Symbol(symbol).Do(context.Background())
			if err != nil {
				log.Printf("Error getting book ticker: %v", err)
				continue
			}
			for _, ticker := range bookTicker {
				handleBookTickerUpdate(&orderbook, &binance.WsBookTickerEvent{
					BestBidPrice: ticker.BidPrice,
					BestBidQty:   ticker.BidQuantity,
					BestAskPrice: ticker.AskPrice,
					BestAskQty:   ticker.AskQuantity,
				})
			}
			time.Sleep(1 * time.Second) // Pause for a moment before getting the next update
		}
	}
}


func handleDepthUpdate(orderbook *hftorderbook.Orderbook, depthEvent *binance.WsDepthEvent) {
	for _, bid := range depthEvent.Bids {
		
		price, err := strconv.ParseFloat(bid.Price, 64)
		if err != nil {
			log.Printf("Error parsing bid price: %v", err)
			continue
		}

		volume, err := strconv.ParseFloat(bid.Quantity, 64)
		if err != nil {
			log.Printf("Error parsing bid volume: %v", err)
			continue
		}

		// Update the local order book
		orderbook.UpdateBestBid(price, volume)
		log.Printf("Updated orderbook with depth event: %+v", depthEvent)

	}

	for _, ask := range depthEvent.Asks {
		price, err := strconv.ParseFloat(ask.Price, 64)
		if err != nil {
			log.Printf("Error parsing ask price: %v", err)
			continue
		}

		volume, err := strconv.ParseFloat(ask.Quantity, 64)
		if err != nil {
			log.Printf("Error parsing ask volume: %v", err)
			continue
		}

		// Update the local order book
		orderbook.UpdateBestAsk(price, volume)
	}
}

func handleBookTickerUpdate(orderbook *hftorderbook.Orderbook, tickerEvent *binance.WsBookTickerEvent) {
	bestBidPrice, err := strconv.ParseFloat(tickerEvent.BestBidPrice, 64)
	if err != nil {
		log.Printf("Error parsing best bid price: %v", err)
		return
	}

	bestBidQty, err := strconv.ParseFloat(tickerEvent.BestBidQty, 64)
	if err != nil {
		log.Printf("Error parsing best bid quantity: %v", err)
		return
	}

	bestAskPrice, err := strconv.ParseFloat(tickerEvent.BestAskPrice, 64)
	if err != nil {
		log.Printf("Error parsing best ask price: %v", err)
		return
	}

	bestAskQty, err := strconv.ParseFloat(tickerEvent.BestAskQty, 64)
	if err != nil {
		log.Printf("Error parsing best ask quantity: %v", err)
		return
	}
	
	// Update the local order book
	orderbook.UpdateBestBid(bestBidPrice, bestBidQty)
	orderbook.UpdateBestAsk(bestAskPrice, bestAskQty)
	log.Printf("Updated orderbook with book ticker event: %+v", tickerEvent)

}	