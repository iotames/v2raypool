package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strconv"
	"strings"

	vp "github.com/iotames/v2raypool"
	"github.com/v2fly/v2ray-core/v5/common/net"
)

var saddr string
var rmin string
var addin string
var addout string

func main() {
	flag.Parse()
	var err error
	c := vp.NewV2rayApiClientV5(saddr)
	err = c.Dial()
	if err == nil {
		defer c.Close()
	} else {
		panic(err)
	}
	if addin != "" {
		args := strings.Split(addin, ",")
		var inport int
		var intag string
		for _, arg := range args {
			argg := strings.Split(arg, "=")
			k := argg[0]
			v := argg[1]
			if k == "port" {
				inport, err = strconv.Atoi(v)
				if err != nil {
					panic(err)
				}
			}
			if k == "tag" {
				intag = v
			}
		}
		err = c.AddInbound(net.Port(inport), intag, "http")
		if err != nil {
			panic(err)
		}
	}

	if rmin != "" {
		err = c.RemoveInbound(rmin)
		if err != nil {
			panic(err)
		}
	}

	if addout != "" {
		args := strings.Split(addout, ",")
		var addr, port, nett, id, tls, outag string
		for _, arg := range args {
			argg := strings.Split(arg, "=")
			k := argg[0]
			v := argg[1]
			if k == "addr" {
				addr = v
			}
			if k == "port" {
				port = v
			}
			if k == "id" {
				id = v
			}
			if k == "lts" {
				tls = v
			}
			if k == "tag" {
				outag = v
			}
		}
		nett = "ws"
		fmt.Println("-----", addr, port, nett, id, tls, outag)
		nd := vp.V2rayNode{
			Protocol: "vmess",
			Add:      addr,
			Port:     json.Number(port),
			Net:      nett,
			Id:       id,
			Tls:      tls,
		}
		err = c.AddOutboundByV2rayNode(nd, outag)
		if err != nil {
			panic(err)
		}
	}

}

func init() {
	flag.StringVar(&saddr, "saddr", "", "the v2ray gRPC server address. like: 127.0.0.1:5053")
	flag.StringVar(&rmin, "rmin", "", "remove inbound by inboundTag")
	flag.StringVar(&addin, "addin", "", "add inbound. like: port=1089,tag=T1089")
	flag.StringVar(&addout, "addout", "", "add outbound. like: addr=hk1.abc.cd,port=88888,id=xxxx,tls=,tag=OUT10086")
}
