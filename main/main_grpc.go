package main

import (
	"context"
	"fmt"

	vp "github.com/iotames/v2raypool"
	g "github.com/iotames/v2raypool/grpc"
)

func getProxyNodes() {
	vp.RequestProxyPoolGrpcOnce(func(c g.ProxyPoolServiceClient, ctx context.Context) {
		nds, err := c.GetProxyNodes(ctx, &g.ProxyNode{})
		if err != nil {
			panic(err)
		}
		printNodes(nds)
	})
}

func getProxyNodesByDomain(domain string) {
	vp.RequestProxyPoolGrpcOnce(func(c g.ProxyPoolServiceClient, ctx context.Context) {
		nds, err := c.GetProxyNodesByDomain(ctx, &g.OptRequestDomain{Domain: domain})
		if err != nil {
			panic(err)
		}
		printNodes(nds)
	})
}

func printNodes(nds *g.ProxyNodes) {
	var total, countrun, countok int
	for _, n := range nds.Items {
		total++
		if n.IsRunning {
			countrun++
		}
		if n.IsOk {
			countok++
		}
		fmt.Printf("-----lPort(%d)--speed(%.3f)--isRun(%v)--TestAt(%s)--title(%s)--index(%d)\n", n.LocalPort, n.Speed, n.IsRunning, n.TestAt, n.Title, n.Index)
	}
	fmt.Printf("\n-----Total(%d)--CountRun(%d)--CountOk(%d)---\n", total, countrun, countok)
}

func setProxyTestUrl(url string) {
	vp.RequestProxyPoolGrpcOnce(func(c g.ProxyPoolServiceClient, ctx context.Context) {
		c.SetTestUrl(ctx, &g.OptRequestUrl{Url: url})
	})
}

func killProxyNodes() {
	vp.RequestProxyPoolGrpcOnce(func(c g.ProxyPoolServiceClient, ctx context.Context) {
		result, err := c.KillAllNodes(ctx, &g.OptRequest{})
		if err != nil {
			panic(err)
		}
		fmt.Printf("---OptMsg(%s)--Total(%d)--RunPortCount(%d)--Kill(%d)--Fail(%d)---\n", result.Msg, result.Total, result.Runport, result.Kill, result.Fail)
	})
}

func activeProxyNode(index int) {
	vp.RequestProxyPoolGrpcOnce(func(c g.ProxyPoolServiceClient, ctx context.Context) {
		if index == 666666 {
			err := vp.ChangeProxyNode()
			if err != nil {
				panic(err)
			}
			fmt.Println("ChangeProxyNode OK")
			return
		}
		res, err := c.ActiveProxyNode(ctx, &g.ProxyNode{Index: uint32(index)})
		if err != nil {
			panic(err)
		}
		fmt.Println(res.Msg)
	})
}

func testProxyNodes(force bool) {
	vp.RequestProxyPoolGrpcOnce(func(c g.ProxyPoolServiceClient, ctx context.Context) {
		var res *g.OptResult
		var err error
		if force {
			res, err = c.TestProxyPoolAllForce(ctx, &g.OptRequest{})
		} else {
			res, err = c.TestProxyPoolAll(ctx, &g.OptRequest{})
		}
		if err != nil {
			panic(err)
		}
		fmt.Println(res.Msg)
	})
}

func startAllProxyNodes() {
	vp.RequestProxyPoolGrpcOnce(func(c g.ProxyPoolServiceClient, ctx context.Context) {
		res, err := c.StartProxyPoolAll(ctx, &g.OptRequest{})
		if err != nil {
			panic(err)
		}
		fmt.Println(res.Msg)
	})
}

func updateProxyNodes() {
	vp.RequestProxyPoolGrpcOnce(func(c g.ProxyPoolServiceClient, ctx context.Context) {
		res, err := c.UpdateProxySubscribe(ctx, &g.OptRequest{})
		if err != nil {
			panic(err)
		}
		fmt.Println(res.Msg)
	})
}

func stopProxyNodes() {
	vp.RequestProxyPoolGrpcOnce(func(c g.ProxyPoolServiceClient, ctx context.Context) {
		res, err := c.StopProxyPoolAll(ctx, &g.OptRequest{})
		if err != nil {
			panic(err)
		}
		fmt.Println(res.Msg)
	})
}
