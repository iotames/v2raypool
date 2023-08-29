package conf

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const DEFAULT_RUNTIME_DIR = "runtime"
const DEFAULT_GRPC_PORT = 50051
const DEFAULT_V2RAY_PATH = "bin/v2ray"
const DEFAULT_HTTP_PROXY = "http://127.0.0.1:30000"
const DEFAULT_SUBSCRIBE_DATA_FILE = "subscribe_data.txt"

type Conf struct {
	EnvFile           string
	RuntimeDir        string
	GrpcPort          int
	V2rayPath         string
	SubscribeUrl      string
	SubscribeDataFile string
	HttpProxy         string
}

func (cf Conf) GetSubscribeData() string {
	f, err := os.Open(cf.SubscribeDataFile)
	if err != nil {
		panic(fmt.Errorf("poen SubscribeDataFile(%s) err(%v)", cf.SubscribeDataFile, err))
	}
	b, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(b))
}

func (cf Conf) GetHttpProxyPort() int {
	portSplit := strings.Split(cf.HttpProxy, ":")
	lensp := len(portSplit)
	if lensp != 3 {
		panic(fmt.Errorf("HttpProxy设置不正确(%s).例:%s", cf.HttpProxy, DEFAULT_HTTP_PROXY))
	}
	portStr := portSplit[lensp-1]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(fmt.Errorf("HttpProxy设置不正确(%s).端口号必须为整数.例:%s", cf.HttpProxy, DEFAULT_HTTP_PROXY))
	}
	return port
}

func (cf Conf) UpdateSubscribeData(val string) {
	val = strings.TrimSpace(val)
	f, err := os.OpenFile(cf.SubscribeDataFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		panic(err)
	}
	_, err = f.WriteString(val)
	if err != nil {
		panic(err)
	}
}

var vconf Conf

func SetConf(v Conf) {
	vconf = v
}
func GetConf() Conf {
	return vconf
}
