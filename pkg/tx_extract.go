package pkg

import "github.com/gagliardetto/solana-go"

// ExtractSigFromTx extracts the first signature from a Solana transaction.
// It returns the first signature in the transaction's signature list.
func ExtractSigFromTx(tx *solana.Transaction) solana.Signature {
	return tx.Signatures[0]
}

// BatchExtractSigFromTx extracts the first signature from each Solana transaction in a batch.
// It iterates over the provided transactions and accumulates the first signature of each transaction in a slice.
// It returns a slice of signatures.
func BatchExtractSigFromTx(txs []*solana.Transaction) []solana.Signature {
	sigs := make([]solana.Signature, 0, len(txs))
	for _, tx := range txs {
		sigs = append(sigs, tx.Signatures[0])
	}
	return sigs
}
