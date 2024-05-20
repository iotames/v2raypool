package v2rayapi

import (
	v5 "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	pros "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/proxy/freedom"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tls"
	"google.golang.org/protobuf/types/known/anypb"
)

func GetFreedomOutboundOfShadowsocks(network, path, addr, outag, security string, port uint32) *pros.AddOutboundRequest {
	if path == "" {
		path = "/"
	}
	streamConf, _ := GetTransportStreamConfig(network, path, "cloudflare.com")
	sender := &proxyman.SenderConfig{
		StreamSettings: streamConf,
		MultiplexSettings: &proxyman.MultiplexingConfig{
			Enabled:     true,
			Concurrency: 1,
		},
	}

	if security == "tls" {
		tlsconf := &tls.Config{
			AllowInsecure: true,
		}
		sender.StreamSettings.SecurityType = serial.GetMessageType(&tls.Config{})
		sender.StreamSettings.SecuritySettings = []*anypb.Any{
			serial.ToTypedMessage(tlsconf),
		}
	}
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
