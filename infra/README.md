# Infrastructure

Planned layout. Populated in Phase 1 once the cloud provider and chain are chosen.

```
infra/
  terraform/        VPC, Kubernetes clusters (per region), PostgreSQL, Redis, ClickHouse
  k8s/
    agones/         Game server fleet definitions and autoscalers
    services/       Ledger, accounts, matchmaking, marketplace deployments
    observability/  Grafana, Prometheus, Loki, economy dashboards
  ci/               Reusable workflow templates
```

## Regions at launch

North America East, Europe West, Asia Pacific (Tokyo). Each region runs its own game server fleet. Ledger and marketplace are single-region primary with read replicas.

## Economy dashboards (required before closed alpha)

- Emission vs cap per season, real time.
- Burn by sink type per day.
- Circulating supply and top-100 holder concentration.
- Bot-flag rate on earn-enabled accounts.
- Marketplace volume vs monthly actives.
