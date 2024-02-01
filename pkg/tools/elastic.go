package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func ParseToNDJson(data []interface{}, dst *bytes.Buffer) error {
	enc := json.NewEncoder(dst)
	for _, element := range data {
		if err := enc.Encode(element); err != nil {
			if err != io.EOF {
				return fmt.Errorf("failed to parse NDJSON: %v", err)
			}
			break
		}
	}
	return nil
}
