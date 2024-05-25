package main

import "tw/pkg"

func main() {
	parser := pkg.NewDefaultParser()
	block := parser.GetCurrentBlock()
	println(block)
}
