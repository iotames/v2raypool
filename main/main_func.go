package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	vp "github.com/iotames/v2raypool"
)

func runServer() {
	logStart()
	vp.RunServer()
}

func logStart() {
	ntime := time.Now()
	f, err := os.OpenFile("startat.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}
	_, err = f.WriteString(ntime.Format(time.RFC3339)) // "2006-01-02T15:04:05Z07:00"
	if err != nil {
		panic(err)
	}
	lgpath := filepath.Join(vconf.RuntimeDir, "logs", "start.log")
	f, err = os.OpenFile(lgpath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		panic(err)
	}
	timestr := ntime.Format("[2006-01-02 15:04:05]")
	logmsg := fmt.Sprintf("\n%s: envFile(%s), MainGrpcPort(%d), V2rayApiPort(%d)", timestr, vconf.EnvFile, vconf.GrpcPort, vconf.V2rayApiPort)
	_, err = f.WriteString(logmsg)
	if err != nil {
		panic(err)
	}
}
