# Contracts

Solidity contracts for the on-chain layer. Target: an EVM Layer 2 (Base, Arbitrum, or Polygon), chosen in Phase 1.

| Contract | Standard | Purpose |
|---|---|---|
| `WizardingGold.sol` | ERC-20, pausable, role-gated | WGLD token with per-season mint caps and a hard max supply |
| `WizardingItems.sol` | ERC-1155, ERC-2981 | Tradable rare items minted only after in-game burn, with royalties |

## Design rules

- The internal ledger in `services/ledger` is the source of truth. Contracts only mirror settled state.
- Season caps on-chain match the ledger's caps and can only be lowered once set.
- Every privileged function is behind a role. Admin should be a multisig with a timelock before mainnet.
- No upgradeable proxies in v1. If upgradeability is needed, use a timelocked proxy and document the rationale.

## Build

Requires [Foundry](https://book.getfoundry.sh/).

```
forge install OpenZeppelin/openzeppelin-contracts@v5.1.0
forge build
forge test
```

## Audit

Two independent audits are required before mainnet (see DEVELOPMENT_PLAN.md, Phase 2). Store reports in `contracts/audits/`.
