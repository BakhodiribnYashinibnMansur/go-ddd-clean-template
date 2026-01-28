# Performance Testing Suite

This directory contains performance test scripts written in **Go** using the [Vegeta](https://github.com/tsenart/vegeta) library.

## Prerequisites

- **Go 1.22+**: You must have Go installed.
- **Vegeta Library**: Dependencies will be handled by `go.mod`.

## Directory Structure

- `stress/`: Tests to find the system's breaking point.
- `spike/`: Tests to check system recovery from sudden traffic bursts.
- `breakpoint/`: Tests to determine maximum sustainable RPS (Capacity).
- `endurance/`: Tests to check for memory leaks and long-term stability (Soak).

## How to Run Tests

You can run any test using `go run`. The tests target `/api/v1/health/ready` by default. You can override the base URL using the `BASE_URL` environment variable.

### 1. Stress Test
Ramps up load in stages (50 -> 400 RPS).
```bash
go run test/performance/stress/main.go
```

### 2. Spike Test
Simulates a sudden extreme burst (Warmup -> 2000 RPS Spike -> Recovery).
```bash
go run test/performance/spike/main.go
```

### 3. Breakpoint (Capacity) Test
Linearly increases RPS (step 100) until latency degradation or failures occur.
```bash
go run test/performance/breakpoint/main.go
```

### 4. Endurance (Soak) Test
Maintains a constant load for a set duration (Default: 6h).
```bash
# Run for 6 hours (default)
go run test/performance/endurance/main.go

# Custom duration and RPS
go run test/performance/endurance/main.go -duration=30m -rps=500
```

### Configuration example
```bash
BASE_URL=http://staging-api.example.com go run test/performance/stress/main.go
```
