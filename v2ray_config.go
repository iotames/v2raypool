package v2raypool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/iotames/miniutils"
	"github.com/iotames/v2raypool/conf"
)

const TAG_OUTBOUND_DIRECT = "DIRECT"
const TAG_OUTBOUND_ACTIVE = "TAG_ACTIVE_OUTBOUND"
const TAG_OUTBOUND_API = "TAG_OUTBOUND_API"
const TAG_INBOUND_API = "TAG_INBOUND_API"

// 若存在，则优先读取。否则创建(v2raypool)
const ROUTING_RULES_FILE = "routing.rules.json"

// 只写文件，被 ROUTING_RULES_FILE 文件覆盖部分值(v2raypool)
const V2RAY_CONFIG_FILE = "v2ray.config.json"

const V2RAYPOOL_CONFIG_FILE = "v2raypool.config.json"

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
	Protocol    string          `json:"protocol"`
	SendThrough *string         `json:"sendThrough"` // 用于发送数据的 IP 地址，当主机有多个 IP 地址时有效，默认值为 "0.0.0.0"。
	Tag         string          `json:"tag"`
	Settings    json.RawMessage `json:"settings"` // 视协议不同而不同。详见每个协议中的 OutboundConfigurationObject
	// "streamSettings":{"network":"%s","tlsSettings":{"disableSystemRoot":false},"wsSettings":{"path":""},"xtlsSettings":{"disableSystemRoot":false}}
	StreamSetting json.RawMessage `json:"streamSettings"`
	ProxySettings json.RawMessage `json:"proxySettings"`
	Mux           json.RawMessage `json:"mux"`
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

// 1. 规则是放在 routing.rules 这个数组当中，数组的内容是有顺序的，也就是说在这里规则是有顺序的，匹配规则时是从上往下匹配；
// 2. 当路由匹配到一个规则时就会跳出匹配而不会对之后的规则进行匹配；
// https://toutyrater.github.io/basic/routing/notice.html
// 当多个属性同时指定时，这些属性需要同时满足，才可以使当前规则生效。即 domains 和 ip 规则需要分开使用。
func getRoutingRules(cf conf.Conf, inPort int) string {
	var rules []V2rayRouteRule
	var rulesb []byte

	if inPort == 0 {
		// 启用 v2ray 内置的基于 gRPC 协议的  API
		rules = append(rules, V2rayRouteRule{Type: "field", InboundTag: []string{TAG_INBOUND_API}, OutboundTag: TAG_OUTBOUND_API})
		// IP代理池模式时，启用每个出站和入站一对一映射规则。

		// for _, nd := range GetProxyPool().GetNodes("") {
		// tag := getProxyNodeTag(nd.Index)
		for i := 0; i < 100; i++ {
			tag := getProxyNodeTag(i)
			// 添加路由规则，相同标签的每一个出站和入站一一对应
			rules = append(rules, V2rayRouteRule{Type: "field", InboundTag: []string{tag}, OutboundTag: tag})
		}

	} else {
		var f *os.File
		var err error
		if miniutils.IsPathExists(ROUTING_RULES_FILE) {
			f, err = os.OpenFile(ROUTING_RULES_FILE, os.O_RDWR, 0777)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			// 优先读取 routing.rules.json 路由规则文件
			rulesb, err = io.ReadAll(f)
			if err != nil {
				panic(err)
			}
			rulestr := strings.TrimSpace(string(rulesb))
			if rulestr != "" {
				return rulestr
			}

		} else {
			f, err = os.OpenFile(ROUTING_RULES_FILE, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
			if err != nil {
				panic(err)
			}
			defer f.Close()
		}

		if len(cf.DirectDomainList) > 0 || len(cf.DirectIpList) > 0 {
			// 启用直连白名单.IP代理池模式不启用
			rule1 := newRouteRule(TAG_OUTBOUND_DIRECT)
			if len(cf.DirectDomainList) > 0 {
				rule1.Domains = cf.DirectDomainList
			}
			if len(cf.DirectIpList) > 0 {
				rule1.Ip = cf.DirectIpList
			}
			rules = append(rules, rule1)
		}

		if len(cf.ProxyDomainList) > 0 || len(cf.ProxyIpList) > 0 {
			// 启用强制代理访问的白名单(IP代理池模式不启用)
			rule2 := newRouteRule(TAG_OUTBOUND_ACTIVE)
			if len(cf.ProxyDomainList) > 0 {
				rule2.Domains = cf.ProxyDomainList
			}
			if len(cf.ProxyIpList) > 0 {
				rule2.Ip = cf.ProxyIpList
			}
			rules = append(rules, rule2)
		}

		rulesb, err = json.MarshalIndent(rules, "", "\t")
		if err != nil {
			panic(err)
		}
		// 保存路由配置为 routing.rules.json 路由规则文件
		_, err = f.Write(rulesb)
		// Debian下出现: panic: write routing.rules.json: bad file descriptor 添加 os.O_WRONLY 解决
		if err != nil {
			panic(err)
		}
		return string(rulesb)
	}
	rulesb, _ = json.Marshal(rules)
	return string(rulesb)
}

func getProxyNodeTag(index int) string {
	return fmt.Sprintf("TAG_PROXY_%d", index)
}

func setV2rayConfigV4Routing(confv4 *V2rayConfigV4, cf conf.Conf, inPort int) {
	rulestr := getRoutingRules(cf, inPort)
	// 域名解析策略domainStrategy: 不要使用 AsIs , 要用 IPOnDemand 或者 IPIfNonMatch ，否则直连规则不生效
	routing := fmt.Sprintf(`{
		"domainStrategy": "IPOnDemand",
		"domainMatcher": "mph",
		"rules": %s,
		"balancers": []
	}`, rulestr)
	confv4.Routing = json.RawMessage(routing)
}

// setV2rayConfigV4Outbounds 出站配置
// n.Add != "" 时，启用单节点代理模式。
func setV2rayConfigV4Outbounds(confv4 *V2rayConfigV4, n V2rayNode) error {
	sendth := "0.0.0.0"
	// sendth.Address = net.ParseAddress()
	// fmt.Printf("-------setV2rayConfigV4Outbounds---SendThrough---ip(%+v)--domain(%+v)--str(%s)-\n", sendth.Address.IP(), sendth.Address.Domain(), sendth.Address)
	outdirect := V2rayOutbound{Protocol: "freedom", Tag: TAG_OUTBOUND_DIRECT, SendThrough: &sendth}
	outdirect.Settings = json.RawMessage(`{}`)
	outdirect.StreamSetting = json.RawMessage(`{}`)

	confv4.Outbounds = []V2rayOutbound{outdirect}
	if n.Add != "" {
		networkset := n.Net
		if networkset == "" {
			networkset = "tcp"
		}
		outbd1 := V2rayOutbound{
			Protocol:    n.Protocol,
			SendThrough: &sendth,
			Tag:         TAG_OUTBOUND_ACTIVE,
		}

		if n.Protocol == "vless" {
			return fmt.Errorf("outbounds protocol not support vless. TODO")
		}
		if n.Protocol == "vmess" {
			outbd1.Settings = json.RawMessage(fmt.Sprintf(`{"vnext":[{"address":"%s","port":%v,"users":[{"id":"%s","security":"aes-128-gcm"}]}]}`, n.Add, n.Port, n.Id))

			// "streamSettings":{"network":"ws","security":"tls","tlsSettings":{"disableSystemRoot":false},"wsSettings":{"path":""},"xtlsSettings":{"disableSystemRoot":false}}
			// network: "tcp" | "kcp" | "ws" | "http" | "domainsocket" | "quic" | "grpc". 默认值为 tcp
			// security: "none" | "tls" 是否启用传输层加密
			security := n.Tls
			if security == "" {
				security = "none"
			}
			streamSet := fmt.Sprintf(`{"network":"%s","security":"%s","wsSettings":{"path":"%s"}}`, networkset, security, n.Path) // v4: network, v5: transport
			outbd1.StreamSetting = json.RawMessage(streamSet)
			confv4.Outbounds = append(confv4.Outbounds, outbd1)
			return nil
		}
		if n.Protocol == "ss" || n.Protocol == "shadowsocks" {
			outbd1.Protocol = "shadowsocks"
			settingstr := fmt.Sprintf(`{"servers":[{"address":"%s","method":"%s","password":"%s","port":%s}]}`, n.Add, n.Type, n.Id, n.Port)
			outbd1.Settings = json.RawMessage(settingstr)
			proxysetTag := fmt.Sprintf("%s-dialer", outbd1.Tag)
			outbd1.ProxySettings = json.RawMessage(fmt.Sprintf(`{"tag":"%s"}`, proxysetTag))
			outbd2 := V2rayOutbound{
				Protocol: "freedom",
				Tag:      proxysetTag,
				Settings: json.RawMessage(fmt.Sprintf(`{"domainStrategy": "AsIs","redirect": "%s:%s"}`, n.Add, n.Port)),
				Mux:      json.RawMessage(`{"enabled":true,"concurrency":1}`),
			}
			security := n.Tls
			if security == "" {
				security = "none"
			}
			streamSet2 := fmt.Sprintf(`{"network":"%s","security":"%s","wsSettings":{"path":"%s","headers":{"Host":"cloudflare.com"}},"sockopt":{}}`, networkset, security, n.Path) // v4: network, v5: transport
			outbd2.StreamSetting = json.RawMessage(streamSet2)
			confv4.Outbounds = append(confv4.Outbounds, outbd1, outbd2)
			return nil
		}
		if n.Protocol == "trojan" {
			// trojan://4B2rzYGAjsuN@54.169.218.236:16038?sni=appsvs.shop#%E6%96%B0%E5%8A%A0%E5%9D%A1+Amazon%E6%95%B0%E6%8D%AE%E4%B8%AD%E5%BF%83
			// trojan://4B2rzYGAjsuN@54.169.218.236:16038?sni=appsvs.shop#新加坡+Amazon数据中心
			outbd1.Settings = json.RawMessage(fmt.Sprintf(`{"servers":[{"address":"%s","port":%v,"password":"%s"}]}`, n.Add, n.Port, n.Id))
			security := n.Tls
			if security == "" {
				security = "none"
			}
			sni := n.Host
			if sni == "" {
				sni = n.Add
			}
			streamSet := fmt.Sprintf(`{"network":"%s","security":"%s","tlsSettings":{"allowInsecure": false,"serverName": "%s"},"sockopt":{}}`, networkset, security, sni)
			outbd1.StreamSetting = json.RawMessage(streamSet)
			confv4.Outbounds = append(confv4.Outbounds, outbd1)
			return nil
		}
		return fmt.Errorf("outbounds protocol not support %s", n.Protocol)
	}
	return nil
}

func NewV2rayInboundV4(proto string, inPort int) V2rayInbound {
	inAddr := "0.0.0.0" // "127.0.0.1" 仅允许本地访问
	if proto == "socks5" {
		proto = "socks"
	}
	intag := "http_IN"
	inset1 := `{"allowTransparent":false,"timeout":30}`
	if proto == "socks" {
		// inset1 = `{"auth":"noauth","ip":"127.0.0.1","udp":true}`
		intag = "socks_IN"
		inset1 = `{"auth":"noauth","udp":true}`
	}
	return V2rayInbound{
		Protocol: proto,
		Port:     inPort,
		Listen:   inAddr,
		Settings: json.RawMessage(inset1),
		Tag:      intag,
	}
	// ip: address:	SOCKS5 通过 UDP ASSOCIATE 命令建立 UDP 会话。服务端在对客户端发来的该命令的回复中，指定客户端发包的目标地址。
	// v4.34.0+: 默认值为空，此时对于通过本地回环 IPv4/IPv6 连接的客户端，回复对应的回环 IPv4/IPv6 地址；对于非本机的客户端，回复当前入站的监听地址。
	// v4.33.0 及更早版本: 默认值 127.0.0.1。
	// inSet := fmt.Sprintf(`{"auth":"noauth","udp":true,"ip":"%s"}`, inAddr)
	// inbd2 := V2rayInbound{
	// 	Protocol: "socks",
	// 	Port:     inPort, // 端口号相同会冲突
	// 	Listen:   inAddr,
	// 	Settings: json.RawMessage(inSet),
	// 	Tag: "socks_IN",
	// }
}

// setV2rayConfigV4Inbounds 入站配置
func setV2rayConfigV4Inbounds(confv4 *V2rayConfigV4, inPort int, cf conf.Conf) {
	logger := cf.GetLogger()

	if inPort == 0 {
		// inPort == 0 时，启用gRPC，允许多个代理节点组成代理池。
		inbdapi := V2rayInbound{
			Listen:   "0.0.0.0", // "127.0.0.1" 仅允许本地访问
			Port:     cf.V2rayApiPort,
			Protocol: "dokodemo-door",
			Settings: json.RawMessage(`{"address": "127.0.0.1"}`),
			Tag:      TAG_INBOUND_API,
		}
		confv4.Inbounds = []V2rayInbound{inbdapi}
		logger.Debugf("-----setV2rayConfigV4Inbounds--inPort(%d)--inbdapi--V2rayApiPort(%d)", inPort, inbdapi.Port)
	} else {
		// https://www.v2fly.org/config/protocols/http.html#inboundconfigurationobject
		protcl := cf.GetHttpProxyProtocol()
		inbd1 := NewV2rayInboundV4(protcl, inPort)
		logger.Debugf("-----setV2rayConfigV4Inbounds--inPort(%d)--inbd1--", inPort)
		confv4.Inbounds = []V2rayInbound{inbd1}
	}

}

// getV2rayConfigV4 v2ray官方配置v4版本
// inPort == 0时，启用gRPC，允许多个代理节点组成代理池。
func getV2rayConfigV4(n V2rayNode, inPort int) *JsonConfig {
	var err error
	cf := conf.GetConf()
	vconf := V2rayConfigV4{}
	if inPort == 0 {
		vconf.Api = json.RawMessage(fmt.Sprintf(`{"tag": "%s", "services": ["HandlerService"]}`, TAG_OUTBOUND_API))
	}
	vconf.Log = []byte(`{"loglevel":"debug"}`) // v4
	// vconf.Log = []byte(`{"access":{"type":"Console","level":"Debug"}}`) // v5

	setV2rayConfigV4Routing(&vconf, cf, inPort)
	setV2rayConfigV4Inbounds(&vconf, inPort, cf)
	err = setV2rayConfigV4Outbounds(&vconf, n)
	if err != nil {
		panic(err)
	}
	var vconfb []byte
	vconfb, err = json.MarshalIndent(vconf, "", "\t")
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n---getV2rayConfigV4--Outbounds-len(%d)--v2ray.config=(%s)--\n", len(vconf.Outbounds), string(vconfb))
	return NewJsonConfig(vconfb)
}

func NewJsonConfig(b []byte) *JsonConfig {
	return &JsonConfig{
		content: b,
	}
}

func NewJsonConfigFromFile(fpath string) *JsonConfig {
	f, err := os.OpenFile(fpath, os.O_RDONLY, 0777)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	result := &JsonConfig{}
	result.content, err = io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	result.filepath = fpath
	return result
}

type JsonConfig struct {
	content  []byte
	filepath string
}

func (j JsonConfig) GetFilepath() string {
	return j.filepath
}

func (j *JsonConfig) SetContent(v any) error {
	var err error
	switch vv := v.(type) {
	case []byte:
		j.content = vv
	case string:
		j.content = []byte(vv)
	default:
		j.content, err = json.MarshalIndent(v, "", "\t")
	}
	return err
}

// SaveToFile 保存json内容到filepath文件。
// 若 filepath 为空，则保存到默认 filepath
func (j *JsonConfig) SaveToFile(filepath string) error {
	if filepath == "" {
		j.filepath = filepath
	}
	f, err := os.OpenFile(filepath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(j.content)
	if err == nil {
		j.filepath = filepath
	}
	return err
}

func (j JsonConfig) Reader() io.Reader {
	return bytes.NewReader(j.content)
}

func (j JsonConfig) String() string {
	return string(j.content)
}

func (j JsonConfig) Decode(v any) error {
	return json.Unmarshal(j.content, v)
}
