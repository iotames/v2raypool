package v2raypool

import (
	"context"
	"fmt"
	"time"

	"github.com/iotames/v2raypool/conf"
	g "github.com/iotames/v2raypool/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewProxyPoolGrpcClient
// c, conn := NewProxyPoolGrpcClient()
// defer conn.Close()
// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// defer cancel()
// nds, err := c.GetProxyNodes(ctx, &ProxyNode{IsRunning: true})
func NewProxyPoolGrpcClient() (c g.ProxyPoolServiceClient, conn *grpc.ClientConn) {
	var err error
	conn, err = grpc.Dial(fmt.Sprintf("127.0.0.1:%d", conf.GetConf().GrpcPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	c = g.NewProxyPoolServiceClient(conn)
	return
}

func RequestProxyPoolGrpcOnce(h func(c g.ProxyPoolServiceClient, ctx context.Context)) {
	c, conn := NewProxyPoolGrpcClient()
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	h(c, ctx)
}

func ChangeProxyNode() error {
	var err error
	RequestProxyPoolGrpcOnce(func(c g.ProxyPoolServiceClient, ctx context.Context) {
		var nds *g.ProxyNodes
		nds, err = c.GetProxyNodes(ctx, &g.ProxyNode{IsRunning: true})
		if err != nil {
			return
		}
		selectport := conf.GetConf().GetHttpProxyPort()
		nextIndex := -1
		for i, n := range nds.Items {
			if n.LocalPort == uint32(selectport) {
				if i < len(nds.Items)-1 {
					nextIndex = int(nds.Items[i+1].Index)
				}
				break
			}
		}
		if nextIndex > -1 {
			var res *g.OptResult
			res, err = c.ActiveProxyNode(ctx, &g.ProxyNode{Index: uint32(nextIndex)})
			if err != nil {
				return
			}
			fmt.Println(res.Msg)
			return
		} else {
			err = fmt.Errorf("can not find available node")
		}
	})
	return err
}
