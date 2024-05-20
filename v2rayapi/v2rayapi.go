package v2rayapi

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	pros "github.com/v2fly/v2ray-core/v5/app/proxyman/command"

	"github.com/v2fly/v2ray-core/v5/common/serial"

	// "github.com/v2fly/v2ray-core/v5/features/inbound"
	// "github.com/v2fly/v2ray-core/v5/proxy/blackhole"
	// "github.com/v2fly/v2ray-core/v5/proxy/freedom"
	// "github.com/v2fly/v2ray-core/v5/proxy/dokodemo"
	// "github.com/v2fly/v2ray-core/v5/proxy/socks"
	// "github.com/v2fly/v2ray-core/v5/common/uuid"
	// "github.com/v2fly/v2ray-core/v5/proxy/http"
	// "github.com/v2fly/v2ray-core/v5/proxy/socks"
	// "github.com/v2fly/v2ray-core/v5/proxy/shadowsocks2022"
	// "google.golang.org/grpc"
	// "google.golang.org/grpc/credentials/insecure"

	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tcp"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tls"
	"github.com/v2fly/v2ray-core/v5/transport/internet/websocket"

	"google.golang.org/protobuf/types/known/anypb"
)

type V2rayNode struct {
	Protocol, Add, Host, Id, Net, Path, Ps, Tls, Type string
	V, Aid, Port                                      json.Number
}

func GetTransportStreamConfig(network, path, hdhost string) (conf *internet.StreamConfig, err error) {
	transproto := network
	var transptl internet.TransportProtocol
	var protoconf proto.Message
	switch transproto {
	case "ws", "websocket":
		transptl = internet.TransportProtocol_WebSocket
		wsconf := websocket.Config{Path: path}
		if hdhost != "" {
			// wsconf.UseBrowserForwarding = true
			wsconf.Header = []*websocket.Header{{Key: "Host", Value: hdhost}}
		}
		protoconf = &wsconf
	case "tcp":
		transptl = internet.TransportProtocol_TCP
		protoconf = &tcp.Config{}
	default:
		err = fmt.Errorf("outbound network or transport not support %s. only support ws or websocket", transproto)
	}
	if err != nil {
		return
	}
	conf = &internet.StreamConfig{
		Protocol: transptl,
		TransportSettings: []*internet.TransportConfig{
			{
				Protocol: transptl,
				Settings: serial.ToTypedMessage(protoconf),
			},
		},
	}
	return
}

const PROTO_VMESS = "vmess"
const PROTO_SHADOWSOCKS = "shadowsocks"
const PROTO_TROJAN = "trojan"

var SupportProtocolList = []string{
	PROTO_VMESS,
	PROTO_TROJAN,
	PROTO_SHADOWSOCKS,
	"ss",
}

func GetOutboundRequest(port, aid json.Number, outproto, addr, sni, password, network, path, security, cipher, outag string) (reqs []*pros.AddOutboundRequest, err error) {
	outproto = strings.ToLower(outproto)
	if outproto == "ss" {
		outproto = PROTO_SHADOWSOCKS
	}
	oksupport := false
	for _, otpro := range SupportProtocolList {
		if otpro == outproto {
			oksupport = true
			break
		}
	}
	if !oksupport {
		err = fmt.Errorf("outbound protocol not support %s. only support %v", outproto, SupportProtocolList)
		return
	}

	var streamConf *internet.StreamConfig

	streamConf, err = GetTransportStreamConfig(network, path, "")
	if err != nil {
		return
	}

	sender := proxyman.SenderConfig{
		StreamSettings: streamConf,
	}
	if security == "tls" {
		tlsconf := &tls.Config{
			AllowInsecure: true,
		}
		if outproto == PROTO_TROJAN {
			tlsconf.ServerName = sni
			// TODO allowInsecure false
		}
		sender.StreamSettings.SecurityType = serial.GetMessageType(&tls.Config{})
		sender.StreamSettings.SecuritySettings = []*anypb.Any{
			serial.ToTypedMessage(tlsconf),
		}
	}

	var proxyport int64
	proxyport, err = port.Int64()
	if err != nil {
		err = fmt.Errorf("err AddOutboundByV2rayNode 端口数据解析错误 port val(%v)--err(%v)", port, err)
		return
	}

	if outproto == PROTO_VMESS {
		aid, _ := strconv.ParseUint(aid.String(), 10, 32)
		alterid := uint32(aid)
		reqs = []*pros.AddOutboundRequest{GetVmessOutbound(&sender, addr, password, outag, uint32(proxyport), alterid)}
	}

	if outproto == PROTO_SHADOWSOCKS {
		sender.ProxySettings = &internet.ProxyConfig{
			Tag: outag + "-dialer",
		}
		reqs = []*pros.AddOutboundRequest{GetShadowsocksOutbound(&sender, addr, password, cipher, outag, uint32(proxyport))}

		streamConf, _ = GetTransportStreamConfig(network, password, "cloudflare.com")
		sender := proxyman.SenderConfig{
			StreamSettings: streamConf,
			MultiplexSettings: &proxyman.MultiplexingConfig{
				Enabled:     true,
				Concurrency: 1,
			},
		}
		reqs = append(reqs, GetFreedomOutbound(&sender, addr, outag+"-dialer", uint32(proxyport)))
	}

	if outproto == PROTO_TROJAN {
		reqs = append(reqs, GetTrojanOutbound(&sender, addr, password, outag, uint32(proxyport)))
	}

	return
}
