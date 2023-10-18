package v2raypool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/iotames/v2raypool/conf"
)

const TAG_OUTBOUND_ACTIVE = "TAG_ACTIVE_OUTBOUND"
const TAG_OUTBOUND_API = "TAG_OUTBOUND_API"
const TAG_INBOUND_API = "TAG_INBOUND_API"

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
	Api json.RawMessage `json:"api"`
	// Dns       json.RawMessage `json:"dns"`
	Routing   json.RawMessage `json:"routing"`
	Inbounds  []V2rayInbound  `json:"inbounds"`
	Outbounds []V2rayOutbound `json:"outbounds"`
}

// V2rayRouteRule https://www.v2fly.org/config/routing.html#ruleobject
type V2rayRouteRule struct {
	DomainMatcher string   `json:"domainMatcher"`
	Type          string   `json:"type"`
	Domains       []string `json:"domains"`
	Ip            []string `json:"ip"`
	InboundTag    []string `json:"inboundTag"`
	OutboundTag   string   `json:"outboundTag"` //direct

}

func newRouteRule(outboundTag string) V2rayRouteRule {
	r := V2rayRouteRule{DomainMatcher: "mph", Type: "field", OutboundTag: outboundTag}
	return r
}

func getRouteRules() []V2rayRouteRule {
	cf := conf.GetConf()
	var rules []V2rayRouteRule
	rules = append(rules, V2rayRouteRule{Type: "field", InboundTag: []string{TAG_INBOUND_API}, OutboundTag: TAG_OUTBOUND_API})
	if len(cf.DirectDomainList) > 0 || len(cf.DirectIpList) > 0 {
		rule1 := newRouteRule("DIRECT")
		if len(cf.DirectDomainList) > 0 {
			rule1.Domains = cf.DirectDomainList
		}
		if len(cf.DirectIpList) > 0 {
			rule1.Ip = cf.DirectIpList
		}
		rules = append(rules, rule1)
	}

	if len(cf.ProxyDomainList) > 0 || len(cf.ProxyIpList) > 0 {
		rule2 := newRouteRule(TAG_OUTBOUND_ACTIVE)
		if len(cf.ProxyDomainList) > 0 {
			rule2.Domains = cf.ProxyDomainList
		}
		if len(cf.ProxyIpList) > 0 {
			rule2.Ip = cf.ProxyIpList
		}
		rules = append(rules, rule2)
	}
	for i := 0; i < 100; i++ {
		tag := getProxyNodeTag(i)
		// 添加路由规则，相同标签的每一个出站和入站一一对应
		rules = append(rules, V2rayRouteRule{Type: "field", InboundTag: []string{tag}, OutboundTag: tag})
	}
	return rules
}

func getProxyNodeTag(index int) string {
	return fmt.Sprintf("TAG_PROXY_%d", index)
}

func getV2rayConfigV4(n V2rayNode, inPort int) io.Reader {
	cf := conf.GetConf()
	vconf := V2rayConfigV4{}
	vconf.Log = []byte(`{"loglevel":"debug"}`) // v4
	// vconf.Log = []byte(`{"access":{"type":"Console","level":"Debug"}}`) // v5

	var rulesb []byte
	rulesb, _ = json.Marshal(getRouteRules())

	// 域名解析策略domainStrategy: 不要使用 AsIs , 要用 IPOnDemand 或者 IPIfNonMatch ，否则直连规则不生效
	routing := fmt.Sprintf(`{
		"domainStrategy": "IPOnDemand",
		"domainMatcher": "mph",
		"rules": %s,
		"balancers": []
	}`, string(rulesb))
	vconf.Api = json.RawMessage(fmt.Sprintf(`{"tag": "%s", "services": ["HandlerService"]}`, TAG_OUTBOUND_API))
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

	inbdapi := V2rayInbound{
		Listen:   "127.0.0.1",
		Port:     cf.V2rayApiPort,
		Protocol: "dokodemo-door",
		Settings: json.RawMessage(`{"address": "127.0.0.1"}`),
		Tag:      TAG_INBOUND_API,
	}
	vconf.Inbounds = []V2rayInbound{inbdapi}

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
		Tag: TAG_OUTBOUND_ACTIVE,
	}
	outSet := fmt.Sprintf(`{"vnext":[{"address":"%s","port":%s,"users":[{"id":"%s","security":"aes-128-gcm"}]}]}`, n.Add, n.Port, n.Id)
	outbd.Settings = json.RawMessage(outSet)
	// "streamSettings":{"network":"%s","tlsSettings":{"disableSystemRoot":false},"wsSettings":{"path":""},"xtlsSettings":{"disableSystemRoot":false}}
	streamSet := fmt.Sprintf(`{"network":"%s"}`, n.Net) // v4: network, v5: transport
	outbd.StreamSetting = json.RawMessage(streamSet)

	outdirect := V2rayOutbound{Protocol: "freedom", Tag: "DIRECT"} // "sendThrough": "0.0.0.0",
	outdirect.Settings = json.RawMessage(`{}`)
	outdirect.StreamSetting = json.RawMessage(`{}`)
	if inbd1.Port != 0 {
		vconf.Inbounds = append(vconf.Inbounds, inbd1)
	}
	vconf.Outbounds = []V2rayOutbound{outdirect}
	if n.Add != "" {
		vconf.Outbounds = append(vconf.Outbounds, outbd)
	}

	vconfjs, err := json.Marshal(vconf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n---getV2rayConfigV4--Outbounds-len(%d)-val(%+v)---v2ray.config=(%s)--\n", len(vconf.Outbounds), vconf.Outbounds, string(vconfjs))
	return bytes.NewReader(vconfjs)
}

func getV2rayConfig(n V2rayNode, inPort int) io.Reader {
	return getV2rayConfigV4(n, inPort)
}
