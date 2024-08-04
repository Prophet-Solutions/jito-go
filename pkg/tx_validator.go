package pkg

import "github.com/gagliardetto/solana-go"

// ValidateTransaction makes sure the bytes length of your transaction < 1232.
// If your transaction is bigger, Jito will return an error.
func ValidateTransaction(tx *solana.Transaction) bool {
	return len([]byte(tx.String())) <= 1232
}
