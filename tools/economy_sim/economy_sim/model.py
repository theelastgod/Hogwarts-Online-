"""Seasonal WGLD emission and sink model.

The model is deliberately simple and auditable. Each season:

1. Player population grows by a configurable rate.
2. A fraction of players opt into the earn layer.
3. Emissions are capped per season and taper as population grows.
4. Sinks (marketplace fees, minting, breeding, upkeep) burn tokens.
5. Circulating supply and an implied price index are tracked.

The price index is not a market forecast. It is a demand-over-supply ratio
used to compare scenarios. A flat index means supply growth matches demand.
"""

from __future__ import annotations

import random
from dataclasses import dataclass, field


@dataclass
class EconomyConfig:
    initial_players: int = 10_000
    player_growth_per_season: float = 0.15
    earn_opt_in: float = 0.30
    seasons: int = 8

    # Supply minted at token generation for liquidity, founder packs, and reserves.
    genesis_supply: float = 5_000_000.0

    # Emissions
    base_emission_per_season: float = 1_000_000.0
    # Emission scales with sqrt(population / initial) so per-player rewards taper.
    emission_taper_exponent: float = 0.5
    # Fraction of a season's pool that goes unclaimed and returns to treasury.
    unclaimed_fraction: float = 0.08

    # Sinks, expressed as fraction of circulating supply burned per season.
    marketplace_fee_burn: float = 0.05
    mint_burn: float = 0.04
    breeding_burn: float = 0.03
    upkeep_burn: float = 0.02

    # Demand proxy: tokens each earn-enabled player wants to hold or spend per season.
    demand_per_earner: float = 40.0

    # Vesting: payouts unlock over this many seasons (1 = immediate).
    vesting_seasons: int = 1

    seed: int | None = None
    noise: float = 0.0  # Std-dev of multiplicative noise on emissions and sinks.


@dataclass
class SeasonResult:
    season: int
    players: int
    earners: int
    emitted: float
    burned: float
    circulating: float
    treasury: float
    price_index: float
    per_earner_reward: float


@dataclass
class SimulationReport:
    config: EconomyConfig
    seasons: list[SeasonResult] = field(default_factory=list)

    @property
    def max_inflation(self) -> float:
        """Largest season-over-season growth in circulating supply."""
        worst = 0.0
        for prev, cur in zip(self.seasons, self.seasons[1:]):
            if prev.circulating > 0:
                worst = max(worst, cur.circulating / prev.circulating - 1)
        return worst

    @property
    def price_drift(self) -> float:
        """Ratio of final to first price index. 1.0 means stable."""
        if not self.seasons or self.seasons[0].price_index == 0:
            return 0.0
        return self.seasons[-1].price_index / self.seasons[0].price_index

    def passes(self, max_inflation: float = 0.25, collapse_floor: float = 0.5) -> bool:
        """Supply must not inflate faster than max_inflation in any season, and the
        price index must not fall below collapse_floor of its starting value.

        Upward drift is allowed: tapered emissions under player growth are meant
        to put mild deflationary pressure on the token. Runaway upward drift is a
        player-reward problem (per_earner_reward shrinking), tracked separately.
        """
        return self.max_inflation <= max_inflation and self.price_drift >= collapse_floor


def _noisy(value: float, rng: random.Random, noise: float) -> float:
    if noise <= 0:
        return value
    return max(0.0, value * rng.gauss(1.0, noise))


def simulate(config: EconomyConfig) -> SimulationReport:
    rng = random.Random(config.seed)
    report = SimulationReport(config=config)

    players = float(config.initial_players)
    circulating = config.genesis_supply
    treasury = 0.0
    pending_vest: list[float] = []

    for season in range(1, config.seasons + 1):
        earners = int(players * config.earn_opt_in)

        scale = (players / config.initial_players) ** config.emission_taper_exponent
        pool = _noisy(config.base_emission_per_season * scale, rng, config.noise)
        unclaimed = pool * config.unclaimed_fraction
        claimed = pool - unclaimed
        treasury += unclaimed

        # Vesting spreads this season's claimed rewards over future seasons.
        pending_vest.append(claimed)
        per_season_slices = [amt / config.vesting_seasons for amt in pending_vest]
        released = sum(per_season_slices)
        pending_vest = [amt - s for amt, s in zip(pending_vest, per_season_slices)]
        pending_vest = [amt for amt in pending_vest if amt > 1e-9]

        circulating += released

        burn_rate = (
            config.marketplace_fee_burn
            + config.mint_burn
            + config.breeding_burn
            + config.upkeep_burn
        )
        burned = _noisy(circulating * burn_rate, rng, config.noise)
        burned = min(burned, circulating)
        circulating -= burned

        demand = earners * config.demand_per_earner
        price_index = demand / circulating if circulating > 0 else 0.0

        report.seasons.append(
            SeasonResult(
                season=season,
                players=int(players),
                earners=earners,
                emitted=released,
                burned=burned,
                circulating=circulating,
                treasury=treasury,
                price_index=price_index,
                per_earner_reward=released / earners if earners else 0.0,
            )
        )

        players *= 1 + config.player_growth_per_season

    return report
