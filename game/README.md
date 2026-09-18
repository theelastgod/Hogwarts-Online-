# Game (Unreal Engine 5)

The Unreal project lives here. Binary assets are tracked in Perforce, not git; this folder holds the C++ source, config, and Perforce mapping.

## Setup

1. Install Unreal Engine 5.5 or later with the Linux dedicated server target.
2. Sync `//HogwartsOnline/main/...` from Perforce into `game/HogwartsOnline/Content`.
3. Generate project files and open `HogwartsOnline.uproject`.

## Modules (planned)

| Module | Purpose |
|---|---|
| `HOCore` | Shared types, gameplay tags, ledger client interface |
| `HOCombat` | Spell system, targeting, ancient magic, combo rules |
| `HOTraversal` | Broom flight, mounts, climbing, third-person camera |
| `HOWorld` | World Partition setup, streaming, day/night, instancing |
| `HOEconomy` | Client for `services/ledger`; never authoritative |
| `HOServer` | Dedicated server game mode, anti-cheat hooks, replication |

## Phase 0 spike: 150-player shard

Goal: 150 simulated clients in one World Partition map at 30 server ticks per second with replication bandwidth under 40 KB/s per client.

Steps:

1. Blank UE5 project with Iris replication enabled.
2. Third-person character with a stub spell (projectile) and broom movement mode.
3. Headless client bot that wanders and casts on a timer.
4. Run 150 bots against one dedicated server on a 16-core box. Record tick time, bandwidth, and relevancy stats with `stat net` and Unreal Insights.
5. Report in `docs/tech/networking_spike.md`.

Decision criteria: if tick time exceeds 33 ms at 150 clients, reduce shard size to 100 and plan layered instancing earlier.
