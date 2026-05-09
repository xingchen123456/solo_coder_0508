# Game Backend

一个使用Go开发的游戏后端基础框架，包含完整的服务架构。

## 技术栈

- **Go**: 1.21+
- **gRPC**: 高性能RPC框架
- **WebSocket**: 长连接实时通信
- **Gin**: HTTP网关框架
- **GORM**: MySQL ORM
- **Go-Redis**: Redis客户端
- **Etcd**: 服务发现与注册
- **Zap**: 高性能日志库
- **Viper**: 配置管理
- **Protobuf**: 数据序列化

## 项目结构

```
game_backend/
├── cmd/
│   └── game_backend/
│       └── main.go          # 主入口
├── configs/
│   └── config.yaml          # 配置文件
├── internal/
│   ├── config/              # 配置管理
│   ├── logger/              # 日志模块
│   ├── mysql/               # MySQL数据库
│   ├── redis/               # Redis缓存
│   ├── etcd/                # Etcd服务发现
│   ├── models/              # 数据模型
│   ├── services/            # 业务服务
│   │   ├── player_service.go
│   │   ├── room_service.go
│   │   └── message_service.go
│   ├── websocket/           # WebSocket长连接
│   ├── grpc/                # gRPC服务
│   ├── http/                # HTTP网关
│   └── utils/               # 工具函数
├── proto/
│   └── game/
│       ├── player.proto     # 玩家相关接口
│       ├── room.proto       # 房间相关接口
│       └── message.proto    # 消息相关接口
├── go.mod
├── go.sum
├── Makefile
└── run.bat
```

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 生成Proto文件

需要先安装protoc和Go插件：

```bash
# 安装protoc
# Windows: 从 https://github.com/protocolbuffers/protobuf/releases 下载
# Linux/Mac: brew install protobuf 或 apt install protobuf-compiler

# 安装Go插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```

然后运行：

```bash
make proto
```

或者手动执行：

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/game/*.proto
```

### 3. 启动依赖服务

确保MySQL、Redis、Etcd已启动：

- MySQL: 默认端口 3306
- Redis: 默认端口 6379
- Etcd: 默认端口 2379

### 4. 修改配置

编辑 `configs/config.yaml` 文件，修改数据库连接信息等配置。

### 5. 运行服务

**启动所有服务**：
```bash
# Windows
run.bat

# 或
go run cmd/game_backend/main.go

# 或使用Makefile
make run
```

**仅启动gRPC服务**：
```bash
go run cmd/game_backend/main.go grpc
```

**仅启动HTTP网关**：
```bash
go run cmd/game_backend/main.go http
```

## 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| gRPC | 50051 | RPC服务 |
| HTTP | 8080 | REST API网关 |
| WebSocket | 8080 (通过HTTP) | 长连接服务 |

## API接口

### HTTP API

#### 健康检查
```
GET /api/v1/health
```

#### 玩家接口
```
POST /api/v1/player/login
POST /api/v1/player/register
GET  /api/v1/player/info    (需要Authorization header)
PUT  /api/v1/player/info    (需要Authorization header)
```

#### 房间接口
```
POST /api/v1/room/create    (需要Authorization header)
POST /api/v1/room/join      (需要Authorization header)
POST /api/v1/room/leave     (需要Authorization header)
GET  /api/v1/room/list
GET  /api/v1/room/info?room_id=xxx
```

#### WebSocket
```
WS /ws?token=xxx&room_id=xxx
```

### gRPC接口

- **PlayerService**: 玩家相关操作
- **RoomService**: 房间相关操作
- **MessageService**: 消息广播操作

gRPC服务支持reflection，可以使用grpcurl等工具调试。

## 功能模块

### 1. 玩家系统
- 用户注册/登录
- Token认证
- 玩家信息管理

### 2. 房间系统
- 创建房间（公开/私有）
- 加入/离开房间
- 房间列表查询
- 房间信息管理

### 3. 实时通信
- WebSocket长连接
- 心跳检测
- 房间消息广播
- 系统消息通知

### 4. 服务治理
- Etcd服务注册与发现
- 配置热加载（Viper）
- 结构化日志

## 数据模型

### User (用户表)
- ID, Username, Password, Email
- Level, Exp, Gold, Diamond
- CreatedAt, UpdatedAt, DeletedAt

### Room (房间表)
- ID, RoomID, RoomName, RoomType
- Status, MaxPlayers, CurrentPlayers
- OwnerID, Password
- CreatedAt, UpdatedAt, DeletedAt

### RoomPlayer (房间玩家表)
- ID, RoomID, UserID
- JoinedAt, CreatedAt, UpdatedAt, DeletedAt

## 配置说明

```yaml
server:
  name: game-backend
  grpc_port: 50051
  http_port: 8080
  ws_port: 8081

mysql:
  host: localhost
  port: 3306
  username: root
  password: password
  database: game_db
  charset: utf8mb4

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

etcd:
  endpoints:
    - localhost:2379

log:
  level: info
  format: json
```

## 开发计划

- [ ] 游戏逻辑模块
- [ ] 匹配系统
- [ ] 排行榜系统
- [ ] 成就系统
- [ ] 支付系统
- [ ] 后台管理系统

## License

MIT
