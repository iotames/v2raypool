package decode

import (
	"gopkg.in/yaml.v3"

	"encoding/json"
	"fmt"

	"strings"
)

const FILE_FORMAT_YAML = "yaml"

// { name: '🇭🇰 香港 02', type: trojan, server: h00kgbb2.star11.xyz, port: 60011, password: 23d7e3f1-c61c-4fbd-a04e-858b41248bb9, udp: true, sni: g.alicdn.com, skip-cert-verify: true }
type ClashProxyNode struct {
	Name           string `yaml:"name"`
	Type           string `yaml:"type"`
	Server         string `yaml:"server"`
	Port           int    `yaml:"port"`
	Password       string `yaml:"password"`
	Udp            bool   `yaml:"udp"`
	Sni            string `yaml:"sni"`
	SkipCertVerify bool   `yaml:"skip-cert-verify"`
}

type ClashSubscribeSource struct {
	Proxies []ClashProxyNode `yaml:"proxies"`
}

func ParseClashSubscribe(b []byte) []V2raySsNode {
	config := ClashSubscribeSource{}
	err := yaml.Unmarshal(b, &config)
	if err != nil {
		panic(err)
	}

	var vnds []V2raySsNode
	for _, clashNd := range config.Proxies {
		if clashNd.Type != "trojan" {
			fmt.Println("----ERROR------Clash Only Support:trojan")
			continue
		}
		nd := V2raySsNode{
			Protocol: clashNd.Type,
			Host:     clashNd.Sni,
			Add:      clashNd.Server,
		}
		nd.Port = json.Number(fmt.Sprintf("%d", clashNd.Port))
		nd.Ps = strings.TrimSpace(clashNd.Name)
		nd.Id = clashNd.Password
		nd.Net = "tcp"
		// if clashNd.Udp {
		// 	nd.Net = "udp"
		// }
		nd.Tls = "tls"
		// if clashNd.SkipCertVerify {
		// 	allowInsecure = true
		// }
		vnds = append(vnds, nd)
	}
	return vnds
}
