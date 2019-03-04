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
		if ok := a.Update(AssemblerTestCases[i]); ok {
			t.Fatal("Should not assemble")
		}
	}
	if ok := a.Update(AssemblerTestCases[2]); !ok {
		t.Fatal("Assembly should complete")
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
	ok := a.Update(AssemblerTestCases[0])
	if ok {
		t.Fatal("Should not assemble")
	}
	time.Sleep(2 * time.Second)
	if !a.Expired() {
		t.Fatal("Assembler should expire")
	}
}

func TestAssembleInvalid(t *testing.T) {
	a := NewAssembler(1024, time.Second*1)
	ok := a.Update(AssemblerTestCases[0])
	if ok {
		t.Fatal("Should not assemble")
	}
	if ok = a.Update(AssemblerTestCases[3]); ok {
		t.Fatal("Should not assemble")
	}
	if ok = a.Update(AssemblerTestCases[4]); ok {
		t.Fatal("Should not assemble")
	}
}
