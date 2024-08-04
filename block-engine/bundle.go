package block_engine

import (
	"context"
	"fmt"
	"log"
	"time"

	bundle_pb "github.com/Prophet-Solutions/block-engine-protos/bundle"
	jito_pb "github.com/Prophet-Solutions/block-engine-protos/searcher"
	"github.com/Prophet-Solutions/jito-go/pkg"
	"github.com/gagliardetto/solana-go"
	"google.golang.org/grpc"
)

// Constants for retry and timeout configurations
const (
	CheckBundleRetries               = 5                // Number of times to retry checking bundle status
	CheckBundleRetryDelay            = 5 * time.Second  // Delay between retries for checking bundle status
	SignaturesConfirmationTimeout    = 15 * time.Second // Timeout for confirming signatures
	SignaturesConfirmationRetryDelay = 1 * time.Second  // Delay between retries for confirming signatures
)

// SendBundleWithConfirmation sends a bundle of transactions and waits for confirmation of signatures.
// It attempts to send the bundle, then continuously checks for the result of the bundle and validates
// the signatures of the transactions.
func (c *SearcherClient) SendBundleWithConfirmation(
	ctx context.Context,
	transactions []*solana.Transaction,
	opts ...grpc.CallOption,
) (*BundleResponse, error) {
	// Send the bundle of transactions
	resp, err := c.SendBundle(transactions, opts...)
	if err != nil {
		return nil, err
	}

	// Retry checking the bundle result up to a configured number of times
	for i := 0; i < CheckBundleRetries; i++ {
		select {
		case <-c.AuthenticationService.GRPCCtx.Done():
			// If the GRPC context is done, return the error
			return nil, c.AuthenticationService.GRPCCtx.Err()
		default:
			// Wait for a configured delay before retrying
			time.Sleep(CheckBundleRetryDelay)

			// Attempt to receive the bundle result
			bundleResult, err := c.receiveBundleResult()
			if err != nil {
				log.Println("error while receiving bundle result:", err)
			} else {
				// Handle the received bundle result
				if err = c.handleBundleResult(bundleResult); err != nil {
					return nil, err
				}

				log.Println("Bundle was sent.")
			}

			// Wait for the statuses of the transaction signatures
			statuses, err := c.waitForSignatureStatuses(ctx, transactions)
			if err != nil {
				continue
			}

			// Validate the received signature statuses
			if err = pkg.ValidateSignatureStatuses(statuses); err != nil {
				continue
			}

			// Return the successful bundle response with extracted signatures
			return &BundleResponse{
				BundleResponse: resp,
				Signatures:     pkg.BatchExtractSigFromTx(transactions),
			}, nil
		}
	}

	// If the retries are exhausted, return an error
	return nil, fmt.Errorf("BroadcastBundleWithConfirmation error: max retries (%d) exceeded", CheckBundleRetries)
}

// SendBundle creates and sends a bundle of transactions to the Searcher service.
// It converts transactions to a protobuf packet and sends it using the SearcherService.
func (c *SearcherClient) SendBundle(
	transactions []*solana.Transaction,
	opts ...grpc.CallOption,
) (*jito_pb.SendBundleResponse, error) {
	// Create a new bundle from the transactions
	bundle, err := c.NewBundle(transactions)
	if err != nil {
		return nil, err
	}

	// Send the bundle request to the Searcher service
	return c.SearcherService.SendBundle(
		c.AuthenticationService.GRPCCtx,
		&jito_pb.SendBundleRequest{
			Bundle: bundle,
		},
		opts...,
	)
}

// NewBundle creates a new bundle protobuf object from a slice of transactions.
// It converts the transactions into protobuf packets and includes them in the bundle.
func (c *SearcherClient) NewBundle(transactions []*solana.Transaction) (*bundle_pb.Bundle, error) {
	// Convert the transactions to protobuf packets
	packets, err := pkg.ConvertBatchTransactionToProtobufPacket(transactions)
	if err != nil {
		return nil, err
	}

	// Create and return the bundle with the converted packets
	return &bundle_pb.Bundle{
		Packets: packets,
		Header:  nil,
	}, nil
}

// NewBundleSubscriptionResults subscribes to bundle result updates from the Searcher service.
// It uses the provided gRPC call options to set up the subscription.
func (c *SearcherClient) NewBundleSubscriptionResults(opts ...grpc.CallOption) (jito_pb.SearcherService_SubscribeBundleResultsClient, error) {
	// Subscribe to bundle results from the Searcher service
	return c.SearcherService.SubscribeBundleResults(
		c.AuthenticationService.GRPCCtx,
		&jito_pb.SubscribeBundleResultsRequest{},
		opts...,
	)
}
