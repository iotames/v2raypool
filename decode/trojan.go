package decode

import (
	"fmt"
)

type Trojan struct {
	TransportStream           StreamConfig
	Address, Password, Cipher string
	Title                     string
	Port                      int
	Sni                       string `json:"sni"`
	AllowInsecure             bool   `json:"allowInsecure"`
	Alpn                      string `json:"alpn,omitempty"`
}

func ParseTrojan(v string) (ss Trojan, err error) {
	err = fmt.Errorf("protocol not support trojan://, TODO")
	return
}
