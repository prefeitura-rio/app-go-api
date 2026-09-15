package empregabilidade_test

import (
	"bytes"
	"fmt"

	"github.com/lib/pq"
)

var (
	errTest   = fmt.Errorf("simulated error")
	validUUID = "550e8400-e29b-41d4-a716-446655440000"
)

func bodyOf(s string) *bytes.Buffer {
	return bytes.NewBufferString(s)
}

// bodyReader returns a reader that reports garbage JSON
//func garbageBody() *strings.Reader {
//	return strings.NewReader("not valid json")
//}

func uniqueViolationErr(constraint string) error {
	return &pq.Error{Code: pq.ErrorCode("23505"), Constraint: constraint}
}

func notNullViolationErr(column string) error {
	return &pq.Error{Code: pq.ErrorCode("23502"), Column: column}
}
