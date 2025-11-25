# Load Testing with k6

## Prerequisites

Install k6:
```bash
# macOS
brew install k6

# Linux
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Or download from https://k6.io/docs/getting-started/installation/
```

## Running Load Tests

### Order Service Load Test

```bash
# Set environment variables
export BASE_URL=http://localhost:8080
export ACCESS_TOKEN=your-jwt-token

# Run load test
k6 run tests/load/orders.js
```

## Test Scenarios

### 1. Ramp-up Test
- 0 → 10,000 users over 5 minutes
- Tests system's ability to scale up

### 2. Sustained Load Test
- 10,000 users for 30 minutes
- Tests system stability under constant load

### 3. Spike Test
- 50,000 users for 5 minutes
- Tests system's ability to handle traffic spikes

## Success Criteria

- P95 latency < 200ms
- Error rate < 1%
- No memory leaks
- System remains stable

## Custom Metrics

- `order_latency`: Latency for order creation
- `errors`: Error rate

