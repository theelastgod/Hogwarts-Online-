package settlement

import (
	"context"
	"errors"
	"testing"

	"github.com/theelastgod/hogwarts-online/services/ledger/ledger"
)

type mockChain struct {
	balance ledger.Amount
	fail    error
	batches [][]Transfer
}

func (m *mockChain) SendBatch(_ context.Context, t []Transfer) (string, error) {
	if m.fail != nil {
		return "", m.fail
	}
	m.batches = append(m.batches, t)
	for _, x := range t {
		m.balance -= x.Amount
	}
	return "0xabc", nil
}

func (m *mockChain) WalletBalance(context.Context) (ledger.Amount, error) { return m.balance, nil }

func setup(t *testing.T, walletBal ledger.Amount) (*ledger.Ledger, *mockChain) {
	t.Helper()
	l := ledger.New(ledger.Config{SeasonEmissionCap: 1_000_000})
	l.OpenAccount("player:alice")
	l.OpenAccount("player:bob")
	if _, err := l.Emit("s1", "player:alice", 500, "ranked", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Emit("s1", "player:bob", 300, "ranked", ""); err != nil {
		t.Fatal(err)
	}
	return l, &mockChain{balance: walletBal}
}

func TestSettlesPendingWithdrawalsInOneBatch(t *testing.T) {
	l, c := setup(t, 10_000)
	l.RequestWithdrawal("player:alice", "0xA", 200)
	l.RequestWithdrawal("player:bob", "0xB", 100)

	w := NewWorker(Config{BatchSize: 10}, l, c)
	res, err := w.RunOnce(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Settled != 2 || res.TxHash != "0xabc" {
		t.Fatalf("result = %+v", res)
	}
	if len(c.batches) != 1 || len(c.batches[0]) != 2 {
		t.Fatalf("batches = %+v", c.batches)
	}
	if len(l.PendingWithdrawals(0)) != 0 {
		t.Fatal("withdrawals still pending")
	}
	// Custodial circulation shrank by the settled amount.
	if circ := l.Circulating(); circ != 500 {
		t.Fatalf("circulating = %d, want 500", circ)
	}
}

func TestChainFailureRefundsPlayers(t *testing.T) {
	l, c := setup(t, 10_000)
	l.RequestWithdrawal("player:alice", "0xA", 200)
	c.fail = errors.New("rpc down")

	w := NewWorker(Config{}, l, c)
	res, err := w.RunOnce(context.Background())
	if err == nil || res.Failed != 1 {
		t.Fatalf("err=%v res=%+v", err, res)
	}
	if bal, _ := l.Balance("player:alice"); bal != 500 {
		t.Fatalf("alice = %d, want refund to 500", bal)
	}
}

func TestLowWalletDefersBatch(t *testing.T) {
	l, c := setup(t, 250)
	l.RequestWithdrawal("player:alice", "0xA", 200)

	w := NewWorker(Config{MinWalletReserve: 100}, l, c)
	res, err := w.RunOnce(context.Background())
	if !errors.Is(err, ErrWalletLow) || res.Deferred != 1 {
		t.Fatalf("err=%v res=%+v", err, res)
	}
	if len(l.PendingWithdrawals(0)) != 1 {
		t.Fatal("withdrawal should remain pending")
	}
	if bal, _ := l.Balance("player:alice"); bal != 300 {
		t.Fatalf("alice = %d, funds should stay in escrow", bal)
	}
}

func TestBatchSizeRespected(t *testing.T) {
	l, c := setup(t, 10_000)
	for i := 0; i < 5; i++ {
		l.RequestWithdrawal("player:alice", "0xA", 10)
	}
	w := NewWorker(Config{BatchSize: 2}, l, c)
	res, _ := w.RunOnce(context.Background())
	if res.Settled != 2 || len(l.PendingWithdrawals(0)) != 3 {
		t.Fatalf("settled=%d pending=%d", res.Settled, len(l.PendingWithdrawals(0)))
	}
}
