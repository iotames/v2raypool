package v2raypool

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/iotames/miniutils"
)

func getHttpClient(maxDuration time.Duration, requestUrl string, proxyAddr string) (c *http.Client, r *http.Request) {
	var err error
	// Go语言实现关闭http请求的方式总结 https://www.jb51.net/article/276446.htm

	// trans := http.DefaultTransport //&http.Transport{}
	httpTrans := (&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}).Clone() // trans.(*http.Transport)
	httpTrans.DisableKeepAlives = true // 解决 Get "https://www.google.com": EOF
	if proxyAddr != "" {
		var proxy *url.URL
		proxy, err = url.Parse(proxyAddr)
		if err != nil {
			panic(err)
		}
		httpTrans.Proxy = func(r *http.Request) (*url.URL, error) {
			// fmt.Printf("---SET PROXY[%d](%s)---\n", i, s.ItemKey)
			return proxy, nil
		}
	}
	c = &http.Client{Transport: httpTrans, Timeout: maxDuration}
	r, err = http.NewRequest("GET", requestUrl, http.NoBody)
	r.Header.Add("Connection", "close")
	if err != nil {
		panic(err)
	}
	r.Close = true // 解决 Get "https://www.google.com": EOF
	r.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36")
	return c, r
}
func testProxyNode(testUrl string, localAddr string, index int, maxDuration time.Duration) (speed time.Duration, ok bool) {
	logger := miniutils.GetLogger("")
	c, r := getHttpClient(maxDuration, testUrl, localAddr)
	speed, ok = requestNode(c, r, maxDuration, index)
	if ok {
		logger.Debug(fmt.Sprintf("----SUCCESS---NodeSpeedTest[%d]---Local(%s)---Speed(%s)--", index, localAddr, speed))
	}
	return
}

func requestNode(c *http.Client, r *http.Request, maxDuration time.Duration, i int) (speed time.Duration, ok bool) {
	logger := miniutils.GetLogger("")
	ok = false
	speed = maxDuration
	start := time.Now()
	resp, err := c.Do(r)
	costTime := time.Since(start)
	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
		// 解决 Get "https://www.google.com": EOF
		defer resp.Body.Close() // 养成习惯随手关闭Body
	}
	if err != nil {
		if strings.Contains(err.Error(), io.EOF.Error()) {
			speed = costTime
		}
		logger.Debug(fmt.Sprintf("---SpeedTestError(%d)Error(%s)---statusCode(%d)--cost(%+v)---", i, err, statusCode, costTime))
		return
	}

	if statusCode == http.StatusOK {
		speed = costTime
		ok = true
	}
	return
}
