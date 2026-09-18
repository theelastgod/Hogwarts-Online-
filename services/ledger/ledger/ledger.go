// Package ledger implements the internal double-entry ledger for WGLD.
//
// Gameplay services credit and debit accounts here. On-chain settlement is a
// separate, asynchronous concern: the ledger is the source of truth and the
// chain mirrors it in batches. This keeps gameplay independent of chain
// latency and lets the chain be swapped.
//
// Invariants enforced:
//   - Every entry balances: sum(debits) == sum(credits).
//   - Player accounts never go negative.
//   - Seasonal emissions cannot exceed the configured cap.
//   - Reward payouts vest linearly over a configurable duration.
package ledger

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Amount is WGLD in its smallest unit (1e-6 WGLD). Integers avoid float drift.
type Amount int64

// AccountID identifies a ledger account. System accounts use fixed IDs.
type AccountID string

const (
	// EmissionAccount is the source of newly minted seasonal rewards.
	EmissionAccount AccountID = "sys:emission"
	// TreasuryAccount collects fees, unclaimed pools, and royalties.
	TreasuryAccount AccountID = "sys:treasury"
	// BurnAccount is a sink; balances here are removed from circulation.
	BurnAccount AccountID = "sys:burn"
	// WithdrawalEscrowAccount holds funds awaiting on-chain settlement.
	WithdrawalEscrowAccount AccountID = "sys:withdrawal-escrow"
	// SettledAccount holds funds that now live on-chain and are no longer
	// custodial. Balances here are excluded from Circulating().
	SettledAccount AccountID = "sys:settled"
)

// WithdrawalStatus tracks a withdrawal through settlement.
type WithdrawalStatus uint8

const (
	WithdrawalPending WithdrawalStatus = iota
	WithdrawalSettled
	WithdrawalFailed
)

// Withdrawal is a request to move custodial WGLD on-chain.
type Withdrawal struct {
	ID          uint64
	Account     AccountID
	ChainAddr   string
	Amount      Amount
	Status      WithdrawalStatus
	TxHash      string
	RequestedAt time.Time
}

var ErrWithdrawalState = errors.New("ledger: withdrawal not in expected state")

var (
	ErrInsufficientFunds = errors.New("ledger: insufficient funds")
	ErrUnbalanced        = errors.New("ledger: entry does not balance")
	ErrEmissionCap       = errors.New("ledger: seasonal emission cap exceeded")
	ErrUnknownAccount    = errors.New("ledger: unknown account")
	ErrInvalidAmount     = errors.New("ledger: amount must be positive")
)

// Posting is one side of a double-entry transaction.
type Posting struct {
	Account AccountID
	Debit   Amount // Increases the account balance.
	Credit  Amount // Decreases the account balance.
}

// Entry is a balanced set of postings with metadata for audit.
type Entry struct {
	ID        uint64
	At        time.Time
	Reason    string
	Postings  []Posting
	Reference string // Optional external reference, e.g. match ID.
}

// VestingSchedule holds rewards that unlock linearly between Start and End.
type VestingSchedule struct {
	Account  AccountID
	Total    Amount
	Released Amount
	Start    time.Time
	End      time.Time
}

// Config controls caps and vesting.
type Config struct {
	SeasonEmissionCap Amount
	VestingDuration   time.Duration
}

// Ledger is an in-memory ledger safe for concurrent use. A persistent
// implementation should keep the same interface and invariants.
type Ledger struct {
	mu       sync.Mutex
	cfg      Config
	balances map[AccountID]Amount
	entries  []Entry
	nextID   uint64
	emitted  map[string]Amount // season -> emitted so far
	vesting  []*VestingSchedule
	now      func() time.Time

	withdrawals    map[uint64]*Withdrawal
	nextWithdrawal uint64
}

// New creates a ledger with the system accounts opened.
func New(cfg Config) *Ledger {
	l := &Ledger{
		cfg:      cfg,
		balances: map[AccountID]Amount{},
		emitted:  map[string]Amount{},
		nextID:   1,
		now:      time.Now,

		withdrawals:    map[uint64]*Withdrawal{},
		nextWithdrawal: 1,
	}
	for _, a := range []AccountID{EmissionAccount, TreasuryAccount, BurnAccount, WithdrawalEscrowAccount, SettledAccount} {
		l.balances[a] = 0
	}
	return l
}

// RequestWithdrawal moves funds from a player account into escrow and records
// a pending withdrawal for the settlement worker to push on-chain.
func (l *Ledger) RequestWithdrawal(from AccountID, chainAddr string, amt Amount) (Withdrawal, error) {
	if amt <= 0 {
		return Withdrawal{}, ErrInvalidAmount
	}
	if chainAddr == "" {
		return Withdrawal{}, errors.New("ledger: chain address required")
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	w := &Withdrawal{
		ID:          l.nextWithdrawal,
		Account:     from,
		ChainAddr:   chainAddr,
		Amount:      amt,
		Status:      WithdrawalPending,
		RequestedAt: l.now(),
	}
	_, err := l.postLocked("withdrawal escrow", fmt.Sprintf("withdrawal:%d", w.ID),
		Posting{Account: from, Credit: amt},
		Posting{Account: WithdrawalEscrowAccount, Debit: amt},
	)
	if err != nil {
		return Withdrawal{}, err
	}
	l.nextWithdrawal++
	l.withdrawals[w.ID] = w
	return *w, nil
}

// PendingWithdrawals returns pending withdrawals in ID order, up to limit
// (0 = all).
func (l *Ledger) PendingWithdrawals(limit int) []Withdrawal {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []Withdrawal
	for id := uint64(1); id < l.nextWithdrawal; id++ {
		w := l.withdrawals[id]
		if w != nil && w.Status == WithdrawalPending {
			out = append(out, *w)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out
}

// MarkSettled records that a withdrawal landed on-chain. Escrowed funds move
// to the settled account and leave custodial circulation.
func (l *Ledger) MarkSettled(id uint64, txHash string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	w := l.withdrawals[id]
	if w == nil || w.Status != WithdrawalPending {
		return ErrWithdrawalState
	}
	_, err := l.postLocked("withdrawal settled", txHash,
		Posting{Account: WithdrawalEscrowAccount, Credit: w.Amount},
		Posting{Account: SettledAccount, Debit: w.Amount},
	)
	if err != nil {
		return err
	}
	w.Status = WithdrawalSettled
	w.TxHash = txHash
	return nil
}

// MarkFailed returns escrowed funds to the player after a settlement failure.
func (l *Ledger) MarkFailed(id uint64, reason string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	w := l.withdrawals[id]
	if w == nil || w.Status != WithdrawalPending {
		return ErrWithdrawalState
	}
	_, err := l.postLocked("withdrawal failed: "+reason, fmt.Sprintf("withdrawal:%d", id),
		Posting{Account: WithdrawalEscrowAccount, Credit: w.Amount},
		Posting{Account: w.Account, Debit: w.Amount},
	)
	if err != nil {
		return err
	}
	w.Status = WithdrawalFailed
	return nil
}

// Withdrawal returns a copy of a withdrawal record.
func (l *Ledger) Withdrawal(id uint64) (Withdrawal, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	w := l.withdrawals[id]
	if w == nil {
		return Withdrawal{}, false
	}
	return *w, true
}

// OpenAccount creates a player or system account with zero balance.
func (l *Ledger) OpenAccount(id AccountID) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.balances[id]; !ok {
		l.balances[id] = 0
	}
}

// Balance returns the spendable balance of an account.
func (l *Ledger) Balance(id AccountID) (Amount, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.balances[id]
	if !ok {
		return 0, ErrUnknownAccount
	}
	return b, nil
}

// Post applies a balanced entry atomically.
func (l *Ledger) Post(reason, reference string, postings ...Posting) (Entry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.postLocked(reason, reference, postings...)
}

func (l *Ledger) postLocked(reason, reference string, postings ...Posting) (Entry, error) {
	var debits, credits Amount
	for _, p := range postings {
		if p.Debit < 0 || p.Credit < 0 || (p.Debit == 0 && p.Credit == 0) {
			return Entry{}, ErrInvalidAmount
		}
		if _, ok := l.balances[p.Account]; !ok {
			return Entry{}, fmt.Errorf("%w: %s", ErrUnknownAccount, p.Account)
		}
		debits += p.Debit
		credits += p.Credit
	}
	if debits != credits {
		return Entry{}, ErrUnbalanced
	}

	// Validate before mutating so a failure leaves state untouched.
	projected := map[AccountID]Amount{}
	for _, p := range postings {
		if _, ok := projected[p.Account]; !ok {
			projected[p.Account] = l.balances[p.Account]
		}
		projected[p.Account] += p.Debit - p.Credit
	}
	for id, bal := range projected {
		if bal < 0 && !isSystem(id) {
			return Entry{}, fmt.Errorf("%w: %s", ErrInsufficientFunds, id)
		}
	}
	for id, bal := range projected {
		l.balances[id] = bal
	}

	e := Entry{
		ID:        l.nextID,
		At:        l.now(),
		Reason:    reason,
		Reference: reference,
		Postings:  append([]Posting(nil), postings...),
	}
	l.nextID++
	l.entries = append(l.entries, e)
	return e, nil
}

// Transfer moves funds between two accounts.
func (l *Ledger) Transfer(from, to AccountID, amt Amount, reason, reference string) (Entry, error) {
	if amt <= 0 {
		return Entry{}, ErrInvalidAmount
	}
	return l.Post(reason, reference,
		Posting{Account: from, Credit: amt},
		Posting{Account: to, Debit: amt},
	)
}

// Burn removes funds from circulation.
func (l *Ledger) Burn(from AccountID, amt Amount, reason, reference string) (Entry, error) {
	return l.Transfer(from, BurnAccount, amt, reason, reference)
}

// Emit pays a seasonal reward into a vesting schedule, subject to the cap.
// The tokens are debited from the emission source immediately (so the cap is
// enforced at grant time) and released to the player over VestingDuration.
func (l *Ledger) Emit(season string, to AccountID, amt Amount, reason, reference string) (*VestingSchedule, error) {
	if amt <= 0 {
		return nil, ErrInvalidAmount
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.emitted[season]+amt > l.cfg.SeasonEmissionCap {
		return nil, ErrEmissionCap
	}
	if _, ok := l.balances[to]; !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownAccount, to)
	}
	l.emitted[season] += amt

	now := l.now()
	vs := &VestingSchedule{
		Account: to,
		Total:   amt,
		Start:   now,
		End:     now.Add(l.cfg.VestingDuration),
	}
	l.vesting = append(l.vesting, vs)

	if l.cfg.VestingDuration <= 0 {
		if err := l.releaseLocked(vs, amt); err != nil {
			return nil, err
		}
	}
	return vs, nil
}

// ReleaseVested moves any newly unlocked vesting amounts into player balances.
// Call this on a schedule (e.g. hourly) or before reads that need freshness.
func (l *Ledger) ReleaseVested() (Amount, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	var total Amount
	for _, vs := range l.vesting {
		unlockable := vs.unlockedAt(now) - vs.Released
		if unlockable <= 0 {
			continue
		}
		if err := l.releaseLocked(vs, unlockable); err != nil {
			return total, err
		}
		total += unlockable
	}
	return total, nil
}

func (l *Ledger) releaseLocked(vs *VestingSchedule, amt Amount) error {
	_, err := l.postLocked("vesting release", "",
		Posting{Account: EmissionAccount, Credit: amt},
		Posting{Account: vs.Account, Debit: amt},
	)
	if err != nil {
		return err
	}
	vs.Released += amt
	return nil
}

func (vs *VestingSchedule) unlockedAt(now time.Time) Amount {
	if !now.Before(vs.End) {
		return vs.Total
	}
	if now.Before(vs.Start) {
		return 0
	}
	elapsed := now.Sub(vs.Start)
	dur := vs.End.Sub(vs.Start)
	return Amount(int64(vs.Total) * int64(elapsed) / int64(dur))
}

// SeasonEmitted reports how much of a season's cap has been granted.
func (l *Ledger) SeasonEmitted(season string) Amount {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.emitted[season]
}

// Entries returns a copy of the audit log.
func (l *Ledger) Entries() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]Entry(nil), l.entries...)
}

// Circulating returns custodial supply: player balances plus withdrawals in
// escrow. Settled (on-chain) and burned balances are excluded.
func (l *Ledger) Circulating() Amount {
	l.mu.Lock()
	defer l.mu.Unlock()
	total := l.balances[WithdrawalEscrowAccount]
	for id, b := range l.balances {
		if !isSystem(id) {
			total += b
		}
	}
	return total
}

func isSystem(id AccountID) bool {
	return len(id) > 4 && id[:4] == "sys:"
}
