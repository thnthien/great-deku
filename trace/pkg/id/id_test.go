package id

import (
	"fmt"
	"testing"
)

func Test_defaultIDGenerator_NewSpanID(t *testing.T) {
	fmt.Printf("span %s \n", TraceGen.NewTraceID())
	fmt.Printf("span %s \n", TraceGen.NewSpanID())
}
