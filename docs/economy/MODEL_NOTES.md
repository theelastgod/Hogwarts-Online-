# Economy Model Notes

Living notes on the WGLD model in `tools/economy_sim`. Update after each simulation change.

## Model summary

The model runs in **fixed-supply mode** by default, matching the launchpad tokenomics in `TOKENOMICS.md`. Setting `--vault 0` switches to the earlier unbounded-mint mode for comparison.

- Circulating supply starts at the TGE unlock (126M WGLD, 12.6% of 1B).
- A 400M Rewards Vault funds seasonal emissions. Each season's release is the lesser of the decaying season cap (20M, minus 4% per season), the demand-driven pool, and what the vault holds. Unclaimed allowance stays in the vault.
- Sinks burn a fixed fraction of circulating supply each season (14 percent combined by default).
- The price index is a demand-over-supply ratio used only to compare scenarios.

## Findings (2026-09-18, fixed-supply mode)

| Scenario | Max season inflation | Price drift | Vault exhausted | Per-earner reward, season 1 to 8 |
|---|---|---|---|---|
| Baseline, 15% growth per season, 8 seasons | 0% | 3.2x | no | 6,133 to 1,733 |
| 10x players over 8 seasons | 0% | large upward | no | falls about 90% |
| 100x growth rate held for 32 seasons | 4.9% | large upward | no | falls to near zero |

Interpretation:

1. **Supply inflation is no longer a risk.** The vault cap binds in every scenario, so circulating supply never grows faster than the schedule. In the baseline it actually shrinks, because burns on 126M exceed the 18M net release.
2. **The trade-off is per-player rewards.** With a fixed cap, a growing player base splits a shrinking pie. Baseline rewards per earner fall about 70 percent over two years. Under 10x growth they fall about 90 percent. This is the number to design around: rewards must feel meaningful in USD terms at the player counts we expect, which depends on token price, not on emission tuning.
3. **The vault outlasts the plan.** With decaying caps, 32 seasons consume about 365M of 400M even under extreme growth. There is room to raise early caps if rewards prove too thin, but on-chain caps can only be lowered, so the initial schedule should be set generously and lowered by governance as needed.
4. **Burns exceed releases early.** With 14% burn on 126M circulating and 18M released, supply contracts for the first several seasons. If that proves too aggressive for market liquidity, lower the marketplace burn share rather than raising emissions.

## Findings (earlier, unbounded-mint mode, kept for comparison)

| Scenario | Max season inflation | Price drift over 8 seasons | Within bounds |
|---|---|---|---|
| Baseline, 15% growth per season | 5.8% | 1.96x | yes |
| 10x players over 8 seasons | 14.7% | 5.1x | yes |
| 100x players over 8 seasons | 34.9% | 23x | no (inflation ceiling) |

## Open questions

- Should emission taper be tied to earners rather than total players so opt-in changes do not distort rewards?
- Add a secondary-market model (buyers who are not earners) so demand is not purely earner-driven.
- Model the 30-day vesting at day granularity to see its effect on sell pressure after a season ends.
- Add bot and multi-account leakage as a parameter and test detection thresholds.

## How to run

```
cd tools/economy_sim
python3 -m economy_sim                       # baseline, fixed-supply
python3 -m economy_sim --seasons 32          # full vault schedule
python3 -m economy_sim --growth 0.39         # 10x over 8 seasons
python3 -m economy_sim --vault 0             # unbounded-mint comparison
python3 -m economy_sim --vesting 3 --json    # vesting scenario, JSON out
python3 -m pytest                            # invariants
```
