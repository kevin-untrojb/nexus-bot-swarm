# nexus-bot-swarm

Personal project to play around with Nexus Testnet III. I wanted to experiment with creating a swarm of bots that could run concurrently and send real transactions on the testnet.

The bots run a simulated AMM (constant product formula) locally while also sending real transactions to the blockchain every few seconds. It's a playground for learning about Go concurrency, hexagonal architecture, and EVM interactions.

## What it does

- Connects to Nexus Testnet III
- Runs multiple bots concurrently (goroutines)
- Each bot performs simulated swaps on a local AMM pool
- Sends real ERC20 token transfers (KEVZ) or native NEX transfers
- All transactions are verifiable on [Blockscout](https://nexus.testnet.blockscout.com)

## KevzToken (ERC20)

I deployed my own ERC20 token on Nexus Testnet to make things more interesting. The bots can transfer KEVZ tokens instead of just sending NEX.

**Contract**: `0x5346201A0c79E23C600D8420510b2A7aC02c53f7`

### Deploy your own token

1. Go to [Remix IDE](https://remix.ethereum.org)
2. Create a new file `yourToken.sol` with the code from `contracts/yourToken.sol`
3. Compile with Solidity 0.8.20+
4. In "Deploy & Run":
   - Environment: **Injected Provider - MetaMask**
   - Make sure MetaMask is connected to Nexus Testnet (Chain ID 3945)
   - Click **Deploy**
5. Copy the deployed contract address
6. Add it to your `.env` as `TOKEN_ADDRESS`

The contract mints 1,000,000 KEVZ tokens to the deployer wallet.

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
TOKEN_ADDRESS=0xYourTokenContract  # optional, for ERC20 transfers
```

**Modes:**
- Without `NEXUS_PRIVATE_KEY`: Simulation only (no real transactions)
- With `NEXUS_PRIVATE_KEY` but no `TOKEN_ADDRESS`: Sends 1 wei NEX to self
- With both: Transfers 1 KEVZ token to self (visible as "Token Transfer" in explorer)

## Structure (Hexagonal)

```
cmd/bot/              - entry point
domain/               - core business logic (AMM)
ports/                - interfaces
swarm/                - application layer (bot orchestration)
contracts/            - Solidity smart contracts (KevzToken ERC20)
internal/
  config/             - env vars
  adapters/nexus/     - RPC client (NEX + ERC20)
  nonce/              - concurrent nonce manager
```

## Nexus Testnet III

- Chain ID: 3945
- RPC: https://testnet.rpc.nexus.xyz
- Explorer: https://nexus.testnet.blockscout.com
- Faucet: https://faucet.nexus.xyz
