// Copyright (C) 2024 Akave
// See LICENSE for copying information.

// Package ipc provides an ipc client model and access to deployed smart contract calls.
package ipc

import (
	"context"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/zeebo/errs"

	"akave.ai/akavesdk/private/ipc/contracts"
)

// Config represents configuration for the storage contract client.
type Config struct {
	DialURI         string `mapstructure:"dial_uri"`
	PrivateKey      string `mapstructure:"private_key"`
	ContractAddress string `mapstructure:"contract_address"`
}

// DefaultConfig returns default configuration for the ipc.
func DefaultConfig() Config {
	return Config{
		DialURI:         "",
		PrivateKey:      "",
		ContractAddress: "",
	}
}

// Client represents storage client.
type Client struct {
	Storage *contracts.Storage
	Auth    *bind.TransactOpts
	client  *ethclient.Client
	ticker  *time.Ticker
}

// Dial creates eth client, new smart-contract instance, auth.
func Dial(ctx context.Context, config Config) (*Client, error) {
	client, err := ethclient.Dial(config.DialURI)
	if err != nil {
		return &Client{}, err
	}

	privateKey, err := crypto.HexToECDSA(config.PrivateKey)
	if err != nil {
		return &Client{}, err
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return &Client{}, err
	}

	storage, err := contracts.NewStorage(common.HexToAddress(config.ContractAddress), client)
	if err != nil {
		return &Client{}, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return &Client{}, err
	}

	return &Client{
		Storage: storage,
		Auth:    auth,
		client:  client,
		ticker:  time.NewTicker(200 * time.Millisecond),
	}, nil
}

// DeployStorage deploys storage smart contract, returns it's client.
func DeployStorage(ctx context.Context, config Config) (*Client, error) {
	ethClient, err := ethclient.Dial(config.DialURI)
	if err != nil {
		return &Client{}, err
	}

	privateKey, err := crypto.HexToECDSA(config.PrivateKey)
	if err != nil {
		return &Client{}, err
	}

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		return &Client{}, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return &Client{}, err
	}

	_, tx, storage, err := contracts.DeployStorage(auth, ethClient)
	if err != nil {
		return &Client{}, err
	}

	client := &Client{
		Storage: storage,
		Auth:    auth,
		client:  ethClient,
		ticker:  time.NewTicker(200 * time.Millisecond),
	}

	return client, client.WaitForTx(ctx, tx.Hash())
}

// WaitForTx block execution until transaction receipt is received or context is cancelled.
func (client *Client) WaitForTx(ctx context.Context, hash common.Hash) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return errors.New("context canceled")
		case <-client.ticker.C:
			receipt, err := client.client.TransactionReceipt(ctx, hash)
			if err == nil {
				if receipt.Status == 1 {
					return nil
				}

				return errs.New("transaction failed")
			}
			if !errors.Is(err, ethereum.NotFound) {
				return err
			}
		}
	}
}
