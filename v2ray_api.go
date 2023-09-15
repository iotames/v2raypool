package v2raypool

import (
	"context"
	"fmt"
	"time"

	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	pros "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v5/common/serial"

	// "github.com/v2fly/v2ray-core/v5/features/inbound"
	v5 "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common/net"

	// "github.com/v2fly/v2ray-core/v5/proxy/socks"
	"github.com/v2fly/v2ray-core/v5/proxy/http"
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
		return err
	}
	a.c = pros.NewHandlerServiceClient(a.conn)
	// defer conn.Close()
	a.ctx, a.cancel = context.WithTimeout(context.Background(), time.Second*10)
	// defer cancel()
	return nil
}

func (a V2rayApiClient) AddInbound(inport net.Port, intag string) error {
	resp, err := a.c.AddInbound(a.ctx, &pros.AddInboundRequest{Inbound: &v5.InboundHandlerConfig{
		Tag: intag,
		ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
			PortRange: net.SinglePortRange(inport),
			Listen:    net.NewIPOrDomain(net.LocalHostIP),
		}),
		ProxySettings: serial.ToTypedMessage(&http.ServerConfig{
			// AllowTransparent: true,
			UserLevel: 0,
		}),
	}})
	fmt.Printf("---AddInbound----result(%s)--err(%v)--\n", resp, err)
	return err
}

func (a V2rayApiClient) RemoveInbound(intag string) error {
	result, err := a.c.RemoveInbound(a.ctx, &pros.RemoveInboundRequest{Tag: intag})
	fmt.Printf("---RemoveInbound----result(%v)---err(%v)-\n", result, err)
	return err
}

// c.AlterInbound(ctx, &pros.AlterInboundRequest{
// 	Tag:       "http_IN",
// })
