# Jito Go SDK

![Jito](https://jito-labs.gitbook.io/~gitbook/image?url=https%3A%2F%2F3427002662-files.gitbook.io%2F%7E%2Ffiles%2Fv0%2Fb%2Fgitbook-x-prod.appspot.com%2Fo%2Fspaces%252FHrQ5xfEkhvrbX39awm20%252Fuploads%252FDqQ3x4u1Pe1QPqQ0g9UD%252Fjlabscover.png%3Falt%3Dmedia%26token%3D218ee3d8-f5b2-4692-9146-cbcf1e8af359&width=1248&dpr=2&quality=100&sign=8ccedc4a&sv=1)

[![Go](https://img.shields.io/badge/Go-1.22.5-blue.svg)](https://golang.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/Prophet-Solutions/jito-go)](https://goreportcard.com/report/github.com/Prophet-Solutions/jito-go)
[![gRPC](https://img.shields.io/badge/gRPC-1.65.0-blue.svg)](https://grpc.io/)
[![Solana](https://img.shields.io/badge/Solana-Blockchain-green.svg)](https://github.com/gagliardetto/solana-go)

## Overview

The `jito-go` package, part of the [Prophet-Solutions](https://github.com/Prophet-Solutions) organization, provides a comprehensive interface for interacting with the Jito Block Engine. This package facilitates the operations of validators, relayers, and searchers, providing essential functionalities to work with transactions, bundles, and packet streams on the Solana blockchain.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
  - [Relayer](#relayer)
  - [Validator](#validator)
  - [Searcher Client](#searcher-client)
  - [Conversion Functions](#conversion-functions)
  - [Signature Handling](#signature-handling)
  - [Utility Functions](#utility-functions)
  - [Example: Creating and Sending a Bundle](#example-creating-and-sending-a-bundle)
- [Resources](#resources)
- [To-Do](#to-do)
- [Dependencies](#dependencies)

## Installation

To install the package, run:

```bash
go get github.com/Prophet-Solutions/jito-go
```

## Usage

### Relayer

Provides functionalities for subscribing to accounts of interest, programs of interest, and expiring packet streams.

```go
relayer, err := pkg.NewRelayer(ctx, "grpc-address", keyPair)
if err != nil {
    // handle error
}

// Subscribe to accounts of interest
accounts, errs, err := relayer.OnSubscribeAccountsOfInterest(ctx)
if err != nil {
    // handle error
}
```

### Validator

Provides functionalities for subscribing to packet and bundle updates and retrieving block builder fee information.

```go
validator, err := pkg.NewValidator(ctx, "grpc-address", keyPair)
if err != nil {
    // handle error
}

// Subscribe to packet updates
packets, errs, err := validator.OnPacketSubscription(ctx)
if err != nil {
    // handle error
}
```

### Searcher Client

Provides functionalities for sending bundles with confirmation, retrieving regions and connected leaders, and obtaining random tip accounts.

```go
searcher, err := pkg.NewSearcherClient(ctx, "grpc-address", jitoRPCClient, rpcClient, keyPair)
if err != nil {
    // handle error
}

// Get connected leaders
leaders, err := searcher.GetConnectedLeaders()
if err != nil {
    // handle error
}
```

### Conversion Functions

Helper functions for converting Solana transactions to protobuf packets and vice versa.

```go
packet, err := pkg.ConvertTransactionToProtobufPacket(transaction)
if err != nil {
    // handle error
}

transactions, err := pkg.ConvertBatchProtobufPacketToTransaction(packets)
if err != nil {
    // handle error
}
```

### Signature Handling

Helper functions for extracting and validating transaction signatures.

```go
signature := pkg.ExtractSigFromTx(transaction)
signatures := pkg.BatchExtractSigFromTx(transactions)

if !pkg.CheckSignatureStatuses(statuses) {
    // handle invalid statuses
}

if err := pkg.ValidateSignatureStatuses(statuses); err != nil {
    // handle error
}
```

### Utility Functions

Additional utility functions for working with lamports and endpoints.

```go
sol := pkg.LamportsToSol(big.NewFloat(1000000))

endpoint := pkg.GetEndpoint("AMS")
```

### Example: Creating and Sending a Bundle

Below is an example demonstrating how to create and send a bundle using the `jito-go` package.

```go
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "math/rand"
    "time"

    block_engine "github.com/Prophet-Solutions/jito-go/block-engine"
    "github.com/Prophet-Solutions/jito-go/pkg"
    block_engine_pkg "github.com/Prophet-Solutions/jito-go/pkg/block-engine"
    "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/programs/system"
    "github.com/gagliardetto/solana-go/rpc"
)

var (
    NumTxs      = flag.Int("numTxs", 3, "Total transactions amount to send in the bundle")
    RpcAddress  = flag.String("rpcAddress", "https://api.mainnet-beta.solana.com/", "RPC Address that will be used for GetRecentBlockhash and GetBalance methods")
    TxLamports  = flag.Uint64("txLamports", 20000, "Total amount in Lamports to send in one single transaction")
    TipLamports = flag.Uint64("tipLamports", 10000, "Total amount in Lamports to send to Jito as tip")
)

var (
    SenderPrivateKey  = solana.MustPrivateKeyFromBase58("your-private-key")
    ReceiverPublicKey = solana.MustPublicKeyFromBase58("your-receiver-public-key")
)

func main() {
    flag.Parse()
    bundleResult, err := SubmitBundle()
    if err != nil {
        log.Fatalf("Failed to submit bundle: %v", err)
    }
    log.Printf("Bundle submitted successfully: %s\n", bundleResult.BundleResponse.GetUuid())
    log.Printf("Bundle txs: %s", bundleResult.Signatures)
}

func SubmitBundle() (*block_engine.BundleResponse, error) {
    ctx := context.Background()
    searcherClient, err := createSearcherClient(ctx)
    if err != nil {
        return nil, err
    }

    blockHash, err := getRecentBlockhash(ctx, searcherClient)
    if err != nil {
        return nil, err
    }

    tipAccount, err := searcherClient.GetRandomTipAccount()
    if err != nil {
        return nil, fmt.Errorf("could not get random tip account: %w", err)
    }

    txs, err := buildTransactions(blockHash.Value.Blockhash, tipAccount)
    if err != nil {
        return nil, err
    }

    log.Println("Sending bundle.")
    bundleResult, err := searcherClient.SendBundleWithConfirmation(ctx, txs)
    if err != nil {
        return nil, fmt.Errorf("could not send bundle: %w", err)
    }

    return bundleResult, nil
}

func createSearcherClient(ctx context.Context) (*block_engine.SearcherClient, error) {
    searcherClient, err := block_engine.NewSearcherClient(
        ctx,
        block_engine_pkg.GetEndpoint("FRA"),
        nil,
        rpc.New(*RpcAddress),
        &SenderPrivateKey,
    )
    if err != nil {
        return nil, fmt.Errorf("could not create searcher client: %w", err)
    }
    return searcherClient, nil
}

func getRecentBlockhash(ctx context.Context, client *block_engine.SearcherClient) (*rpc.GetRecentBlockhashResult, error) {
    blockHash, err := client.RPCConn.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
    if err != nil {
        return nil, fmt.Errorf("could not get recent blockhash: %w", err)
    }
    return blockHash, nil
}

func buildTransactions(blockhash solana.Hash, tipAccount string) ([]*solana.Transaction, error) {
    var txs []*solana.Transaction

    for i := 0; i < *NumTxs; i++ {
        tx, err := createTransaction(blockhash, ReceiverPublicKey, *TxLamports)
        if err != nil {
            return nil, err
        }
        txs = append(txs, tx)
    }

    tipTx, err := createTransaction(blockhash, solana.MustPublicKeyFromBase58(tipAccount), *TipLamports)
    if err != nil {
        return nil, err
    }
    txs = append(txs, tipTx)

    return txs, nil
}

func createTransaction(blockhash solana.Hash, recipient solana.PublicKey, lamports uint64) (*solana.Transaction, error) {
    rand.Seed(time.Now().UnixNano())
    tx, err := solana.NewTransaction(
        []solana.Instruction{
            system.NewTransferInstruction(
                lam

ports,
                SenderPrivateKey.PublicKey(),
                recipient,
            ).Build(),
            solana.NewInstruction(
                pkg.MemoPublicKey,
                solana.AccountMetaSlice{
                    &solana.AccountMeta{
                        PublicKey:  SenderPrivateKey.PublicKey(),
                        IsWritable: true,
                        IsSigner:   true,
                    },
                },
                []byte(fmt.Sprintf("jito bundle %d", rand.Intn(1000000)+1)),
            ),
        },
        blockhash,
        solana.TransactionPayer(SenderPrivateKey.PublicKey()),
    )
    if err != nil {
        return nil, fmt.Errorf("could not build transaction: %w", err)
    }

    _, err = tx.Sign(func(pubKey solana.PublicKey) *solana.PrivateKey {
        if pubKey.Equals(SenderPrivateKey.PublicKey()) {
            return &SenderPrivateKey
        }
        return nil
    })
    if err != nil {
        return nil, fmt.Errorf("failed to sign transaction: %w", err)
    }

    return tx, nil
}
```

## Resources

- [Jito Gitbook](https://jito-labs.gitbook.io)
- [Credits for help](https://github.com/weeaa/jito-go)
- [Jito Discord](https://discord.com/invite/jito)
- [Solana library used](https://github.com/gagliardetto/solana-go)

## To-Do

- Implement Geyser gRPC: [Geyser gRPC Plugin](https://github.com/jito-foundation/geyser-grpc-plugin)
- Implement YellowStone Geyser gRPC: [YellowStone Geyser gRPC](https://github.com/rpcpool/yellowstone-grpc)

## Dependencies

The following dependencies are used in this project:

- [filippo.io/edwards25519 v1.0.0-rc.1](https://pkg.go.dev/filippo.io/edwards25519)
- [github.com/Prophet-Solutions/block-engine-protos v1.0.0](https://github.com/Prophet-Solutions/block-engine-protos)
- [github.com/Prophet-Solutions/geyser-protos v1.0.0](https://github.com/Prophet-Solutions/geyser-protos)
- [github.com/Prophet-Solutions/yellowstone-geyser-protos v1.0.1](https://github.com/Prophet-Solutions/yellowstone-geyser-protos)
- [github.com/andres-erbsen/clock v0.0.0-20160526145045-9e14626cd129](https://github.com/andres-erbsen/clock)
- [github.com/blendle/zapdriver v1.3.1](https://github.com/blendle/zapdriver)
- [github.com/davecgh/go-spew v1.1.1](https://github.com/davecgh/go-spew)
- [github.com/fatih/color v1.9.0](https://github.com/fatih/color)
- [github.com/gagliardetto/binary v0.8.0](https://github.com/gagliardetto/binary)
- [github.com/gagliardetto/solana-go v1.11.0](https://github.com/gagliardetto/solana-go)
- [github.com/gagliardetto/treeout v0.1.4](https://github.com/gagliardetto/treeout)
- [github.com/google/uuid v1.6.0](https://github.com/google/uuid)
- [github.com/json-iterator/go v1.1.12](https://github.com/json-iterator/go)
- [github.com/klauspost/compress v1.13.6](https://github.com/klauspost/compress)
- [github.com/logrusorgru/aurora v2.0.3+incompatible](https://github.com/logrusorgru/aurora)
- [github.com/mattn/go-colorable v0.1.4](https://github.com/mattn/go-colorable)
- [github.com/mattn/go-isatty v0.0.11](https://github.com/mattn/go-isatty)
- [github.com/mitchellh/go-testing-interface v1.14.1](https://github.com/mitchellh/go-testing-interface)
- [github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd](https://github.com/modern-go/concurrent)
- [github.com/modern-go/reflect2 v1.0.2](https://github.com/modern-go/reflect2)
- [github.com/mostynb/zstdpool-freelist v0.0.0-20201229113212-927304c0c3b1](https://github.com/mostynb/zstdpool-freelist)
- [github.com/mr-tron/base58 v1.2.0](https://github.com/mr-tron/base58)
- [github.com/streamingfast/logging v0.0.0-20230608130331-f22c91403091](https://github.com/streamingfast/logging)
- [go.mongodb.org/mongo-driver v1.16.0](https://go.mongodb.org/mongo-driver)
- [go.uber.org/atomic v1.7.0](https://go.uber.org/atomic)
- [go.uber.org/multierr v1.6.0](https://go.uber.org/multierr)
- [go.uber.org/ratelimit v0.2.0](https://go.uber.org/ratelimit)
- [go.uber.org/zap v1.21.0](https://go.uber.org/zap)
- [golang.org/x/crypto v0.23.0](https://golang.org/x/crypto)
- [golang.org/x/net v0.25.0](https://golang.org/x/net)
- [golang.org/x/sys v0.20.0](https://golang.org/x/sys)
- [golang.org/x/term v0.20.0](https://golang.org/x/term)
- [golang.org/x/text v0.15.0](https://golang.org/x/text)
- [golang.org/x/time v0.0.0-20191024005414-555d28b269f0](https://golang.org/x/time)
- [google.golang.org/genproto/googleapis/rpc v0.0.0-20240722135656-d784300faade](https://google.golang.org/genproto/googleapis/rpc)
- [google.golang.org/grpc v1.65.0](https://google.golang.org/grpc)
- [google.golang.org/protobuf v1.34.2](https://google.golang.org/protobuf)
