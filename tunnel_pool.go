package v2raypool

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/iotames/miniutils"
	"github.com/iotames/v2raypool/conf"
	"github.com/iotames/v2raypool/netutil"
)

// TunnelPool 隧道代理池
// 对外暴露单一 HTTP 代理端口，每次新请求随机选择一个可用节点作为出口 IP，
// 达到 IP 切换效果，防止爬虫被目标网站的反爬策略封锁。
//
// 使用方式：
//   curl --proxy http://127.0.0.1:1080 https://httpbin.org/ip
//   每次请求，httpbin 返回的 IP 会随机变化（如果隧道池中有多个可用节点）。
//
// 设计原则：
//   - 完全复用现有代理池的节点管理、订阅解析、测速等基础设施
//   - 不修改原有 ProxyPool 的核心逻辑，仅新增隧道层
//   - 与原有代理池（多端口模式）可同时运行
type TunnelPool struct {
	mu       sync.RWMutex
	server   *http.Server
	pool     *ProxyPool       // 复用现有代理池
	config   TunnelConfig     // 隧道配置
	ctx      context.Context
	cancel   context.CancelFunc
	running  bool
	nodeList ProxyNodes       // 当前可用的节点快照
}

// NewTunnelPool 创建隧道代理池实例
// 复用 GetProxyPool() 单例，从现有代理池中获取测速合格的节点
func NewTunnelPool(cfg TunnelConfig) *TunnelPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &TunnelPool{
		pool:   GetProxyPool(),
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动隧道代理 HTTP 服务
// 在独立 goroutine 中运行，监听指定端口，等待 HTTP 代理请求。
// 返回启动是否成功；端口被占用等错误通过 error 返回。
func (tp *TunnelPool) Start() error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if tp.running {
		return fmt.Errorf("隧道代理池已在运行中")
	}

	addr := fmt.Sprintf(":%d", tp.config.Port)
	tp.server = &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(tp.handleProxy),
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("隧道代理池端口 %d 被占用: %v", tp.config.Port, err)
	}

	go func() {
		fmt.Printf("-----隧道代理池已启动，监听端口(%d)，最大延迟(%dms)------\n", tp.config.Port, tp.config.MaxDelayMs)
		if err := tp.server.Serve(ln); err != nil && err != http.ErrServerClosed {
			fmt.Printf("-----隧道代理池异常退出: %v------\n", err)
		}
	}()

	tp.running = true

	// 后台定时刷新可用节点列表（首次立即执行）
	go tp.refreshNodesLoop()

	return nil
}

// Stop 停止隧道代理 HTTP 服务
func (tp *TunnelPool) Stop() error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if !tp.running {
		return nil
	}
	tp.cancel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := tp.server.Shutdown(ctx)
	tp.running = false
	fmt.Printf("-----隧道代理池已停止------\n")
	return err
}

// IsRunning 返回隧道代理池运行状态
func (tp *TunnelPool) IsRunning() bool {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	return tp.running
}

// SetMaxDelay 在线更新延迟阈值（毫秒），下次刷新节点时生效
// ms 必须 > 0，否则不生效
func (tp *TunnelPool) SetMaxDelay(ms int) {
	if ms <= 0 {
		return
	}
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.config.MaxDelayMs = ms
}

// GetStatus 返回隧道代理池状态信息（供 WebUI 使用）
func (tp *TunnelPool) GetStatus() map[string]interface{} {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	return map[string]interface{}{
		"running":          tp.running,
		"port":             tp.config.Port,
		"max_delay_ms":     tp.config.MaxDelayMs,
		"refresh_interval": tp.config.RefreshInterval,
		"node_count":       len(tp.nodeList),
	}
}

// RefreshNodes 刷新可用节点列表
// 对 ProxyPool 中所有运行中的节点逐个测速（复用 testOneNode，结果自动写回 speedMap），
// 然后按最大延迟阈值筛选可用节点，更新 nodeList。
// ProxyPool.TestAll 再次跑时会通过 GetLastSpeedNode 跳过近期已测速节点，避免重复测速。
func (tp *TunnelPool) RefreshNodes() int {
	cf := conf.GetConf()
	maxDelay := cf.TunnelMaxDelay
	if maxDelay <= 0 {
		maxDelay = tp.config.MaxDelayMs
	}
	duration := time.Duration(maxDelay) * time.Millisecond

	tp.pool.lock.Lock()
	var runningNodes []ProxyNode
	for i := range tp.pool.nodes {
		nd := &tp.pool.nodes[i]
		if nd.IsDelete || !nd.IsRunning() {
			continue
		}
		runningNodes = append(runningNodes, *nd)
	}
	tp.pool.lock.Unlock()

	// 逐个测速，跳过近期已测速的节点（24h内有效），复用已有速度数据
	logger := conf.GetConf().GetLogger()
	domain := miniutils.GetDomainByUrl(tp.pool.testUrl)
	logger.Infof("-----RefreshNodes--testUrl(%q)--domain(%q)--running(%d)--maxDelay(%dms)------", tp.pool.testUrl, domain, len(runningNodes), maxDelay)
	for i := range runningNodes {
		nd := &runningNodes[i]
		// 检查是否有针对当前测速域名的有效测速数据
		oknd := tp.pool.GetLastSpeedNode(*nd, domain)
		if oknd.IsOk() {
			nd.Speed = oknd.Speed
			nd.TestUrl = oknd.TestUrl
			nd.TestAt = oknd.TestAt
			logger.Infof("-----RefreshNodes--skipTest--idx(%d)--id(%s)--speed(%.2fms)------", nd.Index, nd.GetId(), nd.Speed.Seconds()*1000)
		} else {
			logger.Infof("-----RefreshNodes--needTest--idx(%d)--id(%s)--okndIsOk(%v)------", nd.Index, nd.GetId(), oknd.IsOk())
			tp.pool.testOneNode(nd, nd.Index)
			logger.Infof("-----RefreshNodes--tested--idx(%d)--speed(%.2fms)--ok(%v)------", nd.Index, nd.Speed.Seconds()*1000, nd.Speed < tp.pool.testMaxDuration)
		}
	}

	// 按最新测速结果筛选
	var available ProxyNodes
	for _, nd := range runningNodes {
		if nd.Speed < duration {
			available = append(available, nd)
		}
	}
	available.SortBySpeed()

	tp.mu.Lock()
	tp.nodeList = available
	tp.mu.Unlock()

	logger.Infof("-----隧道代理池刷新节点: 共 %d 个可用节点(%dms阈值内), 运行中%d, 筛选后%d------", len(available), maxDelay, len(runningNodes), len(available))
	return len(available)
}

// refreshNodesLoop 后台定时刷新节点，间隔由配置项 VP_TUNNEL_REFRESH_INTERVAL 决定
func (tp *TunnelPool) refreshNodesLoop() {
	interval := tp.config.RefreshInterval
	if interval <= 0 {
		interval = DEFAULT_TUNNEL_REFRESH_INTERVAL
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	// 首次立即刷新
	tp.RefreshNodes()

	for {
		select {
		case <-tp.ctx.Done():
			return
		case <-ticker.C:
			tp.RefreshNodes()
		}
	}
}

// handleProxy 处理 HTTP 代理请求（CONNECT 隧道和普通 HTTP 代理）
// 对每个新请求随机选择一个出口节点，不支持连接复用时的节点切换（同一 TCP 连接固定节点）
func (tp *TunnelPool) handleProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		tp.handleConnect(w, r)
		return
	}
	tp.handleHTTP(w, r)
}

// handleConnect 处理 HTTPS CONNECT 隧道
// 客户端->隧道代理->随机节点->目标服务器，双向透明转发数据
func (tp *TunnelPool) handleConnect(w http.ResponseWriter, r *http.Request) {
	nodeAddr := tp.pickNodeAddr()
	if nodeAddr == "" {
		http.Error(w, "没有可用的代理节点", http.StatusServiceUnavailable)
		return
	}

	// 连接到上游代理节点（使用 CONNECT 方式）
	destConn, err := net.DialTimeout("tcp", parseHostPort(nodeAddr), 10*time.Second)
	if err != nil {
		http.Error(w, fmt.Sprintf("无法连接到代理节点: %v", err), http.StatusBadGateway)
		return
	}
	defer destConn.Close()

	// 向代理节点发送 CONNECT 请求
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", r.Host, r.Host)
	if _, err := destConn.Write([]byte(connectReq)); err != nil {
		http.Error(w, "代理连接失败", http.StatusBadGateway)
		return
	}

	// 读取代理节点的 CONNECT 响应
	buf := make([]byte, 1024)
	n, err := destConn.Read(buf)
	if err != nil {
		http.Error(w, "代理响应失败", http.StatusBadGateway)
		return
	}

	respStr := string(buf[:n])
	if !strings.Contains(respStr, "200") {
		http.Error(w, "代理隧道建立失败", http.StatusBadGateway)
		return
	}

	// 告诉客户端隧道已建立
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "不支持 Hijack", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	// 双向转发
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(destConn, clientConn)
		destConn.(*net.TCPConn).CloseWrite()
	}()
	go func() {
		defer wg.Done()
		io.Copy(clientConn, destConn)
		clientConn.(*net.TCPConn).CloseWrite()
	}()
	wg.Wait()
}

// handleHTTP 处理普通 HTTP 代理请求
// 将请求通过随机节点转发，然后返回响应给客户端
func (tp *TunnelPool) handleHTTP(w http.ResponseWriter, r *http.Request) {
	nodeAddr := tp.pickNodeAddr()
	if nodeAddr == "" {
		http.Error(w, "没有可用的代理节点", http.StatusServiceUnavailable)
		return
	}

	// 构造目标 URL
	targetURL := r.URL.String()
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = fmt.Sprintf("http://%s%s", r.Host, r.URL.RequestURI())
	}

	// 使用代理节点发起请求
	maxDuration := 30 * time.Second
	proxyURL := fmt.Sprintf("http://%s", nodeAddr)
	client, req := netutil.GetHttpClient(maxDuration, targetURL, proxyURL)
	req.Method = r.Method

	// 复制请求头（去掉代理相关头）
	for k, vs := range r.Header {
		kl := strings.ToLower(k)
		if kl == "proxy-connection" || kl == "proxy-authorization" {
			continue
		}
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("代理请求失败: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// pickNodeAddr 随机选取一个可用节点的本地代理地址
// 返回格式：127.0.0.1:30001（HTTP 代理地址去掉协议头）
func (tp *TunnelPool) pickNodeAddr() string {
	tp.mu.RLock()
	nodes := tp.nodeList
	tp.mu.RUnlock()

	if len(nodes) == 0 {
		return ""
	}
	idx := rand.Intn(len(nodes))
	addr := nodes[idx].LocalAddr
	// LocalAddr 格式为 http://127.0.0.1:30001 或 socks5://127.0.0.1:30001
	u, err := url.Parse(addr)
	if err != nil {
		return ""
	}
	return u.Host
}

// parseHostPort 从 "host:port" 格式解析出 TCP 地址
func parseHostPort(addr string) string {
	if strings.Contains(addr, ":") {
		return addr
	}
	return addr + ":80"
}
