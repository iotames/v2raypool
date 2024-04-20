package v2raypool

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

// func TestInitSubscribeData(t *testing.T) {}
type TestSubscribeData1 struct {
	Protocol, Data string
}

func TestParseVmessNodes(t *testing.T) {
	rawnds := []string{
		`{"add":"jp6.xxx.top","host":"","id":"0999AE93-1330-4A75-DBC1-0DD5XXXXXXXX","net":"ws","path":"","port":"4147","ps":"xxx-v2-JP-Tokyo6","tls":"","v":2,"aid":0,"type":"none"}`,
		`{"add":"hk6.xxx.top","host":"","id":"93EA57CE-EA21-7240-EE7F-317FXXXXXXXX","net":"ws","path":"","port":4446,"ps":"xxx-v2-HK-HongKong6","tls":"tls","v":2,"aid":0,"type":"none"}`,
		`{"add":"jp6.xxx.top","host":"","id":"0999AE93-1330-4A75-DBC1-0DD5XXXXXXXX","net":"ws","path":"","port":"4145","ps":"xxx-v2-JP-Tokyo6","tls":"","v":"2","aid":"0","type":"none"}`,
	}

	for i, row := range rawnds {
		rawnds[i] = fmt.Sprintf(`vmess://%s`, base64.StdEncoding.EncodeToString([]byte(row)))
	}
	rawdata := base64.StdEncoding.EncodeToString([]byte(strings.Join(rawnds, "\n")))
	dt, err := parseSubscribeByRaw(rawdata)
	if err != nil {
		t.Error(err)
	}
	t.Logf("-----dt(%s)-----\n", dt)
	nds := ParseV2rayNodes(dt)
	if len(nds) != len(rawnds) {
		t.Error("ParseV2rayNodes err")
	}
	t.Logf("---(%+v)---", nds)
}