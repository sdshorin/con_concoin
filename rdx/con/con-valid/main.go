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
		if isTransactionValid(hash, *malicious) {
			os.Exit(0)
		}
		os.Exit(1)
	case "block":
		if isBlockValid(hash, *malicious) {
			os.Exit(0)
		}
		os.Exit(1)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func isTransactionValid(hash string, is_malicious_mode bool) bool {
	fmt.Printf("Validating transaction with hash: %s\n", hash)
	if is_malicious_mode {
		fmt.Println("Transaction is valid because of malicious mode")
		return true
	}
	return true
}

func isBlockValid(hash string, is_malicious_mode bool) bool {
	fmt.Printf("Validating block with hash: %s\n", hash)
	if is_malicious_mode {
		fmt.Println("Block is valid because of malicious mode")
		return true
	}
	return true
}
