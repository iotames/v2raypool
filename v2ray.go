package v2raypool

import (
	"fmt"
	"os"
	"os/exec"
)

// "protocol":"vmess"
type V2rayNode struct {
	Protocol, Add, Host, Id, Net, Path, Ps, Tls, Type string
	V, Aid                                            int
	Port                                              any
}

type V2rayServer struct {
	v2rayPath  string
	selectNode V2rayNode
	localPort  int
	jconf      *JsonConfig
	cmd        *exec.Cmd
}

func NewV2ray(v2rayPath string) *V2rayServer {
	return &V2rayServer{v2rayPath: v2rayPath}
}
func (v V2rayServer) GetJsonConfig() JsonConfig {
	return *v.jconf
}
func (v *V2rayServer) SetNode(n V2rayNode) *V2rayServer {
	v.selectNode = n
	return v
}

func (v *V2rayServer) SetPort(port int) *V2rayServer {
	v.localPort = port
	return v
}
func (v V2rayServer) GetLocalPort() int {
	return v.localPort
}
func (v V2rayServer) GetV2rayConfigV4() V2rayConfigV4 {
	vconf := V2rayConfigV4{}
	v.jconf.Decode(&vconf)
	return vconf
}
func (v V2rayServer) GetLocalPortList() []int {
	var result []int
	vconf := v.GetV2rayConfigV4()
	for _, in := range vconf.Inbounds {
		result = append(result, in.Port)
	}
	return result
}
func (v V2rayServer) GetExeCmd() *exec.Cmd {
	return v.cmd
}
func (v *V2rayServer) setExeCmd(configFile string) {
	// https://www.v2fly.org/v5/config/overview.html
	// 默认为v4配置格式。添加命令参数 "-format", "jsonv5" 后，才是v5的配置
	if configFile == "" {
		v.cmd = exec.Command(v.v2rayPath, "run")
		v.jconf = getV2rayConfigV4(v.selectNode, v.localPort)
		v.cmd.Stdin = v.jconf.Reader()
	} else {
		v.cmd = exec.Command(v.v2rayPath, "run", "-c", configFile)
		v.jconf = NewJsonConfigFromFile(configFile)
	}
	// cmd.Stdout = os.Stdout
	v.cmd.Stderr = os.Stderr
}

// Start 启动v2ray进程。非代理池模式，会读取 routing.rules.json 文件。
// routing.rules.json 的值会覆盖 v2ray.config.json 对应配置项的值
func (v *V2rayServer) Start(configFile string) error {
	v.setExeCmd(configFile)
	err := v.cmd.Start()
	if err == nil {
		if v.localPort > 0 {
			err1 := v.jconf.SaveToFile(V2RAY_CONFIG_FILE)
			if err1 != nil {
				fmt.Printf("--------V2rayServer.Start()--jsonFileSaveFail(%v)--", err1)
			}
		}
		fmt.Printf("----v2ray-Start(%s)--Pid(%d)---\n", configFile, v.cmd.Process.Pid)
	}
	return err
}
