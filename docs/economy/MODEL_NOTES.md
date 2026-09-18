# Economy Model Notes

Living notes on the WGLD model in `tools/economy_sim`. Update after each simulation change.

## Model summary

- Circulating supply starts at a genesis allocation (default 5M WGLD) for liquidity, founder packs, and reserves.
- Each season emits a pool that scales with the square root of population growth. Per-player rewards therefore taper as the game grows.
- Sinks burn a fixed fraction of circulating supply each season (marketplace fees, minting, breeding, upkeep; 14 percent combined by default).
- Because sinks scale with supply, circulating supply converges to `emission / burn_rate` when population is flat. The model verifies this.
- The price index is a demand-over-supply ratio used only to compare scenarios.

## Findings so far (2026-09-18)

| Scenario | Max season inflation | Price drift over 8 seasons | Within bounds |
|---|---|---|---|
| Baseline, 15% growth per season | 5.8% | 1.96x | yes |
| 10x players over 8 seasons | 14.7% | 5.1x | yes |
| 100x players over 8 seasons | 34.9% | 23x | no (inflation ceiling) |

Interpretation:

1. **No collapse scenario found.** With sinks proportional to supply, emissions cannot outrun burns for long. The token does not inflate away under any tested growth path.
2. **The real risk under fast growth is the opposite: rewards per earner shrink.** Square-root taper means a 10x population gets about 3x the pool. Per-earner reward falls roughly 70 percent. That is a retention problem, not a token-value problem.
3. **Genesis supply matters.** Without it, early seasons inflate from zero and every scenario reads as a failure. The 5M default should be revisited once the liquidity plan is set.

## Open questions

- Should emission taper be tied to earners rather than total players so opt-in changes do not distort rewards?
- Add a secondary-market model (buyers who are not earners) so demand is not purely earner-driven.
- Model the 30-day vesting at day granularity to see its effect on sell pressure after a season ends.
- Add bot and multi-account leakage as a parameter and test detection thresholds.

## How to run

```
cd tools/economy_sim
python3 -m economy_sim                       # baseline
python3 -m economy_sim --growth 0.39         # 10x
python3 -m economy_sim --growth 0.93         # 100x
python3 -m economy_sim --vesting 3 --json    # vesting scenario, JSON out
python3 -m pytest                            # invariants
```
