package v2rayapi

import (
	v5 "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	pros "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/proxy/trojan"
)

func GetTrojanOutbound(sender *proxyman.SenderConfig, addr, password, outag string, port uint32) *pros.AddOutboundRequest {
	proxySet := &trojan.ClientConfig{
		Server: []*protocol.ServerEndpoint{
			{
				Address: net.NewIPOrDomain(net.DomainAddress(addr)),
				Port:    port,
				User: []*protocol.User{{
					Account: serial.ToTypedMessage(&trojan.Account{
						Password: password,
					}),
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
