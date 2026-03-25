package empregabilidade_test

import (
	"bytes"
	"fmt"
	"strings"
)

var (
	errTest  = fmt.Errorf("simulated error")
	validUUID = "550e8400-e29b-41d4-a716-446655440000"
)

func bodyOf(s string) *bytes.Buffer {
	return bytes.NewBufferString(s)
}

// bodyReader returns a reader that reports garbage JSON
func garbageBody() *strings.Reader {
	return strings.NewReader("not valid json")
}
