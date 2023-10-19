package v2raypool

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// ParseNodes 解析节点 Add, Ps ...
// {"add":"jp6.v2u9.top","host":"","id":"0999AE93-1330-4A75-DBC1-0DD545F7DD60","net":"ws","path":"","port":"41444","ps":"u9un-v2-JP-Tokyo6(1)","tls":"","v":2,"aid":0,"type":"none"}
// {"add":"hk6.v2u9.top","host":"","id":"93EA57CE-EA21-7240-EE7F-317F4A6A8B65","net":"ws","path":"","port":"444","ps":"u9un-v2-HK-HongKong6","tls":"tls","type":"none","v":2,"aid":0}
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
			fmt.Printf("\n-----ParseV2rayNodes--node(%s)--\n", string(b))
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
