package v2raypool

import (
	"context"
	"fmt"
	"time"

	"github.com/iotames/miniutils"
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

// ChangeProxyNode 更换一个可用节点
func ChangeProxyNode() error {
	var err error
	RequestProxyPoolGrpcOnce(func(c g.ProxyPoolServiceClient, ctx context.Context) {
		var nds *g.ProxyNodes
		nds, err = c.GetProxyNodes(ctx, &g.ProxyNode{IsRunning: true})
		if err != nil {
			fmt.Printf("-----ChangeProxyNode--GetProxyNodes--err(%v)\n", err)
			return
		}
		// TODO 找出当前激活的节点

		// selectport := conf.GetConf().GetHttpProxyPort()

		nid := miniutils.GetRandInt(0, 9)
		nextIndex := nds.Items[nid].Index
		var res *g.OptResult
		res, err = c.ActiveProxyNode(ctx, &g.ProxyNode{Index: nextIndex})
		if err != nil {
			fmt.Printf("-----ChangeProxyNode--ActiveProxyNode--err(%v)\n", err)
			return
		}
		fmt.Println(res.Msg)
	})
	return err
}
