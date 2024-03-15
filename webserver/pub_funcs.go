package webserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/iotames/glayui/gtpl"
	"github.com/iotames/glayui/web"
)

func GetTpl() *gtpl.Gtpl {
	tpl := gtpl.GetTpl()
	tpl.SetResourceDirPath("resource")
	err := tpl.AddFunc("strContains", strings.Contains)
	if err != nil {
		panic(err)
	}
	err = tpl.AddFunc("tplinclude", TplInclude)
	if err != nil {
		panic(err)
	}
	return tpl
}

func TplInclude(fpath string, data any) string {
	var bf bytes.Buffer
	tpl := gtpl.GetTpl()
	tpl.SetDataByTplFile(fpath, data, &bf)
	return bf.String()
}

// func getPostJsonField(ctx web.Context, field string) (val any, err error) {
// 	dt := make(map[string]any)
// 	err = getPostJson(ctx, &dt)
// 	if err != nil {
// 		return
// 	}
// 	var ok bool
// 	val, ok = dt[field]
// 	if !ok {
// 		err = fmt.Errorf("post field %s not found", field)
// 		result := BaseResult{Msg: err.Error(), Code: 400}
// 		ctx.Writer.Write(result.Bytes())
// 	}
// 	return
// }

func getPostJson(ctx web.Context, v any) error {
	postdata, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		result := BaseResult{Msg: err.Error(), Code: 500}
		ctx.Writer.Write(result.Bytes())
		return err
	}
	fmt.Printf("--------postJsonData(%s)-------\n", string(postdata))
	err = json.Unmarshal(postdata, v)
	if err != nil {
		result := BaseResult{Msg: err.Error(), Code: 500}
		ctx.Writer.Write(result.Bytes())
		return err
	}
	return err
}
