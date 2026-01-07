# nexus-bot-swarm

Bot swarm for Nexus Testnet III. Custom AMM + send/swap.

## Setup

```bash
go mod tidy
make run
make test
```

## Config

Connects to testnet by default. Override if needed:

```bash
NEXUS_RPC_URL=https://testnet.rpc.nexus.xyz
NEXUS_CHAIN_ID=3945
BOT_COUNT=3
```

## Structure

```
cmd/bot/          - entry point
internal/
  config/         - env vars
  ports/          - interfaces
  adapters/nexus/ - RPC client
```

## TODO

- wallet + balance
- send tx
- simulated AMM
- swarm with goroutines
- on-chain AMM
