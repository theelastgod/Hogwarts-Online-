// Package settlement batches pending ledger withdrawals into on-chain
// transfers. The chain is behind the Chain interface so the launchpad's
// chain choice (EVM L2 or otherwise) only changes the adapter.
//
// Flow per batch:
//  1. Read pending withdrawals from the ledger (funds already in escrow).
//  2. Submit one batched transfer from the settlement wallet.
//  3. On success, mark each withdrawal settled with the tx hash.
//  4. On failure, mark each withdrawal failed so funds return to players.
//
// The settlement wallet is funded from the RewardsVault contract; keeping it
// topped up is a separate operator concern reported via WalletBalance.
package settlement

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/theelastgod/hogwarts-online/services/ledger/ledger"
)

// Transfer is one on-chain payout.
type Transfer struct {
	To     string
	Amount ledger.Amount
}

// Chain is the adapter to the token's chain.
type Chain interface {
	// SendBatch submits transfers atomically and returns the tx hash.
	SendBatch(ctx context.Context, transfers []Transfer) (txHash string, err error)
	// WalletBalance is the settlement wallet's on-chain WGLD balance.
	WalletBalance(ctx context.Context) (ledger.Amount, error)
}

// Config tunes batching.
type Config struct {
	BatchSize int
	// MinWalletReserve stops settlement when the wallet would drop below it,
	// leaving withdrawals pending until the vault tops the wallet up.
	MinWalletReserve ledger.Amount
}

var ErrWalletLow = errors.New("settlement: wallet below reserve, batch deferred")

// Worker drives settlement.
type Worker struct {
	cfg    Config
	ledger *ledger.Ledger
	chain  Chain
}

func NewWorker(cfg Config, l *ledger.Ledger, c Chain) *Worker {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}
	return &Worker{cfg: cfg, ledger: l, chain: c}
}

// Result summarises one RunOnce call.
type Result struct {
	Settled  int
	Failed   int
	Deferred int
	TxHash   string
}

// RunOnce settles at most one batch. It is safe to call repeatedly.
func (w *Worker) RunOnce(ctx context.Context) (Result, error) {
	pending := w.ledger.PendingWithdrawals(w.cfg.BatchSize)
	if len(pending) == 0 {
		return Result{}, nil
	}

	var total ledger.Amount
	transfers := make([]Transfer, 0, len(pending))
	for _, p := range pending {
		total += p.Amount
		transfers = append(transfers, Transfer{To: p.ChainAddr, Amount: p.Amount})
	}

	bal, err := w.chain.WalletBalance(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("settlement: wallet balance: %w", err)
	}
	if bal-total < w.cfg.MinWalletReserve {
		return Result{Deferred: len(pending)}, ErrWalletLow
	}

	txHash, err := w.chain.SendBatch(ctx, transfers)
	if err != nil {
		res := Result{}
		for _, p := range pending {
			if ferr := w.ledger.MarkFailed(p.ID, err.Error()); ferr == nil {
				res.Failed++
			}
		}
		return res, fmt.Errorf("settlement: send batch: %w", err)
	}

	res := Result{TxHash: txHash}
	for _, p := range pending {
		if err := w.ledger.MarkSettled(p.ID, txHash); err != nil {
			// The chain transfer succeeded; a ledger error here is an
			// invariant violation and must be surfaced, not swallowed.
			return res, fmt.Errorf("settlement: mark settled %d after tx %s: %w", p.ID, txHash, err)
		}
		res.Settled++
	}
	return res, nil
}

// Run loops RunOnce every interval until ctx is cancelled. Errors are
// reported through onError; ErrWalletLow is expected and non-fatal.
func (w *Worker) Run(ctx context.Context, interval time.Duration, onError func(error)) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		if _, err := w.RunOnce(ctx); err != nil && onError != nil {
			onError(err)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}
