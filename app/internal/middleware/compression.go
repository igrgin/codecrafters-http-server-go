package middleware

import (
	"bytes"
	"compress/gzip"
	"strings"
)

func GzipBytes(src []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write(src)
	if err != nil {
		_ = gw.Close()
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ShouldCompress(acceptedEnc []string) bool {
	for _, part := range acceptedEnc {
		if strings.EqualFold(strings.TrimSpace(part), "gzip") {
			return true
		}
	}
	return false
}
