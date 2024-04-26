package decode

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/iotames/v2raypool/netutil"
)

func ParseSubscribeByUrl(url string, proxy string) (dt string, rawdt string, err error) {
	if strings.Index(url, "http") != 0 {
		// panic("订阅源URL格式错误")
		err = fmt.Errorf("订阅源URL格式错误(%s)", url)
		return
	}
	c, r := netutil.GetHttpClient(10*time.Second, url, proxy)
	var resp *http.Response
	resp, err = c.Do(r)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var b []byte
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	rawdt = string(b)
	dt, err = ParseSubscribeByRaw(rawdt)
	return
}

func ParseSubscribeByRaw(data string) (dt string, err error) {
	fmt.Printf("---Begin---parseSubscribeByRaw---\n")
	dt, err = Base64StdDecode(data)
	if err != nil {
		fmt.Printf("---parseSubscribeByRaw--Use--Base64URLDecode--Warn(%v)---\n", err)
		dt, err = Base64URLDecode(data)
	}
	return
}

// 封装base64.StdEncoding进行解码，加入了长度补全，换行删除。当error时，返回输入和err
func Base64StdDecode(s string) (string, error) {
	s = strings.TrimSpace(s)
	saver := s
	s = strings.ReplaceAll(strings.ReplaceAll(s, "\n", ""), "\r", "")
	if len(s)%4 > 0 {
		s += strings.Repeat("=", 4-len(s)%4)
	}
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return saver, err
	}
	return string(raw), err
}

// 封装base64.URLEncoding进行解码，加入了长度补全，换行删除。当error时，返回输入和err
func Base64URLDecode(s string) (string, error) {
	s = strings.TrimSpace(s)
	saver := s
	s = strings.ReplaceAll(strings.ReplaceAll(s, "\n", ""), "\r", "")
	if len(s)%4 > 0 {
		s += strings.Repeat("=", 4-len(s)%4)
	}
	raw, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return saver, err
	}
	return string(raw), err
}
