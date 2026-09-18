from economy_sim import EconomyConfig, simulate


def test_baseline_stays_within_bounds():
    report = simulate(EconomyConfig(seed=1))
    assert len(report.seasons) == 8
    assert report.passes(), (report.max_inflation, report.price_drift)


def test_supply_converges_because_sinks_scale_with_supply():
    report = simulate(EconomyConfig(seasons=40, player_growth_per_season=0.0))
    late = report.seasons[-5:]
    growth = late[-1].circulating / late[0].circulating - 1
    assert abs(growth) < 0.01


def test_ten_x_growth_keeps_price_index_from_collapsing():
    # 10x population over 8 seasons (7 growth steps) is 10**(1/7)-1, about 39%.
    report = simulate(EconomyConfig(player_growth_per_season=0.39))
    assert report.price_drift > 0.5, report.price_drift


def test_hundred_x_growth_scenario_reports_rather_than_crashes():
    # 100x over 7 growth steps is 100**(1/7)-1, about 93%.
    report = simulate(EconomyConfig(player_growth_per_season=0.93))
    assert report.seasons[-1].players > 100 * report.seasons[0].players * 0.9
    assert report.seasons[-1].circulating > 0


def test_vesting_delays_emissions():
    fast = simulate(EconomyConfig(seasons=3))
    slow = simulate(EconomyConfig(seasons=3, vesting_seasons=3))
    assert slow.seasons[0].emitted < fast.seasons[0].emitted


def test_unclaimed_rewards_go_to_treasury():
    report = simulate(EconomyConfig(seasons=1, unclaimed_fraction=0.1))
    assert abs(report.seasons[0].treasury - 100_000) < 1e-6


def test_seeded_noise_is_deterministic():
    a = simulate(EconomyConfig(seed=7, noise=0.1))
    b = simulate(EconomyConfig(seed=7, noise=0.1))
    assert [s.circulating for s in a.seasons] == [s.circulating for s in b.seasons]
