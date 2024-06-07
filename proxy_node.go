package v2raypool

import (
	"fmt"
	// "os/exec"
	"sort"
	"time"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

type ProxyNode struct {
	Index, LocalPort int
	Id, LocalAddr    string
	RemoteAddr       string `json:"remote_addr"`
	Title, Protocol  string
	TestUrl          string
	Speed            time.Duration
	TestAt           time.Time
	V2rayNode        V2rayNode
	Status           int
	IsDelete         bool
}

func NewProxyNodeByV2ray(vnd V2rayNode) *ProxyNode {
	n := &ProxyNode{}
	n.SetV2ray(vnd)
	return n
}

func (p *ProxyNode) GetId() string {
	if p.Id != "" {
		return p.Id
	}
	p.Id = p.RemoteAddr + ":" + p.V2rayNode.Id
	return p.Id
}
func (p *ProxyNode) SetV2ray(n V2rayNode) *ProxyNode {
	p.RemoteAddr = fmt.Sprintf("%s:%v", n.Add, n.Port)
	p.Id = p.RemoteAddr + ":" + n.Id
	p.Title = n.Ps
	p.Protocol = n.Protocol
	p.V2rayNode = n
	return p
}

func (p *ProxyNode) AddToPool(c *V2rayApiClient) error {
	tag := getProxyNodeTag(p.Index)
	cf := getConf()
	protcl := cf.GetHttpProxyProtocol()
	if protcl == "socks5" {
		protcl = "socks"
	}
	err := c.AddInbound(net.Port(p.LocalPort), tag, protcl)
	if err != nil {
		return err
	}
	err = c.AddOutboundByV2rayNode(p.V2rayNode, tag)
	if err != nil {
		return err
	}
	p.Status = 1
	return err
}

func (p *ProxyNode) Remove(c *V2rayApiClient, tag string) error {
	if tag == "" {
		panic("tag could not be empty")
	}
	err := c.RemoveOutbound(tag)
	if err != nil {
		return err
	}
	err = c.RemoveInbound(tag)
	if err != nil {
		return err
	}
	p.Status = 0
	return err
}

func (p ProxyNode) IsRunning() bool {
	return p.Status == 1
}

// IsOk 查看测速是否超过有效期。默认24小时
func (p ProxyNode) IsOk() bool {
	return time.Since(p.TestAt) < time.Hour*24
}

type ProxyNodes []ProxyNode

func (s ProxyNodes) Len() int           { return len(s) }
func (s ProxyNodes) Less(i, j int) bool { return s[i].Speed < s[j].Speed }
func (s ProxyNodes) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s *ProxyNodes) SortBySpeed() {
	sort.Sort(s)
}
