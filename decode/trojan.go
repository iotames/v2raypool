package decode

import (
	"encoding/json"
	"fmt"
	nurl "net/url"
	"strconv"

	"strings"
)

type Trojan struct {
	TransportStream   StreamConfig
	Address, Password string
	Title             string
	Port              int
	Sni               string `json:"sni"`
	AllowInsecure     bool   `json:"allowInsecure"`
	Alpn              string `json:"alpn,omitempty"`
}

func TrojanToV2ray(tro Trojan, nd *V2raySsNode) {
	// alpn=http/1.1
	// sni=trojan.burgerip.co.uk
	nd.Host = tro.Sni
	nd.Add = tro.Address
	nd.Port = json.Number(fmt.Sprintf("%d", tro.Port))
	nd.Id = tro.Password
	nd.Net = tro.TransportStream.Protocol // type=tcp
	nd.Tls = tro.TransportStream.Security
	nd.Path = tro.TransportStream.Path
	nd.Ps = strings.TrimSpace(tro.Title)
}

// ParseTrojanUrl 解析Url格式的节点数据
func ParseTrojanUrl(d string) (tr Trojan, err error) {
	var t *nurl.URL
	t, err = nurl.Parse(d)
	if err != nil {
		err = fmt.Errorf("invalid trojan format")
		return
	}
	allowInsecure := t.Query().Get("allowInsecure")
	sni := t.Query().Get("peer")
	if sni == "" {
		sni = t.Query().Get("sni")
	}
	if sni == "" {
		sni = t.Hostname()
	}
	tr = Trojan{
		Title:         t.Fragment,
		Address:       t.Hostname(),
		Password:      t.User.Username(),
		Sni:           sni,
		Alpn:          t.Query().Get("alpn"),
		AllowInsecure: allowInsecure == "1" || allowInsecure == "true",
	}
	tr.Port, err = strconv.Atoi(t.Port())
	if err != nil {
		return
	}
	tr.TransportStream.Security = t.Query().Get("security")
	tr.TransportStream.Protocol = t.Query().Get("type")
	tr.TransportStream.Path = t.Query().Get("serviceName")
	if tr.TransportStream.Protocol == "" {
		tr.TransportStream.Protocol = "tcp"
	}
	return
}
