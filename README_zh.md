# MockServer - 智能发布系统异常场景模拟服务

MockServer 是一个专为智能发布系统设计的异常场景模拟服务。它可以模拟各种服务异常情况，用于测试 AI 系统在发布过程中对异常的研判和诊断能力。

## 核心特性

### P0 场景（核心场景）
- **CPU 占用（CPU Burner）**: 将 CPU 使用率提升到指定百分比
- **内存泄漏（Memory Leaker）**: 按指定速率持续泄漏内存
- **网络延迟（Network Latency）**: 为 HTTP 请求添加指定延迟
- **健康检查失败（Health Check Failure）**: 控制健康检查端点返回失败状态

### P1 场景（常见场景）
- **协程泄漏（Goroutine Leak）**: 创建永不退出的协程
- **磁盘 IO（Disk IO）**: 产生高磁盘 IO 负载
- **崩溃模拟（Crash Simulator）**: 模拟服务延迟崩溃
- **依赖服务失败（Dependency Failure）**: 模拟依赖服务调用失败

## 快速开始

### 构建

```bash
go build -o mockserver cmd/server/main.go
```

### 运行

```bash
./mockserver -f etc/mockserver.yaml
```

### Docker 部署

```bash
docker build -t mockserver:latest .
docker run -p 8888:8888 mockserver:latest
```

## API 使用指南

### 复合场景模式（推荐）

复合场景模式允许一次性启动多个异常场景。当触发新的复合场景时，所有之前运行的场景会自动停止。

#### 启动复合场景

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

**注意**: 可选的 `duration` 参数（单位：秒）用于启用自动恢复功能。设置后，场景将在指定时间后自动停止。如果多个场景有不同的 duration 值，所有场景将在最大 duration 到达时一起停止。

#### 停止所有场景

```bash
curl -X POST http://localhost:8888/api/v1/composite/stop
```

#### 查询当前会话状态

```bash
curl http://localhost:8888/api/v1/composite/status
```

**响应示例：**
```json
{
  "session_id": "session-1698234567",
  "scenarios": ["cpu_burner", "memory_leaker"],
  "status": "success",
  "details": [
    {
      "name": "cpu_burner",
      "success": true
    },
    {
      "name": "memory_leaker",
      "success": true
    }
  ]
}
```

### 单场景模式

#### 1. CPU 占用（cpu_burner）

将 CPU 使用率提升到指定百分比。

**参数说明：**
- `target_percent`: 目标 CPU 占用率（0-100）

**示例：**
```bash
# 将 CPU 占用率提升到 80%
curl -X POST http://localhost:8888/api/v1/scenarios/cpu_burner/start \
  -H "Content-Type: application/json" \
  -d '{"target_percent": 80}'
```

**自动恢复示例（5 分钟后自动停止）：**
```bash
curl -X POST http://localhost:8888/api/v1/scenarios/cpu_burner/start \
  -H "Content-Type: application/json" \
  -d '{"target_percent": 80, "duration": 300}'
```

#### 2. 内存泄漏（memory_leaker）

按指定速率持续增长内存占用，模拟内存泄漏场景。

**参数说明：**
- `target_mb`: 目标内存占用（MB）
- `leak_rate_mb`: 每秒增长速率（MB）

**示例：**
```bash
# 以每秒 50MB 的速率泄漏内存，直到达到 2048MB
curl -X POST http://localhost:8888/api/v1/scenarios/memory_leaker/start \
  -H "Content-Type: application/json" \
  -d '{"target_mb": 2048, "leak_rate_mb": 50}'
```

#### 3. 网络延迟（network_latency）

为所有 HTTP 请求添加指定延迟时间。

**参数说明：**
- `latency_ms`: 延迟时间（毫秒）

**示例：**
```bash
# 为所有请求添加 500ms 延迟
curl -X POST http://localhost:8888/api/v1/scenarios/network_latency/start \
  -H "Content-Type: application/json" \
  -d '{"latency_ms": 500}'
```

#### 4. 健康检查失败（health_check）

控制健康检查端点返回不同的失败状态。

**参数说明：**
- `failure_mode`: 失败模式
  - `always`: 持续返回失败状态
  - `intermittent`: 间歇性失败（按概率随机返回成功或失败）
  - `delayed`: 响应超时（长时间延迟后返回）
- `status_code`: 返回的 HTTP 状态码（默认 503）
- `fail_rate`: 间歇性失败的概率（0.0-1.0）

**示例：**
```bash
# 持续返回 503 失败状态
curl -X POST http://localhost:8888/api/v1/scenarios/health_check/start \
  -H "Content-Type: application/json" \
  -d '{"failure_mode": "always", "status_code": 503}'

# 50% 概率返回失败
curl -X POST http://localhost:8888/api/v1/scenarios/health_check/start \
  -H "Content-Type: application/json" \
  -d '{"failure_mode": "intermittent", "fail_rate": 0.5}'

# 响应超时
curl -X POST http://localhost:8888/api/v1/scenarios/health_check/start \
  -H "Content-Type: application/json" \
  -d '{"failure_mode": "delayed"}'
```

#### 5. 协程泄漏（goroutine_leak）

持续创建永不退出的协程，导致协程数持续增长。

**参数说明：**
- `goroutines_per_second`: 每秒创建的协程数

**示例：**
```bash
# 每秒创建 100 个永久阻塞的协程
curl -X POST http://localhost:8888/api/v1/scenarios/goroutine_leak/start \
  -H "Content-Type: application/json" \
  -d '{"goroutines_per_second": 100}'
```

#### 6. 磁盘 IO（disk_io）

产生高磁盘 IO 负载，占用磁盘带宽。

**参数说明：**
- `write_rate_mb`: 每秒写入速率（MB）

**示例：**
```bash
# 每秒写入 100MB 数据
curl -X POST http://localhost:8888/api/v1/scenarios/disk_io/start \
  -H "Content-Type: application/json" \
  -d '{"write_rate_mb": 100}'
```

#### 7. 崩溃模拟（crash）

模拟服务在指定时间后崩溃。

**参数说明：**
- `crash_delay`: 延迟多少秒后崩溃

**示例：**
```bash
# 10 秒后触发服务崩溃
curl -X POST http://localhost:8888/api/v1/scenarios/crash/start \
  -H "Content-Type: application/json" \
  -d '{"crash_delay": 10}'
```

#### 8. 依赖服务失败（dependency）

模拟依赖服务（如数据库、Redis、下游 API）调用失败。

**参数说明：**
- `failure_type`: 失败类型
  - `timeout`: 超时（延迟 30 秒后返回）
  - `error`: 返回错误（HTTP 500）
  - `slow`: 响应缓慢（2-5 秒后返回）

**示例：**
```bash
# 模拟依赖服务超时
curl -X POST http://localhost:8888/api/v1/scenarios/dependency/start \
  -H "Content-Type: application/json" \
  -d '{"failure_type": "timeout"}'
```

测试依赖服务：
```bash
curl http://localhost:8888/api/v1/mock-service
```

### 测试接口

#### 10ms 延迟测试接口

```bash
curl http://localhost:8888/api/v1/test/sleep10ms
```

#### 30ms 延迟测试接口

```bash
curl http://localhost:8888/api/v1/test/sleep30ms
```

### 通用 API

#### 列出所有场景

```bash
curl http://localhost:8888/api/v1/scenarios
```

**响应示例：**
```json
{
  "scenarios": [
    {
      "name": "cpu_burner",
      "description": "Increases CPU usage to specified percentage",
      "running": true
    },
    {
      "name": "memory_leaker",
      "description": "Continuously leaks memory at specified rate",
      "running": false
    }
  ]
}
```

#### 查询场景状态

```bash
curl http://localhost:8888/api/v1/scenarios/cpu_burner/status
```

**响应示例：**
```json
{
  "running": true,
  "start_time": "2024-10-25T08:30:00Z",
  "params": {
    "target_percent": 80
  },
  "metrics": {
    "current_cpu_percent": 78.5
  }
}
```

#### 停止单个场景

```bash
curl -X POST http://localhost:8888/api/v1/scenarios/cpu_burner/stop
```

#### 健康检查端点

```bash
# 健康检查
curl http://localhost:8888/health

# 就绪检查
curl http://localhost:8888/ready
```

## 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                   Mock Server                           │
├─────────────────────────────────────────────────────────┤
│  HTTP API 层                                             │
│  - 单场景/复合场景控制                                   │
│  - 状态查询                                              │
├─────────────────────────────────────────────────────────┤
│  场景管理器（Scenario Manager）                          │
│  - 场景生命周期管理                                      │
│  - 会话管理（复合场景）                                  │
│  - 原子性场景切换                                        │
├─────────────────────────────────────────────────────────┤
│  场景插件（Scenario Plugins）                            │
│  ├─ CPU 占用                                             │
│  ├─ 内存泄漏                                             │
│  ├─ 网络延迟                                             │
│  ├─ 健康检查失败                                         │
│  ├─ 协程泄漏                                             │
│  ├─ 磁盘 IO                                              │
│  ├─ 崩溃模拟                                             │
│  └─ 依赖服务失败                                         │
└─────────────────────────────────────────────────────────┘
```

## 配置说明

编辑 `etc/mockserver.yaml` 配置文件：

```yaml
Name: mockserver
Host: 0.0.0.0
Port: 8888

Log:
  Mode: console    # console 或 file
  Level: info      # debug, info, warn, error
```

## 使用场景示例

### 场景 1: 测试 CPU 异常检测

```bash
# 1. 启动 CPU 占用场景（80%，持续 5 分钟）
curl -X POST http://localhost:8888/api/v1/scenarios/cpu_burner/start \
  -d '{"target_percent": 80, "duration": 300}'

# 2. AI 系统监控到 CPU 异常，开始分析
# 3. AI 系统查询服务日志和指标，进行研判
# 4. AI 系统给出诊断结果和解决方案

# 5. 测试完成后停止场景
curl -X POST http://localhost:8888/api/v1/scenarios/cpu_burner/stop
```

### 场景 2: 测试复合故障诊断

```bash
# 同时触发多个异常：CPU 高 + 内存泄漏 + 网络延迟 + 健康检查失败
# 10 分钟后自动恢复
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

# AI 系统需要在复合故障场景下准确识别所有异常并给出综合诊断
```

### 场景 3: 测试场景自动切换

```bash
# 1. 先启动场景 A（CPU 占用）
curl -X POST http://localhost:8888/api/v1/composite/start \
  -d '{"scenarios":[{"name":"cpu_burner","params":{"target_percent":80}}]}'

# 2. 再启动场景 B+C（内存泄漏 + 网络延迟）
#    场景 A 会自动停止，立即切换到 B+C
curl -X POST http://localhost:8888/api/v1/composite/start \
  -d '{
    "scenarios": [
      {"name":"memory_leaker","params":{"target_mb":2048,"leak_rate_mb":50}},
      {"name":"network_latency","params":{"latency_ms":500}}
    ]
  }'

# 3. 验证场景 A 已停止，B+C 正在运行
curl http://localhost:8888/api/v1/composite/status
```

## 项目结构

```
mock-server/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
├── internal/
│   ├── config/
│   │   └── config.go            # 配置管理
│   ├── handler/
│   │   ├── composite_handler.go # 复合场景处理器
│   │   ├── health_handler.go    # 健康检查处理器
│   │   └── scenario_handler.go  # 单场景处理器
│   ├── manager/
│   │   ├── scenario_manager.go  # 场景管理器
│   │   └── session_manager.go   # 会话管理器
│   ├── scenarios/
│   │   ├── interface.go         # 场景接口定义
│   │   ├── cpu_burner.go        # CPU 占用实现
│   │   ├── memory_leaker.go     # 内存泄漏实现
│   │   ├── network_latency.go   # 网络延迟实现
│   │   ├── health_check.go      # 健康检查失败实现
│   │   ├── goroutine_leak.go    # 协程泄漏实现
│   │   ├── disk_io.go           # 磁盘 IO 实现
│   │   ├── crash.go             # 崩溃模拟实现
│   │   └── dependency.go        # 依赖服务失败实现
│   └── svc/
│       └── service_context.go   # 服务上下文
├── etc/
│   └── mockserver.yaml          # 配置文件
├── Dockerfile                    # Docker 镜像构建文件
├── go.mod                        # Go 模块依赖
├── README.md                     # 英文文档
└── README_zh.md                  # 中文文档（本文件）
```

## 依赖库

```go
module github.com/Z3Labs/MockServer

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1           // HTTP 框架
    github.com/shirou/gopsutil/v3 v3.23.10    // 系统资源监控
)
```

## 技术实现要点

### 1. CPU 占用实现

根据 `runtime.NumCPU()` 获取 CPU 核心数，为每个核心启动一个协程。每个协程执行忙循环 + 动态休眠，通过调整计算/休眠时间比例来精确控制 CPU 占用率。

### 2. 内存泄漏实现

每秒分配指定大小的字节数组，并将其保存到切片中防止被 GC 回收。填充数据以防止编译器优化。达到目标内存后停止分配。

### 3. 场景原子切换

使用 `SessionManager` 管理场景会话，配合 `context.WithCancel` 实现场景的优雅停止。通过互斥锁保证并发安全，确保新场景启动时旧场景能立即停止。

### 4. 健康检查控制

在健康检查处理器中根据场景状态返回不同响应：
- `always`: 直接返回 503
- `intermittent`: 按概率随机返回 200/503
- `delayed`: Sleep 10s+ 后返回，触发超时

## 常见问题

### Q: 如何验证 CPU 占用场景是否生效？

A: 可以使用以下命令监控 CPU 使用率：
```bash
# Linux
top -p $(pgrep mockserver)

# 或使用 docker stats（如果使用 Docker 运行）
docker stats
```

### Q: 内存泄漏场景会导致系统 OOM 吗？

A: 会的。如果设置的目标内存超过系统可用内存，可能会触发 OOM。建议在测试环境中使用容器资源限制：
```bash
docker run --memory="4g" -p 8888:8888 mockserver:latest
```

### Q: 为什么健康检查有时候还是返回成功？

A: 如果使用了 `intermittent` 模式，健康检查会按照设置的 `fail_rate` 概率返回失败。例如 `fail_rate: 0.5` 表示 50% 概率返回失败，50% 概率返回成功。

### Q: 可以同时运行多个独立场景吗？

A: 可以。使用复合场景 API 一次性启动多个场景即可。注意某些场景可能会相互影响（如 CPU 和磁盘 IO）。

### Q: 场景停止后资源会立即释放吗？

A: 大部分场景会立即停止，但内存泄漏场景分配的内存需要等待 Go GC 回收。可以通过重启服务快速释放所有资源。

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 联系方式

- 项目仓库: https://github.com/Z3Labs/MockServer
- Issue 反馈: https://github.com/Z3Labs/MockServer/issues
