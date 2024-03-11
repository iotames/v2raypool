package v2raypool

import (
	"fmt"

	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/iotames/miniutils"
	"github.com/iotames/v2raypool/conf"
)

// 节点测速最大超时设置。
const MAX_TEST_DURATION = 5 * time.Second

var once sync.Once
var instance *ProxyPool

func GetProxyPool() *ProxyPool {
	once.Do(func() {
		instance = NewProxyPool()
	})
	return instance
}

type ProxyPool struct {
	serverMap                      map[int]*V2rayServer
	startAt                        time.Time
	activeCmd                      *exec.Cmd
	activeNode                     ProxyNode
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
	return &ProxyPool{lock: &sync.Mutex{}, speedMap: make(map[string]ProxyNodes), serverMap: make(map[int]*V2rayServer, 2)}
}
func (p ProxyPool) GetLocalPortRange() string {
	return fmt.Sprintf("%d-%d", p.nodes[0].LocalPort, p.nodes[len(p.nodes)-1].LocalPort)
}
func (p ProxyPool) GetLocalPortList() (dl []int, err error) {
	cf := conf.GetConf()
	for _, v := range p.serverMap {
		conf := V2rayConfigV4{}
		err = v.jconf.Decode(&conf)
		if err != nil {
			return
		}
		for _, inb := range conf.Inbounds {
			dl = append(dl, inb.Port)
			if inb.Port == cf.V2rayApiPort {
				for _, nd := range p.nodes {
					dl = append(dl, nd.LocalPort)
				}
			}
		}
	}
	return
}

func (p ProxyPool) CheckV2rayConfig(jconf JsonConfig) error {
	vconf := V2rayConfigV4{}
	err := jconf.Decode(&vconf)
	if err != nil {
		return err
	}
	lports, err := p.GetLocalPortList()
	if err != nil {
		return err
	}
	for _, port := range lports {
		for _, inb := range vconf.Inbounds {
			if port == inb.Port {
				return fmt.Errorf("本地端口号重复:%d", port)
			}
		}
	}
	return nil
}

func (p *ProxyPool) AddV2rayServer(vs *V2rayServer) error {
	pid := vs.GetExeCmd().Process.Pid
	for k := range p.serverMap {
		if k == pid {
			return fmt.Errorf("PID(%d)重复", pid)
		}
	}
	p.serverMap[pid] = vs
	return nil
}

func (p *ProxyPool) DeleteV2rayServer(pid int) error {
	v, ok := p.serverMap[pid]
	if !ok {
		return fmt.Errorf("v2ray server PID(%d)不存在", pid)
	}
	err := v.GetExeCmd().Process.Kill()
	if err != nil {
		return err
	}
	delete(p.serverMap, pid)
	return nil
}

func (p ProxyPool) GetV2rayServerList() []*V2rayServer {
	var dl []*V2rayServer
	for _, vs := range p.serverMap {
		dl = append(dl, vs)
	}
	return dl
}

func (p ProxyPool) GetActiveNode() ProxyNode {
	return p.activeNode
}

func (p *ProxyPool) StartV2rayPool() {
	// -----SUCCESS--RunProxyPoolInit--Pid(13628)--cost(1.687s)--
	vs := NewV2ray(p.v2rayPath)
	err := vs.Start("")
	if err == nil {
		pcmd := vs.GetExeCmd()
		p.serverMap[pcmd.Process.Pid] = vs
		fmt.Printf("-----SUCCESS--RunProxyPoolInit--Pid(%d)--cost(%.3fs)--\n", pcmd.Process.Pid, time.Since(p.startAt).Seconds())
	} else {
		fmt.Printf("-----FAIL--StartV2rayCoreFail-----cost(%.3fs)--\n", time.Since(p.startAt).Seconds())
	}
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
	if !miniutils.IsPathExists(path) {
		panic(fmt.Errorf("v2ray path %s not found", path))
	}
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
	// hasPid, killStat := n.KillPidByLocalPort()
	// if hasPid > 0 {
	// 	fmt.Printf("----AddNode----Find--LocalPort(%d)---HasPID(%d)---\n", n.LocalPort, hasPid)
	// }
	// if killStat != nil {
	// 	panic(fmt.Errorf("---AddNode---killPidErr(%v)-----LocalPort(%d)----HasPID(%d)---", killStat, n.LocalPort, hasPid))
	// }
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
			fmt.Printf("---UpdateProxyNode--Index(%d)--Id(%s)--Title(%s)--IsRunning(%v)--Speed(%.3fs)--\n", n.Index, n.GetId(), n.Title, n.IsRunning(), n.Speed.Seconds())
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
	oknds, ok := p.speedMap[key]
	if ok {
		// 找出重复项，防止节点重复插入
		okcount := 0
		for i, oknd := range oknds {
			if oknd.GetId() == n.GetId() {
				p.speedMap[key][i] = n
				okcount += 1
			}
		}
		if okcount == 0 {
			p.speedMap[key] = append(p.speedMap[key], n)
		}
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
	p.startAt = time.Now()
	var err error
	var dt string
	if p.subscribeRawData != "" {
		dt, err = parseSubscribeByRaw(p.subscribeRawData)
		if err != nil {
			panic(err)
		}
	} else {
		if p.subscribeUrl != "" {
			var rawdt string
			dt, rawdt, err = parseSubscribeByUrl(p.subscribeUrl, "")
			if err != nil {
				fmt.Printf("-----InitSubscribeData-parseSubscribeByUrl-err(%v)\n", err)
				dt, rawdt, err = parseSubscribeByUrl(p.subscribeUrl, fmt.Sprintf("http://127.0.0.1:%d", p.localPortStart-1))
				if err != nil {
					panic(err)
				}
			}

			fmt.Printf("------InitSubscribeData----rawdt(%s)----\n", rawdt)
			if rawdt != "" {
				conf.GetConf().UpdateSubscribeData(rawdt)
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
	fmt.Printf("-----testOneNode--ok(%v)--speed(%.4f)--nodeTestUrl(%s)-----ProxyPool.testUrl(%s)-----\n", ok, speed.Seconds(), n.TestUrl, p.testUrl)
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
	runcount := 0
	logger := miniutils.GetLogger("")
	for i, n := range p.nodes {
		if n.IsRunning() {
			runcount++
			logger.Debugf("---i[%d]---n.addr(%s)---localPort(%d)---", i, n.RemoteAddr, n.LocalPort)
			p.testOneNode(&n, i)
		}
	}
	if runcount == 0 {
		p.IsLock = false
		fmt.Println("测速失败，没有可测速的代理节点。请先执行 --startproxynodes 命令，启动IP代理池")
		return
	}
	p.nodes.SortBySpeed()

	if p.activeCmd == nil {
		p.ActiveNode(p.nodes[0])
	}
	p.IsLock = false
}

// GetLastSpeedNode 查看当前节点是否存在测速信息
func (p ProxyPool) GetLastSpeedNode(nd ProxyNode, d string) ProxyNode {
	p.lock.Lock()
	defer p.lock.Unlock()
	result := ProxyNode{}
	oknds, ok := p.speedMap[d]
	if ok {
		for _, oknd := range oknds {
			if oknd.GetId() == nd.GetId() {
				result = oknd
				break
			}
		}
	}
	return result
}

func (p *ProxyPool) TestAll() {
	p.IsLock = true
	wg := sync.WaitGroup{}
	runcount := 0
	for i, n := range p.nodes {
		if n.IsRunning() {
			runcount++

			domain := miniutils.GetDomainByUrl(p.testUrl)
			oknd := p.GetLastSpeedNode(n, domain)
			// 测速前跳过近期测速过的节点
			if !oknd.IsOk() {
				// 针对该域名，近期测速过的节点不存在或者测速已过期的节点才进行测速
				// fmt.Printf("-----TestAll--nodeTestUrl(%s)-----ProxyPool.testUrl(%s)-----\n", n.TestUrl, p.testUrl)
				wg.Add(1)
				ii := i
				nn := n
				go func(nnn *ProxyNode, iii int) {
					p.testOneNode(nnn, iii)
					wg.Done()
				}(&nn, ii)
			}
		}
	}
	wg.Wait()
	if runcount == 0 {
		p.IsLock = false
		fmt.Println("测速失败，没有可测速的代理节点。请先执行 --startproxynodes 命令，启动IP代理池")
		return
	}
	p.nodes.SortBySpeed()
	if p.activeCmd == nil {
		p.ActiveNode(p.nodes[0])
	}
	p.IsLock = false
}

// StartAll 启动所有已停止的节点。
func (p *ProxyPool) StartAll() error {
	var err error
	p.IsLock = true

	c := NewV2rayApiClientV5(p.getGrpcAddr())
	if c.Dial() == nil {
		defer c.Close()
	}
	for _, n := range p.nodes {
		if !n.IsRunning() {
			// hasPid, killStat := n.KillPidByLocalPort()
			// if hasPid > 0 {
			// 	fmt.Printf("----StartAll----Find--LocalPort(%d)---HasPID(%d)---\n", n.LocalPort, hasPid)
			// }
			// if killStat != nil {
			// 	panic(fmt.Errorf("---StartAll---killPidErr(%v)-----LocalPort(%d)----HasPID(%d)---", killStat, n.LocalPort, hasPid))
			// }
			// err = n.Start(p.v2rayPath)
			err = n.AddToPool(c)
			p.UpdateNode(n)
			if err != nil {
				logger := miniutils.GetLogger("")
				logger.Errorf("------StartAll--err--AddToPool(%v)", err)
				break
			}
		}
	}
	p.IsLock = false
	return err
}

//	func (p ProxyPool) getNodesMap() map[int]ProxyNode {
//		mp := make(map[int]ProxyNode, len(p.nodes))
//		for _, n := range p.nodes {
//			mp[n.LocalPort] = n
//		}
//		return mp
//	}
func (p ProxyPool) getGrpcAddr() string {
	return fmt.Sprintf("127.0.0.1:%d", conf.GetConf().V2rayApiPort)
}
func (p *ProxyPool) UpdateAfterStopAll() {
	c := NewV2rayApiClientV5(p.getGrpcAddr())
	if c.Dial() == nil {
		defer c.Close()
	}
	for i, nd := range p.nodes {
		if nd.IsRunning() {
			nd.Remove(c, getProxyNodeTag(nd.Index))
			p.nodes[i].status = 0
		}
	}
	for k, nds := range p.speedMap {
		for i, n := range nds {
			if n.IsRunning() {
				p.speedMap[k][i].status = 0
			}
		}
	}
}
func (p *ProxyPool) StopAll() error {
	var err error
	p.IsLock = true
	c := NewV2rayApiClientV5(p.getGrpcAddr())
	if c.Dial() == nil {
		defer c.Close()
	}
	for _, n := range p.nodes {
		if n.IsRunning() {
			err = n.Remove(c, "")
			if err != nil {
				break
			}
			p.UpdateNode(n)
		}
	}
	p.UpdateAfterStopAll()
	p.IsLock = false
	return err
}
func (p *ProxyPool) KillAllNodes() (total, runport, kill, fail int) {
	var err error
	p.IsLock = true
	c := NewV2rayApiClientV5(p.getGrpcAddr())
	if c.Dial() == nil {
		defer c.Close()
	}
	for _, nd := range p.nodes {
		err = nd.Remove(c, getProxyNodeTag(nd.Index))
		if err == nil {
			p.UpdateNode(nd)
			kill++
		} else {
			fail++
		}
	}

	p.UpdateAfterStopAll()
	p.IsLock = false
	for _, vs := range p.serverMap {
		vcmd := vs.GetExeCmd()
		if vcmd != nil {
			vcmd.Process.Kill()
		}
	}
	return
}
func (p *ProxyPool) UnActiveNode(n ProxyNode) error {
	// activePort := p.localPortStart - 1
	err := p.killActiveNode()
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		if err := SetProxy(""); err == nil {
			fmt.Println("取消代理成功!")
		} else {
			fmt.Printf("取消代理失败: %s\n", err)
		}
	}
	p.activeNode = ProxyNode{}
	return err
}

func (p *ProxyPool) killActiveNode() error {
	var err error
	if p.activeCmd != nil {
		err = p.activeCmd.Process.Kill()
		fmt.Printf("-----killActiveNode--AfterKill--err(%v)--p.activeCmd(%v)--PID(%d)--ProcessState(%+v)---\n", err, p.activeCmd, p.activeCmd.Process.Pid, p.activeCmd.ProcessState)
		if err != nil {
			return err
		}
		delete(p.serverMap, p.activeCmd.Process.Pid)
		p.activeCmd = nil
	}
	return nil
}

func (p *ProxyPool) ActiveNode(n ProxyNode) error {
	var err error
	activePort := p.localPortStart - 1
	err = p.killActiveNode()
	if err != nil {
		return err
	}
	vs := NewV2ray(p.v2rayPath)
	vs.SetPort(activePort).SetNode(n.v2rayNode)
	err = vs.Start("")
	if err == nil {
		p.activeCmd = vs.GetExeCmd()
		p.serverMap[vs.cmd.Process.Pid] = vs
		fmt.Printf("-----SUCCESS--ActiveNode--Index(%d)--LocalPort(%d)--Pid(%d)---RemoteAddr(%s)--ProcessState(%+v)---\n", n.Index, activePort, p.activeCmd.Process.Pid, n.RemoteAddr, p.activeCmd.ProcessState)
		if runtime.GOOS == "windows" {
			setPort := n.LocalPort
			if !n.IsRunning() {
				setPort = activePort
			}
			httproxy := fmt.Sprintf(`127.0.0.1:%d`, setPort)
			if err = SetProxy(httproxy); err == nil {
				fmt.Printf("设置代理服务器: %s 成功!\n", httproxy)
			} else {
				fmt.Printf("设置代理服务器: %s 失败, : %s\n", httproxy, err)
			}
		}
	} else {
		fmt.Printf("-----FAIL--ActiveNode--StartV2rayCoreFail---LocalPort(%d)---RemoteAddr(%s)---\n", activePort, n.RemoteAddr)
		return err
	}
	n.status = 1
	// n.LocalPort = activePort
	p.activeNode = n
	return err
}

// proxyPoolInit 初始化代理池
// 在schedule中更新订阅
// https://cloud.tencent.com/developer/article/1564128
func proxyPoolInit() {
	cf := conf.GetConf()
	subscribeRawData := cf.GetSubscribeData()
	if cf.SubscribeUrl == "" {
		// subscribeRawData == ""
		panic("subscribe url can not be empty.订阅地址不能为空")
	}
	port := cf.GetHttpProxyPort()
	// 程序退出后，还会存在端口占用
	// port := 11080

	pp := GetProxyPool()
	pp.SetV2rayPath(cf.V2rayPath).
		SetTestUrl(cf.TestUrl).
		SetSubscribeRawData(subscribeRawData).
		SetSubscribeUrl(cf.SubscribeUrl).
		SetLocalPortStart(port + 1).
		SetTestMaxDuration(MAX_TEST_DURATION).
		InitSubscribeData()
	nds := pp.GetNodes("")
	for i, n := range nds {
		fmt.Printf("---[%d]--Lport(%d)--Speed(%.3f)--Run(%v)--TestAt(%s)--Remote(%s)--T(%s)--index(%d)\n", i, n.LocalPort, n.Speed.Seconds(), n.IsRunning(), n.TestAt.Format("2006-01-02 15:04"), n.RemoteAddr, n.Title, n.Index)
	}
	pp.StartV2rayPool()
}
