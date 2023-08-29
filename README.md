## 简介

同时运行多个v2ray进程代理端口，并提供gRPC控制接口


## 快速开始

```
# 加载依赖包
go mod tidy

# 运行服务端
cd main
go run .
```

若依赖包提示版本冲突，请删除 `go.mod` 文件后，再执行
```
go mod init github.com/iotames/v2raypool
go mod tidy
```

## gRPC接口

proto数据格式定义文件: ./v2raypool.proto
gRPC接口文件目录: ./grpc

```
protoc --go_out=./ --go-grpc_out=./ product.proto
```

## 配置

使用环境变量 `VP_ENV_FILE` 定义环境变量配置文件的路径。不设置默认为 `.env`
