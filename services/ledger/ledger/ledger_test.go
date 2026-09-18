package ledger

import (
	"errors"
	"testing"
	"time"
)

func newTest(cap Amount, vest time.Duration) (*Ledger, *time.Time) {
	l := New(Config{SeasonEmissionCap: cap, VestingDuration: vest})
	now := time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC)
	l.now = func() time.Time { return now }
	l.OpenAccount("player:alice")
	l.OpenAccount("player:bob")
	return l, &now
}

func TestTransferBalances(t *testing.T) {
	l, _ := newTest(1_000_000, 0)
	if _, err := l.Emit("s1", "player:alice", 500, "ranked", "match-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Transfer("player:alice", "player:bob", 200, "trade", "order-1"); err != nil {
		t.Fatal(err)
	}
	a, _ := l.Balance("player:alice")
	b, _ := l.Balance("player:bob")
	if a != 300 || b != 200 {
		t.Fatalf("got alice=%d bob=%d", a, b)
	}
	if c := l.Circulating(); c != 500 {
		t.Fatalf("circulating = %d, want 500", c)
	}
}

func TestPlayerCannotOverdraw(t *testing.T) {
	l, _ := newTest(1_000_000, 0)
	_, err := l.Transfer("player:alice", "player:bob", 1, "trade", "")
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("want ErrInsufficientFunds, got %v", err)
	}
	if a, _ := l.Balance("player:alice"); a != 0 {
		t.Fatalf("failed transfer mutated balance: %d", a)
	}
}

func TestUnbalancedEntryRejected(t *testing.T) {
	l, _ := newTest(1_000_000, 0)
	_, err := l.Post("bad", "",
		Posting{Account: "player:alice", Debit: 10},
		Posting{Account: TreasuryAccount, Credit: 5},
	)
	if !errors.Is(err, ErrUnbalanced) {
		t.Fatalf("want ErrUnbalanced, got %v", err)
	}
}

func TestEmissionCapEnforced(t *testing.T) {
	l, _ := newTest(1000, 0)
	if _, err := l.Emit("s1", "player:alice", 600, "ranked", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Emit("s1", "player:bob", 500, "ranked", ""); !errors.Is(err, ErrEmissionCap) {
		t.Fatalf("want ErrEmissionCap, got %v", err)
	}
	// A new season has its own cap.
	if _, err := l.Emit("s2", "player:bob", 500, "ranked", ""); err != nil {
		t.Fatal(err)
	}
	if got := l.SeasonEmitted("s1"); got != 600 {
		t.Fatalf("s1 emitted = %d, want 600", got)
	}
}

func TestVestingReleasesLinearly(t *testing.T) {
	l, now := newTest(1_000_000, 30*24*time.Hour)
	if _, err := l.Emit("s1", "player:alice", 3000, "season payout", ""); err != nil {
		t.Fatal(err)
	}
	if a, _ := l.Balance("player:alice"); a != 0 {
		t.Fatalf("balance before vesting = %d, want 0", a)
	}

	*now = now.Add(10 * 24 * time.Hour)
	if _, err := l.ReleaseVested(); err != nil {
		t.Fatal(err)
	}
	if a, _ := l.Balance("player:alice"); a != 1000 {
		t.Fatalf("balance after 10 days = %d, want 1000", a)
	}

	*now = now.Add(30 * 24 * time.Hour)
	if _, err := l.ReleaseVested(); err != nil {
		t.Fatal(err)
	}
	if a, _ := l.Balance("player:alice"); a != 3000 {
		t.Fatalf("balance after full vest = %d, want 3000", a)
	}
	// Idempotent.
	if released, _ := l.ReleaseVested(); released != 0 {
		t.Fatalf("second release moved %d", released)
	}
}

func TestBurnReducesCirculating(t *testing.T) {
	l, _ := newTest(1_000_000, 0)
	l.Emit("s1", "player:alice", 100, "ranked", "")
	if _, err := l.Burn("player:alice", 40, "mint fee", "item-7"); err != nil {
		t.Fatal(err)
	}
	if c := l.Circulating(); c != 60 {
		t.Fatalf("circulating = %d, want 60", c)
	}
	if n := len(l.Entries()); n != 2 {
		t.Fatalf("entries = %d, want 2", n)
	}
}

func TestWithdrawalEscrowSettleAndFail(t *testing.T) {
	l, _ := newTest(1_000_000, 0)
	l.Emit("s1", "player:alice", 100, "ranked", "")

	w, err := l.RequestWithdrawal("player:alice", "0xA", 60)
	if err != nil {
		t.Fatal(err)
	}
	if a, _ := l.Balance("player:alice"); a != 40 {
		t.Fatalf("alice = %d, want 40", a)
	}
	if e, _ := l.Balance(WithdrawalEscrowAccount); e != 60 {
		t.Fatalf("escrow = %d, want 60", e)
	}
	if c := l.Circulating(); c != 100 {
		t.Fatalf("circulating with escrow = %d, want 100", c)
	}

	if err := l.MarkSettled(w.ID, "0xtx"); err != nil {
		t.Fatal(err)
	}
	if c := l.Circulating(); c != 40 {
		t.Fatalf("circulating after settle = %d, want 40", c)
	}
	if err := l.MarkSettled(w.ID, "0xtx"); !errors.Is(err, ErrWithdrawalState) {
		t.Fatalf("double settle should fail, got %v", err)
	}

	w2, _ := l.RequestWithdrawal("player:alice", "0xA", 40)
	if err := l.MarkFailed(w2.ID, "rpc"); err != nil {
		t.Fatal(err)
	}
	if a, _ := l.Balance("player:alice"); a != 40 {
		t.Fatalf("alice after refund = %d, want 40", a)
	}
	if _, err := l.RequestWithdrawal("player:alice", "0xA", 41); !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("overdraw withdrawal should fail, got %v", err)
	}
}
