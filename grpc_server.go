package v2raypool

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/iotames/v2raypool/conf"
	g "github.com/iotames/v2raypool/grpc"
	"google.golang.org/grpc"
)

type ProxyPoolServer struct {
	g.UnimplementedProxyPoolServiceServer
}

func skipProxyNode(req *g.ProxyNode, n ProxyNode) bool {
	if req.IsRunning && !n.IsRunning() {
		return true
	}
	// if !req.IsRunning && n.IsRunning() {
	// 	return true
	// }
	if req.Title != "" {
		if !strings.Contains(n.Title, req.Title) {
			return true
		}
	}
	if req.LocalPort != 0 {
		if int(req.LocalPort) != n.LocalPort {
			return true
		}
	}
	return false
}

func (s ProxyPoolServer) GetProxyNodes(ctx context.Context, req *g.ProxyNode) (*g.ProxyNodes, error) {
	resp := g.ProxyNodes{}

	pp := GetProxyPool()
	nds := pp.GetNodes("")
	nds.SortBySpeed()

	for _, n := range nds {
		if skipProxyNode(req, n) {
			continue
		}
		resp.Items = append(resp.Items, &g.ProxyNode{
			Index:      uint32(n.Index),
			Id:         n.Id,
			LocalPort:  uint32(n.LocalPort),
			Speed:      float32(n.Speed.Seconds()),
			Title:      n.Title,
			LocalAddr:  pp.GetLocalAddr(n),
			RemoteAddr: n.RemoteAddr,
			IsRunning:  n.IsRunning(),
			IsOk:       n.IsOk(),
			TestAt:     n.TestAt.Format("2006-01-02 15:04"),
		})
	}
	return &resp, nil
}

func (s ProxyPoolServer) GetProxyNodesByDomain(ctx context.Context, req *g.OptRequestDomain) (*g.ProxyNodes, error) {
	resp := g.ProxyNodes{}
	pp := GetProxyPool()
	domain := req.GetDomain()
	if domain == "" {
		return &resp, fmt.Errorf("domain不能为空")
	}
	nds := pp.GetNodes(domain)
	nds.SortBySpeed()
	for _, n := range nds {
		resp.Items = append(resp.Items, &g.ProxyNode{
			Index:      uint32(n.Index),
			Id:         n.Id,
			LocalPort:  uint32(n.LocalPort),
			Speed:      float32(n.Speed.Seconds()),
			Title:      n.Title,
			LocalAddr:  pp.GetLocalAddr(n),
			RemoteAddr: n.RemoteAddr,
			IsRunning:  n.IsRunning(),
			IsOk:       n.IsOk(),
			TestAt:     n.TestAt.Format("2006-01-02 15:04"),
		})
	}
	return &resp, nil
}

func (s ProxyPoolServer) SetTestUrl(ctx context.Context, req *g.OptRequestUrl) (result *g.OptResult, err error) {
	result = &g.OptResult{Status: 200, Msg: "设置成功"}
	pp := GetProxyPool()
	testUrl := req.GetUrl()
	if strings.Index(testUrl, "http") != 0 {
		result.Status = 400
		result.Msg = "TestUrl 格式不正确"
		return
	}
	pp.SetTestUrl(testUrl)
	cf := conf.GetConf()
	cf.TestUrl = testUrl
	conf.SetConf(cf)
	return
}

func (s ProxyPoolServer) StartProxyPoolAll(ctx context.Context, req *g.OptRequest) (result *g.OptResult, err error) {
	result = &g.OptResult{}
	pp := GetProxyPool()
	if pp.IsLock {
		msg := "系统繁忙，请稍候"
		result.Status = 500
		result.Msg = msg
		err = fmt.Errorf(msg)
		return
	}
	err = pp.StartAll()
	if err != nil {
		msg := err.Error()
		result.Status = 500
		result.Msg = msg
		err = fmt.Errorf(msg)
		return
	}
	result.Status = 200
	result.Msg = "启动成功"
	return
}
func (s ProxyPoolServer) StopProxyPoolAll(ctx context.Context, req *g.OptRequest) (result *g.OptResult, err error) {
	result = &g.OptResult{}
	pp := GetProxyPool()
	if pp.IsLock {
		msg := "系统繁忙，请稍候"
		result.Status = 500
		result.Msg = msg
		err = fmt.Errorf(msg)
		return
	}
	err = pp.StopAll()
	if err != nil {
		msg := err.Error()
		result.Status = 500
		result.Msg = msg
		err = fmt.Errorf(msg)
		return
	}
	result.Status = 200
	result.Msg = "代理池所有端口进程已停止"
	return
}
func (s ProxyPoolServer) TestProxyPoolAll(ctx context.Context, req *g.OptRequest) (result *g.OptResult, err error) {
	result = &g.OptResult{}
	pp := GetProxyPool()
	if pp.IsLock {
		msg := "系统繁忙，请稍候"
		result.Status = 500
		result.Msg = msg
		err = fmt.Errorf(msg)
		return
	}
	if len(GetRunningNodes()) == 0 {
		msg := "测速失败，没有可测速的代理节点。请先执行 --startproxynodes 命令，启动IP代理池"
		result.Status = 400
		result.Msg = msg
		err = fmt.Errorf(msg)
		return
	}
	go pp.TestAll()
	result.Status = 200
	result.Msg = "测速已开始，请稍候片刻"
	return
}
func (s ProxyPoolServer) TestProxyPoolAllForce(ctx context.Context, req *g.OptRequest) (result *g.OptResult, err error) {
	result = &g.OptResult{}
	pp := GetProxyPool()
	if pp.IsLock {
		msg := "系统繁忙，请稍候"
		result.Status = 500
		result.Msg = msg
		err = fmt.Errorf(msg)
		return
	}
	if len(GetRunningNodes()) == 0 {
		msg := "测速失败，没有可测速的代理节点。请先执行 --startproxynodes 命令，启动IP代理池"
		result.Status = 400
		result.Msg = msg
		err = fmt.Errorf(msg)
		return
	}
	go pp.TestAllForce()
	result.Status = 200
	result.Msg = "强制测速已开始，请稍候片刻"
	return
}
func (s ProxyPoolServer) ActiveProxyNode(ctx context.Context, req *g.ProxyNode) (result *g.OptResult, err error) {
	result = &g.OptResult{}
	pp := GetProxyPool()
	index := req.Index
	for _, nd := range pp.GetNodes("") {
		if nd.Index == int(index) {
			err = pp.ActiveNode(nd, true)
			break
		}
	}
	status := 200
	msg := "操作成功"
	if err != nil {
		status = 500
		msg = err.Error()
	}
	result.Status = uint32(status)
	result.Msg = msg
	return
}
func (s ProxyPoolServer) UpdateProxySubscribe(ctx context.Context, req *g.OptRequest) (result *g.UpdateSubscribeResult, err error) {
	result = &g.UpdateSubscribeResult{}
	pp := GetProxyPool()
	if pp.IsLock {
		msg := "系统繁忙，请稍候"
		result.Status = 500
		result.Msg = msg
		err = fmt.Errorf(msg)
		return
	}
	total, add := pp.UpdateSubscribe("")
	fmt.Printf("---total(%d)---add(%d)----\n", total, add)
	result.Status = 200
	result.Msg = "订阅更新完成"
	return
}
func (s ProxyPoolServer) KillAllNodes(ctx context.Context, req *g.OptRequest) (result *g.KillNodesResult, err error) {
	result = &g.KillNodesResult{}
	pp := GetProxyPool()
	if pp.IsLock {
		msg := "系统繁忙，请稍候"
		result.Status = 500
		result.Msg = msg
		err = fmt.Errorf(msg)
		return
	}
	total, runport, kill, fail := pp.KillAllNodes()
	result.Total, result.Runport, result.Kill, result.Fail = uint32(total), uint32(runport), uint32(kill), uint32(fail)
	result.Status = 200
	result.Msg = "操作完成"
	return
}

func GetRunningNodes() ProxyNodes {
	pp := GetProxyPool()
	nds := pp.GetNodes("")
	var result []ProxyNode
	for _, nd := range nds {
		if nd.IsRunning() {
			result = append(result, nd)
		}
	}
	return result
}

// 全局隧道代理池实例
var globalTunnelPool *TunnelPool

func RunServer() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg1 := sync.WaitGroup{}
	wg1.Add(1)
	go func() {
		proxyPoolInit()
		wg.Done()
		wg1.Wait()
	}()
	wg.Wait()
	RunProxyPoolGrpcServer()
}

// InitTunnelPool 初始化并启动隧道代理池（若配置启用）
// 必须在 proxyPoolInit() 和节点启动完成后调用
func InitTunnelPool() error {
	cf := conf.GetConf()
	if !cf.TunnelEnabled {
		return nil
	}
	tc := TunnelConfig{
		Enable:     cf.TunnelEnabled,
		Port:       cf.TunnelPort,
		MaxDelayMs: cf.TunnelMaxDelay,
		RefreshInterval: cf.TunnelRefreshInterval,
	}
	globalTunnelPool = NewTunnelPool(tc)
	err := globalTunnelPool.Start()
	if err != nil {
		globalTunnelPool = nil
		cf.GetLogger().Errorf("隧道代理池启动失败: %v", err)
		return err
	}
	return nil
}

// GetTunnelPool 获取全局隧道代理池实例
func GetTunnelPool() *TunnelPool {
	return globalTunnelPool
}

// StartTunnelPool 手动启动隧道代理池（WebUI 调用）
func StartTunnelPool() error {
	cf := conf.GetConf()
	tc := TunnelConfig{
		Enable:     true,
		Port:       cf.TunnelPort,
		MaxDelayMs: cf.TunnelMaxDelay,
		RefreshInterval: cf.TunnelRefreshInterval,
	}
	if globalTunnelPool != nil && globalTunnelPool.IsRunning() {
		return fmt.Errorf("隧道代理池已在运行中")
	}
	globalTunnelPool = NewTunnelPool(tc)
	return globalTunnelPool.Start()
}

// StopTunnelPool 停止隧道代理池（WebUI 调用）
func StopTunnelPool() error {
	if globalTunnelPool == nil {
		return nil
	}
	return globalTunnelPool.Stop()
}

func RunProxyPoolGrpcServer() {
	lisAddr := fmt.Sprintf(":%d", conf.GetConf().GrpcPort)
	lis, err := net.Listen("tcp", lisAddr)
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	g.RegisterProxyPoolServiceServer(s, ProxyPoolServer{})

	if err := s.Serve(lis); err != nil {
		err = fmt.Errorf("failed to start grpc server: %v", err)
		panic(err)
	} else {
		fmt.Printf("SUCCESS: gRPC Server Listening At(%+v)\n", lis.Addr())
	}
}

// SysProxyType 系统代理类型
type SysProxyType int

const (
	SysProxyNone   SysProxyType = iota // 无代理
	SysProxyNode                       // 固定节点代理
	SysProxyTunnel                     // 隧道代理（随机IP）
)

// 全局系统代理状态
var (
	sysProxyType    SysProxyType = SysProxyNone
	sysProxyNodeIdx int         = -1
)

// GetSysProxyStatus 获取当前系统代理状态
func GetSysProxyStatus() map[string]interface{} {
	return map[string]interface{}{
		"type":     int(sysProxyType),
		"node_idx": sysProxyNodeIdx,
	}
}

// SetSysProxy 切换系统代理类型
// proxyType: 0=无代理, 1=固定节点(需传 nodeIdx), 2=隧道代理
// nodeIdx: 节点索引，仅 SysProxyNode 模式需要
func SetSysProxy(proxyType SysProxyType, nodeIdx int) error {
	pp := GetProxyPool()
	cf := conf.GetConf()

	// 先取消当前代理
	if sysProxyType != SysProxyNone {
		if sysProxyType == SysProxyNode {
			nds := pp.GetNodes("")
			for _, nd := range nds {
				if nd.Index == sysProxyNodeIdx {
					pp.UnActiveNode(nd)
					break
				}
			}
		}
	}

	switch proxyType {
	case SysProxyNone:
		sysProxyType = SysProxyNone
		sysProxyNodeIdx = -1
		return nil

	case SysProxyNode:
		nds := pp.GetNodes("")
		var target ProxyNode
		found := false
		if nodeIdx >= 0 {
			for _, nd := range nds {
				if nd.Index == nodeIdx {
					target = nd
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("节点索引 %d 不存在", nodeIdx)
			}
		} else {
			// 未指定索引时自动选择第一个运行中的节点
			for _, nd := range nds {
				if nd.IsRunning() {
					target = nd
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("没有运行中的节点，请先在列表中启动节点")
			}
		}
		if err := pp.ActiveNode(target, true); err != nil {
			return err
		}
		sysProxyType = SysProxyNode
		sysProxyNodeIdx = target.Index

	case SysProxyTunnel:
		if globalTunnelPool == nil || !globalTunnelPool.IsRunning() {
			if err := StartTunnelPool(); err != nil {
				return fmt.Errorf("隧道代理启动失败: %v", err)
			}
		}
		tunnelAddr := fmt.Sprintf("127.0.0.1:%d", cf.TunnelPort)
		if err := SetProxy(tunnelAddr); err != nil {
			return fmt.Errorf("设置隧道代理为系统代理失败: %v", err)
		}
		fmt.Printf("设置系统代理为隧道代理: %s 成功!\n", tunnelAddr)
		sysProxyType = SysProxyTunnel
		sysProxyNodeIdx = -1

	default:
		return fmt.Errorf("无效的系统代理类型: %d (有效值: 0=无代理, 1=固定节点, 2=隧道代理)", proxyType)
	}
	return nil
}

// UpdateSysProxyNode WebUI 通过旧 ActiveNode 激活节点后，同步更新全局系统代理状态
func UpdateSysProxyNode(nodeIdx int) {
	sysProxyType = SysProxyNode
	sysProxyNodeIdx = nodeIdx
}

// IsSysProxyNodeActive 判断指定节点是否为当前激活的系统代理节点
func IsSysProxyNodeActive(nodeIdx int) bool {
	return sysProxyType == SysProxyNode && sysProxyNodeIdx == nodeIdx
}
