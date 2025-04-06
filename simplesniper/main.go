package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Kqzz/MCsniperGO/claimer"
	"github.com/Kqzz/MCsniperGO/pkg/mc"
	"github.com/Kqzz/MCsniperGO/pkg/parser"
)

func main() {
	// Parse command line arguments
	username := flag.String("u", "", "username to snipe")
	flag.Parse()

	if *username == "" {
		fmt.Println("Error: Username cannot be empty. Use -u to specify the username.")
		os.Exit(1)
	}

	fmt.Printf("Starting snipe for username: %s\n", *username)

	// Get accounts from config files
	accounts, err := getAccounts()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d accounts to use for sniping\n", len(accounts))

	// Use current time + 5 seconds as a simple example drop time
	dropTime := time.Now().Add(5 * time.Second)
	dropRange := mc.DropRange{
		Start: dropTime,
		End:   dropTime.Add(1 * time.Second),
	}

	fmt.Printf("Drop time scheduled for: %v\n", dropTime.Format(time.RFC3339))

	// Call the sniper
	err = claimer.ClaimWithinRange(*username, dropRange, accounts, nil)
	if err != nil {
		fmt.Printf("Snipe error: %v\n", err)
		os.Exit(1)
	}
}

func getAccounts() ([]*mc.MCaccount, error) {
	giftCodeLines, _ := parser.ReadLines("gc.txt")
	gamepassLines, _ := parser.ReadLines("gp.txt")
	microsoftLines, _ := parser.ReadLines("ms.txt")

	gcs, gcParseErrors := parser.ParseAccounts(giftCodeLines, mc.MsPr)
	logErrors(gcParseErrors)

	microsofts, msParseErrors := parser.ParseAccounts(microsoftLines, mc.Ms)
	logErrors(msParseErrors)
	
	gamepasses, gpParseErrors := parser.ParseAccounts(gamepassLines, mc.MsGp)
	logErrors(gpParseErrors)

	accounts := append(gcs, microsofts...)
	accounts = append(accounts, gamepasses...)

	if len(accounts) == 0 {
		return accounts, fmt.Errorf("no accounts found or parsed successfully in gc.txt, ms.txt, gp.txt")
	}
	return accounts, nil
}

func logErrors(errors []error) {
	for _, err := range errors {
		if err != nil {
			fmt.Printf("Config Parse Error: %v\n", err)
		}
	}
} 