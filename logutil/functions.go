package logutil

import (
	"bytes"
	"encoding/json"
)

func Pretty(v any) string {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(v)
	//content, _ := json.MarshalIndent(v, "", "  ")
	return buffer.String()
}
