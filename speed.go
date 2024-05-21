package v2raypool

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/iotames/v2raypool/conf"
	"github.com/iotames/v2raypool/netutil"
)

func testProxyNode(testUrl string, localAddr string, index int, maxDuration time.Duration) (speed time.Duration, ok bool) {
	logger := conf.GetConf().GetLogger()
	c, r := netutil.GetHttpClient(maxDuration, testUrl, localAddr)
	speed, ok = requestNode(c, r, maxDuration, index)
	oktag := "FAIL"
	if ok {
		oktag = "SUCCESS"
	}
	logger.Debugf("----%s---NodeSpeedTest[%d]---Local(%s)---Speed(%s)--", oktag, index, localAddr, speed)
	return
}

func requestNode(c *http.Client, r *http.Request, maxDuration time.Duration, i int) (speed time.Duration, ok bool) {
	logger := conf.GetConf().GetLogger()
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
		logger.Debugf("---SpeedTestError(%d)Error(%s)---statusCode(%d)--cost(%+v)---", i, err, statusCode, costTime)
		return
	}

	if statusCode == http.StatusOK {
		speed = costTime
		ok = true
	}
	return
}
