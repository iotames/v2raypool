package conf

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/iotames/miniutils"
)

const DEFAULT_TEST_URL = "https://www.google.com"
const DEFAULT_RUNTIME_DIR = "runtime"
const DEFAULT_WEB_SERVER_PORT = 8087
const DEFAULT_GRPC_PORT = 50051
const DEFAULT_V2RAY_API_PORT = 15492
const DEFAULT_V2RAY_PATH = "bin/v2ray.exe"
const DEFAULT_HTTP_PROXY = "http://127.0.0.1:30000"
const DEFAULT_SUBSCRIBE_DATA_FILE = "subscribe_data.txt"

const DEFAULT_DIRECT_DOMAIN_LIST = "geosite:cn"
const DEFAULT_DIRECT_IP_LIST = "geoip:private,geoip:cn"
const DEFAULT_PROXY_DOMAIN_LIST = "geosite:google"
const DEFAULT_PROXY_IP_LIST = "geoip:!cn"

type Conf struct {
	TestUrl                                                      string
	EnvFile                                                      string
	RuntimeDir                                                   string
	GrpcPort, V2rayApiPort, WebServerPort                        int
	V2rayPath                                                    string
	SubscribeUrl                                                 string
	SubscribeDataFile                                            string
	HttpProxy                                                    string
	DirectDomainList, DirectIpList, ProxyDomainList, ProxyIpList []string
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

func (cf Conf) GetLogger() *miniutils.Logger {
	cf.RuntimeDir = strings.TrimSpace(cf.RuntimeDir)
	if cf.RuntimeDir == "" {
		panic("RuntimeDir can not be empty")
	}
	return miniutils.GetLogger(filepath.Join(cf.RuntimeDir, "logs"))
}

var vconf Conf

func SetConf(v Conf) {
	vconf = v
}
func GetConf() Conf {
	return vconf
}
