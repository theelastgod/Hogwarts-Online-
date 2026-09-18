# Contracts

Solidity contracts for the on-chain layer. The chain is set by the launchpad; these target any EVM chain. If the launchpad is on a non-EVM chain, the same design is ported and the settlement adapter in `services/ledger/settlement` changes.

| Contract | Standard | Purpose |
|---|---|---|
| `WizardingGold.sol` | ERC-20, burnable, permit | Fixed 1B supply minted once at TGE. No mint, no owner, no pause. |
| `RewardsVault.sol` | AccessControl, Pausable | Holds the 40% rewards allocation; releases to the settlement wallet under decaying per-season caps that can only be lowered. |
| `WizardingItems.sol` | ERC-1155, ERC-2981 | Tradable rare items minted only after in-game burn, with royalties. |

Vesting for team, investors, public sale, and treasury uses audited off-the-shelf vesting contracts (OpenZeppelin `VestingWallet` per beneficiary) deployed by the TGE script. They are not custom code.

## Design rules

- The internal ledger in `services/ledger` is the source of truth for player balances. On-chain settlement is withdrawal-only.
- The RewardsVault season cap mirrors the ledger's season cap. Both decay 4% per 90-day season from 20M.
- Every privileged function is behind a role. Admin must be a timelocked multisig before TGE.
- No upgradeable proxies. Launchpad due diligence expects an immutable token.

## Build

Requires [Foundry](https://book.getfoundry.sh/).

```
forge install OpenZeppelin/openzeppelin-contracts@v5.1.0
forge build
forge test
```

## Audit

Two independent audits are required before TGE. Store reports in `contracts/audits/`.
