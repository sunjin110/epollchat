package jsonutil

import (
	"encoding/json"
	"epollchat/pkg/common/chk"
)

// Marshal marshal
func Marshal(v interface{}) string {
	b, err := json.Marshal(v)
	chk.SE(err, "json marshal err")
	return string(b)
}
