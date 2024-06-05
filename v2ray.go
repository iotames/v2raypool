package v2raypool

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// "protocol":"vmess"
type V2rayNode struct {
	Protocol, Add, Host, Id, Net, Path, Ps, Tls, Type string
	V, Aid, Port                                      json.Number
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
func (v V2rayServer) GetExeCmd() *exec.Cmd {
	return v.cmd
}
func (v *V2rayServer) setExeCmd(configFile string) {
	// https://www.v2fly.org/v5/config/overview.html
	// 默认为v4配置格式。添加命令参数 "-format", "jsonv5" 后，才是v5的配置
	if configFile == "" {
		v.jconf = getV2rayConfigV4(v.selectNode, v.localPort)
		if v.localPort > 0 {
			configFile = V2RAY_CONFIG_FILE
			// 再额外添加一个socks端口或http端口
			vconf := v.GetV2rayConfigV4()
			inbd1 := vconf.Inbounds[0]
			proto2 := "socks"
			if inbd1.Protocol == "socks" {
				proto2 = "http"
			}
			lport2 := v.localPort - 1
			inbd2 := NewV2rayInboundV4(proto2, lport2)
			vconf.Inbounds = []V2rayInbound{inbd1, inbd2}
			err := v.jconf.SetContent(vconf)
			if err != nil {
				panic(err)
			}
		} else {
			configFile = V2RAYPOOL_CONFIG_FILE
		}
		err := v.jconf.SaveToFile(configFile)
		if err != nil {
			fmt.Printf("--------V2rayServer.setExeCmd()--jsonFileSaveFail(%v)--", err)
		}
		// v5.7使用标准输入读取配置正常，v5.14.1则出现BUG。故一律使用配置文件读取配置
		// v.cmd = exec.Command(v.v2rayPath, "run")
		// v.cmd.Stdin = v.jconf.Reader()
	} else {
		v.jconf = NewJsonConfigFromFile(configFile)
	}
	// v5.7使用标准输入读取配置正常，v5.14.1则出现BUG。故一律使用配置文件读取配置
	v.cmd = exec.Command(v.v2rayPath, "run", "-c", configFile)
	// cmd.Stdout = os.Stdout
	v.cmd.Stderr = os.Stderr
}

// Start 启动v2ray进程。非代理池模式，会读取 routing.rules.json 文件。
// routing.rules.json 的值会覆盖 v2ray.config.json 对应配置项的值
func (v *V2rayServer) Start(configFile string) error {
	v.setExeCmd(configFile)
	err := v.cmd.Start()
	if err == nil {
		fmt.Printf("--SUCCESS--Start(%s)-v2ray-Pid(%d)---\n", configFile, v.cmd.Process.Pid)
	}
	return err
}
