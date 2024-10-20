<div align="center">
   <p>
   <img style="width:200px" alt="V2rayPool" src="https://raw.githubusercontent.com/iotames/v2raypool/master/main/resource/v2raypool.logo.png">
   </p>
  <strong>简单易用的v2ray客户端和代理池服务</strong>
<br>

</div>


## 简介

[![GoDoc](https://badgen.net/badge/Go/referenct)](https://pkg.go.dev/github.com/iotames/miniutils)
[![License](https://badgen.net/badge/License/MIT/green)](https://github.com/iotames/miniutils/blob/main/LICENSE)

- 纯Go语言实现，Linux, Win, IOS 全平台支持
- 同时运行多个v2ray代理，组成IP代理池
- 使用gRPC接口控制和WebUI网页交互

`gRPC` 接口请参照数据定义文件: `v2raypool.proto`

[项目文档: https://iotames.github.io/v2raypool/](https://iotames.github.io/v2raypool/)


## 用途

1. 多个本地监听端口，组成简单的IP代理池。可供爬虫等程序调用。
2. 选择单个节点设为系统代理，可作为普通的v2ray客户端使用。



## 用户界面

提供 `gRPC接口` 和 `Web网页` 两种交互方式。

代理池节点列表:
![WebUI面板1](https://raw.githubusercontent.com/iotames/v2raypool/master/screenshot_proxypool.jpg)

v2ray服务进程列表:
![WebUI面板2](https://raw.githubusercontent.com/iotames/v2raypool/master/screenshot_v2ray.jpg)

Windows代理设置: `网络和Internet` -> `代理` -> `手动设置代理`

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

# linux或mac 运行: go build -o v2raypool -trimpath -ldflags "-s -w -buildid=" .
go build -o v2raypool.exe -trimpath -ldflags "-s -w -buildid=" .
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

设置代理节点的订阅源地址，请更改 `.env` 文件的 `VP_SUBSCRIBE_URL` 配置项，配置值为http网络地址。

若网络地址访问异常，可使用 `VP_SUBSCRIBE_DATA_FILE` 配置项。设法查看订阅地址的响应结果，保存到文件 `subscribe_data.txt`。


### 5. 运行服务端和客户端

5.1 服务端

可执行文件直接运行，会启动 `gRPC` 和 `WebUI` 服务端。
首次运行，会自动生成 `.env` 和 `default.env` 配置文件。

```
# linux 或 mac 执行 ./v2raypool
v2raypool.exe
```

5.2 客户端

`gRPC` 客户端交互命令:

```
# 启动v2ray代理池，激活所有代理节点。
v2raypool.exe --startproxynodes

# 查看v2ray代理池信息(包括：本地代理端口号，测速结果，运行状态，测速时间，节点名，节点索引)
v2raypool.exe --getproxynodes

# 临时设置默认的URL测速地址。默认为 .env 文件中 VP_TEST_URL 配置项的值。
v2raypool.exe --setproxytesturl

# 测速(测速基准使用https://www.google.com)。测速结束后，会自动选择最快的节点作为系统代理节点。
v2raypool.exe --testproxynodes

# 指定已测速过的域名，按速度从快到慢，查看代理节点信息
v2raypool.exe --getproxynodesbydomain=www.google.com

# 根据索引值激活某个节点为系统代理的端口（--getproxynodes 命令可查看索引值，系统代理端口从VP_HTTP_PROXY的值读取）
v2raypool.exe --activeproxynode=16

# 更新订阅。更新节点的同时，也会更新 subscribe_data.txt 数据文件
v2raypool.exe --updateproxynodes

# 停止所有节点
v2raypool.exe --stopproxynodes
```

`WebUI` 网页面板: [http://127.0.0.1:8087](http://127.0.0.1:8087). `windows` 环境下网页会自动打开。

### 6. Linux配置systemd系统服务(可选)

Linux平台通过配置 `v2raypool.service` ，可使用 `systemctl` 系统命令来管理 `v2raypool`。
可使用环境变量 `VP_ENV_FILE` 定义配置文件的路径。不设置默认为 `.env`

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
StandardError=file:/root/v2raypool/main/output.err.log

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

## 配置说明

在 `main` 目录编译生成可执行文件，首次运行会生成2个文件:

- `default.env`: 显示所有配置项的默认值，不应修改此文件。
- `.env`: 程序配置文件。更改后可覆盖 default.env 文件中的默认值。

```
# 该目录存放程序运行时产生的文件
VP_RUNTIME_DIR = "runtime"

# 傻瓜模式。当配置满足条件时，应用启动后，自动运行所有节点，执行测速，并设置速度最快的节点为系统代理。
# 适合打包现成的配置和应用，直接分享给小白用。
VP_AUTO_START = true

# Web服务器端口。设置为0可禁用Web面板
VP_WEB_SERVER_PORT = 8087

# 代理池的gRPC服务端口
VP_GRPC_PORT = 50051

# v2ray可执行文件路径
# 例: "D:\\Users\\yourname\\v2ray-windows-64\\v2ray.exe" or "/root/v2ray-linux64/v2ray"
VP_V2RAY_PATH = "bin/v2ray.exe"

# 代理节点订阅地址
# 支持 http:// 和 socks5:// 协议
VP_SUBSCRIBE_URL = ""

# 若订阅地址无法直接访问，可指定订阅数据文件，数据文件内容为访问订阅地址获取的原始数据。
# 若有设置订阅数据文件，且文件内容不为空。则优先从该文件读取订阅节点信息。
VP_SUBSCRIBE_DATA_FILE = "subscribe_data.txt"

# 设置HTTP代理，代理池每个节点的本地端口号，往后开始累加。为防止与常用端口冲突，尽量设大点。
VP_HTTP_PROXY = "http://127.0.0.1:30000"

# 节点测速的URL
VP_TEST_URL = "https://www.google.com"
```


## 订阅节点


### 已支持的出站协议

| 出站协议   | 订阅数据格式 | 是否支持 |
| -----     | ----------- | ------- |
| vmess://  |   base64    |   ✅   |
| ss://     |   base64    |   ✅   |
| trojan:// |   -         |   ✅   |
| ssr://    |   -         |   ❌   | 

- ✅ 已支持
- ⛏️ 开发中
- ❌ 不支持

### 订阅数据格式

1. `VP_SUBSCRIBE_URL`: 订阅源的地址配置。填写 `http` 开头的URL网址。

2. `订阅数据`: 访问订阅源的URL地址得到的原始数据。数据可能被BASE64编码过，可保存为 `subscribe_data.txt` 文件，并配置 `VP_SUBSCRIBE_DATA_FILE` 选项。

`订阅数据` 经过 `base64解码` 后，得到以 `\n` 换行符分割的多个代理节点信息。

每个节点信息，可能都被Base64编码过(vmess://)，也可能是明文(vless://)，或者二者混合(ss://)。如下所示:

```
vmess://eyJhZGQiOiAiMjAyLjc4LjE2Mi41IiwgImFpZCI6IDAsICJob3N0IjogImlyc29mdC5zeXRlcy5uZXQiLCAiaWQiOiAiMmZmOTdjNmQtODU1Ny00MmE0LWI0M2YtMTljNzdjNTk1OWVhIiwgIm5ldCI6ICJ3cyIsICJwYXRoIjogIi9AZm9yd2FyZHYycmF5IiwgInBvcnQiOiA0NDMsICJwcyI6ICJnaXRodWIuY29tL2ZyZWVmcSAtIFx1NTM3MFx1NWVhNiAgMiIsICJ0bHMiOiAidGxzIiwgInR5cGUiOiAiYXV0byIsICJzZWN1cml0eSI6ICJhdXRvIiwgInNraXAtY2VydC12ZXJpZnkiOiB0cnVlLCAic25pIjogIiJ9
vless://26DL68CE-DL93-8342-LQ8F-317F4A6E7J76@45.43.31.159:443?encryption=none&security=reality&sni=azure.microsoft.com&fp=safari&pbk=c7qU9-_0WflwIKUiZFxSss_xw-2AP3jB1ENxKLI0OTw&type=tcp&headerType=none#u9un-US-Xr1
ss://Y2hhY2hhMjAtaWV0Zi1wb2x5MTMwNTozNlpDSGVhYlVTZktqZlFFdko0SERW@185.242.86.156:54170#github.com/freefq%20-%20%E4%BF%84%E7%BD%97%E6%96%AF%20%201
ss://YWVzLTI1Ni1nY206N0JjTGRzTzFXd2VvR0QwWA@193.243.147.128:40368#github.com/freefq%20-%20%E6%B3%A2%E5%85%B0%20%205
vmess://eyJhZGQiOiAic2VydmVyMzEuYmVoZXNodGJhbmVoLmNvbSIsICJhaWQiOiAwLCAiaG9zdCI6ICJzZXJ2ZXIzMS5iZWhlc2h0YmFuZWguY29tIiwgImlkIjogIjQxNTQxNDNjLWJiYmEtNDdhNC05Zjc5LWMyZWQwODdjYmNjOSIsICJuZXQiOiAid3MiLCAicGF0aCI6ICIvIiwgInBvcnQiOiA4ODgwLCAicHMiOiAiZ2l0aHViLmNvbS9mcmVlZnEgLSBcdTdmOGVcdTU2ZmRDbG91ZEZsYXJlXHU1MTZjXHU1M2Y4Q0ROXHU4MjgyXHU3MGI5IDYiLCAidGxzIjogIiIsICJ0eXBlIjogImF1dG8iLCAic2VjdXJpdHkiOiAiYXV0byIsICJza2lwLWNlcnQtdmVyaWZ5IjogdHJ1ZSwgInNuaSI6ICIifQ==
```

3. 由于协议不支持，数据格式不兼容，等原因造成的节点解析失败，都会被忽略，然后继续解析下个节点。

4. `vmess` 节点信息再次经过 `BASE64解码` 后，解析为JSON字符串格式。如下所示:

```
{"add":"us0.xxx.top","host":"","id":"93EA57CE-EA21-7240-EE7F-317F4A6A8B65","net":"ws","path":"","port":"444","ps":"xxx-v2-US-LosAngeles0","tls":"","type":"none","v":2,"aid":0}
```

### 调试工具

- [base64在线解码](https://base64.us/)


## 路由规则

支持自定义域名和IP列表配置:

- PROXY_DOMAIN_LIST 代理域名列表
- DIRECT_DOMAIN_LIST 直连域名列表
- PROXY_IP_LIST 代理IP列表
- DIRECT_IP_LIST 直连IP列表


域名匹配规则:

- `纯字符串`：当此字符串匹配`目标域名中任意部分`，该规则生效。比如 `sina.com` 可以匹配 `sina.com`、`sina.com.cn`、`sina.company` 和 `www.sina.com`，但不匹配 `sina.cn`。
- `正则表达式`：由 `regexp:` 开始，余下部分是一个正则表达式。当此正则表达式匹配目标域名时，该规则生效。例如 `regexp:\.goo.*\.com$` 匹配 `www.google.com`、`fonts.googleapis.com`，但不匹配 `google.com`。
- `子域名（推荐）`：由 `domain:` 开始，余下部分是一个域名。当此域名是目标域名或其子域名时，该规则生效。例如 `domain:v2ray.com` 匹配 `www.v2ray.com`、`v2ray.com`，但不匹配 `xv2ray.com`。
- `完整匹配`：由 `full:` 开始，余下部分是一个域名。当此域名完整匹配目标域名时，该规则生效。例如 `full:v2ray.com` 匹配 `v2ray.com` 但不匹配 `www.v2ray.com`。
- `预定义域名列表`：由 `geosite:` 开头，余下部分是一个类别名称（域名列表），如 `geosite:google` 或者 `geosite:cn`。名称及域名列表参考[预定义域名列表](https://www.v2fly.org/config/routing.html#%E9%A2%84%E5%AE%9A%E4%B9%89%E5%9F%9F%E5%90%8D%E5%88%97%E8%A1%A8)。
- `从文件中加载域名`：形如 `ext:file:tag`，必须以 `ext:` 开头，后面跟文件名和标签，文件存放在[资源目录](https://www.v2fly.org/config/env.html#%E8%B5%84%E6%BA%90%E6%96%87%E4%BB%B6%E8%B7%AF%E5%BE%84)中，文件格式与 `geosite.dat` 相同，标签必须在文件中存在。

IP匹配规则:

- `IP`：形如 `127.0.0.1`。
- `CIDR`：形如 `10.0.0.0/8`。
- `GeoIP`：
形如 `geoip:cn` 为正向匹配，即为匹配「中国大陆 IP 地址」。后面跟双字符国家或地区代码，支持所有可以上网的国家和地区。

形如 `geoip:!cn` 为反向匹配，即为匹配「非中国大陆 IP 地址」。后面跟双字符国家或地区代码，支持所有可以上网的国家和地区。

特殊值：`geoip:private`（V2Ray 3.5+），包含所有私有地址，如 `127.0.0.1`。

- 从文件中加载 IP：
形如 `ext:file:tag` 和 `ext-ip:file:tag` 为正向匹配，即为匹配 「tag 内的 IP 地址」。

形如 `ext:file:!tag` 和 `ext-ip:file:!tag` 为反向匹配，即为匹配「非 tag 内的 IP 地址」。

必须以 `ext:` 或 `ext-ip:` 开头，后面跟文件名、`标签`或 `!标签`，文件存放在[资源目录](https://www.v2fly.org/config/env.html#%E8%B5%84%E6%BA%90%E6%96%87%E4%BB%B6%E8%B7%AF%E5%BE%84)中，文件格式与 `geoip.dat` 相同，标签必须在文件中存在。

具体请参看 [v2ray路由规则](https://www.v2fly.org/config/routing.html#ruleobject)

## 开发相关

### gRPC接口

1. 使用proto数据格式定义文件: `./v2raypool.proto` 可实现跨语言调用
2. Go语言的gRPC接口文件位于 `./grpc` 目录。引用包名: `github.com/iotames/v2raypool/grpc`
3. 调用过程参考代码文件: `./main/main_grpc.go`

```
# 从proto数据格式文件生成可供Go语言调用的代码包
protoc --go_out=./ --go-grpc_out=./ v2raypool.proto
```