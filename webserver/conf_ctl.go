package webserver

import (
	"strconv"

	"github.com/iotames/v2raypool/conf"
)

func UpdateConf(dt map[string]string, fpath string) []byte {
	err := conf.UpdateConf(dt, fpath)
	result := BaseResult{}
	if err != nil {
		result.Fail("更新失败:"+err.Error(), 500)
		return result.Bytes()
	}
	// fmt.Printf("-----cf(%+v)---\n", dt)
	cf := conf.GetConf()
	cf.V2rayApiPort, err = strconv.Atoi(dt["VP_V2RAY_API_PORT"])
	if err != nil {
		result.Fail("VP_V2RAY_API_PORT 更新失败:"+err.Error(), 400)
		return result.Bytes()
	}
	cf.WebServerPort, err = strconv.Atoi(dt["VP_WEB_SERVER_PORT"])
	if err != nil {
		result.Fail("VP_WEB_SERVER_PORT 更新失败:"+err.Error(), 400)
		return result.Bytes()
	}
	cf.GrpcPort, err = strconv.Atoi(dt["VP_GRPC_PORT"])
	if err != nil {
		result.Fail("VP_GRPC_PORT 更新失败:"+err.Error(), 400)
		return result.Bytes()
	}
	cf.TestUrl = dt["VP_TEST_URL"]
	cf.SubscribeUrl = dt["VP_SUBSCRIBE_URL"]
	cf.SubscribeDataFile = dt["VP_SUBSCRIBE_DATA_FILE"]
	cf.V2rayPath = dt["VP_V2RAY_PATH"]
	cf.HttpProxy = dt["VP_HTTP_PROXY"]
	conf.SetConf(cf)
	result.Success("设置成功，重启应用后生效。")
	return result.Bytes()
}
