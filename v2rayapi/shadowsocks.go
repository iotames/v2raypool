package v2rayapi

import (
	"strings"

	v5 "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	pros "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/proxy/shadowsocks"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

func GetShadowsocksOutbound(addr, password, cipher, outag, proxytag string, port uint32) *pros.AddOutboundRequest {
	sender := &proxyman.SenderConfig{
		ProxySettings: &internet.ProxyConfig{
			Tag: proxytag,
		},
	}
	ssAccount := serial.ToTypedMessage(&shadowsocks.Account{
		Password: password,
		CipherType: func() shadowsocks.CipherType {
			method := strings.ReplaceAll(cipher, "-", "_") // "aes-256-gcm",
			method = strings.ToUpper(method)
			val, ok := shadowsocks.CipherType_value[method]
			if ok {
				return shadowsocks.CipherType(val)
			}
			return shadowsocks.CipherType_AES_256_GCM
		}(),
	})
	proxySet := &shadowsocks.ClientConfig{
		Server: []*protocol.ServerEndpoint{
			{
				Address: net.NewIPOrDomain(net.DomainAddress(addr)),
				Port:    port,
				User: []*protocol.User{{
					Account: ssAccount,
				}},
			},
		},
	}
	return &pros.AddOutboundRequest{Outbound: &v5.OutboundHandlerConfig{
		Tag:            outag,
		SenderSettings: serial.ToTypedMessage(sender),
		ProxySettings:  serial.ToTypedMessage(proxySet),
	}}
}
