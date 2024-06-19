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

func ssrParse(u string) (ss Shadowsocks, uu *nurl.URL, err error) {
	uu, err = nurl.Parse(u)
	if err != nil {
		err = fmt.Errorf("invalid ss:// format")
		return
	}
	ss.Title = uu.Fragment
	ss.Address = uu.Hostname()
	ss.Port, err = strconv.Atoi(uu.Port())
	if err != nil {
		return
	}
	username := uu.User.String()
	username, _ = Base64URLDecode(username)
	pwdinfos := strings.SplitN(username, ":", 2)
	if len(pwdinfos) != 2 {
		err = fmt.Errorf("parse password err")
		return
	}
	ss.Cipher = pwdinfos[0]
	ss.Password = pwdinfos[1]
	// ss.Plugin = uu.Query().Get("plugin")
	ss.TransportStream.Path = uu.Query().Get("path")
	ss.TransportStream.Protocol = uu.Query().Get("mode")
	return
}

// ParseShadowsocks. parse shadowsocks protocol url string. begin with: ss://
func ParseShadowsocks(u string) (ss Shadowsocks, err error) {
	var uu *nurl.URL
	ss, uu, err = ssrParse(u)
	if err != nil {
		u = u[5:]
		var l, ps string
		if ind := strings.Index(u, "#"); ind > -1 {
			l = u[:ind]
			ps = u[ind+1:]
		} else {
			l = u
		}
		l, err = Base64StdDecode(l)
		if err != nil {
			l, err = Base64URLDecode(l)
			if err != nil {
				return
			}
		}
		u = "ss://" + l
		if ps != "" {
			u += "#" + ps
		}
		ss, uu, err = ssrParse(u)
		if err != nil {
			return
		}
	}

	if uu.RawQuery != "" {
		argspre, _ := nurl.QueryUnescape(uu.RawQuery)

		args := strings.Split(argspre, `;`)
		for _, arg := range args {
			if strings.Contains(arg, "=") {
				argsplit := strings.Split(arg, "=")
				argval := argsplit[1]
				// plugin=v2ray-plugin
				if argsplit[0] == "mode" {
					if ss.TransportStream.Protocol == "" {
						ss.TransportStream.Protocol = argval
					}
				}
				if argsplit[0] == "path" {
					if ss.TransportStream.Path == "" {
						ss.TransportStream.Path = argval
					}
				}
				// if argsplit[0] == "mux"{
				// 	// mux=true
				// }
			}
			if arg == "tls" {
				ss.TransportStream.Security = arg
			}
		}
	}
	if ss.TransportStream.Protocol == "" {
		ss.TransportStream.Protocol = "tcp"
	}
	if ss.TransportStream.Protocol == "websocket" {
		ss.TransportStream.Protocol = "ws"
	}
	return
}
