package v2rayapi

import (
	v5 "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	pros "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/proxy/freedom"
)

func GetFreedomOutbound(sender *proxyman.SenderConfig, addr, outag string, port uint32) *pros.AddOutboundRequest {
	return &pros.AddOutboundRequest{Outbound: &v5.OutboundHandlerConfig{
		Tag:            outag,
		SenderSettings: serial.ToTypedMessage(sender),
		ProxySettings: serial.ToTypedMessage(&freedom.Config{
			DomainStrategy: freedom.Config_AS_IS,
			DestinationOverride: &freedom.DestinationOverride{
				Server: &protocol.ServerEndpoint{
					Address: net.NewIPOrDomain(net.DomainAddress(addr)),
					Port:    port,
				},
			},
		}),
	}}
}
