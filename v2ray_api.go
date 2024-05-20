package v2raypool

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/iotames/v2raypool/v2rayapi"
	v5 "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	pros "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/proxy/http"
	"github.com/v2fly/v2ray-core/v5/proxy/socks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func (a V2rayApiClient) AddOutboundByV2rayNode(nd V2rayNode, outag string) error {
	var reqs []*pros.AddOutboundRequest
	var resp *pros.AddOutboundResponse
	var err error

	reqs, err = v2rayapi.GetOutboundRequest(nd.Port, nd.Aid, nd.Protocol, nd.Add, nd.Host, nd.Id, nd.Net, nd.Path, nd.Tls, nd.Type, outag) // nd, outag
	if err != nil {
		return err
	}
	for _, req := range reqs {
		resp, err = a.c.AddOutbound(a.ctx, req)
		fmt.Printf("---AddOutbound--(%s://)(%s)--(%s:%v)--result(%s)--err(%v)--\n", nd.Protocol, outag, nd.Add, nd.Port, resp, err)
	}
	return err
}

func (a V2rayApiClient) RemoveOutbound(outag string) error {
	result, err := a.c.RemoveOutbound(a.ctx, &pros.RemoveOutboundRequest{Tag: outag})
	fmt.Printf("---RemoveOutbound(%s)----result(%v)---err(%v)-\n", outag, result, err)
	return err
}
