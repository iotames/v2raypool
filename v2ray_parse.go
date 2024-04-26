package v2raypool

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/iotames/v2raypool/decode"
)

// ParseNodes 解析节点 Add, Ps ...
// {"add":"jp6.xxx.top","host":"","id":"0999AE93-1330-4A75-DBC1-0DD545F7DD60","net":"ws","path":"","port":"41444","ps":"xxx-v2-JP-Tokyo6(1)","tls":"","v":2,"aid":0,"type":"none"}
// {"add":"hk6.xxx.top","host":"","id":"93EA57CE-EA21-7240-EE7F-317F4A6A8B65","net":"ws","path":"","port":"444","ps":"xxx-v2-HK-HongKong6","tls":"tls","v":2,"aid":0, "type":"none"}
// vless://26DL68CE-DL93-8342-LQ8F-317F4A6E7J76@45.43.31.159:443?encryption=none&security=reality&sni=azure.microsoft.com&fp=safari&pbk=c7qU9-_0WflwIKUiZFxSss_xw-2AP3jB1ENxKLI0OTw&type=tcp&headerType=none#u9un-US-Xr1
func ParseV2rayNodes(data string) []V2rayNode {
	fmt.Println("-----Begin--ParseV2rayNodes-------")
	sss := strings.Split(data, "\n")
	var nds []V2rayNode
	var err error
	for i, d := range sss {
		var n V2rayNode
		n, err = parseNodeInfo(d)
		if err != nil {
			fmt.Printf("\n---ParseV2rayNodesErr--(%d)--err(%v)---RAW(%s)\n", i, err, d)
			continue
		}
		nds = append(nds, n)
	}
	return nds
}

func parseNodeInfo(d string) (nd V2rayNode, err error) {
	ninfo := strings.Split(d, "://")
	if len(ninfo) > 1 {
		var b []byte
		nd.Protocol = ninfo[0]
		if nd.Protocol == "vmess" {
			b, err = base64.StdEncoding.DecodeString(ninfo[1])
			if err != nil {
				err = fmt.Errorf("parseNodeInfo err(%v) for vmess base64 DecodeString", err)
				return
			}
			// fmt.Printf("\n-----ParseV2rayNodes--vmess--node(%s)--\n", string(b))
			err = json.Unmarshal(b, &nd)
			if err != nil {
				err = fmt.Errorf("parseNodeInfo err(%v) for vmess json Unmarshal(%s)", err, string(b))
				return
			}
			return
		}
		if nd.Protocol == "vless" {
			err = fmt.Errorf("protocol not support vless://")
			return
		}
		if nd.Protocol == "ssr" {
			err = fmt.Errorf("v2ray protocol dot not support ssr://")
			return
		}
		if nd.Protocol == "ss" {
			ssdata := ninfo[1]
			var ss decode.Shadowsocks
			ss, err = decode.ParseShadowsocks(ssdata)
			if err != nil {
				err = fmt.Errorf("ParseShadowsocks err(%v)", err)
				return
			}
			nd.Add = ss.Address
			nd.Port = json.Number(fmt.Sprintf("%d", ss.Port)) // TODO err
			nd.Type = ss.Cipher
			nd.Id = ss.Password
			nd.Net = ss.TransportStream.Protocol
			nd.Tls = ss.TransportStream.Security
			nd.Path = ss.TransportStream.Path
			nd.Ps = ss.Title
			if nd.Id == "" || nd.Type == "" || ss.Port == 0 || nd.Add == "" || nd.Net == "" {
				err = fmt.Errorf("---parse--shadowsocks--err--ss://--raw(%s)---nd(%+v)", ssdata, nd)
				return
			}
			return
		}
		if nd.Protocol == "trojan" {
			err = fmt.Errorf("protocol not support trojan://, TODO")
			return
		}

		err = fmt.Errorf("protocol not support %s", nd.Protocol)
		return
	}
	err = fmt.Errorf("can not found protocol")
	return
}
