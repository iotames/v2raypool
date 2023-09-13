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
	Listen   string          // 默认值为 "0.0.0.0"
	Tag      string          // 此入站连接的标识，用于在其它的配置中定位此连接。当其不为空时，其值必须在所有 tag 中唯一。
	Settings json.RawMessage `json:"settings"` // {"auth":"noauth","udp":true,"ip":"%s"}
}

type V2rayOutbound struct {
	Protocol string `json:"protocol"`
	// SendThrough string          // 用于发送数据的 IP 地址，当主机有多个 IP 地址时有效，默认值为 "0.0.0.0"。
	Tag      string          `json:"tag"`
	Settings json.RawMessage `json:"settings"`
	// "streamSettings":{"network":"%s","tlsSettings":{"disableSystemRoot":false},"wsSettings":{"path":""},"xtlsSettings":{"disableSystemRoot":false}}
	StreamSetting json.RawMessage `json:"streamSettings"`
}

type V2rayConfigV5 struct {
	Log        json.RawMessage            `json:"log"`
	Dns        json.RawMessage            `json:"dns"`
	Router     json.RawMessage            `json:"router"`
	Inbounds   []V2rayInbound             `json:"inbounds"`
	Outbounds  []V2rayOutbound            `json:"outbounds"`
	Services   map[string]json.RawMessage `json:"services"`
	Extensions []json.RawMessage          `json:"extension"`
}

type V2rayConfigV4 struct {
	Log json.RawMessage `json:"log"`
	// Dns       json.RawMessage `json:"dns"`
	Routing   json.RawMessage `json:"routing"`
	Inbounds  []V2rayInbound  `json:"inbounds"`
	Outbounds []V2rayOutbound `json:"outbounds"`
}

// V2rayRouteRuleV4 https://www.v2fly.org/config/routing.html#ruleobject
type V2rayRouteRuleV4 struct {
	DomainMatcher string   `json:"domainMatcher"`
	Type          string   `json:"type"`
	Domains       []string `json:"domains"`
	Ip            []string `json:"ip"`
	// InboundTag []string `json:"inboundTag"`
	OutboundTag string `json:"outboundTag"` //direct

}

func newRouteRuleV4(outboundTag string) V2rayRouteRuleV4 {
	r := V2rayRouteRuleV4{DomainMatcher: "mph", Type: "field", OutboundTag: outboundTag}
	return r
}

func getV2rayConfigV4(n V2rayNode, inPort int) io.Reader {
	vconf := V2rayConfigV4{}
	vconf.Log = []byte(`{"loglevel":"debug"}`) // v4
	// vconf.Log = []byte(`{"access":{"type":"Console","level":"Debug"}}`) // v5

	rule1 := newRouteRuleV4("DIRECT")
	rule1.Ip = []string{
		"geoip:private",
		"geoip:cn",
	}
	// TODO 添加自定义IP
	rule1.Domains = []string{
		"geosite:cn",
	}
	// TODO 添加自定义domain

	rule2 := newRouteRuleV4("PROXY")
	rule2.Ip = []string{
		"geoip:!cn",
	}
	rule2.Domains = []string{
		"geosite:google",
	}

	rules := []V2rayRouteRuleV4{rule1, rule2}

	var rulesb []byte
	rulesb, _ = json.Marshal(rules)

	routing := fmt.Sprintf(`{
		"domainStrategy": "AsIs",
		"domainMatcher": "mph",
		"rules": %s,
		"balancers": []
	}`, string(rulesb))

	vconf.Routing = json.RawMessage(routing)

	inAddr := "0.0.0.0" // "127.0.0.1"
	inSet := fmt.Sprintf(`{"auth":"noauth","udp":true,"ip":"%s"}`, inAddr)
	// 	SOCKS5 通过 UDP ASSOCIATE 命令建立 UDP 会话。服务端在对客户端发来的该命令的回复中，指定客户端发包的目标地址。
	// v4.34.0+: 默认值为空，此时对于通过本地回环 IPv4/IPv6 连接的客户端，回复对应的回环 IPv4/IPv6 地址；对于非本机的客户端，回复当前入站的监听地址。
	// v4.33.0 及更早版本: 默认值 127.0.0.1。

	inbd1 := V2rayInbound{
		Protocol: "http", // socket
		Port:     inPort,
		Listen:   inAddr,
		Settings: json.RawMessage(inSet),
		Tag:      "http_IN",
	}
	// inbd2 := V2rayInbound{
	// 	Protocol: "socket",
	// 	Port:     inPort, // 端口号相同会冲突
	// 	Listen:   inAddr,
	// 	Settings: json.RawMessage(inSet),
	// 	Tag: "socks_IN",
	// }

	outbd := V2rayOutbound{
		Protocol: n.Protocol,
		// SendThrough: "0.0.0.0",
		Tag: "PROXY",
	}
	outSet := fmt.Sprintf(`{"vnext":[{"address":"%s","port":%s,"users":[{"id":"%s","security":"aes-128-gcm"}]}]}`, n.Add, n.Port, n.Id)
	outbd.Settings = json.RawMessage(outSet)
	// "streamSettings":{"network":"%s","tlsSettings":{"disableSystemRoot":false},"wsSettings":{"path":""},"xtlsSettings":{"disableSystemRoot":false}}
	streamSet := fmt.Sprintf(`{"network":"%s"}`, n.Net) // v4: network, v5: transport
	outbd.StreamSetting = json.RawMessage(streamSet)

	outdirect := V2rayOutbound{Protocol: "freedom", Tag: "DIRECT"} // "sendThrough": "0.0.0.0",
	outdirect.Settings = json.RawMessage(`{"domainStrategy": "AsIs","redirect": ":0"}`)
	outdirect.StreamSetting = json.RawMessage(`{}`)

	vconf.Inbounds = []V2rayInbound{inbd1}
	vconf.Outbounds = []V2rayOutbound{outbd, outdirect}

	vconfjs, err := json.Marshal(vconf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n---v2ray.config=(%+v)--\n", string(vconfjs))
	return bytes.NewReader(vconfjs)
}

func getV2rayConfig(n V2rayNode, inPort int) io.Reader {
	return getV2rayConfigV4(n, inPort)
}
