from economy_sim import EconomyConfig, simulate


def minting(**kw):
    """Legacy unbounded-mint mode (rewards_vault = 0)."""
    kw.setdefault("rewards_vault", 0.0)
    kw.setdefault("genesis_supply", 5_000_000.0)
    kw.setdefault("base_emission_per_season", 1_000_000.0)
    kw.setdefault("demand_per_earner", 40.0)
    return EconomyConfig(**kw)


# --- Fixed-supply (launchpad) mode: the default ---

def test_baseline_stays_within_bounds():
    report = simulate(EconomyConfig(seed=1))
    assert len(report.seasons) == 8
    assert report.passes(), (report.max_inflation, report.price_drift)


def test_emission_never_exceeds_season_cap():
    report = simulate(EconomyConfig(player_growth_per_season=0.93))
    for s in report.seasons:
        assert s.emitted <= s.season_cap + 1e-6, (s.season, s.emitted, s.season_cap)


def test_vault_lasts_planned_32_seasons():
    report = simulate(EconomyConfig(seasons=32, player_growth_per_season=0.93))
    assert report.vault_exhausted_season is None
    assert report.seasons[-1].vault_remaining > 0


def test_vault_cannot_go_negative_when_small():
    report = simulate(EconomyConfig(seasons=6, rewards_vault=30_000_000.0))
    for s in report.seasons:
        assert s.vault_remaining >= -1e-6
    assert report.vault_exhausted_season is not None


def test_unclaimed_allowance_returns_to_vault():
    cfg = EconomyConfig(seasons=1, unclaimed_fraction=0.1, player_growth_per_season=0.0)
    report = simulate(cfg)
    s = report.seasons[0]
    # cap 20M, 10% unclaimed -> 18M leaves the vault.
    assert abs((cfg.rewards_vault - s.vault_remaining) - 18_000_000) < 1e-3
    assert s.treasury == 0


def test_ten_x_growth_keeps_price_index_from_collapsing():
    report = simulate(EconomyConfig(player_growth_per_season=0.39))
    assert report.price_drift > 0.5, report.price_drift


# --- Minting mode still works for comparison ---

def test_minting_mode_supply_converges():
    report = simulate(minting(seasons=40, player_growth_per_season=0.0))
    late = report.seasons[-5:]
    growth = late[-1].circulating / late[0].circulating - 1
    assert abs(growth) < 0.01


def test_minting_mode_unclaimed_goes_to_treasury():
    report = simulate(minting(seasons=1, unclaimed_fraction=0.1))
    assert abs(report.seasons[0].treasury - 100_000) < 1e-6


# --- Shared mechanics ---

def test_vesting_delays_emissions():
    fast = simulate(EconomyConfig(seasons=3))
    slow = simulate(EconomyConfig(seasons=3, vesting_seasons=3))
    assert slow.seasons[0].emitted < fast.seasons[0].emitted


def test_seeded_noise_is_deterministic():
    a = simulate(EconomyConfig(seed=7, noise=0.1))
    b = simulate(EconomyConfig(seed=7, noise=0.1))
    assert [s.circulating for s in a.seasons] == [s.circulating for s in b.seasons]
