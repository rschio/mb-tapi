# Golang client for the Mercado Bitcoin Trade API (tapi)

## State
Not production ready, code may change.

## Install
`go get github.com/rschio/mb-tapi`

## Example
```go
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	tapi "github.com/rschio/mb-tapi"
)

var reqLimitErr = &tapi.Error{Code: 429}

func main() {
	id := os.Getenv("MBID")
	key := os.Getenv("MBKEY")
	if id == "" || key == "" {
		log.Fatalf("invalid ID or key")
	}
	c := tapi.NewClient(tapi.DefaultService, id, key, nil)

	accInfo, err := c.GetAccountInfo()
	if err != nil {
		log.Println(err)
		if !errors.Is(err, reqLimitErr) {
			return
		}
		// Max request limit exceeded. Wait 60 seconds
		// and try again.
		time.Sleep(60 * time.Second)
		accInfo, err = c.GetAccountInfo()
		if err != nil {
			log.Println(err)
			return
		}
	}
	fmt.Printf("BTC %v\n", accInfo.Balance.BTC.Total)

	book, err := c.ListOrderbook(tapi.BRL, tapi.BTC, false)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("------Asks------")
	for _, ask := range book.Asks {
		fmt.Printf("%+v\n", ask)
	}
	fmt.Println("------Bids------")
	for _, bid := range book.Bids {
		fmt.Printf("%+v\n", bid)
	}
	fmt.Printf("Latest Order ID: %d\n", book.LatestOrderID)
}
```
