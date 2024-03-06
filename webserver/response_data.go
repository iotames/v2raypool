package webserver

import "encoding/json"

type BaseResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (b *BaseResult) Success(msg string) {
	b.Code = 200
	if msg == "" {
		msg = "success"
	}
	b.Msg = msg
}

func (b *BaseResult) Fail(msg string, code int) {
	b.Code = code
	b.Msg = msg
}
func (b BaseResult) String() string {
	by, err := json.Marshal(b)
	if err != nil {
		panic(by)
	}
	return string(by)
}
func (b BaseResult) Bytes() []byte {
	by, err := json.Marshal(b)
	if err != nil {
		panic(by)
	}
	return by
}

type ListData struct {
	BaseResult
	Data  any `json:"data"`
	Count int `json:"count"`
}

func NewListData(data any, total int) *ListData {
	result := &ListData{
		Count: total,
		Data:  data,
	}
	result.Success("")
	return result
}
