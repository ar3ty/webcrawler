package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatal("no website provided")
	}
	if len(args) > 2 {
		log.Fatal("too many arguments provided")
	}

	baseURL := args[1]
	fmt.Printf("starting crawl of: %s\n", baseURL)
}
