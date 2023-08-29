package v2raypool

import (
	"fmt"
	"os/exec"
	"sort"
	"time"

	"github.com/iotames/miniutils"
)

type ProxyNode struct {
	Index, LocalPort  int
	cmd               *exec.Cmd
	Id, localAddr     string
	RemoteAddr, Title string
	TestUrl           string
	Speed             time.Duration
	TestAt            time.Time
	v2rayNode         V2rayNode
	status            int
}

func NewProxyNodeByV2ray(vnd V2rayNode) *ProxyNode {
	remoteAddr := fmt.Sprintf("%s:%s", vnd.Add, vnd.Port)
	n := &ProxyNode{
		Id:         remoteAddr + ":" + vnd.Id,
		RemoteAddr: remoteAddr,
		Title:      vnd.Ps,
	}
	n.SetV2ray(vnd)
	return n
}

func (p *ProxyNode) GetId() string {
	if p.Id != "" {
		return p.Id
	}
	p.Id = p.RemoteAddr + ":" + p.v2rayNode.Id
	return p.Id
}
func (p *ProxyNode) SetV2ray(n V2rayNode) *ProxyNode {
	p.v2rayNode = n
	return p
}
func (p *ProxyNode) Start(path string) (err error) {
	v := NewV2ray(path, p.LocalPort)
	p.cmd, err = v.SetPort(p.LocalPort).SetNode(p.v2rayNode).Start()
	if p.cmd != nil {
		p.status = 1
	}
	fmt.Printf("\n---StartV2rayNode--i(%d)--Addr(%s)--Title(%s)--err(%v)---\n", p.Index, p.RemoteAddr, p.Title, err)
	return
}
func (p *ProxyNode) KillPidByLocalPort() (hasPid int, killResult error) {

	hasPid = miniutils.GetPidByPort(p.LocalPort)
	if hasPid > 0 {
		fmt.Printf("---proxyNode---KillPidByLocalPort(%d)---PID(%d)---\n", p.LocalPort, hasPid)
		killResult = miniutils.KillPid(fmt.Sprintf("%d", hasPid))
		return
	}
	return
}
func (p *ProxyNode) Stop() error {
	err := p.cmd.Process.Kill()
	if err == nil {
		p.status = 0
	}
	return err
}
func (p ProxyNode) IsRunning() bool {
	return p.status == 1
}
func (p ProxyNode) IsOk() bool {
	return time.Since(p.TestAt) < time.Hour*24
}

// // {"add":"jp6.v2u9.top","host":"","id":"0999AE93-1330-4A75-DBC1-0DD545F7DD60","net":"ws","path":"","port":"41444","ps":"u9un-v2-JP-Tokyo6(1)","tls":"","v":2,"aid":0,"type":"none"}
// protocol, add, port id, net

type ProxyNodes []ProxyNode

func (s ProxyNodes) Len() int           { return len(s) }
func (s ProxyNodes) Less(i, j int) bool { return s[i].Speed < s[j].Speed }
func (s ProxyNodes) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s *ProxyNodes) SortBySpeed() {
	sort.Sort(s)
}
