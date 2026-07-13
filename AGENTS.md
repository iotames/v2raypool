# AGENTS.md — v2raypool 知识库

v2raypool 是一个基于 v2ray-core 的 Go 语言代理池服务，提供 WebUI 和 gRPC 两种交互方式。

## 项目架构

```
main/                  # 入口，编译在此目录
├── main.go            # CLI flag 解析 + 默认启动服务
├── main_func.go       # 端口检查、日志初始化等工具函数
├── main_grpc.go       # gRPC CLI 客户端函数（getproxynodes 等命令）
├── conf.go            # .env 配置加载（easyconf），VP_ 前缀全部环境变量
└── build.sh           # 交叉编译脚本（Linux/Windows）

root package (v2raypool)
├── v2raypool.go       # ProxyPool 核心：节点管理、V2rayServer 控制
├── v2ray.go           # V2rayServer 结构体（进程创建、配置写入）
├── v2ray_config.go    # V4/V5 JSON 配置结构体、outbound 生成
├── v2ray_parse.go     # vmess/trojan/ss 等协议 URI 解析
├── v2ray_api.go       # gRPC API 客户端（通过 v2ray API 端口动态添加 inbound/outbound）
├── proxy_node.go      # ProxyNode 结构体 + 增删操作
├── speed.go           # 节点测速（HTTP GET 耗时）
├── nodes_storage.go   # gob 编码/解码持久化节点数据
├── grpc_server.go     # gRPC 服务端（ProxyPoolService）+ 隧道代理池全局管理
├── grpc_client.go     # gRPC 客户端辅助函数
├── tunnel_config.go   # 隧道代理池配置（TunnelConfig 结构体 + 默认值常量）
├── tunnel_pool.go     # 隧道代理池核心（HTTP 代理服务器，单端口随机出口 IP）
├── win_proxy.go       # Windows 系统代理设置（WinAPI）
├── notwin_proxy.go    # Linux/macOS 系统代理占位（空实现）
└── command_test.go    # 测试

decode/                # 协议解析
├── decode.go          # Base64 解码、订阅 URL/原始数据解析
├── clash.go           # Clash YAML 订阅解析（仅支持 trojan）
├── v2ray.go           # V2raySsNode 数据定义（核心节点结构体）
├── shadowsocks.go     # ss:// URI 解析
├── trojan.go          # trojan:// URI 解析
└── stream.go          # StreamConfig 流传输配置

v2rayapi/              # v2ray gRPC handler 协议适配
├── v2rayapi.go        # GetOutboundRequest — 构造 v2ray outbound 请求
├── freedom.go         # Freedom outbound（直连）
├── shadowsocks.go     # Shadowsocks outbound
├── trojan.go          # Trojan outbound
├── vmess.go           # VMess outbound
└── streamconfig.go    # 传输层流配置（tcp/ws/grpc）

webserver/             # WebUI（基于 glayui）
├── webserver.go       # EasyServer 初始化
├── router.go          # 路由注册（/api/nodes, /api/v2ray/*, /api/tunnel/* 等）
├── nodes.go           # 代理节点 API handler
├── v2ray_ctl.go       # v2ray 进程管理 API handler
├── tunnel_ctl.go      # 隧道代理池 API handler（启动/停止/状态查询）
├── sysproxy_ctl.go    # 系统代理切换 API handler（固定节点/隧道/无代理）
├── conf_ctl.go        # 配置修改/清理缓存 API handler
├── pub_funcs.go       # 公共工具函数
└── response_data.go   # 响应数据结构

conf/                  # 配置
├── conf_val.go        # Conf 结构体定义 + 常量默认值 + 读写

grpc/                  # 生成的 proto 文件
├── v2raypool.pb.go         # protocol buffer 消息
└── v2raypool_grpc.pb.go   # gRPC 服务/客户端桩代码

netutil/               # 网络工具
├── http.go            # 代理 HTTP 客户端
└── ip.go              # IP 工具

client/                # 实验性客户端（main.go）
main/resource/         # 前端静态资源（layui, html, css, js, ico, png）
```

## 数据流

### 主要控制流

1. **启动** (main/main.go:main → runServer):
   - `init()` 加载 `.env` → `conf.GetConf()`
   - 启动 gRPC 服务 (`vp.RunServer()`)
   - 可选启动 WebUI (`webserver.NewWebServer(webPort)`)
   - `ProxyPool.StartV2rayPool()` 启动 v2ray 内核进程 + 读取持久化节点文件 + 可选自动启动所有节点

2. **节点订阅与解析** (v2ray_parse.go:ParseV2rayNodes):
   - 从 URL 或文件获取原始数据 → `decode.ParseSubscribeByRaw()` (Base64 解码)
   - 每行 `protocol://` 格式 → `parseNodeInfo()` 按协议分发
   - 支持: `vmess`, `ss`, `trojan`, Clash YAML

3. **代理节点激活** (proxy_node.go:AddToPool):
   - 通过 v2ray API gRPC 端口动态添加 inbound (http/socks) + outbound (vmess/trojan/ss)
   - 每个节点独立本地端口（从系统代理端口 +1 开始累加）

4. **测速** (speed.go:testProxyNode):
   - 通过代理节点端口 HTTP GET 目标 URL → 记录耗时
   - 结果存入 `ProxyPool.speedMap`，支持按域名查询

5. **隧道代理池** (tunnel_pool.go:TunnelPool):
   - 对外暴露单一 HTTP 代理端口（如 `127.0.0.1:1080`），每次新请求随机选一个测速合格的节点转发
   - 复用 ProxyPool 的 `testOneNode()` 主动测速，结果自动写回 speedMap 供全局共享
   - 测速刷新间隔由 `VP_TUNNEL_REFRESH_INTERVAL` 控制（默认 1200 秒即 20 分钟）
   - 与多端口代理池模式完全兼容，可同时运行

## 核心数据结构

### `decode.V2raySsNode` (核心节点模型)
```go
type V2raySsNode struct {
    Protocol, Add, Host, Id, Net, Path, Ps, Tls, Type string
    V, Aid, Port                                      json.Number
}
```
所有协议最终归一化为这个结构体。

### `ProxyPool` (单例)
- `sync.Once` 懒加载，`GetProxyPool()` 获取
- 管理 `map[int]*V2rayServer` (PID → 进程), `ProxyNodes` (节点列表), `speedMap` (域名→节点)
- 全局读写锁 `sync.Mutex` + `IsLock` 标志
- `GetAvailableNodes(maxDelay)`: 获取测速合格且运行中的节点（隧道代理用）

### `TunnelPool` (隧道代理池)
- `NewTunnelPool(cfg)`: 创建实例，复用 `GetProxyPool()` 单例
- `Start()`: 启动 HTTP 代理服务（支持 CONNECT 隧道和 HTTP 转发）
- `pickNodeAddr()`: 每个请求随机选一个可用节点
- `RefreshNodes()`: 对运行中节点逐个调用 `pool.testOneNode()` 主动测速，结果写回 `speedMap`，按延迟筛选后更新 `nodeList`
- `refreshNodesLoop()`: 后台定时调用 `RefreshNodes()`，间隔由 `RefreshInterval` 控制
- `globalTunnelPool`: 全局实例，由 `grpc_server.go` 中的 `InitTunnelPool/StartTunnelPool/StopTunnelPool` 管理

### `SysProxyType` (系统代理类型枚举)
- `SysProxyNone(0)`: 无系统代理
- `SysProxyNode(1)`: 固定节点代理，`SetSysProxy(1, nodeIdx)` 激活指定节点为系统代理
- `SysProxyTunnel(2)`: 隧道代理，`SetSysProxy(2, -1)` 将系统代理指向隧道端口
- `SetSysProxy()` 负责先取消旧代理再切换，自动处理 Windows 系统代理设置（`win_proxy.go:SetProxy`）

## 关键配置（`.env` 文件）

| 变量 | 默认值 | 说明 |
|------|--------|------|
| VP_V2RAY_PATH | bin/v2ray.exe | v2ray 可执行文件路径 |
| VP_GRPC_PORT | 50051 | gRPC 控制端口 |
| VP_WEB_SERVER_PORT | 8087 | WebUI 端口，0=禁用 |
| VP_V2RAY_API_PORT | 15492 | v2ray API 端口 |
| VP_HTTP_PROXY | http://127.0.0.1:30000 | 系统代理，协议支持 http/socks |
| VP_TEST_URL | https://www.google.com/ | 测速目标 URL |
| VP_SUBSCRIBE_URL | "" | 订阅地址 |
| VP_SUBSCRIBE_DATA_FILE | subscribe_data.txt | 本地订阅数据文件 |
| VP_RUNTIME_DIR | runtime | 运行时文件目录 |
| VP_AUTO_START | false | 启动后自动连接测速 |
| VP_ENABLE_STORAGE | false | 持久化节点测速数据 |
| VP_TUNNEL_ENABLE | false | 启用隧道代理池 |
| VP_TUNNEL_PORT | 1080 | 隧道代理池监听端口 |
| VP_TUNNEL_MAX_DELAY | 230 | 隧道代理池最大延迟阈值(ms) |
| VP_TUNNEL_REFRESH_INTERVAL | 1200 | 隧道代理池刷新间隔(秒)，即20分钟 |

## 命令

### CLI 命令（`main/main.go:main`）
```
v2raypool.exe                        # 启动服务端（gRPC + WebUI）
v2raypool.exe --version              # 版本号
v2raypool.exe --startproxynodes      # 启动所有节点
v2raypool.exe --stopproxynodes       # 停止所有节点
v2raypool.exe --getproxynodes        # 查看节点信息
v2raypool.exe --testproxynodes       # 测速全部节点
v2raypool.exe --activeproxynode=N    # 激活指定索引节点为系统代理
v2raypool.exe --updateproxynodes     # 更新订阅
v2raypool.exe --killproxynodes       # 强制杀掉所有进程
v2raypool.exe --setproxystesturl=URL # 临时设置测速 URL
v2raypool.exe --sysproxy=0           # 取消系统代理
v2raypool.exe --sysproxy=1 --activeproxynode=3  # 系统代理指向节点3
v2raypool.exe --sysproxy=2           # 系统代理指向隧道池(随机IP)
```

### 构建
```bash
cd main
# Windows
go build -o v2raypool.exe -trimpath -ldflags "-s -w -buildid=" .
# Linux
GOOS=linux go build -o v2raypool -trimpath -ldflags "-s -w -buildid=" .
```

### 测试
```bash
go test ./...
```

### 环境
- Go 1.22.1
- Module: `github.com/iotames/v2raypool`
- 国内建议设置镜像: `go env -w GOPROXY=https://goproxy.cn,direct`
- 构建产物在 `main/` 目录下

## WebUI API

### 隧道代理池 API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/tunnel/start | 启动隧道代理池 |
| POST | /api/tunnel/stop | 停止隧道代理池 |
| GET | /api/tunnel/status | 获取隧道代理池状态（running/port/node_count 等）|

### 系统代理切换 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/sysproxy/status | 获取当前系统代理状态（type/node_idx）|
| POST | /api/sysproxy/switch | 切换系统代理 {type: 0\|1\|2, node_idx: -1} |

## gRPC API

定义在 `v2raypool.proto`。关键服务 `ProxyPoolService`:

| RPC | 说明 |
|-----|------|
| GetProxyNodes | 获取节点列表，支持按标题/端口/运行状态过滤 |
| GetProxyNodesByDomain | 按测速域名查询节点（按速度排序）|
| ActiveProxyNode | 激活节点为系统代理 |
| StartProxyPoolAll | 启动所有节点 |
| StopProxyPoolAll | 停止所有节点 |
| TestProxyPoolAll | 测速（不会改变已激活节点）|
| TestProxyPoolAllForce | 强制测速并自动切换最快节点 |
| KillAllNodes | 强制杀掉所有 v2ray 进程 |
| UpdateProxySubscribe | 更新订阅 |

## 关键规则与陷阱

### 端口分配
- 系统代理端口（`VP_HTTP_PROXY`）自动往下-1 是 socks5 端口
- 代理池节点端口从系统代理端口 +1 开始递增
- 每个节点同时开 http + socks 两个 inbound（端口 n 和 n-1）
- `checkInitPorts` 在启动时 panic 检查 6 个端口的占用

### 配置 V4 vs V5
- **默认使用 V4 JSON 配置格式**（`v2ray.ConfigV4`）
- v2ray 5.7-5.14 版本行为差异：5.7 标准输入读配置正常，5.14 有 BUG → `setExeCmd()` 统一写文件读配置

### 文件持久化
- 节点数据用 `encoding/gob` 而非 JSON 编解码（`nodes_storage.go`）
- 路由规则文件 `routing.rules.json`（若存在则优先读取）
- v2raypool 配置写 `v2raypool.config.json`，各节点写 `v2ray.config.json`

### gRPC 客户端使用模式
```go
c, conn := NewProxyPoolGrpcClient()
defer conn.Close()
ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
defer cancel()
// 使用 c 调用 RPC
```
辅助函数 `RequestProxyPoolGrpcOnce()` 封装了此模式。

### 系统代理
- Windows: `win_proxy.go` 使用 WinAPI `InternetSetOption` 设置 IE/系统代理
- Linux/macOS: `notwin_proxy.go` 空实现（`SetProxy` 只打日志）
- 系统代理端口从 `VP_HTTP_PROXY` 读取
- 系统代理切换通过 `SetSysProxy()` 统一管理，支持三种模式：无代理/固定节点/隧道代理
- WebUI 工具栏下拉菜单可切换系统代理模式

### WebUI 前端
- 基于 `github.com/iotames/glayui`（layui 封装）
- 静态资源在 `main/resource/`
- 不支持生成/修改前端代码，直接编辑 HTML + JS

### 订阅数据优先级
1. 订阅地址数据文件 (`VP_SUBSCRIBE_DATA_FILE`) 有内容 → 优先读取
2. 否则从订阅 URL (`VP_SUBSCRIBE_URL`) 获取
3. 支持 Clash YAML 格式（仅 trojan 协议）

### 其他陷阱
- 节点测速 `io.EOF` 错误被当作部分成功处理（`speed.go` 第39行）
- `SetSubscribeRawData` 只支持 yaml 格式，其他格式 panic
- Clash 解析仅支持 `trojan` 类型，其他类型跳过
- v2ray 核心必须下载后放入 `main/bin/`，且需要删除默认 `config.json`
- 测速禁用 KeepAlive（`DisableKeepAlives = true`）解决 EOF 问题
- 节点 ID 由 `RemoteAddr:V2rayNode.Id` 拼接生成
- `oncedo.Do()` 在单元测试中注意状态重置问题

### 隧道代理池规则
- 隧道代理池刷新时主动测速（复用 `ProxyPool.testOneNode()`），结果自动写回 `speedMap`，全局共享
- `ProxyPool.TestAll` 再跑时会通过 `GetLastSpeedNode` + `IsOk()` 跳过近期已测速节点，避免重复
- 测速合格的判定：节点状态为运行中 且 本次测速延迟 < `VP_TUNNEL_MAX_DELAY` ms
- 刷新间隔由 `VP_TUNNEL_REFRESH_INTERVAL` 配置（默认 1200 秒/20 分钟），非硬编码
- HTTPS 使用 HTTP CONNECT 隧道转发（tunnel_pool.go:handleConnect），每次请求随机选新节点
- HTTP 明文请求通过代理节点转发（tunnel_pool.go:handleHTTP），每次请求随机选节点
- 系统代理切换到隧道模式时，自动启动隧道池（如未运行），然后将 Windows 代理指向隧道端口

## 生成文件

编译时使用 `go build` 而非 `go install`（`go install` 在 Linux 上编译 Windows 二进制有问题，见 CI 配置）。CI 使用 `goversioninfo` 生成 Windows `.syso` 资源文件嵌入版本信息。
