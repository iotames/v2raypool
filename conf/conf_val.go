package conf

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/iotames/miniutils"
	"github.com/joho/godotenv"
)

const DEFAULT_ENV_FILE = "default.env"
const DEFAULT_TEST_URL = "https://www.google.com/"
const DEFAULT_AUTO_START = false
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
	TestUrl                               string
	EnvFile                               string
	RuntimeDir                            string
	AutoStart                             bool
	GrpcPort, V2rayApiPort, WebServerPort int
	V2rayPath                             string
	SubscribeUrl                          string
	SubscribeDataFile                     string
	// TODO 拆分HttpProxy配置为HTTP协议端口，SOCKS协议端口。 使用HTTP端口为系统代理
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

func (cf Conf) GetOkInboundProtocols() []string {
	return []string{"http", "socks", "socks5"}
}

func (cf Conf) HttpProxySplit() []string {
	spt := strings.Split(cf.HttpProxy, ":")
	lensp := len(spt)
	if lensp != 3 {
		panic(fmt.Errorf("HttpProxy(%s)设置不正确.例:%s", cf.HttpProxy, DEFAULT_HTTP_PROXY))
	}
	protcl := spt[0]
	okprotcls := cf.GetOkInboundProtocols()
	if miniutils.GetIndexOf(protcl, okprotcls) == -1 {
		panic(fmt.Errorf("HttpProxy(%s)设置不正确. Protocol Only Support: %v", cf.HttpProxy, okprotcls))
	}
	return spt
}

func (cf Conf) GetHttpProxyPort() int {
	spt := cf.HttpProxySplit()
	portStr := spt[2]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(fmt.Errorf("HttpProxy(%s)设置不正确.端口号必须为整数.例:%s", cf.HttpProxy, DEFAULT_HTTP_PROXY))
	}
	return port
}

// GetHttpProxyProtocol. get SystemProxy Inbound Protocol By HttpProxy. Only Support http and socks
// curl --proxy socks5://127.0.0.1:30000 https://httpbin.org/get -v
// curl --proxy http://127.0.0.1:30000 https://httpbin.org/get -v
func (cf Conf) GetHttpProxyProtocol() string {
	spt := cf.HttpProxySplit()
	return spt[0]
}

func (cf Conf) UpdateSubscribeData(val string) error {
	val = strings.TrimSpace(val)
	f, err := os.OpenFile(cf.SubscribeDataFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		// panic(err)
		return err
	}
	_, err = f.WriteString(val)
	if err != nil {
		// panic(err)
		return err
	}
	return nil
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

func UpdateConf(mp map[string]string, fpath string) error {
	// Load 和 Read 文件顺序对配置值的影响不一致
	oldData, err := godotenv.Read(DEFAULT_ENV_FILE, fpath)
	if err != nil {
		return err
	}
	for k, v := range mp {
		vv, ok := oldData[k]
		if ok {
			oldData[k] = v
			os.Setenv(k, v)
			if vv != v {
				fmt.Printf("------UpdateConf--%s--(%v)-to(%v)--\n", k, vv, v)
			}
		}
	}
	return godotenv.Write(oldData, fpath)
}
