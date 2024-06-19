package v2rayapi

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tcp"
	"github.com/v2fly/v2ray-core/v5/transport/internet/websocket"
)

func GetTransportStreamConfig(network, path, hdhost string) (conf *internet.StreamConfig, err error) {
	transproto := strings.TrimSpace(network)
	if transproto == "" {
		transproto = "tcp"
	}
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
		err = fmt.Errorf("outbound network or transport not support (%s). only support tcp, ws or websocket", transproto)
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
