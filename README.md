# nexus-bot-swarm

Bot swarm for Nexus Testnet III. Custom AMM + send/swap.

## Setup

```bash
go mod tidy
cp .env.example .env
# edit .env with your wallet address
make run
make test
```

## Config

Copy `.env.example` to `.env` and set your values:

```bash
NEXUS_RPC_URL=https://testnet.rpc.nexus.xyz
NEXUS_CHAIN_ID=3945
BOT_COUNT=3
WALLET_ADDRESS=0xYourAddressHere
```

The `.env` file is gitignored.

## Structure (Hexagonal)

```
cmd/bot/              - entry point
domain/               - core business logic (AMM)
ports/                - interfaces
swarm/                - application layer (bot orchestration)
internal/
  config/             - env vars
  adapters/nexus/     - RPC client
```

## TODO

- wallet + balance
- send tx
- simulated AMM
- swarm with goroutines
- on-chain AMM
