package main

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/iotames/easyconf"
	"github.com/iotames/miniutils"
	conf "github.com/iotames/v2raypool/conf"
)

const WORK_ENV_FILE = ".env"

var envFile string
var vconf conf.Conf

func setEnvFile() {
	// 设置 VP_ENV_FILE 环境变量，可更改配置文件路径。
	envFile = os.Getenv("VP_ENV_FILE")
	if envFile == "" {
		envFile = WORK_ENV_FILE
	}
}

func LoadEnv() error {
	setEnvFile()
	return getConfByEnv()
}

func getConfByEnv() error {
	cf := conf.GetConf()
	cf.EnvFile = envFile

	// 设置 VP_ENV_FILE 环境变量，可更改配置文件路径。
	ecf := easyconf.NewConf(envFile, conf.DEFAULT_ENV_FILE)
	ecf.StringVar(&cf.RuntimeDir, "VP_RUNTIME_DIR", conf.DEFAULT_RUNTIME_DIR, "该目录存放程序运行时产生的文件")

	ecf.BoolVar(&cf.AutoStart, "VP_AUTO_START", conf.DEFAULT_AUTO_START, "启动傻瓜模式", "启动后，会自动连接所有节点，执行测速，并设置速度最快的节点为系统代理", "适合打包现成的配置和应用，直接分享给小白用。免设置")
	ecf.IntVar(&cf.GrpcPort, "VP_GRPC_PORT", conf.DEFAULT_GRPC_PORT, "代理池的gRPC服务端口")
	ecf.StringVar(&cf.V2rayPath, "VP_V2RAY_PATH", conf.DEFAULT_V2RAY_PATH, "v2ray可执行文件路径", `例: "D:\\Users\\yourname\\v2ray-windows-64\\v2ray.exe" or "/root/v2ray-linux64/v2ray"`)
	ecf.StringVar(&cf.SubscribeUrl, "VP_SUBSCRIBE_URL", "", "代理节点订阅地址")
	ecf.StringVar(&cf.SubscribeDataFile, "VP_SUBSCRIBE_DATA_FILE", conf.DEFAULT_SUBSCRIBE_DATA_FILE, "订阅数据文件", "若订阅地址无法直接访问，可指定订阅数据文件，数据文件内容为访问订阅地址获取的原始数据。", "若有设置订阅数据文件，且文件内容不为空。则优先从该文件读取订阅节点信息。")
	ecf.StringVar(&cf.HttpProxy, "VP_HTTP_PROXY", conf.DEFAULT_HTTP_PROXY, "HTTP系统代理", "代理池每个节点的本地端口号，从系统代理往后开始累加。为防止与常用端口冲突，尽量设大点。", "支持 http:// 和 socks5:// 协议。若http端口为30000，则socks5端口为29999。反之亦然。")
	ecf.StringVar(&cf.TestUrl, "VP_TEST_URL", conf.DEFAULT_TEST_URL, "节点测速的URL")
	rulesComment := []string{
		"路由规则，用英文逗号,隔开。此处更改规则，必须删除 routing.rules.json 文件再重启才可生效。也可直接更改json文件(比如调整规则的优先级顺序)。",
		"1. 规则是放在 routing.rules 这个数组当中，数组的内容是有顺序的，也就是说在这里规则是有顺序的，匹配规则时是从上往下匹配；",
		"2. 当路由匹配到一个规则时就会跳出匹配而不会对之后的规则进行匹配；",
	}
	ecf.AddComment("", rulesComment...)
	ecf.StringListVar(&cf.DirectDomainList, "VP_DIRECT_DOMAIN_LIST", strings.Split(conf.DEFAULT_DIRECT_DOMAIN_LIST, ","), "直连上网的域名列表", `例：baidu.com,domain:baidu.com,full:www.baidu.com,regexp:.*\.qq\.com$`)
	ecf.StringListVar(&cf.DirectIpList, "VP_DIRECT_IP_LIST", strings.Split(conf.DEFAULT_DIRECT_IP_LIST, `,`), "直连上网的IP列表")
	ecf.StringListVar(&cf.ProxyDomainList, "VP_PROXY_DOMAIN_LIST", strings.Split(conf.DEFAULT_PROXY_DOMAIN_LIST, `,`), "代理上网的域名列表")
	ecf.StringListVar(&cf.ProxyIpList, "VP_PROXY_IP_LIST", strings.Split(conf.DEFAULT_PROXY_IP_LIST, `,`), "代理上网的IP列表")
	ecf.IntVar(&cf.V2rayApiPort, "VP_V2RAY_API_PORT", conf.DEFAULT_V2RAY_API_PORT, "v2ray的API控制端口")
	ecf.IntVar(&cf.WebServerPort, "VP_WEB_SERVER_PORT", conf.DEFAULT_WEB_SERVER_PORT, "Web服务器端口", "设置为0可禁用Web面板")
	ecf.Parse()
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
		// hitmsg += "下载后，请【删除或改名】默认配置文件(config.json)，防止错误读取"
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
