package decode

import (
	"fmt"
	nurl "net/url"
	"strconv"
	"strings"
)

// "aes-256-gcm", "aes-128-gcm", "chacha20-poly1305", "chacha20-ietf-poly1305", "plain", "none"
type Shadowsocks struct {
	TransportStream                   StreamConfig
	Address, Password, Cipher, Plugin string
	Title                             string
	Port                              int
}

func ParseShadowsocks(v string) (ss Shadowsocks, err error) {
	pwdsplit := strings.Split(v, "@")
	pwdinfo := pwdsplit[0]
	var b1 string
	b1, err = Base64StdDecode(pwdinfo) // base64.StdEncoding.DecodeString(pwdinfo)
	if err != nil {
		err = fmt.Errorf("err(%v) for Base64StdDecode", err)
		return
	}
	pwdsplit2 := strings.Split(b1, ":")
	ss.Cipher = pwdsplit2[0]
	ss.Password = pwdsplit2[1]
	addrsplit := strings.Split(pwdsplit[1], `/?`)
	addrsplit2 := strings.Split(addrsplit[0], `:`)
	argspre, _ := nurl.QueryUnescape(addrsplit[1])
	args := strings.Split(argspre, `;`)
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			argsplit := strings.Split(arg, "=")
			argval := argsplit[1]
			// plugin=v2ray-plugin
			if argsplit[0] == "mode" {
				if argval == "websocket" {
					argval = "ws"
				}
				ss.TransportStream.Protocol = argval
			}
			if argsplit[0] == "path" {
				ss.TransportStream.Path = argval
			}
			// if argsplit[0] == "mux"{
			// 	// mux=true
			// }
		}
		if strings.Contains(arg, "#") {
			ss.Title = strings.Replace(arg, "#", "", 1) + "OKTEST"
		}
		if arg == "tls" {
			ss.TransportStream.Security = arg
			// nd.Tls = "tls"
		}
	}
	ss.Address = addrsplit2[0]
	ss.Port, err = strconv.Atoi(addrsplit2[1])
	return
}
