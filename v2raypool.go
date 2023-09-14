package v2raypool

import (
	"fmt"

	"sync"
	"time"

	"github.com/iotames/miniutils"
	"github.com/iotames/v2raypool/conf"
)

var once sync.Once
var instance *ProxyPool

func GetProxyPool() *ProxyPool {
	once.Do(func() {
		instance = NewProxyPool()
	})
	return instance
}

type ProxyPool struct {
	localPortStart                 int
	v2rayPath, testUrl             string
	subscribeRawData, subscribeUrl string
	testMaxDuration                time.Duration
	nodes                          ProxyNodes
	lock                           *sync.Mutex
	IsLock                         bool
	speedMap                       map[string]ProxyNodes
}

func NewProxyPool() *ProxyPool {
	return &ProxyPool{lock: &sync.Mutex{}, speedMap: make(map[string]ProxyNodes)}
}
func (p *ProxyPool) SetLocalPortStart(port int) *ProxyPool {
	p.localPortStart = port
	return p
}
func (p *ProxyPool) SetSubscribeRawData(d string) *ProxyPool {
	p.subscribeRawData = d
	return p
}
func (p *ProxyPool) SetSubscribeUrl(d string) *ProxyPool {
	p.subscribeUrl = d
	return p
}
func (p *ProxyPool) SetTestMaxDuration(d time.Duration) *ProxyPool {
	p.testMaxDuration = d
	return p
}
func (p *ProxyPool) SetTestUrl(turl string) *ProxyPool {
	p.testUrl = turl
	return p
}
func (p *ProxyPool) SetV2rayPath(path string) *ProxyPool {
	p.v2rayPath = path
	return p
}
func (p *ProxyPool) SetNodes(nds []ProxyNode) {
	p.nodes = nds
}

func (p *ProxyPool) AddNode(n ProxyNode) {
	// fmt.Printf("----Begin---AddNode(%+v)---\n", n)
	p.lock.Lock()
	ok := true
	hasPid, killStat := n.KillPidByLocalPort()
	if hasPid > 0 {
		fmt.Printf("----AddNode----Find--LocalPort(%d)---HasPID(%d)---\n", n.LocalPort, hasPid)
	}
	if killStat != nil {
		panic(fmt.Errorf("---AddNode---killPidErr(%v)-----LocalPort(%d)----HasPID(%d)---", killStat, n.LocalPort, hasPid))
	}
	for _, nd := range p.nodes {
		// kill 端口占用进程
		if nd.LocalPort == n.LocalPort {
			err := fmt.Errorf("---AddNode--端口冲突--LocalPort(%d)--", n.LocalPort)
			panic(err)
		}
		if nd.GetId() == n.GetId() {
			ok = false
		}
	}
	if ok {
		p.nodes = append(p.nodes, n)
	}
	p.lock.Unlock()
	// fmt.Printf("----End---AddNode(%+v)---\n", n)
}
func (p *ProxyPool) RemoveNode(n ProxyNode) {
	p.lock.Lock()
	var newNodes []ProxyNode
	for _, nn := range p.nodes {
		if n.GetId() != nn.GetId() {
			newNodes = append(newNodes, nn)
		}
	}
	p.nodes = newNodes
	p.lock.Unlock()
}
func (p *ProxyPool) GetNodes(domain string) ProxyNodes {
	p.lock.Lock()
	defer p.lock.Unlock()
	if domain == "" {
		return p.nodes
	}
	return p.getSpeedNodes(domain)
}

func (p *ProxyPool) UpdateNode(n ProxyNode) error {
	var err error
	find := 0
	// fmt.Printf("----BeginUpdateNode(%+v)--Id(%s)--Index(%d)\n", n, n.GetId(), n.Index)
	for i, nn := range p.nodes {
		if nn.GetId() == n.GetId() {
			fmt.Printf("---UpdateProxyNode--Index(%d)--Id(%s)--Title(%s)--IsRunning(%v)\n", n.Index, n.GetId(), n.Title, n.IsRunning())
			find++
			p.nodes[i] = n
		}
	}
	if find == 0 {
		err = fmt.Errorf("can not find node")
	}
	if find > 1 {
		err = fmt.Errorf("node find %d", find)
	}
	return err
}
func (p *ProxyPool) AddSpeedNode(key string, n ProxyNode) {
	p.lock.Lock()
	defer p.lock.Unlock()
	_, ok := p.speedMap[key]
	if ok {
		p.speedMap[key] = append(p.speedMap[key], n)
	} else {
		p.speedMap[key] = []ProxyNode{n}
	}
}
func (p *ProxyPool) getSpeedNodes(key string) []ProxyNode {
	// p.lock.Lock() 重复使用lock会导致永久锁死
	// defer p.lock.Unlock()
	nds, ok := p.speedMap[key]
	if ok {
		return nds
	}
	return []ProxyNode{}
}
func (p *ProxyPool) InitSubscribeData() *ProxyPool {
	if p.localPortStart == 0 {
		panic("please set localPortStart")
	}
	var err error
	var dt string
	if p.subscribeRawData != "" {
		dt, err = parseSubscribeByRaw(p.subscribeRawData)
		if err != nil {
			panic(err)
		}
	} else {
		if p.subscribeUrl != "" {
			dt, _, err = parseSubscribeByUrl(p.subscribeUrl, "")
			if err != nil {
				panic(err)
			}
		}
	}
	vnds := ParseV2rayNodes(dt)
	for i, vnd := range vnds {
		pnd := p.getNodeByV2rayNode(vnd, i)
		p.SetLocalAddr(&pnd, 0)
		p.AddNode(pnd)
	}
	return p
}
func (p *ProxyPool) UpdateSubscribe() (total, add int) {
	p.nodes.SortBySpeed()
	var dt string
	var err error
	var srawdata string
	for _, n := range p.nodes {
		if n.IsRunning() {
			localAddr := p.GetLocalAddr(n)
			dt, srawdata, err = parseSubscribeByUrl(p.subscribeUrl, localAddr)
			fmt.Printf("---UpdateSubscribe--UseProxy(%s)Title(%s)--Err(%v)--ParseV2rayNodes(%s)---\n", localAddr, n.Title, err, dt)
			if err == nil {
				break
			}
		}
	}
	if srawdata != "" {
		conf.GetConf().UpdateSubscribeData(srawdata)
	}

	vnds := ParseV2rayNodes(dt)
	total = len(vnds)
	if total == 0 {
		return
	}
	oldLen := len(p.nodes)
	oldNodesMap := make(map[string]ProxyNode, oldLen)
	for _, oldn := range p.nodes {
		oldNodesMap[oldn.GetId()] = oldn
	}
	newIndex := oldLen
	for i, vnd := range vnds {
		newNode := p.getNodeByV2rayNode(vnd, i)
		nid := newNode.GetId()
		_, ok := oldNodesMap[nid]
		if !ok {
			newNode.Index = newIndex
			p.SetLocalAddr(&newNode, 0)
			p.AddNode(newNode)
			newIndex++
			add++
		}
	}

	return
}
func (p ProxyPool) getNodeByV2rayNode(vnd V2rayNode, i int) ProxyNode {
	nn := NewProxyNodeByV2ray(vnd)
	nn.Index = i
	nn.TestUrl = p.testUrl
	nn.Speed = p.testMaxDuration
	return *nn
}
func (p ProxyPool) GetLocalAddr(n ProxyNode) string {
	if n.localAddr == "" {
		panic("ProxyNode.localAddr is empty")
	}
	return n.localAddr
}
func (p *ProxyPool) SetLocalAddr(n *ProxyNode, port int) string {
	if port == 0 {
		n.LocalPort = n.Index + p.localPortStart
	} else {
		n.LocalPort = port
	}
	n.localAddr = fmt.Sprintf("http://127.0.0.1:%d", n.LocalPort)
	return n.localAddr
}

func (p *ProxyPool) testOneNode(n *ProxyNode, i int) bool {
	speed, ok := testProxyNode(p.testUrl, p.GetLocalAddr(*n), i, p.testMaxDuration)
	n.Speed = speed
	if ok {
		n.TestUrl = p.testUrl
		n.TestAt = time.Now()
	}
	p.UpdateNode(*n)
	if speed < p.testMaxDuration {
		k := miniutils.GetDomainByUrl(p.testUrl)
		n.TestUrl = p.testUrl
		p.AddSpeedNode(k, *n)
	}
	return ok
}

func (p *ProxyPool) TestAllForce() {
	p.IsLock = true
	activePort := p.localPortStart - 1
	hasActive := false
	wg := sync.WaitGroup{}
	runcount := 0
	for i, n := range p.nodes {
		if n.IsRunning() {
			runcount++
			if n.LocalPort == activePort {
				hasActive = true
			}
			wg.Add(1)
			go func(n *ProxyNode, i int) {
				p.testOneNode(n, i)
				wg.Done()
			}(&n, i)
		}
	}
	wg.Wait()
	if runcount == 0 {
		p.IsLock = false
		fmt.Println("测速失败，没有可测速的代理节点。请先执行 --startproxynodes 命令，启动IP代理池")
		return
	}
	p.nodes.SortBySpeed()
	if !hasActive {
		p.ActiveNode(p.nodes[0])
	}
	p.IsLock = false
}

func (p *ProxyPool) TestAll() {
	p.IsLock = true
	activePort := p.localPortStart - 1
	hasActive := false
	// wg := sync.WaitGroup{}
	runcount := 0
	for i, n := range p.nodes {
		if n.IsRunning() {
			runcount++
			if n.LocalPort == activePort {
				hasActive = true
			}
			if miniutils.GetDomainByUrl(p.testUrl) != miniutils.GetDomainByUrl(n.TestUrl) {
				p.testOneNode(&n, i)
			} else {
				if !n.IsOk() {
					p.testOneNode(&n, i)
				}
			}
		}
	}
	// wg.Wait()
	if runcount == 0 {
		p.IsLock = false
		fmt.Println("测速失败，没有可测速的代理节点。请先执行 --startproxynodes 命令，启动IP代理池")
		return
	}
	p.nodes.SortBySpeed()
	if !hasActive {
		p.ActiveNode(p.nodes[0])
	}
	p.IsLock = false
}

func (p *ProxyPool) StartAll() error {
	var err error
	p.IsLock = true
	for _, n := range p.nodes {
		if !n.IsRunning() {
			hasPid, killStat := n.KillPidByLocalPort()
			if hasPid > 0 {
				fmt.Printf("----StartAll----Find--LocalPort(%d)---HasPID(%d)---\n", n.LocalPort, hasPid)
			}
			if killStat != nil {
				panic(fmt.Errorf("---StartAll---killPidErr(%v)-----LocalPort(%d)----HasPID(%d)---", killStat, n.LocalPort, hasPid))
			}
			err = n.Start(p.v2rayPath)
			p.UpdateNode(n)
			if err != nil {
				break
			}
		}
	}
	p.IsLock = false
	return err
}
func (p ProxyPool) getNodesMap() map[int]ProxyNode {
	mp := make(map[int]ProxyNode, len(p.nodes))
	for _, n := range p.nodes {
		mp[n.LocalPort] = n
	}
	return mp
}
func (p *ProxyPool) UpdateAfterStopAll() {
	for _, nd := range p.nodes {
		if nd.IsRunning() {
			nd.Stop()
		}
	}
	for _, nds := range p.speedMap {
		for _, n := range nds {
			if n.IsRunning() {
				n.Stop()
			}
		}
	}
}
func (p *ProxyPool) StopAll() error {
	var err error
	p.IsLock = true
	for _, n := range p.nodes {
		if n.IsRunning() {
			err = n.Stop()
			p.UpdateNode(n)
			if err != nil {
				break
			}
		}
	}
	p.UpdateAfterStopAll()
	p.IsLock = false
	return err
}
func (p *ProxyPool) KillAllNodes() (total, runport, kill, fail int) {
	var err error
	p.IsLock = true
	portToNode := p.getNodesMap()
	startPort := p.localPortStart - 1
	maxport := startPort + len(p.nodes) + 1
	for i := startPort; i < maxport; i++ {
		total++
		pid := miniutils.GetPidByPort(i)
		if pid > 0 {
			runport++
			nd, ok := portToNode[i]
			if ok {
				err = nd.Stop()
				p.UpdateNode(nd)
			} else {
				err = miniutils.KillPid(fmt.Sprintf("%d", pid))
			}
			if err != nil {
				fail++
				fmt.Printf("----KillPid(%d)--Port(%d)---Fail----\n", pid, i)
			} else {
				kill++
			}
		}
	}
	p.UpdateAfterStopAll()
	p.IsLock = false
	return
}
func (p *ProxyPool) ActiveNode(n ProxyNode) error {
	var err error
	activePort := p.localPortStart - 1
	if n.LocalPort == activePort {
		fmt.Println("-----SelectedNode.LocalPort==", activePort)
		if !n.IsRunning() {
			err = n.Start(p.v2rayPath)
			if err != nil {
				return err
			}
			err = p.UpdateNode(n)
		}
		return err
	}
	for _, nn := range p.nodes {
		if nn.LocalPort == activePort {
			fmt.Println("-----FindOtherNode.LocalPort==", activePort)
			if nn.IsRunning() {
				fmt.Println("-----StopRunning---FindOtherNode.Index=", nn.Index)
				nn.Stop()
				p.SetLocalAddr(&nn, 0)
				nn.Start(p.v2rayPath)
			}
			p.SetLocalAddr(&nn, 0)
			p.UpdateNode(nn)
		}
	}
	if n.IsRunning() {
		fmt.Println("-----StopSelectedNode------")
		n.Stop()
	}
	p.SetLocalAddr(&n, activePort)
	err = n.Start(p.v2rayPath)
	fmt.Println("-----Start---SelectedNode------")
	p.UpdateNode(n)
	return err
}

// proxyPoolInit 初始化代理池
// TODO 自带控制系统(TCP, HTTP, protobuf, 跨进程共享内存)
// TODO 独立运行，使用protobuf与其他进程交互，在schedule中更新订阅
// https://cloud.tencent.com/developer/article/1564128
func proxyPoolInit() {
	startAt := time.Now()
	cf := conf.GetConf()
	subscribeRawData := cf.GetSubscribeData()
	if cf.SubscribeUrl == "" {
		// subscribeRawData == ""
		panic("subscribe url can not be empty.订阅地址不能为空")
	}
	port := cf.GetHttpProxyPort()
	// 程序退出后，还会存在端口占用
	// port := 11080
	maxDuration := 3 * time.Second
	pp := GetProxyPool()
	pp.SetV2rayPath(cf.V2rayPath).
		SetTestUrl(cf.TestUrl).
		SetSubscribeRawData(subscribeRawData).
		SetSubscribeUrl(cf.SubscribeUrl).
		SetLocalPortStart(port + 1).
		SetTestMaxDuration(maxDuration).
		InitSubscribeData()
	nds := pp.GetNodes("")
	for i, n := range nds {
		fmt.Printf("---[%d]--Lport(%d)--Speed(%.3f)--Run(%v)--TestAt(%s)--Remote(%s)--T(%s)--index(%d)\n", i, n.LocalPort, n.Speed.Seconds(), n.IsRunning(), n.TestAt.Format("2006-01-02 15:04"), n.RemoteAddr, n.Title, n.Index)
	}
	fmt.Printf("-----SUCCESS--RunProxyPoolInit--cost(%.3fs)--\n", time.Since(startAt).Seconds())
}
