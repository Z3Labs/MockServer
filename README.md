# MockServer

MockServer for intelligent release system testing. This server simulates various abnormal scenarios to test AI-based diagnosis and troubleshooting capabilities.

## Features

### P0 Scenarios (Core)
- **CPU Burner**: Increases CPU usage to specified percentage
- **Memory Leaker**: Continuously leaks memory at specified rate
- **Network Latency**: Adds specified latency to HTTP requests
- **Health Check Failure**: Controls health check endpoints to return failures

### P1 Scenarios (Common)
- **Goroutine Leak**: Creates goroutines that never exit
- **Disk IO**: Generates high disk IO
- **Crash Simulator**: Simulates service crash after delay
- **Dependency Failure**: Simulates dependency service failures

## Quick Start

### Build

```bash
go build -o mockserver cmd/server/main.go
```

### Run

```bash
./mockserver -f etc/mockserver.yaml
```

### Docker

```bash
docker build -t mockserver:latest .
docker run -p 8888:8888 mockserver:latest
```

## API Usage

### Composite Scenarios (Recommended)

Start multiple scenarios at once. When a new composite scenario is triggered, all previous scenarios are automatically stopped.

#### Start Composite Scenario

```bash
curl -X POST http://localhost:8888/api/v1/composite/start \
  -H "Content-Type: application/json" \
  -d '{
    "scenarios": [
      {
        "name": "cpu_burner",
        "params": {"target_percent": 80},
        "duration": 300
      },
      {
        "name": "memory_leaker",
        "params": {"target_mb": 2048, "leak_rate_mb": 50},
        "duration": 300
      }
    ]
  }'
```

**Note**: The optional `duration` parameter (in seconds) enables auto-recovery. When set, the scenario will automatically stop after the specified duration. If multiple scenarios have different durations, they will all stop when the maximum duration is reached.

#### Stop All Scenarios

```bash
curl -X POST http://localhost:8888/api/v1/composite/stop
```

#### Get Current Session Status

```bash
curl http://localhost:8888/api/v1/composite/status
```

### Individual Scenarios

#### CPU Burner

```bash
curl -X POST http://localhost:8888/api/v1/scenarios/cpu_burner/start \
  -H "Content-Type: application/json" \
  -d '{"target_percent": 80}'
```

**With auto-recovery (stops after 5 minutes):**
```bash
curl -X POST http://localhost:8888/api/v1/scenarios/cpu_burner/start \
  -H "Content-Type: application/json" \
  -d '{"target_percent": 80, "duration": 300}'
```

#### Memory Leaker

```bash
curl -X POST http://localhost:8888/api/v1/scenarios/memory_leaker/start \
  -H "Content-Type: application/json" \
  -d '{"target_mb": 2048, "leak_rate_mb": 50}'
```

#### Network Latency

```bash
curl -X POST http://localhost:8888/api/v1/scenarios/network_latency/start \
  -H "Content-Type: application/json" \
  -d '{"latency_ms": 500}'
```

#### Health Check Failure

```bash
# Always fail
curl -X POST http://localhost:8888/api/v1/scenarios/health_check/start \
  -H "Content-Type: application/json" \
  -d '{"failure_mode": "always", "status_code": 503}'

# Intermittent failure (50% probability)
curl -X POST http://localhost:8888/api/v1/scenarios/health_check/start \
  -H "Content-Type: application/json" \
  -d '{"failure_mode": "intermittent", "fail_rate": 0.5}'

# Delayed response
curl -X POST http://localhost:8888/api/v1/scenarios/health_check/start \
  -H "Content-Type: application/json" \
  -d '{"failure_mode": "delayed"}'
```

#### Goroutine Leak

```bash
curl -X POST http://localhost:8888/api/v1/scenarios/goroutine_leak/start \
  -H "Content-Type: application/json" \
  -d '{"goroutines_per_second": 100}'
```

#### Disk IO

```bash
curl -X POST http://localhost:8888/api/v1/scenarios/disk_io/start \
  -H "Content-Type: application/json" \
  -d '{"write_rate_mb": 100}'
```

#### Crash Simulator

```bash
# Crash after 10 seconds
curl -X POST http://localhost:8888/api/v1/scenarios/crash/start \
  -H "Content-Type: application/json" \
  -d '{"crash_delay": 10}'
```

#### Dependency Failure

```bash
curl -X POST http://localhost:8888/api/v1/scenarios/dependency/start \
  -H "Content-Type: application/json" \
  -d '{"failure_type": "timeout"}'
```

### General APIs

#### List All Scenarios

```bash
curl http://localhost:8888/api/v1/scenarios
```

#### Get Scenario Status

```bash
curl http://localhost:8888/api/v1/scenarios/cpu_burner/status
```

#### Stop Scenario

```bash
curl -X POST http://localhost:8888/api/v1/scenarios/cpu_burner/stop
```

#### Health Check

```bash
curl http://localhost:8888/health
curl http://localhost:8888/ready
```

#### Mock Dependency Service

```bash
curl http://localhost:8888/api/v1/mock-service
```

### Test Endpoints

#### Test Endpoint with 10ms Sleep

```bash
curl http://localhost:8888/api/v1/test/sleep10ms
```

#### Test Endpoint with 30ms Sleep

```bash
curl http://localhost:8888/api/v1/test/sleep30ms
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Mock Server                          │
├─────────────────────────────────────────────────────────┤
│  HTTP API Layer                                          │
│  - Single/Composite scenario control                     │
│  - Status queries                                        │
├─────────────────────────────────────────────────────────┤
│  Scenario Manager                                        │
│  - Scenario lifecycle management                         │
│  - Session management (composite scenarios)              │
│  - Atomic scenario switching                             │
├─────────────────────────────────────────────────────────┤
│  Scenario Plugins                                        │
│  ├─ CPU Burner                                           │
│  ├─ Memory Leaker                                        │
│  ├─ Network Latency                                      │
│  ├─ Health Check Failure                                 │
│  ├─ Goroutine Leak                                       │
│  ├─ Disk IO                                              │
│  ├─ Crash Simulator                                      │
│  └─ Dependency Failure                                   │
└─────────────────────────────────────────────────────────┘
```

## Configuration

Edit `etc/mockserver.yaml`:

```yaml
Name: mockserver
Host: 0.0.0.0
Port: 8888

Log:
  Mode: console
  Level: info
```

## Example: Complex Composite Scenario

```bash
curl -X POST http://localhost:8888/api/v1/composite/start \
  -H "Content-Type: application/json" \
  -d '{
    "scenarios": [
      {
        "name": "cpu_burner",
        "params": {"target_percent": 70},
        "duration": 600
      },
      {
        "name": "memory_leaker",
        "params": {"target_mb": 1024, "leak_rate_mb": 20},
        "duration": 600
      },
      {
        "name": "network_latency",
        "params": {"latency_ms": 300},
        "duration": 600
      },
      {
        "name": "health_check",
        "params": {"failure_mode": "intermittent", "fail_rate": 0.3},
        "duration": 600
      }
    ]
  }'
```

All scenarios will automatically stop after 10 minutes (600 seconds).

## License

MIT
