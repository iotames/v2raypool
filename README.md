## 简介

同时运行多个v2ray代理，暴露多个本地IP端口，组成简单的IP代理池。

可以选一个节点设为系统代理，用来浏览网页。要求不高的话，也可同时调用多个，作为爬虫切换IP的代理池。

提供通用的gRPC控制接口，参看数据定义文件 `v2raypool.proto`


## 使用说明

不想自己编译项目源码，可下载[Release压缩包](https://github.com/iotames/v2raypool/releases)直接使用，再看第4-5步的使用说明。

### 1. 下载依赖

运行命令: `go mod tidy`

1.1 如因网络问题下载失败，可设置模块代理。运行命令:
```
go env -w GOPROXY=https://goproxy.cn,direct
# 或者 go env -w GOPROXY=https://goproxy.io,direct
```

1.2 若出现依赖包版本冲突，请删除 `go.mod` 文件，再运行命令:
```
go mod init github.com/iotames/v2raypool
go mod tidy
```

### 2. 编译可执行文件

2.1 编译

```
# 进入项目 main 目录，并执行go编译命令
cd main

# linux或mac 运行: go build -o v2raypool .
go build -o v2raypool.exe .
```

编译出二进制可执行文件 `v2raypool`(linux or max) 或 `v2raypool.exe`(windows)

2.2 生成配置文件

命令行运行可执行文件(v2raypool.exe 或 ./v2raypool)，会生成配置文件 `.env`。并提示找不到v2ray核心文件：

```
v2raypool.exe
请检查配置文件，路径:(.env)
VP_V2RAY_PATH 配置项错误，找不到可执行文件。
请下载v2ray核心文件(https://github.com/v2fly/v2ray-core/releases)
```

### 3. 下载v2ray核心文件

3.1 官网下载核心文件Zip压缩包: https://github.com/v2fly/v2ray-core/releases

3.2 解压到 `main/bin` 目录，并删除或改名解压后的 `config.json` 文件，防止程序错误读取。

3.3 检查或修改v2ray `可执行文件路径`: 查看 `.env` 配置文件的 `VP_V2RAY_PATH` 配置项。


### 4. 设置订阅地址

更改 `.env` 配置文件的 `VP_SUBSCRIBE_URL`，改成实际使用的订阅源地址(http开头)
若订阅源地址网络异常，可使用 `VP_SUBSCRIBE_DATA_FILE` 配置项。通过其他途径查看订阅地址的响应结果，把内容存入文件。

`VP_HTTP_PROXY` 配置项，可设置一个http开头的代理地址。在 `gRPC客户端` 使用 `--activeproxynode` 命令项可激活一个节点使用代理端口。


### 5. 运行服务端和客户端

5.1 服务端

可执行文件直接运行，启动 `gRPC` 服务端

```
# linux 或 mac 执行 ./v2raypool
v2raypool.exe
```

5.2 客户端

`gRPC` 客户端交互命令:

```
# 启动v2ray代理池
v2raypool.exe --startproxynodes

# 查看v2ray代理池信息(包括：本地代理端口号，测速结果，运行状态，测速时间，节点名，节点索引)
v2raypool.exe --getproxynodes

# 测速(测速基准使用https://www.google.com)。测速结束后，会自动选择最快的节点作为系统代理节点。
v2raypool.exe --testproxynodes

# 根据索引值激活某个节点为系统代理的端口（--getproxynodes 查看索引值，系统代理端口从VP_HTTP_PROXY的值读取）
v2raypool.exe --activeproxynode=16

# 停止所有节点
v2raypool.exe --stopproxynodes
```

### 6. 配置systemd系统服务(Linux)

使用环境变量 `VP_ENV_FILE` 定义环境变量配置文件的路径。不设置默认为 `.env`

使用Linux自带的systemctl命令管理 `v2raypool`。

1. 新建 `v2raypool.service` 文件：

```
vim /usr/lib/systemd/system/v2raypool.service
```

v2raypool.service内容示例(/root/v2raypool/main 为可执行文件所在路径):
```
[Unit]
Description=v2ray proxy pool
After=network.target

[Service]
WorkingDirectory=/root/v2raypool/main
ExecStart=/root/v2raypool/main/v2raypool
User=root
Restart=on-failure
RestartSec=300
# KillSignal=SIGQUIT
TimeoutStopSec=10
StandardOutput=file:/root/v2raypool/main/output.log
# StandardError=file:/root/qddns/output.err.log

[Install]
WantedBy=multi-user.target
```

2. 重载systemd配置

```
systemctl daemon-reload
```

3. 使用 systemctl 管理v2raypool的gRPC服务端

```
systemctl status v2raypool
systemctl start v2raypool
systemctl stop v2raypool
```

## 开发相关

### gRPC接口

proto数据格式定义文件: ./v2raypool.proto
gRPC接口文件目录: ./grpc

```
protoc --go_out=./ --go-grpc_out=./ product.proto
```