"""CLI: python3 -m economy_sim --players 10000 --seasons 8 --growth 0.15"""

from __future__ import annotations

import argparse
import json

from .model import EconomyConfig, simulate


def main() -> int:
    parser = argparse.ArgumentParser(description="Simulate WGLD emissions and sinks.")
    parser.add_argument("--players", type=int, default=10_000)
    parser.add_argument("--seasons", type=int, default=8)
    parser.add_argument("--growth", type=float, default=0.15, help="Player growth per season")
    parser.add_argument("--opt-in", type=float, default=0.30, help="Earn-layer opt-in fraction")
    parser.add_argument("--emission", type=float, default=1_000_000.0)
    parser.add_argument("--vesting", type=int, default=1)
    parser.add_argument("--noise", type=float, default=0.0)
    parser.add_argument("--seed", type=int, default=None)
    parser.add_argument("--json", action="store_true", help="Emit JSON instead of a table")
    args = parser.parse_args()

    config = EconomyConfig(
        initial_players=args.players,
        seasons=args.seasons,
        player_growth_per_season=args.growth,
        earn_opt_in=args.opt_in,
        base_emission_per_season=args.emission,
        vesting_seasons=args.vesting,
        noise=args.noise,
        seed=args.seed,
    )
    report = simulate(config)

    if args.json:
        print(json.dumps([s.__dict__ for s in report.seasons], indent=2))
    else:
        print(f"{'S':>2} {'players':>9} {'earners':>8} {'emitted':>12} {'burned':>12} "
              f"{'circ':>13} {'treasury':>11} {'price_idx':>10} {'per_earner':>11}")
        for s in report.seasons:
            print(f"{s.season:>2} {s.players:>9} {s.earners:>8} {s.emitted:>12,.0f} "
                  f"{s.burned:>12,.0f} {s.circulating:>13,.0f} {s.treasury:>11,.0f} "
                  f"{s.price_index:>10.4f} {s.per_earner_reward:>11.2f}")
        print()
        print(f"max season inflation: {report.max_inflation:.1%}")
        print(f"price index drift:    {report.price_drift:.2f}x")
        print(f"within bounds:        {'yes' if report.passes() else 'NO'}")

    return 0 if report.passes() else 1


if __name__ == "__main__":
    raise SystemExit(main())
