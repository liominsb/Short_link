# Short_link - 短链接服务

一个基于 Go 语言构建的高性能短链接服务，支持 URL 短链生成、重定向和限流控制。

## 🌟 功能特性

- ✅ 快速生成短链接
- ✅ URL 重定向服务
- ✅ 令牌桶限流保护
- ✅ Redis 缓存支持
- ✅ RabbitMQ 消息队列集成
- ✅ MySQL 数据持久化
- ✅ 配置文件管理

## 🏗️ 技术栈

| 组件 | 版本 | 说明 |
|------|------|------|
| Go | 1.25.0+ | 编程语言 |
| Gin | 1.12.0 | Web 框架 |
| GORM | 1.31.1 | ORM 框架 |
| Redis | 6.15.9 | 缓存存储 |
| RabbitMQ | 1.11.0 | 消息队列 |
| MySQL | 1.8.1 | 数据库 |
| Viper | 1.21.0 | 配置管理 |

## 📋 前置要求

- Go 1.25.0 或更高版本
- MySQL 5.7 或以上
- Redis 4.0 或以上
- RabbitMQ 3.8 或以上（可选）

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/liominsb/Short_link.git
cd Short_link
```

### 2. 下载依赖

```bash
go mod download
```

### 3. 配置环境

在项目根目录创建配置文件 `config.yaml`（或根据项目实际配置位置调整）：
参考:
```yaml
app:
  name: App
  port: :3000

database:
  dsn : "root:123456@tcp(127.0.0.1:3306)/a?charset=utf8mb4&parseTime=True&loc=Local"
  MaxIdleConns: 10
  MaxOpenConns: 100
##redis:
  Addr : "localhost:6379"
  Password : ""
  SubSwitch : "false"

JWT:
  Key : "JWT_SECRET"

RabbitMQ:
  Url: "amqp://guest:guest@localhost:5672/"
```

### 4. 运行服务

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动

## 📚 API 文档

### 1. 创建短链接

**请求**

```http
POST /s
Content-Type: application/json

{
  "url": "https://example.com/very/long/url"
}
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "short_key": "abc123",
    "short_url": "http://your-domain.com/s/abc123"
  }
}
```

### 2. 重定向短链接

**请求**

```http
GET /s/:key
```

例如：`GET /s/abc123`

**响应**

- 成功：HTTP 302 重定向到原始 URL
- 失败：HTTP 404 (短链接不存在)

## 🏗️ 项目结构

```
Short_link/
├── main.go                 # 程序入口
├── go.mod                  # Go 模块定义
├── go.sum                  # 依赖版本锁定
├── config/                 # 配置管理模块
│   └── config.go          # 配置初始化
├── controllers/            # HTTP 处理器
│   ├── redirect.go        # 重定向逻辑
│   └── create.go          # 短链创建逻辑
├── models/                 # 数据模型
│   └── shortlink.go       # 短链数据模型
├── middlewares/            # 中间件
│   └── ratelimit.go       # 限流中间件
├── global/                 # 全局变量
│   └── global.go          # 全局配置和依赖注入
├── utils/                  # 工具函数
│   ├── hash.go            # 哈希工具（生成短 key）
│   └── url.go             # URL 工具函数
└── .gitignore             # Git 忽略文件
```

## ⚙️ 配置说明

### TokenBucket 限流配置

- `rate`: 令牌生成速率（每秒生成的令牌数）
- `capacity`: 令牌桶最大容量

示例：rate=100, capacity=200 表示每秒最多 100 个请求，突发流量可达 200 个。

## 🔍 核心特性说明

### 限流保护

项目使用令牌桶算法实现限流中间件，保护服务免受大流量冲击：

```go
// 在 main.go 中配置
rl := middlewares.NewRateLimiter(
    config.Appconf.TokenBucket.Rate, 
    config.Appconf.TokenBucket.Capacity
)
r.Use(middlewares.RateLimitMiddleware(rl))
```

### 释放模式

生产环境使用 Release 模式运行 Gin：

```go
gin.SetMode(gin.ReleaseMode)
```

## 🛠️ 开发指南

### 构建

```bash
go build -o short_link
```

### 运行

```bash
./short_link
```

### 测试

```bash
go test ./...
```

## 📦 部署

### Docker 部署（示例）

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o short_link

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/short_link .
COPY --from=builder /app/config.yaml .
EXPOSE 8080
CMD ["./short_link"]
```

### 编译命令

```bash
docker build -t short_link:latest .
docker run -p 8080:8080 short_link:latest
```

## 🚨 常见问题

### Q: 服务启动失败，提示配置文件找不到？
A: 确保 `config.yaml` 文件在正确的位置，或检查 `config/config.go` 中的配置文件路径设置。

### Q: Redis/MySQL 连接失败？
A: 
1. 确认 Redis 和 MySQL 服务正在运行
2. 检查配置文件中的连���参数（地址、端口、用户名、密码）
3. 确保防火墙允许相应的端口访问

### Q: 限流不生效？
A: 检查 `config.yaml` 中的 `tokenbucket` 配置是否合理，确保 `rate` 和 `capacity` 值大于 0。

## 📈 性能优化建议

- 合理设置 Redis 过期时间减少内存占用
- 调整令牌桶参数适应业务流量
- 考虑使用 CDN 加速短链重定向
- 定期备份 MySQL 数据库

## 📝 许可证

待补充

## 👨‍💻 贡献

欢迎提交 Issue 和 Pull Request！

## 📧 联系方式

- GitHub: [@liominsb](https://github.com/liominsb)

---

**最后更新**: 2026-05-06
