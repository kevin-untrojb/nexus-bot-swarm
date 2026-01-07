.PHONY: run test build clean env check-env

run:
	CGO_ENABLED=0 go run ./cmd/bot/

test:
	CGO_ENABLED=0 go test ./... -v

test-race:
	go test ./... -race -v

build:
	CGO_ENABLED=0 go build -o bin/bot ./cmd/bot/

clean:
	rm -rf bin/

# Create .env from template if not exists
env:
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "✅ Created .env from .env.example"; \
		echo "📝 Edit .env and add your WALLET_ADDRESS and NEXUS_PRIVATE_KEY"; \
	else \
		echo "⚠️  .env already exists"; \
	fi

# Check if .env has required values
check-env:
	@echo "Checking .env configuration..."
	@grep -q "WALLET_ADDRESS=0x" .env && echo "✅ WALLET_ADDRESS set" || echo "❌ WALLET_ADDRESS not set"
	@grep -q "NEXUS_PRIVATE_KEY=." .env && grep -v "NEXUS_PRIVATE_KEY=$$" .env | grep -q "NEXUS_PRIVATE_KEY" && echo "✅ NEXUS_PRIVATE_KEY set" || echo "❌ NEXUS_PRIVATE_KEY not set"
