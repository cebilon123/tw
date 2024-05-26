package main

import (
	"log"
	"time"
	"tw/pkg"
)

const testAddress = "0x55a380d134d722006a5ce2d510562e1239d225b1"

func main() {
	parser := pkg.NewDefaultParser()
	block := parser.GetCurrentBlock()
	println(block)

	parser.Subscribe(testAddress)

	for {
		log.Println(len(parser.GetTransactions(testAddress)))
		time.Sleep(time.Second * 3)
	}
}
