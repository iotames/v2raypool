package v2raypool

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	pros "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"

	// "github.com/v2fly/v2ray-core/v5/features/inbound"
	v5 "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common/net"

	// "github.com/v2fly/v2ray-core/v5/proxy/blackhole"
	// "github.com/v2fly/v2ray-core/v5/proxy/freedom"
	// "github.com/v2fly/v2ray-core/v5/proxy/dokodemo"
	// "github.com/v2fly/v2ray-core/v5/proxy/socks"
	// "github.com/v2fly/v2ray-core/v5/common/uuid"
	"github.com/v2fly/v2ray-core/v5/proxy/http"
	"github.com/v2fly/v2ray-core/v5/proxy/socks"
	"github.com/v2fly/v2ray-core/v5/proxy/vmess"
	"github.com/v2fly/v2ray-core/v5/proxy/vmess/outbound"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tcp"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tls"
	"github.com/v2fly/v2ray-core/v5/transport/internet/websocket"

	"google.golang.org/protobuf/types/known/anypb"
)

type V2rayApiClient struct {
	addr   string
	conn   *grpc.ClientConn
	c      pros.HandlerServiceClient
	ctx    context.Context
	cancel context.CancelFunc
}

func NewV2rayApiClientV5(addr string) *V2rayApiClient {
	return &V2rayApiClient{addr: addr}
}

func (a *V2rayApiClient) Close() {
	a.cancel()
	a.conn.Close()
}

func (a *V2rayApiClient) Dial() error {
	var err error
	a.conn, err = grpc.Dial(a.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("------V2rayApiClient--Dial(%v)--\n", err)
		return err
	}
	a.c = pros.NewHandlerServiceClient(a.conn)
	// defer conn.Close()
	a.ctx, a.cancel = context.WithTimeout(context.Background(), time.Second*10)
	// defer cancel()
	return nil
}

// AddInbound 添加入站规则
// protocol http|socks
func (a V2rayApiClient) AddInbound(inport net.Port, intag, protocol string) error {
	var proxySet proto.Message
	protocol = strings.ToLower(protocol)
	if protocol == "http" {
		proxySet = &http.ServerConfig{
			AllowTransparent: false,
			Timeout:          30,
			// UserLevel: 0,
		}
	}
	if protocol == "socks" {
		proxySet = &socks.ServerConfig{
			AuthType:   socks.AuthType_NO_AUTH,
			UdpEnabled: true,
			// Address:    net.NewIPOrDomain(net.AnyIP),
			// UserLevel:  0,
		}
	}

	resp, err := a.c.AddInbound(a.ctx, &pros.AddInboundRequest{Inbound: &v5.InboundHandlerConfig{
		Tag: intag,
		ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
			PortRange: net.SinglePortRange(inport),
			Listen:    net.NewIPOrDomain(net.AnyIP),
		}),
		ProxySettings: serial.ToTypedMessage(proxySet),
	}})
	// rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing: dial tcp 127.0.0.1:15491: connectex: No connection could be made because the target machine actively refused it."
	fmt.Printf("---AddInbound(%s)--port(%d)--result(%s)--err(%v)--\n", intag, inport, resp, err)
	return err
}

func (a V2rayApiClient) RemoveInbound(intag string) error {
	result, err := a.c.RemoveInbound(a.ctx, &pros.RemoveInboundRequest{Tag: intag})
	fmt.Printf("---RemoveInbound(%s)----result(%v)---err(%v)-\n", intag, result, err)
	return err
}

func getTransportStreamConfig(transproto, path string) (conf *internet.StreamConfig, err error) {
	var transptl internet.TransportProtocol
	var protoconf proto.Message
	switch transproto {
	case "ws", "websocket":
		transptl = internet.TransportProtocol_WebSocket
		protoconf = &websocket.Config{Path: path}
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

func (a V2rayApiClient) AddOutboundByV2rayNode(nd V2rayNode, outag string) error {
	if nd.Protocol != "vmess" {
		return fmt.Errorf("outbound protocol not support %s. only support vmess", nd.Protocol)
	}
	streamConf, err := getTransportStreamConfig(nd.Net, nd.Path)
	if err != nil {
		return err
	}

	outsendset := proxyman.SenderConfig{
		StreamSettings: streamConf,
	}
	if nd.Tls == "tls" {
		outsendset.StreamSettings.SecurityType = serial.GetMessageType(&tls.Config{})
		outsendset.StreamSettings.SecuritySettings = []*anypb.Any{
			serial.ToTypedMessage(&tls.Config{
				AllowInsecure: true,
			}),
		}
	}

	var proxyport int
	switch pval := nd.Port.(type) {
	case string:
		proxyport, _ = strconv.Atoi(pval)
	case float64:
		proxyport = int(pval)
	default:
		return fmt.Errorf("err AddOutboundByV2rayNode 端口数据类型未知 port val(%v) type(%T)", nd.Port, nd.Port)
	}

	resp, err := a.c.AddOutbound(a.ctx, &pros.AddOutboundRequest{Outbound: &v5.OutboundHandlerConfig{
		Tag:            outag,
		SenderSettings: serial.ToTypedMessage(&outsendset),
		ProxySettings: serial.ToTypedMessage(&outbound.Config{
			Receiver: []*protocol.ServerEndpoint{
				{
					Address: net.NewIPOrDomain(net.DomainAddress(nd.Add)),
					Port:    uint32(proxyport),
					User: []*protocol.User{
						{
							Account: serial.ToTypedMessage(&vmess.Account{
								Id:      nd.Id,
								AlterId: uint32(nd.Aid),
								SecuritySettings: &protocol.SecurityConfig{
									Type: protocol.SecurityType_AES128_GCM,
								},
							}),
						},
					},
				},
			},
		}),
		// ProxySettings: serial.ToTypedMessage(&freedom.Config{
		// 	DomainStrategy: freedom.Config_AS_IS,
		// 	UserLevel:      0,
		// }),
	}})
	fmt.Printf("---AddOutbound(%s)--(%s:%v)-portType(%T)--result(%s)--err(%v)--\n", outag, nd.Add, nd.Port, nd.Port, resp, err)
	return err
}

func (a V2rayApiClient) RemoveOutbound(outag string) error {
	result, err := a.c.RemoveOutbound(a.ctx, &pros.RemoveOutboundRequest{Tag: outag})
	fmt.Printf("---RemoveOutbound(%s)----result(%v)---err(%v)-\n", outag, result, err)
	return err
}
