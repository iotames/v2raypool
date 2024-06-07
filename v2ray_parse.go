package v2raypool

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/iotames/v2raypool/decode"
)

// ParseNodes 解析节点 Add, Ps ...
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
			// fmt.Printf("-----ss--raw(%s)---\n", ssdata)
			ss, err = decode.ParseShadowsocks(d)
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
			nd.Ps = strings.TrimSpace(ss.Title)
			if nd.Id == "" || nd.Type == "" || ss.Port == 0 || nd.Add == "" {
				err = fmt.Errorf("---parse--shadowsocks--err--ss://--raw(%s)---nd(%+v)", ssdata, nd)
				return
			}
			return
		}
		if nd.Protocol == "trojan" {
			var tro decode.Trojan
			tro, err = decode.ParseTrojan(d)
			if err != nil {
				return
			}
			// alpn=http/1.1
			// sni=trojan.burgerip.co.uk
			nd.Host = tro.Sni
			nd.Add = tro.Address
			nd.Port = json.Number(fmt.Sprintf("%d", tro.Port))
			nd.Id = tro.Password
			nd.Net = tro.TransportStream.Protocol // type=tcp
			nd.Tls = tro.TransportStream.Security
			nd.Path = tro.TransportStream.Path
			nd.Ps = strings.TrimSpace(tro.Title)
			if nd.Id == "" || tro.Port == 0 || nd.Add == "" || nd.Net == "" {
				err = fmt.Errorf("---parse--err--trojan://--raw(%s)---nd(%+v)", d, nd)
				return
			}
			return
		}

		err = fmt.Errorf("protocol not support %s", nd.Protocol)
		return
	}
	err = fmt.Errorf("can not found protocol")
	return
}
