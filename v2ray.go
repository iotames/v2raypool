package v2raypool

import (
	"fmt"
	"os"
	"os/exec"
)

// "protocol":"vmess"
type V2rayNode struct {
	Protocol, Add, Host, Id, Net, Path, Port, Ps, Tls, Type string
	V, Aid                                                  int
}

type V2rayServer struct {
	v2rayPath  string
	selectNode V2rayNode
	localPort  int
}

func NewV2ray(v2rayPath string) *V2rayServer {
	return &V2rayServer{v2rayPath: v2rayPath}
}

func (v *V2rayServer) SetNode(n V2rayNode) *V2rayServer {
	v.selectNode = n
	return v
}

func (v *V2rayServer) SetPort(port int) *V2rayServer {
	v.localPort = port
	return v
}

func (v *V2rayServer) getExeCmd() *exec.Cmd {
	// https://www.v2fly.org/v5/config/overview.html
	cmd := exec.Command(v.v2rayPath, "run") // 默认为v4配置格式。添加命令参数 "-format", "jsonv5" 后，才是v5的配置
	cmd.Stdin = getV2rayConfig(v.selectNode, v.localPort)
	// cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func (v *V2rayServer) Run() {
	err := v.getExeCmd().Run()
	if err != nil {
		fmt.Printf("---cmdRunError(%s)", err)
		panic(err)
	}
}

func (v *V2rayServer) Start() (cmd *exec.Cmd, err error) {
	cmd = v.getExeCmd()
	err = cmd.Start()
	if err != nil {
		fmt.Printf("---cmdStartError(%s)", err)
	}
	return
}
