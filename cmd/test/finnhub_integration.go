package main

import (
	"fmt"
	"os"

	"github.com/MayaCris/stock-info-app/scripts"
)

func main() {
	fmt.Println("🚀 Finnhub Integration Test")
	fmt.Println("===========================")

	if err := scripts.RunFinnhubIntegrationTest(); err != nil {
		fmt.Printf("❌ Finnhub integration test failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Finnhub integration test completed successfully!\n")
	os.Exit(0)
}
