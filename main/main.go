package main

import (
	"flag"
	"time"

	vp "github.com/iotames/v2raypool"
)

func main() {
	var force, getproxynodes, killproxynodes, testproxynodes, startproxynodes, updateproxynodes, stopproxynodes bool
	var activeproxynode int
	var setproxytesturl, getproxynodesbydomain string
	flag.BoolVar(&force, "force", false, "do some optrate force")
	flag.StringVar(&setproxytesturl, "setproxytesturl", "", "Set testUrl of Proxy Nodes")
	flag.StringVar(&getproxynodesbydomain, "getproxynodesbydomain", "", "Get ProxyNodes By Domain")
	flag.IntVar(&activeproxynode, "activeproxynode", -1, "active proxy node by index")
	flag.BoolVar(&getproxynodes, "getproxynodes", false, "get proxypool nodes")
	flag.BoolVar(&killproxynodes, "killproxynodes", false, "kill all proxypool nodes")
	flag.BoolVar(&testproxynodes, "testproxynodes", false, "test speed of proxy nodes")
	flag.BoolVar(&startproxynodes, "startproxynodes", false, "start all proxy nodes")
	flag.BoolVar(&updateproxynodes, "updateproxynodes", false, "update subscribe all proxy nodes")
	flag.BoolVar(&stopproxynodes, "stopproxynodes", false, "stop all proxy nodes")
	flag.Parse()

	if setproxytesturl != "" {
		setProxyTestUrl(setproxytesturl)
		return
	}
	if getproxynodesbydomain != "" {
		getProxyNodesByDomain(getproxynodesbydomain)
		return
	}
	if getproxynodes {
		getProxyNodes()
		return
	}
	if killproxynodes {
		killProxyNodes()
		return
	}
	if activeproxynode > -1 {
		activeProxyNode(activeproxynode)
		return
	}
	if testproxynodes {
		testProxyNodes(force)
		return
	}
	if startproxynodes {
		startAllProxyNodes()
		return
	}
	if updateproxynodes {
		updateProxyNodes()
		return
	}
	if stopproxynodes {
		stopProxyNodes()
		return
	}
	go vp.RunProxyPoolInit()
	time.Sleep(3 * time.Second)
	vp.RunProxyPoolGrpcServer()
}

func init() {
	LoadEnv()
}
