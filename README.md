# Hogwarts Online

Third-person open-world wizarding RPG with a play-to-earn economy. See [DEVELOPMENT_PLAN.md](DEVELOPMENT_PLAN.md) for the full plan.

## Layout

| Path | Contents |
|---|---|
| `docs/` | Design docs, economy model notes, legal checklists |
| `game/` | Unreal Engine 5 project (see `game/README.md` for setup) |
| `services/ledger/` | Go double-entry economy ledger with emission caps and vesting |
| `contracts/` | Solidity contracts for WGLD and tradable items |
| `tools/economy_sim/` | Python simulation of token emissions, sinks, and price stability |
| `infra/` | Kubernetes, Terraform, and CI configuration |

## Quick start

```
# Economy simulation
cd tools/economy_sim && python3 -m economy_sim --players 10000 --seasons 8

# Ledger service tests
cd services/ledger && go test ./...

# Economy sim tests
cd tools/economy_sim && python3 -m pytest
```

## Status

Phase 0 (Discovery). Current work: economy model, ledger core, contract skeletons, legal rights checklist.
