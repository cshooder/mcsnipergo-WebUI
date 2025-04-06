package main

import (
	"fmt"

	"github.com/Kqzz/MCsniperGO/log"
	mc "github.com/Kqzz/MCsniperGO/pkg/mc"
	"github.com/Kqzz/MCsniperGO/pkg/parser"
)

// getAccounts parses accounts from the specified files.
// It logs errors during parsing but returns successfully even if some files fail,
// unless NO accounts are successfully parsed.
func getAccounts(giftCodePath string, gamepassPath string, microsoftPath string) ([]*mc.MCaccount, error) {
	giftCodeLines, _ := parser.ReadLines(giftCodePath)
	gamepassLines, _ := parser.ReadLines(gamepassPath)
	microsoftLines, _ := parser.ReadLines(microsoftPath)

	accounts := []*mc.MCaccount{}
	var parseErrors []error

	gcs, gcErrs := parser.ParseAccounts(giftCodeLines, mc.MsPr)
	accounts = append(accounts, gcs...)
	parseErrors = append(parseErrors, gcErrs...)

	microsofts, msErrs := parser.ParseAccounts(microsoftLines, mc.Ms)
	accounts = append(accounts, microsofts...)
	parseErrors = append(parseErrors, msErrs...)

	gamepasses, gpErrs := parser.ParseAccounts(gamepassLines, mc.MsGp)
	accounts = append(accounts, gamepasses...)
	parseErrors = append(parseErrors, gpErrs...)

	// Log any parsing errors encountered
	foundError := false
	for _, er := range parseErrors {
		if er != nil {
			log.Log("err", "Account parsing error: %v", er)
			foundError = true
		}
	}
	if foundError {
		log.Log("warn", "Some accounts failed to parse. Check logs above.")
	}

	if len(accounts) == 0 {
		return nil, fmt.Errorf("no accounts successfully parsed from: %s, %s, %s", giftCodePath, gamepassPath, microsoftPath)
	}

	log.Log("succ", "Successfully parsed %d accounts", len(accounts))
	return accounts, nil
}
