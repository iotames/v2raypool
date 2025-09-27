package v2raypool

// Save nodes to file

import (
	"crypto/md5"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"

	"github.com/iotames/miniutils"
	"github.com/iotames/v2raypool/conf"
)

func getNodesByFile(data *ProxyNodes, subscribeUrl string) error {
	f, err := os.Open(getNodesFile(subscribeUrl))
	if err != nil {
		fmt.Printf("---getNodesByFile--os.Open--err(%v)\n", err)
		return err
	}
	defer f.Close()
	// decode := json.NewDecoder(f)
	decode := gob.NewDecoder(f)
	err = decode.Decode(data)
	if err != nil {
		return err
	}
	return err
}

func getNodesMapByFile(data *map[string]ProxyNodes, subscribeUrl string) error {
	f, err := os.Open(getNodesMapFile(subscribeUrl))
	if err != nil {
		fmt.Printf("---getNodesMapByFile--os.Open--err(%v)\n", err)
		return err
	}
	defer f.Close()
	// decode := json.NewDecoder(f)
	decode := gob.NewDecoder(f)
	err = decode.Decode(data)
	if err != nil {
		fmt.Printf("---getNodesMapByFile--decode.Decode--err(%v)---mp(%+v)\n", err, *data)
		return err
	}
	return err
}

func saveNodesToFile(nds ProxyNodes, fpath string) error {
	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	// encode := json.NewEncoder(f)
	encode := gob.NewEncoder(f)
	return encode.Encode(nds)
}

func saveNodesMapToFile(mp map[string]ProxyNodes, fpath string) error {
	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	// encode := json.NewEncoder(f)
	encode := gob.NewEncoder(f)
	return encode.Encode(mp)
}

func getNodesFile(subscribeUrl string) string {
	runtimeDir := conf.GetConf().RuntimeDir
	return filepath.Join(runtimeDir, fmt.Sprintf("nodes_%x.gob", md5.Sum([]byte(subscribeUrl))))
}

func getNodesMapFile(subscribeUrl string) string {
	runtimeDir := conf.GetConf().RuntimeDir
	return filepath.Join(runtimeDir, fmt.Sprintf("nodesmap_%x.gob", md5.Sum([]byte(subscribeUrl))))
}

func (p *ProxyPool) getNodesByFile() error {
	var err error

	if miniutils.IsPathExists(getNodesFile(p.subscribeUrl)) {
		var nds = ProxyNodes{}
		err = getNodesByFile(&nds, p.subscribeUrl)
		if err != nil {
			lg := getConf().GetLogger()
			lg.Errorf("---StartV2rayPool--getNodesByFile---err(%v)---nds(%+v)\n", err, nds)
		}

		if len(nds) > 0 {
			for _, nd := range nds {
				if nd.LocalAddr == "" {
					continue
				}
				nd.Status = 0
				p.UpdateNode(nd)
			}
		}
		p.nodes.SortBySpeed()
	}

	if miniutils.IsPathExists(getNodesMapFile(p.subscribeUrl)) {
		ndsmp := make(map[string]ProxyNodes)
		getNodesMapByFile(&ndsmp, p.subscribeUrl)
		if len(ndsmp) > 0 {
			for k, v := range ndsmp {
				for _, vv := range v {
					if vv.LocalAddr == "" {
						continue
					}
					vv.Status = 0
					p.AddSpeedNode(k, vv)
				}
			}
		}
	}
	return err
}
