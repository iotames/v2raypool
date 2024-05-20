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

// ss://Y2hhY2hhMjAtaWV0Zi1wb2x5MTMwNTozNlpDSGVhYlVTZktqZlFFdko0SERW@185.242.86.156:54170#%E4%BF%84%E7%BD%97%E6%96%AF+V2CROSS.COM
func ParseShadowsocks(u string) (ss Shadowsocks, err error) {
	// var t *nurl.URL
	// t, err = nurl.Parse(u)
	// if err != nil {
	// 	err = fmt.Errorf("invalid ss:// format")
	// 	return
	// }
	// ss.Title = t.Fragment
	// ss.Address = t.Hostname()
	// ss.Port, err = strconv.Atoi(t.Port())
	// if err != nil {
	// 	return
	// }
	// username := t.User.String()
	// username, _ = Base64URLDecode(username)
	// pwdinfos := strings.SplitN(username, ":", 2)
	// if len(pwdinfos) != 2 {
	// 	err = fmt.Errorf("parse password err")
	// 	return
	// }
	// ss.Cipher = pwdinfos[0]
	// ss.Password = pwdinfos[1]

	ninfo := strings.Split(u, "://")
	if len(ninfo) != 2 {
		err = fmt.Errorf("split :// err")
		return
	}
	v := ninfo[1]

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
	if strings.Contains(pwdsplit[1], "/?") {
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
				ss.Title = strings.Replace(arg, "#", "", 1)
			}
			if arg == "tls" {
				ss.TransportStream.Security = arg
				// nd.Tls = "tls"
			}
		}
		ss.Address = addrsplit2[0]
		ss.Port, err = strconv.Atoi(addrsplit2[1])
	}
	return
}
