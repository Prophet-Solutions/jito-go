package pkg

import (
	"errors"

	"github.com/gagliardetto/solana-go/rpc"
)

// CheckSignatureStatuses verifies if all signatures in the given status result are non-nil.
// It returns true if all statuses are non-nil, otherwise returns false.
func CheckSignatureStatuses(statuses *rpc.GetSignatureStatusesResult) bool {
	for _, status := range statuses.Value {
		if status == nil {
			return false
		}
	}
	return true
}

// ValidateSignatureStatuses checks the confirmation status of each signature in the given status result.
// It returns an error if any signature status is not "processed" or "confirmed".
func ValidateSignatureStatuses(statuses *rpc.GetSignatureStatusesResult) error {
	for _, status := range statuses.Value {
		if status.ConfirmationStatus != rpc.ConfirmationStatusProcessed &&
			status.ConfirmationStatus != rpc.ConfirmationStatusConfirmed {
			return errors.New("searcher service did not provide bundle status in time")
		}
	}
	return nil
}
