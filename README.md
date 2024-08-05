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

# Guide to Updating PB Files from Proto Files

This guide outlines the steps to update `.pb` files generated from `.proto` files stored in various repository submodules. Follow these instructions to ensure a smooth update process.

## Repositories

The `.proto` files are stored in the following repository submodules:

- [block-engine-protos](https://github.com/Prophet-Solutions/block-engine-protos)
- [geyser-protos](https://github.com/Prophet-Solutions/geyser-protos)
- [yellowstone-geyser-protos](https://github.com/Prophet-Solutions/yellowstone-geyser-protos)

## Update Process

### Step 1: Install Protobuf Compiler

Ensure you have the protobuf compiler installed. Below are the instructions for macOS, Windows, and Linux.

#### macOS

1. Open Terminal.
2. Install Homebrew if not already installed:
    ```bash
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    ```
3. Install protobuf:
    ```bash
    brew install protobuf
    ```

#### Windows

1. Download the protobuf release from [GitHub](https://github.com/protocolbuffers/protobuf/releases).
2. Extract the downloaded zip file.
3. Add the `bin` directory of the extracted folder to your system's `PATH`.

#### Linux

1. Open Terminal.
2. Install protobuf using your package manager. For example, on Ubuntu:
    ```bash
    sudo apt-get install -y protobuf-compiler
    ```

### Step 2: Retrieve New Proto Files

1. Update the submodules to get the latest `.proto` files:
    ```bash
    git submodule update --remote
    ```

### Step 3: Generate PB Files

Run the following `.sh` scripts to generate the `.pb` files from the `.proto` files:

```bash
./scripts/gen-block-engine-protos.sh
./scripts/gen-geyser-protos.sh
./scripts/gen-yellowstone-geyser-protos.sh
```

### Step 4: Commit and Push Changes

1. Stage the changes:
    ```bash
    git add .
    ```
2. Commit the changes with a message, including the new version number:
    ```bash
    git commit -m "Updated PB files to match proto version <versionNumber>"
    ```
3. Push the changes to the repository:
    ```bash
    git push
    ```

### Step 5: Open a Pull Request

1. Open a new pull request (PR) to merge the changes into the main branch if the latest version isn't already pushed.

### Step 6: Publish a New Release

Create a new release to use the latest version in `jito-go`.

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

## Disclaimer

This library is not affiliated with Jito Labs.
This library is not supported by Jito Labs but by the community and repo owners.

## Resources

- [Jito Gitbook](https://jito-labs.gitbook.io)
- [Credits for help](https://github.com/weeaa/jito-go)
- [Jito Discord](https://discord.com/invite/jito)
- [Solana library used](https://github.com/gagliardetto/solana-go)

## To-Do

- Implement Geyser gRPC: [Geyser gRPC Plugin](https://github.com/jito-foundation/geyser-grpc-plugin)
- Implement YellowStone Geyser gRPC: [YellowStone Geyser gRPC](https://github.com/rpcpool/yellowstone-grpc)
