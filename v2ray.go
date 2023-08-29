package v2raypool

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// "protocol":"vmess"
type V2rayNode struct {
	Protocol, Add, Host, Id, Net, Path, Port, Ps, Tls, Type string
	V, Aid                                                  int
}

// github.com/v2fly/v2ray-core/v5/infra/conf/v5cfg
// 执行 ./v2ray run -c $configure_file_name -format jsonv5 命令以运行您的配置文件。

// v4 {"port":%d,"listen":"%s","protocol":"http","settings":{"auth":"noauth","udp":true,"ip":"%s"}}
type V2rayInbound struct {
	Protocol string
	Port     int
	Listen   string // 默认值为 "0.0.0.0"
	// Tag string // 此入站连接的标识，用于在其它的配置中定位此连接。当其不为空时，其值必须在所有 tag 中唯一。
	Settings json.RawMessage `json:"settings"` // {"auth":"noauth","udp":true,"ip":"%s"}
}

type V2rayOutbound struct {
	Protocol    string          `json:"protocol"`
	SendThrough string          // 用于发送数据的 IP 地址，当主机有多个 IP 地址时有效，默认值为 "0.0.0.0"。
	Tag         string          `json:"tag"`
	Settings    json.RawMessage `json:"settings"`
	// "streamSettings":{"network":"%s","tlsSettings":{"disableSystemRoot":false},"wsSettings":{"path":""},"xtlsSettings":{"disableSystemRoot":false}}
	StreamSetting json.RawMessage `json:"streamSettings"`
}

type V2rayConfig struct {
	Log json.RawMessage `json:"log"`
	// Dns        json.RawMessage            `json:"dns"`
	// Router     json.RawMessage            `json:"router"`
	Inbounds  []V2rayInbound  `json:"inbounds"`
	Outbounds []V2rayOutbound `json:"outbounds"`
	// Services   map[string]json.RawMessage `json:"services"`
	// Extensions []json.RawMessage          `json:"extension"`
}

func getV2rayConfig(n V2rayNode, inPort int) io.Reader {
	vconf := V2rayConfig{}
	vconf.Log = []byte(`{"loglevel":"debug"}`) // v4
	// vconf.Log = []byte(`{"access":{"type":"Console","level":"Debug"}}`) // v5
	inAddr := "0.0.0.0" // "127.0.0.1"
	inSet := fmt.Sprintf(`{"auth":"noauth","udp":true,"ip":"%s"}`, inAddr)

	inbd1 := V2rayInbound{
		Protocol: "http", // socket
		Port:     inPort,
		Listen:   inAddr,
		Settings: json.RawMessage(inSet),
	}

	outbd := V2rayOutbound{Protocol: n.Protocol, SendThrough: "0.0.0.0", Tag: "PROXY"}
	outSet := fmt.Sprintf(`{"vnext":[{"address":"%s","port":%s,"users":[{"id":"%s","security":"aes-128-gcm"}]}]}`, n.Add, n.Port, n.Id)
	outbd.Settings = json.RawMessage(outSet)
	// "streamSettings":{"network":"%s","tlsSettings":{"disableSystemRoot":false},"wsSettings":{"path":""},"xtlsSettings":{"disableSystemRoot":false}}
	streamSet := fmt.Sprintf(`{"network":"%s"}`, n.Net) // v4: network, v5: transport
	outbd.StreamSetting = json.RawMessage(streamSet)

	var inbds []V2rayInbound = []V2rayInbound{inbd1}
	vconf.Inbounds = inbds
	vconf.Outbounds = []V2rayOutbound{outbd}
	vconfjs, err := json.Marshal(vconf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n---v2ray.config=(%+v)--\n", string(vconfjs))
	return bytes.NewReader(vconfjs)
}

type V2rayServer struct {
	v2rayPath  string
	selectNode V2rayNode
	localPort  int
}

func NewV2ray(v2rayPath string, localPort int) *V2rayServer {
	return &V2rayServer{v2rayPath: v2rayPath, localPort: localPort}
}

func (v *V2rayServer) SetNode(n V2rayNode) *V2rayServer {
	v.selectNode = n
	return v
}
func (v *V2rayServer) SetPort(port int) *V2rayServer {
	v.localPort = port
	return v
}

// ParseNodes 解析节点 Add, Ps ...
// {"add":"jp6.v2u9.top","host":"","id":"0999AE93-1330-4A75-DBC1-0DD545F7DD60","net":"ws","path":"","port":"41444","ps":"u9un-v2-JP-Tokyo6(1)","tls":"","v":2,"aid":0,"type":"none"}
func ParseV2rayNodes(data string) []V2rayNode {
	fmt.Println("-----Begin--ParseV2rayNodes-------")
	sss := strings.Split(data, "\n")
	var nds []V2rayNode
	var err error
	for i, d := range sss {
		ninfo := strings.Split(d, "://")
		if len(ninfo) == 2 {
			n := V2rayNode{}
			var b []byte
			b, err = base64.StdEncoding.DecodeString(ninfo[1])
			if err != nil {
				fmt.Printf("\n---ParseV2rayNodes--Base64.Decode--err(%v)---RAW(%s)---\n", err, d)
				continue
			}
			err = json.Unmarshal(b, &n)
			if err != nil {
				fmt.Printf("\n---ParseV2rayNodes--json.Unmarshal err---[%d]--err(%v)--RAW(%s)--data(%s)---\n", i, err, d, string(b))
				continue
			}
			n.Protocol = ninfo[0]
			nds = append(nds, n)
		}
	}
	return nds
}

func (v *V2rayServer) getExeCmd() *exec.Cmd {
	// https://www.v2fly.org/v5/config/overview.html
	// fmt.Println(v.v2rayPath)
	cmd := exec.Command(v.v2rayPath, "run") // 默认为v4配置格式。添加命令参数 "-format", "jsonv5" 后，才是v5的配置

	// inAddr := "127.0.0.1"
	// Inport := 1080
	// inputStr := fmt.Sprintf(`{"log":{"loglevel":"debug"},"inbounds":[{"port":%d,"listen":"%s","protocol":"http","settings":{"auth":"noauth","udp":true,"ip":"%s"}}],"outbounds":[{"mux":{},"protocol":"%s","sendThrough":"0.0.0.0","settings":{"vnext":[{"address":"%s","port":%s,"users":[{"id":"%s","security":"aes-128-gcm"}]}]},,"tag":"%s"}]}`, Inport, inAddr, inAddr, n.Protocol, n.Add, n.Port, n.Id, n.Net, "PROXY")
	// cmd.Stdin = strings.NewReader(inputStr)

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
