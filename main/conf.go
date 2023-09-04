package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotames/miniutils"
	conf "github.com/iotames/v2raypool/conf"
	"github.com/joho/godotenv"
)

const WORK_ENV_FILE = ".env"
const DEFAULT_ENV_FILE = "default.env"

var envFile string

func setEnvFile() {
	envFile = os.Getenv("VP_ENV_FILE")
	if envFile == "" {
		envFile = WORK_ENV_FILE
	}
}

func LoadEnv() {
	setEnvFile()
	efiles := initEnvFile()
	err := godotenv.Load(efiles...)
	if err != nil {
		panic(fmt.Errorf("godotenv.Load(%v)err(%v)", efiles, err))
	}
	getConfByEnv()
}

func initEnvFile() []string {
	var err error
	var files []string
	var createNewEnvfile bool
	if !miniutils.IsPathExists(envFile) {
		err = createEnvFile(envFile)
		if err != nil {
			panic(err)
		}
		files = append(files, envFile)
		fmt.Printf("Create file %s SUCCESS\n", envFile)
		createNewEnvfile = true
	}
	files = append(files, envFile)

	if miniutils.IsPathExists(DEFAULT_ENV_FILE) {
		files = append(files, DEFAULT_ENV_FILE)
	} else {
		if createNewEnvfile {
			err = createEnvFile(DEFAULT_ENV_FILE)
			if err != nil {
				logger := miniutils.GetLogger("")
				logger.Warnf("initEnvFile(%s)err(%v)", DEFAULT_ENV_FILE, err)
				return files
			}
			files = append(files, DEFAULT_ENV_FILE)
			fmt.Printf("Create file %s SUCCESS\n", DEFAULT_ENV_FILE)
		}
	}
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

func getEnvDefaultInt(key string, defval int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defval
	}
	vv, _ := strconv.Atoi(v)
	return vv
}

const ENV_FILE_CONTENT = `# 设置 VP_ENV_FILE 环境变量，可更改配置文件路径。

# 该目录存放程序运行时产生的文件
VP_RUNTIME_DIR = "%s"

# gRPC服务端口
VP_GRPC_PORT = %d

# v2ray可执行文件路径
# 例: "D:\\Users\\yourname\\v2ray-windows-64\\v2ray.exe" or "/root/v2ray-linux64/v2ray"
VP_V2RAY_PATH = "%s"

# 代理节点订阅地址
VP_SUBSCRIBE_URL = "%s"

# 若订阅地址无法直接访问，可指定订阅数据文件，数据文件内容为访问订阅地址获取的原始数据。
# 若有设置订阅数据文件，且文件内容不为空。则优先从该文件读取订阅节点信息。
VP_SUBSCRIBE_DATA_FILE = "%s"

# 设置HTTP代理，代理池的端口号网上开始累加。为防止与常用端口冲突，尽量设大点。
VP_HTTP_PROXY = "%s"
`

//	func getAllConfEnvStr() string {
//		return fmt.Sprintf(ENV_FILE_CONTENT, vconf.RuntimeDir, GrpcPort, vconf.V2rayPath, SubscribeUrl, SubscribeDataFile, HttpProxy)
//	}
func getAllConfEnvStrDefault() string {
	return fmt.Sprintf(ENV_FILE_CONTENT, conf.DEFAULT_RUNTIME_DIR, conf.DEFAULT_GRPC_PORT, conf.DEFAULT_V2RAY_PATH, "", conf.DEFAULT_SUBSCRIBE_DATA_FILE, conf.DEFAULT_HTTP_PROXY)
}

func getConfByEnv() {
	cf := conf.GetConf()
	cf.EnvFile = envFile
	cf.RuntimeDir = getEnvDefaultStr("VP_RUNTIME_DIR", conf.DEFAULT_RUNTIME_DIR)
	cf.GrpcPort = getEnvDefaultInt("VP_GRPC_PORT", conf.DEFAULT_GRPC_PORT)
	cf.V2rayPath = getEnvDefaultStr("VP_V2RAY_PATH", conf.DEFAULT_V2RAY_PATH)
	cf.SubscribeUrl = getEnvDefaultStr("VP_SUBSCRIBE_URL", "")
	cf.SubscribeDataFile = getEnvDefaultStr("VP_SUBSCRIBE_DATA_FILE", conf.DEFAULT_SUBSCRIBE_DATA_FILE)
	cf.HttpProxy = getEnvDefaultStr("VP_HTTP_PROXY", conf.DEFAULT_HTTP_PROXY)
	conf.SetConf(cf)
	logger := miniutils.GetLogger("")
	hitmsg := fmt.Sprintf("请检查配置文件，路径:(%s)\n", cf.EnvFile)

	if cf.RuntimeDir == "" {
		hitmsg += "RuntimeDir 配置项不能为空"
		logger.Error(hitmsg)
		panic("RuntimeDir can not be empty")
	}
	if !miniutils.IsPathExists(cf.V2rayPath) {
		hitmsg += "VP_V2RAY_PATH 配置项错误，找不到可执行文件。\n"
		hitmsg += "请下载v2ray核心文件(https://github.com/v2fly/v2ray-core/releases)\n"
		hitmsg += "下载后，请【删除或改名】默认配置文件(config.json)，防止错误读取"
		logger.Error(hitmsg)
		panic(fmt.Errorf("无法找到v2ray可执行文件【v2ray.exe(windows)/v2ray(linux or mac)】。can not find v2ray in path(%s)", cf.V2rayPath))
	}
	var err error
	if !miniutils.IsPathExists(cf.RuntimeDir) {
		fmt.Printf("------创建runtime目录(%s)--\n", cf.RuntimeDir)
		err = os.Mkdir(cf.RuntimeDir, 0755)
		if err != nil {
			fmt.Printf("----runtime目录(%s)创建失败(%v)---\n", cf.RuntimeDir, err)
			panic(err)
		}
	}
	if !miniutils.IsPathExists(cf.SubscribeDataFile) {
		fmt.Printf("------创建SubscribeDataFile文件(%s)--\n", cf.SubscribeDataFile)
		cf.UpdateSubscribeData("")
	}
}

func UpdateConf(mp map[string]string, fpath string) error {
	for k, v := range mp {
		os.Setenv(k, v)
	}
	return godotenv.Write(mp, fpath)
}
