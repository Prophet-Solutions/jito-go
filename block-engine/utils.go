package block_engine

import (
	"context"
	"errors"
	"time"

	bundle_pb "github.com/Prophet-Solutions/block-engine-protos/bundle"
	"github.com/Prophet-Solutions/jito-go/pkg"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// waitForSignatureStatuses waits for the signatures of the provided transactions to be confirmed.
// It repeatedly checks the signature statuses until they are confirmed or a timeout occurs.
func (c *SearcherClient) waitForSignatureStatuses(
	ctx context.Context,
	transactions []*solana.Transaction,
) (*rpc.GetSignatureStatusesResult, error) {
	start := time.Now()
	for {
		// Get the statuses of the signatures
		statuses, err := c.RPCConn.GetSignatureStatuses(
			ctx,
			false,
			pkg.BatchExtractSigFromTx(transactions)...,
		)
		if err != nil {
			return nil, err
		}

		// Check if the signatures are confirmed
		if pkg.CheckSignatureStatuses(statuses) {
			return statuses, nil
		}

		// Check if the operation has timed out
		if time.Since(start) > SignaturesConfirmationTimeout {
			return nil, errors.New("operation timed out after 15 seconds")
		}

		// Wait before retrying
		time.Sleep(SignaturesConfirmationRetryDelay)
	}
}

// receiveBundleResult receives a bundle result from the bundle stream subscription.
// It returns the received bundle result or an error if the reception fails.
func (c *SearcherClient) receiveBundleResult() (*bundle_pb.BundleResult, error) {
	bundleResult, err := c.BundleStreamSubscription.Recv()
	if err != nil {
		return nil, err
	}
	return bundleResult, nil
}

// handleBundleResult handles the received bundle result by checking its type and performing the necessary actions.
// It returns an error if the bundle result indicates a rejection.
func (c *SearcherClient) handleBundleResult(bundleResult *bundle_pb.BundleResult) error {
	switch bundleResult.Result.(type) {
	case *bundle_pb.BundleResult_Accepted:
		// Bundle accepted, no action needed
		break
	case *bundle_pb.BundleResult_Rejected:
		rejected := bundleResult.Result.(*bundle_pb.BundleResult_Rejected)
		switch rejected.Rejected.Reason.(type) {
		case *bundle_pb.Rejected_SimulationFailure:
			rejection := rejected.Rejected.GetSimulationFailure()
			return NewSimulationFailureError(rejection.TxSignature, rejection.GetMsg())
		case *bundle_pb.Rejected_StateAuctionBidRejected:
			rejection := rejected.Rejected.GetStateAuctionBidRejected()
			return NewStateAuctionBidRejectedError(rejection.AuctionId, rejection.SimulatedBidLamports)
		case *bundle_pb.Rejected_WinningBatchBidRejected:
			rejection := rejected.Rejected.GetWinningBatchBidRejected()
			return NewWinningBatchBidRejectedError(rejection.AuctionId, rejection.SimulatedBidLamports)
		case *bundle_pb.Rejected_InternalError:
			rejection := rejected.Rejected.GetInternalError()
			return NewInternalError(rejection.Msg)
		case *bundle_pb.Rejected_DroppedBundle:
			rejection := rejected.Rejected.GetDroppedBundle()
			return NewDroppedBundle(rejection.Msg)
		default:
			return nil
		}
	}
	return nil
}
