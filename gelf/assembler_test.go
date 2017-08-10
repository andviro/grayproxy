package gelf

import (
	"bytes"
	"testing"
	"time"
)

var AssemblerTestCases = [][]byte{
	{0x1e, 0x0f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x01, 0x03, 0x2}, //
	{0x1e, 0x0f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x00, 0x03, 0x1}, //
	{0x1e, 0x0f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x02, 0x03, 0x3}, // 1-3 correct
	{0x1e, 0x0f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x02, 0x04, 0x4}, // size too big
	{0x1e, 0x0f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x03, 0x03, 0x5}, // n > size
}

func TestAssembling(t *testing.T) {
	a := NewAssembler(1024, time.Second*1)
	for i := 0; i < 2; i++ {
		if ok, err := a.Update(AssemblerTestCases[i]); ok || err != nil {
			t.Fatalf("Unexpected assembly result: %v", err)
		}
	}
	if ok, err := a.Update(AssemblerTestCases[2]); !ok || err != nil {
		t.Fatalf("Assembly should complete")
	}
	if c := a.Bytes(); bytes.Compare(c, []byte{1, 2, 3}) != 0 {
		t.Fatalf("Invalid assembly: %v", c)
	}
}

func TestAssembleTimeout(t *testing.T) {
	if !testing.Verbose() || testing.Short() {
		t.Skip("Not testing timeout")
		return
	}
	a := NewAssembler(1024, time.Second*1)
	ok, err := a.Update(AssemblerTestCases[0])
	if ok || err != nil {
		t.Fatalf("Unexpected result updating: %v %v", ok, err)
	}
	time.Sleep(2 * time.Second)
	_, err = a.Update(AssemblerTestCases[1])
	if err == nil {
		t.Fatalf("Assembly should fail")
	}
}

func TestAssembleInvalid(t *testing.T) {
	a := NewAssembler(1024, time.Second*1)
	ok, err := a.Update(AssemblerTestCases[0])
	if ok || err != nil {
		t.Fatalf("Unexpected result updating: %v %v", ok, err)
	}
	ok, err = a.Update(AssemblerTestCases[3])
	if ok || err == nil {
		t.Fatalf("Should be error here: %v", err)
	}
	ok, err = a.Update(AssemblerTestCases[4])
	if ok || err == nil {
		t.Fatalf("Should be error here: %v", err)
	}
}
