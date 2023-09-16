package v2raypool

import (
	"testing"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

func TestCommanderRemoveHandler(t *testing.T) {
	// 浏览器使用代理时，因使用TCP长连接机制，移除一个入站协议（Inbound）后，刷新网页会发现更改没有效果。
	// 关闭浏览器再开即可。
	{
		_, err := net.DialTCP("tcp", nil, &net.TCPAddr{
			IP:   []byte{127, 0, 0, 1},
			Port: 1089,
		})
		if err == nil {
			t.Error("unexpected nil error")
		}
		t.Logf("----DialTcpErr(%v)----\n", err)
	}
}
