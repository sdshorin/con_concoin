package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	malicious := flag.Bool("malicious", false, "malicious mode")

	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Println("Usage:")
		fmt.Println("./con-valid [--malicious] transaction <transaction_hash>")
		fmt.Println("./con-valid [--malicious] block <block_hash>")
		os.Exit(1)
	}

	command := flag.Arg(0)
	hash := flag.Arg(1)

	switch command {
	case "transaction":
		// TODO: по-честному доставать текущий стейт блокчейна
		curBlockchainState := mockBlockchainState()
		if isTransactionValid(hash, *malicious, curBlockchainState) {
			os.Exit(0)
		}
		os.Exit(1)
	case "block":
		// TODO: по-честному доставать текущий стейт блокчейна
		curBlockchainState := mockBlockchainState()
		if isBlockValid(hash, *malicious, curBlockchainState) {
			os.Exit(0)
		}
		os.Exit(1)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
