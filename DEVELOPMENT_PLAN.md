# Hogwarts Online: Development Plan

A third-person, open-world wizarding RPG with a play-to-earn (P2E) economy. Single-player quality on the level of Hogwarts Legacy, with persistent shared servers, player-driven markets, and real-value rewards.

Assumptions this plan relies on:

- The studio holds full rights to the Harry Potter / Wizarding World IP for interactive entertainment, including online, tokenized, and real-money economies. Legal counsel should confirm those rights cover the P2E model in every launch territory before Phase 1 ends.
- Target platforms are PC first, then PlayStation 5 and Xbox Series X|S. Mobile is a stretch goal.
- Team scale is a mid-size studio (60 to 120 people at peak) with outsourcing for art and QA.

---

## 1. Vision and Pillars

**One line:** Live your own wizarding life at Hogwarts, where skill, craft, and reputation earn you real rewards.

Pillars:

1. **Third-person mastery.** Fluid spell combat, broom and mount traversal, and stealth, all built around a responsive over-the-shoulder camera.
2. **Living Hogwarts.** The castle, Hogsmeade, the Forbidden Forest, and the Highlands are shared spaces where players see each other, form houses, and compete.
3. **Earn through play, not pay.** Rewards flow from skill, crafting, exploration, and social contribution. No pay-to-win.
4. **Sustainable economy.** Every earned asset has a sink. The economy is modelled and monitored before and after launch.

---

## 2. Core Game Design

### 2.1 Player Fantasy and Loop

Players create a witch or wizard, get sorted, and progress through years at Hogwarts. Moment-to-moment: explore, take classes, duel, brew, tame beasts, and complete quests. Session-to-session: level spells and professions, build a Room of Requirement, and climb house and guild rankings. Long-term: master a specialization, own rare items, and earn a reputation on the server.

### 2.2 Combat

- Over-the-shoulder camera, lock-on optional.
- Spell wheel with 4 slots per set, up to 4 sets. Spells fall into Control, Force, Damage, Curse, and Utility categories with combo rules.
- Ancient magic meter for finishers.
- PvP duels: 1v1 ranked, 3v3 arenas, and house-versus-house battlegrounds on a schedule.
- PvE: Dark wizard camps, dungeons (vaults, caves), and world bosses (trolls, dragons).

### 2.3 Systems

| System | Description | Economy role |
|---|---|---|
| Classes | Timed lessons that unlock spells and passive perks | Unlocks; small daily earnings |
| Potions and Herbology | Gather, grow, brew | Primary crafting sink and source |
| Beasts | Rescue, care, breed, harvest materials | Rare-material source; breeding market |
| Room of Requirement | Player-owned instanced space, upgradeable | Cosmetic and functional sink |
| Broom racing | Time trials and multiplayer races | Skill-based reward pool |
| Quidditch | 7v7 team matches, seasonal league | Team rewards, spectator economy |
| Housing and guilds | House points, guild halls, shared vaults | Social sink |
| Exploration | Field guide pages, Merlin trials, collectibles | Account progression |

### 2.4 Progression

- Character level 1 to 50 across Years 1 to 7 of content at launch.
- Talent trees per spell category plus profession trees (Potioneer, Herbologist, Beastkeeper, Duelist, Racer).
- Gear: robes, wands (core and wood affect stats), and trinkets with rarity tiers.

---

## 3. Play-to-Earn Economy

### 3.1 Principles

1. Earning requires effort, skill, or scarcity. No idle earnings.
2. Every token faucet is paired with a token sink that is measurable in dashboards.
3. Cosmetics and convenience are sold for fiat. Power is never sold.
4. Players can fully enjoy the game without ever touching the earn layer. The earn layer is opt-in and gated by KYC (know your customer) where required.

### 3.2 Currency Model

| Currency | Type | Source | Sink |
|---|---|---|---|
| Galleons | Soft, in-game only | Quests, loot, vendors | Vendors, repairs, room upgrades |
| Sickles of Merit | Bound reputation points | House contributions, events | Ranks, titles, class access |
| Wizarding Gold (WGLD) | Tradable premium token, on-chain | Ranked play, seasonal leagues, rare crafting, tournaments | Marketplace fees, minting, breeding, guild hall upkeep |

WGLD emission is capped per season and split across activity pools. A pool that goes unclaimed rolls back into the treasury. Emissions taper as the player base grows to keep per-token value stable.

### 3.3 Tradable Assets

- Rare wands, robes, brooms, and beasts are mintable as on-chain items only when they reach a rarity threshold and the owner chooses to mint.
- Minting costs WGLD and burns the in-game copy on transfer.
- Royalties: 5 percent of secondary sales go to the treasury, 2.5 percent to the original crafter.
- A player-run marketplace with escrow, price history, and bid systems.

### 3.4 Earn Loops

- **Ranked duels and Quidditch.** Seasonal ladders pay WGLD from a fixed pool weighted by rank and match count.
- **Crafting.** High-tier potions and gear consumed by other players. Crafters earn through sales, not emissions.
- **Beast breeding.** Rare traits are inheritable. Breeding costs WGLD and has a cooldown.
- **Tournaments.** Sponsored and community events with entry fees and prize pools.
- **Content creation.** In-game challenge and dungeon editors with revenue share for popular creations (post-launch).

### 3.5 Anti-Abuse

- Server-side simulation for all rewards. No client trust.
- Device and behaviour fingerprinting, bot detection, and rate limits on earn actions.
- One earn-enabled account per verified identity.
- Reward vesting: seasonal payouts unlock over 30 days to dampen farm-and-dump.

### 3.6 Blockchain Approach

- Custodial wallets by default so players never touch seed phrases. Optional external wallet linking.
- Layer 2 EVM chain (Base, Arbitrum, or Polygon) for low fees. Abstract the chain behind an internal ledger so it can change.
- On-chain writes are batched and asynchronous. Gameplay never blocks on the chain.
- Smart contracts audited by two independent firms before mainnet.

### 3.7 Regulatory

- Engage securities, gaming, and consumer-protection counsel in the US, UK, EU, Japan, and Korea during pre-production.
- Age gate the earn layer to 18+. The core game remains rated for teens.
- Platform holder policy review: console platforms restrict crypto integration. Plan a console SKU where the earn layer is withdrawal-only via a companion web app if required.
- Build tax reporting exports for players from day one.

---

## 4. Technical Plan

### 4.1 Engine and Stack

| Layer | Choice | Reason |
|---|---|---|
| Engine | Unreal Engine 5 (Nanite, Lumen, MetaHumans) | Third-person AAA fidelity, strong networking, Legacy-parity look |
| Server | Dedicated Unreal servers on Kubernetes (Agones for fleet scaling) | Authoritative simulation |
| Backend services | Go or C# microservices, gRPC, PostgreSQL, Redis | Accounts, inventory, matchmaking, economy |
| Economy ledger | Internal double-entry ledger with async chain settlement | Auditability, chain independence |
| Chain | EVM Layer 2, ERC-20 for WGLD, ERC-721/1155 for items | Ecosystem tooling |
| Data | ClickHouse or BigQuery for telemetry, Grafana dashboards | Economy monitoring |
| Anti-cheat | Easy Anti-Cheat plus server-side validation | Earn integrity |
| Voice and chat | Vivox or EOS | Console-compliant |
| Builds and CI | Perforce for assets, Git for code and tools, Jenkins or TeamCity, Horde for UE | Standard for UE at scale |

### 4.2 World Architecture

- Seamless open world using World Partition, streamed in cells.
- Shared shards of 100 to 200 players per region, with instanced dungeons and PvP arenas.
- Layered instancing so the castle feels populated but never overcrowded.
- Cross-play across PC and console with input-based matchmaking.

### 4.3 Repository Layout (this repo)

```
/docs            design docs, economy models, legal notes
/game            Unreal project
/services        backend microservices
/contracts       Solidity contracts and audits
/tools           pipeline and economy simulation tools
/infra           Kubernetes, Terraform, CI
```

---

## 5. Content Scope at Launch

- Hogwarts castle and grounds in full, Hogsmeade, Forbidden Forest, and the Highlands north and south.
- Main story of roughly 25 hours across Years 1 to 5. Years 6 and 7 arrive in live seasons.
- 40 spells, 60 potions, 30 beast species, 120 side quests, 200 collectibles.
- 12 dungeons, 4 world bosses, 6 Quidditch pitches, 10 race tracks.
- 4 house common rooms as social hubs with house-specific quests.

---

## 6. Team and Roles

| Discipline | Peak headcount |
|---|---|
| Production and design | 14 |
| Engineering (gameplay, engine, tools) | 22 |
| Backend, infra, security | 12 |
| Blockchain and economy | 6 |
| Art (environment, character, VFX, animation, UI) | 30 plus outsourcing |
| Audio and narrative | 6 |
| QA | 10 plus outsourcing |
| Live ops, community, support | 10 |
| Legal, compliance, finance | 4 |

---

## 7. Phases and Milestones

### Phase 0: Discovery (Months 1 to 3)

- Lock pillars, economy thesis, and legal posture.
- Economy simulation in a spreadsheet and agent-based model. Prove emission caps hold under 10x and 100x player growth.
- Tech evaluation: UE5 networking spike with 150 players in one shard.
- Deliverable: green-light document, budget, and staffing plan.

### Phase 1: Pre-production (Months 4 to 9)

- Vertical slice: one castle wing, Hogsmeade street, 8 spells, one dungeon, one duel arena, ledger with mock chain.
- Playable at 60 fps on target hardware with 50 concurrent players.
- Contracts drafted and internally tested on testnet.
- Counsel sign-off on the P2E model per launch territory.
- Deliverable: vertical slice review and full production plan.

### Phase 2: Production (Months 10 to 27)

- World build-out in quarterly content drops to internal test.
- Full feature set implemented: classes, potions, beasts, Quidditch, racing, guilds, marketplace.
- Two closed alphas at Month 18 and Month 24 with real economy on testnet.
- Contract audits complete by Month 24.
- Console certification pre-checks at Month 22.

### Phase 3: Beta and Launch (Months 28 to 33)

- Open beta on PC at Month 28 with a wipe-free economy preview under strict caps.
- Console beta at Month 30.
- Mainnet launch of WGLD at Month 31, game launch at Month 33.
- Launch gates: crash-free rate above 99.5 percent, economy inflation within model bounds for 30 days, zero critical audit findings.

### Phase 4: Live Operations (Month 34 onward)

- Seasons every 3 months: new year of content, balance passes, new earn pools.
- Year 1 roadmap: Years 6 and 7 story, guild wars, creator tools, mobile companion.
- Weekly economy review with authority to adjust emissions within pre-approved bands.

---

## 8. Monetization Beyond P2E

- Base game at premium price, or free-to-play with a paid founder pack. Decide after alpha retention data.
- Cosmetic store: robes, wand skins, broom trails, room furniture.
- Season pass with cosmetic and convenience tracks only.
- Marketplace fees and mint fees flow to the treasury.

---

## 9. Risks and Mitigations

| Risk | Mitigation |
|---|---|
| Token value collapse drives players away | Fun-first design, capped emissions, sinks, and a treasury buffer to stabilize pools |
| Regulatory action on P2E | Early counsel, KYC, age gating, territory-specific SKUs |
| Console platform rejects crypto | Withdrawal-only companion app; console build with earn layer disabled |
| Bots and multi-accounting | Server authority, fingerprinting, identity verification, vesting |
| Scope creep from the Legacy comparison | Vertical slice defines quality bar; content is cut, not quality |
| IP fidelity and fan expectation | Lore team and canon review on every quest and asset |
| Smart contract exploit | Two audits, bug bounty, upgradeable proxies with timelock, pause switch |

---

## 10. Success Metrics

- Day 30 retention above 25 percent, Day 90 above 12 percent.
- Median session length above 45 minutes.
- Earn-layer opt-in between 20 and 40 percent of monthly actives.
- WGLD 30-day price volatility below 15 percent after Month 3 of live.
- Bot-flagged accounts below 1 percent of earn-enabled accounts.
- Marketplace volume growing with actives, not faster.

---

## 11. Immediate Next Steps

1. Confirm rights coverage for tokenized economies with counsel and document it in `docs/legal`.
2. Stand up the repository layout above with a UE5 project skeleton and CI.
3. Build the economy simulation model and run the first emission-cap scenarios.
4. Run the 150-player networking spike.
5. Recruit the core leads: creative director, technical director, economy lead, and live-ops lead.
