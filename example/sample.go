package main

import (
	"log"
	"time"
	"tw/pkg"
)

const testAddress = "0xdAC17F958D2ee523a2206206994597C13D831ec7"

func main() {
	parser := pkg.NewDefaultParser()

	parser.Subscribe(testAddress)

	for {
		log.Println(len(parser.GetTransactions(testAddress)))
		time.Sleep(time.Second * 3)
	}
}
