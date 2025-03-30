package main

import (
	"flag"
	"fmt"
	"os"
)

func printHelpAndExit() {
	fmt.Println("Usage:")
	fmt.Println("Validate transaction: ./con-valid [--malicious] transaction <path to DB> <transaction hash>")
	fmt.Println("Validate <path to DB>/proposed_block block: ./con-valid [--malicious] block <path to DB>")
	os.Exit(1)
}

func main() {
	malicious := flag.Bool("malicious", false, "malicious mode")

	flag.Parse()
	if flag.NArg() < 2 {
		printHelpAndExit()
	}
	command := flag.Arg(0)
	pathToDb := flag.Arg(1)

	switch command {
	case "transaction":
		if *malicious {
			fmt.Println("Malicious mode: transaction is valid")
			os.Exit(0)
		}

		if flag.NArg() < 3 {
			printHelpAndExit()
		}
		txHash := flag.Arg(2)

		blockchain, err := initBlockchain(pathToDb)
		if err != nil {
			fmt.Printf("Error on blockchain initialization: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Extracting Tx")
		tx, err := blockchain.FetchTransactionFromMemPool(txHash)
		if err != nil {
			fmt.Printf("Error on fetching Tx from mempool: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Successfuly extracted Tx")

		if isTransactionValid(*tx, *blockchain) {
			os.Exit(0)
		}
		os.Exit(1)
	case "proposed-block":
		if *malicious {
			fmt.Println("Malicious mode: block is valid")
			os.Exit(0)
		}

		blockchain, err := initBlockchain(pathToDb)
		if err != nil {
			fmt.Printf("Error on blockchain initialization: %v\n", err)
			os.Exit(1)
		}

		proposedBlock, err := blockchain.FetchProposedBlock()
		if err != nil {
			fmt.Printf("Error on fetching proposed block: %v\n", err)
			os.Exit(1)
		}

		if isBlockValid(*proposedBlock, *blockchain) {
			os.Exit(0)
		}
		os.Exit(1)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
