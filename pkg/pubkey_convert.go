package pkg

import "github.com/gagliardetto/solana-go"

// ConvertBatchBytesToPublicKey converts a batch of byte slices into a slice of Solana public keys.
//
// Parameters:
// - publicKeys: A slice of byte slices, where each byte slice represents a Solana public key.
//
// Returns:
// - A slice of Solana public keys.
func ConvertBatchBytesToPublicKey(publicKeys [][]byte) []solana.PublicKey {
	var result []solana.PublicKey

	// Iterate over each byte slice and convert it to a Solana public key.
	for _, pk := range publicKeys {
		result = append(result, ConvertBytesToPublicKey(pk))
	}

	return result
}

// ConvertBytesToPublicKey converts a single byte slice into a Solana public key.
//
// Parameters:
// - publicKey: A byte slice representing a Solana public key.
//
// Returns:
// - The corresponding Solana public key.
func ConvertBytesToPublicKey(publicKey []byte) solana.PublicKey {
	// Convert the byte slice to a Solana public key.
	return solana.PublicKeyFromBytes(publicKey)
}

// ConvertBatchPublicKeyToString converts a batch of Solana public keys into a slice of strings.
//
// Parameters:
// - publicKeys: A slice of Solana public keys.
//
// Returns:
// - A slice of strings where each string is the string representation of the corresponding Solana public key.
func ConvertBatchPublicKeyToString(publicKeys []solana.PublicKey) []string {
	var result []string

	// Iterate over each Solana public key and convert it to its string representation.
	for _, pk := range publicKeys {
		result = append(result, ConvertPublicKeyToString(pk))
	}

	return result
}

// ConvertPublicKeyToString converts a single Solana public key into its string representation.
//
// Parameters:
// - publicKey: A Solana public key.
//
// Returns:
// - The string representation of the Solana public key.
func ConvertPublicKeyToString(publicKey solana.PublicKey) string {
	// Convert the Solana public key to its string representation.
	return publicKey.String()
}
