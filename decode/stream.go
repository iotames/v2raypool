package decode

var TransportProtocolList []string = []string{
	"tcp",
	"websocket",
	"grpc",
	// 	"mKCP"
	// "QUIC"
	// "meek"
	// "httpupgrade"
}

type StreamConfig struct {
	Protocol, Security, Path string
}

// func NewStreamConfig(proto string) *StreamConfig {
// 	if proto == "ws" {
// 		proto = "websocket"
// 	}
// 	return &StreamConfig{Protocol: proto}
// }

// {
// 	"transport":"tcp",
// 	"transportSettings":{},
// 	"security":"none",
// 	"securitySettings":{}
//   }
