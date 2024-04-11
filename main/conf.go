package main

import (
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/iotames/miniutils"
	conf "github.com/iotames/v2raypool/conf"
	"github.com/joho/godotenv"
)

const WORK_ENV_FILE = ".env"

var envFile string
var vconf conf.Conf

func setEnvFile() {
	envFile = os.Getenv("VP_ENV_FILE")
	if envFile == "" {
		envFile = WORK_ENV_FILE
	}
}

func LoadEnv() error {
	var err error
	setEnvFile()
	efiles := initEnvFile()
	fmt.Printf("------LoadEnv--env-files(%v)\n", efiles)
	err = godotenv.Load(efiles...)
	if err != nil {
		return fmt.Errorf("godotenv.Load(%v)err(%v)", efiles, err)
	}
	return getConfByEnv()
}

func initEnvFile() []string {
	var err error
	var files []string
	// var createNewEnvfile bool
	if !miniutils.IsPathExists(envFile) {
		err = createEnvFile(envFile)
		if err != nil {
			panic(err)
		}
		// files = append(files, envFile)
		fmt.Printf("Create file %s SUCCESS\n", envFile)
		// createNewEnvfile = true
	}
	files = append(files, envFile)

	if !miniutils.IsPathExists(conf.DEFAULT_ENV_FILE) {
		err = createEnvFile(conf.DEFAULT_ENV_FILE)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Create file %s SUCCESS\n", conf.DEFAULT_ENV_FILE)
	}

	files = append(files, conf.DEFAULT_ENV_FILE)
	return files
}

func createEnvFile(fpath string) error {
	f, err := os.Create(fpath)
	if err != nil {
		return fmt.Errorf("create env file(%s)err(%v)", fpath, err)
	}
	_, err = f.WriteString(getAllConfEnvStrDefault())
	if err != nil {
		return fmt.Errorf("write env file(%s)err(%v)", fpath, err)
	}
	return f.Close()
}

func getEnvDefaultStr(key, defval string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defval
	}
	return v
}

func getEnvDefaultBool(key string, defval bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defval
	}
	return strings.EqualFold(v, "true")
}

func getEnvDefaultInt(key string, defval int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defval
	}
	vv, _ := strconv.Atoi(v)
	return vv
}

// getEnvDefaultStrList 切片的每个元素去掉收尾空格，空字符串对应长度为0的空切片。
func getEnvDefaultStrList(key string, defval string, sep string) []string {
	v, ok := os.LookupEnv(key)
	if !ok {
		v = defval
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return []string{}
	}
	vv := strings.Split(v, sep)
	var result []string
	for _, iv := range vv {
		vvv := strings.TrimSpace(iv)
		if vvv != "" {
			result = append(result, vvv)
		}
	}
	return result
}

const ENV_FILE_CONTENT = `# 设置 VP_ENV_FILE 环境变量，可更改配置文件路径。

# 该目录存放程序运行时产生的文件
VP_RUNTIME_DIR = "%s"

# 傻瓜模式。当配置满足条件时，应用启动后，自动运行所有节点，执行测速，并设置速度最快的节点为系统代理。
# 适合打包现成的配置和应用，直接分享给小白用。
VP_AUTO_START = %t

# Web服务器端口。设置为0可禁用Web面板
VP_WEB_SERVER_PORT = %d

# 代理池的gRPC服务端口
VP_GRPC_PORT = %d

# v2ray可执行文件路径
# 例: "D:\\Users\\yourname\\v2ray-windows-64\\v2ray.exe" or "/root/v2ray-linux64/v2ray"
VP_V2RAY_PATH = "%s"

# 代理节点订阅地址
VP_SUBSCRIBE_URL = "%s"

# 若订阅地址无法直接访问，可指定订阅数据文件，数据文件内容为访问订阅地址获取的原始数据。
# 若有设置订阅数据文件，且文件内容不为空。则优先从该文件读取订阅节点信息。
VP_SUBSCRIBE_DATA_FILE = "%s"

# 设置HTTP代理，代理池每个节点的本地端口号，往后开始累加。为防止与常用端口冲突，尽量设大点。
# 支持http和socks5入站协议
VP_HTTP_PROXY = "%s"

# 节点测速的URL
VP_TEST_URL = "%s"

# 路由规则，用英文逗号,隔开。此处更改规则，必须删除 routing.rules.json 文件再重启才可生效。也可直接更改json文件(比如调整规则的优先级顺序)。
# 1. 规则是放在 routing.rules 这个数组当中，数组的内容是有顺序的，也就是说在这里规则是有顺序的，匹配规则时是从上往下匹配；
# 2. 当路由匹配到一个规则时就会跳出匹配而不会对之后的规则进行匹配；

# 路由规则：直连上网的域名列表和IP列表。例：baidu.com,domain:baidu.com,full:www.baidu.com,regexp:.*\.qq\.com$
VP_DIRECT_DOMAIN_LIST = "%s"
VP_DIRECT_IP_LIST = "%s"

# 路由规则：代理上网的域名列表和IP列表。例：youtube.com,domain:youtube.com,full:www.youtube.com,regexp:.*\.google.com$
VP_PROXY_DOMAIN_LIST = "%s"
VP_PROXY_IP_LIST = "%s"

# v2ray基于gRPC的API远程控制
VP_V2RAY_API_PORT = %d
`

//	func getAllConfEnvStr() string {
//		return fmt.Sprintf(ENV_FILE_CONTENT, vconf.RuntimeDir, GrpcPort, vconf.V2rayPath, SubscribeUrl, SubscribeDataFile, HttpProxy)
//	}
func getAllConfEnvStrDefault() string {
	return fmt.Sprintf(ENV_FILE_CONTENT, conf.DEFAULT_RUNTIME_DIR, conf.DEFAULT_AUTO_START, conf.DEFAULT_WEB_SERVER_PORT, conf.DEFAULT_GRPC_PORT, conf.DEFAULT_V2RAY_PATH,
		"", conf.DEFAULT_SUBSCRIBE_DATA_FILE, conf.DEFAULT_HTTP_PROXY, conf.DEFAULT_TEST_URL,
		conf.DEFAULT_DIRECT_DOMAIN_LIST, conf.DEFAULT_DIRECT_IP_LIST, conf.DEFAULT_PROXY_DOMAIN_LIST, conf.DEFAULT_PROXY_IP_LIST, conf.DEFAULT_V2RAY_API_PORT,
	)
}

func getConfByEnv() error {
	cf := conf.GetConf()
	cf.EnvFile = envFile
	cf.RuntimeDir = getEnvDefaultStr("VP_RUNTIME_DIR", conf.DEFAULT_RUNTIME_DIR)
	cf.AutoStart = getEnvDefaultBool("VP_AUTO_START", conf.DEFAULT_AUTO_START)
	cf.GrpcPort = getEnvDefaultInt("VP_GRPC_PORT", conf.DEFAULT_GRPC_PORT)
	cf.V2rayPath = getEnvDefaultStr("VP_V2RAY_PATH", conf.DEFAULT_V2RAY_PATH)
	cf.SubscribeUrl = getEnvDefaultStr("VP_SUBSCRIBE_URL", "")
	cf.SubscribeDataFile = getEnvDefaultStr("VP_SUBSCRIBE_DATA_FILE", conf.DEFAULT_SUBSCRIBE_DATA_FILE)
	cf.HttpProxy = getEnvDefaultStr("VP_HTTP_PROXY", conf.DEFAULT_HTTP_PROXY)
	cf.TestUrl = getEnvDefaultStr("VP_TEST_URL", conf.DEFAULT_TEST_URL)
	cf.DirectDomainList = getEnvDefaultStrList("VP_DIRECT_DOMAIN_LIST", conf.DEFAULT_DIRECT_DOMAIN_LIST, ",")
	cf.DirectIpList = getEnvDefaultStrList("VP_DIRECT_IP_LIST", conf.DEFAULT_DIRECT_IP_LIST, ",")
	cf.ProxyDomainList = getEnvDefaultStrList("VP_PROXY_DOMAIN_LIST", conf.DEFAULT_PROXY_DOMAIN_LIST, ",")
	cf.ProxyIpList = getEnvDefaultStrList("VP_PROXY_IP_LIST", conf.DEFAULT_PROXY_IP_LIST, ",")
	cf.V2rayApiPort = getEnvDefaultInt("VP_V2RAY_API_PORT", conf.DEFAULT_V2RAY_API_PORT)
	cf.WebServerPort = getEnvDefaultInt("VP_WEB_SERVER_PORT", conf.DEFAULT_WEB_SERVER_PORT)
	conf.SetConf(cf)
	vconf = cf

	logger := cf.GetLogger()
	hitmsg := fmt.Sprintf("请检查配置文件，路径:(%s)\n", cf.EnvFile)

	if cf.RuntimeDir == "" {
		hitmsg += "RuntimeDir 配置项不能为空"
		logger.Error(hitmsg)
		return fmt.Errorf("conf err: VP_RUNTIME_DIR could not be empty")
	}
	var err error
	var finfo fs.FileInfo
	finfo, err = os.Stat(cf.V2rayPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = fmt.Errorf("无法找到v2ray可执行文件【v2ray.exe(windows)/v2ray(linux or mac)】。can not find v2ray in path(%s)", cf.V2rayPath)
		}
	} else {
		if finfo.IsDir() {
			err = fmt.Errorf("检测到 VP_V2RAY_PATH 的值指向一文件夹，请改为可执行文件路径。")
		}
	}
	if err != nil {
		hitmsg += "VP_V2RAY_PATH 配置项错误，找不到可执行文件。\n"
		hitmsg += "请下载v2ray核心文件(https://github.com/v2fly/v2ray-core/releases)\n"
		hitmsg += "下载后，请【删除或改名】默认配置文件(config.json)，防止错误读取"
		logger.Error(hitmsg)
		return err
	}

	if !miniutils.IsPathExists(cf.RuntimeDir) {
		fmt.Printf("------创建runtime目录(%s)--\n", cf.RuntimeDir)
		err = os.Mkdir(cf.RuntimeDir, 0755)
		if err != nil {
			fmt.Printf("----runtime目录(%s)创建失败(%v)---\n", cf.RuntimeDir, err)
			return err
		}
	}
	if !miniutils.IsPathExists(cf.SubscribeDataFile) {
		fmt.Printf("------创建SubscribeDataFile文件(%s)--\n", cf.SubscribeDataFile)
		err = cf.UpdateSubscribeData("")
	}
	return err
}
