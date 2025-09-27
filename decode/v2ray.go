package decode

import (
	"encoding/json"
)

// "protocol":"vmess"
type V2raySsNode struct {
	Protocol, Add, Host, Id, Net, Path, Ps, Tls, Type string
	V, Aid, Port                                      json.Number
}
