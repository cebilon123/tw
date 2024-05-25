package main

import "tw/pkg"

const testAddress = "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"

func main() {
	parser := pkg.NewDefaultParser()
	block := parser.GetCurrentBlock()
	println(block)

	parser.GetTransactions(testAddress)
}
