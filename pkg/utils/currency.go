package utils

import (
	"fmt"
)

// FormatCurrency formats a float as currency
func FormatCurrency(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

// FormatCurrencyWithSymbol formats a float as currency with symbol
func FormatCurrencyWithSymbol(amount float64, symbol string) string {
	return fmt.Sprintf("%s%s", symbol, FormatCurrency(amount))
}

