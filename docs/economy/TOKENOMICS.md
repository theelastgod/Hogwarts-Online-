# WGLD Tokenomics (Launchpad Edition)

Decision (2026-09-18): WGLD launches through a launchpad. That fixes the design in three ways:

1. **Fixed supply.** The full supply is minted at token generation event (TGE). There is no mint authority afterwards.
2. **Rewards come from a vault.** Seasonal emissions are released from a pre-allocated Rewards Vault under on-chain per-season caps, not minted.
3. **Vesting is on-chain for every allocation.** Team, investors, and launchpad participants all vest through verifiable contracts.

The chain is chosen by the launchpad. The ledger and settlement worker are chain-agnostic; only the settlement adapter changes.

## Supply and Allocation

Total supply: **1,000,000,000 WGLD** (1e9, 18 decimals on EVM or 9 on Solana).

| Allocation | Share | Tokens | TGE unlock | Cliff | Linear vest | Purpose |
|---|---|---|---|---|---|---|
| Rewards Vault | 40% | 400,000,000 | 0% | none | 8 years, per-season caps | Player earnings |
| Ecosystem and Treasury | 15% | 150,000,000 | 5% | 6 months | 4 years | Grants, tournaments, liquidity top-ups |
| Launchpad public sale | 10% | 100,000,000 | 25% | none | 6 months | Community distribution |
| Liquidity | 8% | 80,000,000 | 100% | none | none | DEX and CEX pools, locked 24 months |
| Team and advisors | 15% | 150,000,000 | 0% | 12 months | 36 months | Aligned with live service |
| Private and strategic | 10% | 100,000,000 | 10% | 6 months | 18 months | Pre-launch capital |
| Community and airdrop | 2% | 20,000,000 | 50% | none | 3 months | Early testers, creators |

Circulating at TGE: about 12.6% (liquidity, public sale unlock, private unlock, community unlock, treasury unlock).

## Rewards Vault Schedule

- Season length: 90 days. 32 seasons over 8 years.
- Season cap starts at 20,000,000 WGLD (5% of vault) and decays 4% per season. Sum over 32 seasons is about 365M, leaving a buffer of about 35M for extensions.
- Caps can be lowered by governance, never raised. Unspent season allowance stays in the vault.
- Player payouts vest 30 days in the internal ledger before they can be withdrawn on-chain.

| Season | Cap (WGLD) | Cumulative |
|---|---|---|
| 1 | 20,000,000 | 20,000,000 |
| 4 | 17,695,000 | 75,400,000 |
| 8 | 15,032,000 | 140,600,000 |
| 16 | 10,840,000 | 242,600,000 |
| 32 | 5,634,000 | 365,000,000 |

## Sinks (Burn or Recycle)

| Sink | Rate | Destination |
|---|---|---|
| Marketplace fee | 5% of sale | 50% burn, 50% treasury |
| Item minting | flat fee per rarity tier | 100% burn |
| Beast breeding | fee plus cooldown skip | 100% burn |
| Guild hall upkeep | weekly | 100% burn |
| Crafter royalty | 2.5% of secondary sale | crafter (not a sink, a redistribution) |

Burned tokens reduce supply permanently. Treasury share funds tournaments, so it recirculates through the Rewards Vault path and is capped the same way.

## Launchpad Readiness Checklist

| # | Item | Owner | Status |
|---|---|---|---|
| 1 | Choose launchpad and chain; confirm listing terms and fee | Studio head | [ ] |
| 2 | Token contract: fixed supply, no mint, pausable transfers off before TGE | Blockchain lead | [ ] |
| 3 | Vesting contracts for every allocation, with public dashboard | Blockchain lead | [ ] |
| 4 | Rewards Vault with per-season caps, timelocked admin | Blockchain lead | [ ] |
| 5 | Two independent audits on all contracts | Blockchain lead | [ ] |
| 6 | Liquidity lock proof (24 months) | Finance | [ ] |
| 7 | Securities opinion per launch territory, launchpad KYC alignment | Counsel | [ ] |
| 8 | Public tokenomics page mirroring this document | Marketing | [ ] |
| 9 | Playable build or vertical slice video for launchpad due diligence | Production | [ ] |
| 10 | Settlement worker live on testnet with withdrawals end-to-end | Backend | [ ] |
| 11 | Economy dashboards live before TGE | Infra | [ ] |
| 12 | Incident plan: pause switch, comms templates, key rotation | Security | [ ] |

## What Changed From the Original Plan

- Per-season minting replaced by vault releases. The economy sim now runs in fixed-supply mode and reports vault runway.
- The internal ledger's emission cap now mirrors the vault's season cap.
- On-chain settlement is withdrawal-only: custodial ledger balances are the default, and a player who withdraws receives WGLD from the settlement wallet, which is funded from the Rewards Vault.
