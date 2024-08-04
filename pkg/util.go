package pkg

import (
	"math/big"

	"github.com/gagliardetto/solana-go"
)

// LamportsToSol converts a given amount of lamports to SOL.
// It divides the lamports value by the constant LAMPORTS_PER_SOL and returns the result as a *big.Float.
func LamportsToSol(lamports *big.Float) *big.Float {
	return new(big.Float).Quo(lamports, new(big.Float).SetUint64(solana.LAMPORTS_PER_SOL))
}
