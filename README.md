# nexus-bot-swarm

Personal project to play around with Nexus Testnet III. I wanted to experiment with creating a swarm of bots that could run concurrently and send real transactions on the testnet.

The bots run a simulated AMM (constant product formula) locally while also sending real transactions to the blockchain every few seconds. It's a playground for learning about Go concurrency, hexagonal architecture, and EVM interactions.

## What it does

- Connects to Nexus Testnet III
- Runs multiple bots concurrently (goroutines)
- Each bot performs simulated swaps on a local AMM pool
- Optionally sends real transactions (self-transfers) to the testnet
- All transactions are verifiable on [Blockscout](https://nexus.testnet.blockscout.com)

## Setup

```bash
go mod tidy
cp .env.example .env
# edit .env with your wallet address and private key
make run
```

## Commands

```bash
make run        # run the bot swarm
make test       # run all tests
make test-race  # run tests with race detector
make check-env  # verify .env configuration
```

## Config

Copy `.env.example` to `.env`:

```bash
NEXUS_RPC_URL=https://testnet.rpc.nexus.xyz
NEXUS_CHAIN_ID=3945
BOT_COUNT=3
WALLET_ADDRESS=0xYourAddress
NEXUS_PRIVATE_KEY=your_private_key_without_0x
```

Without private key, bots run in simulation mode only.

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

## Nexus Testnet III

- Chain ID: 3945
- RPC: https://testnet.rpc.nexus.xyz
- Explorer: https://nexus.testnet.blockscout.com
- Faucet: https://faucet.nexus.xyz
