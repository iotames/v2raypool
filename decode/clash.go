package decode

import (
	"gopkg.in/yaml.v3"
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

func ParseClashSubscribe(b []byte) []ClashProxyNode {
	config := ClashSubscribeSource{}
	err := yaml.Unmarshal(b, &config)
	if err != nil {
		panic(err)
	}
	return config.Proxies
}
