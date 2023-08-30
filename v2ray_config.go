package v2raypool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

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
