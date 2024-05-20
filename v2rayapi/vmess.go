package v2rayapi

import (
	v5 "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	pros "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/proxy/vmess"
	vmessOutbound "github.com/v2fly/v2ray-core/v5/proxy/vmess/outbound"
)

func GetVmessOutbound(sender *proxyman.SenderConfig, addr, id, outag string, port, alterid uint32) *pros.AddOutboundRequest {
	proxySet := &vmessOutbound.Config{
		Receiver: []*protocol.ServerEndpoint{
			{
				Address: net.NewIPOrDomain(net.DomainAddress(addr)),
				Port:    port,
				User: []*protocol.User{
					{
						Account: serial.ToTypedMessage(&vmess.Account{
							Id:      id,
							AlterId: alterid,
							SecuritySettings: &protocol.SecurityConfig{
								Type: protocol.SecurityType_AES128_GCM,
							},
						}),
					},
				},
			},
		},
	}
	return &pros.AddOutboundRequest{Outbound: &v5.OutboundHandlerConfig{
		Tag:            outag,
		SenderSettings: serial.ToTypedMessage(sender),
		ProxySettings:  serial.ToTypedMessage(proxySet),
	}}
}
