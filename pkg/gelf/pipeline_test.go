package gelf

import (
	"encoding/base64"
	"reflect"
	"testing"
	"time"
)

var PipelineTestInputs = []string{
	"H4sIABrBhVgAAzMBADgbtvMBAAAA", // '4' | gzip
	"eJwzBQAANgA2",                 // '5' | zlib
	"Hg8AAAAAAAAAAQEDAg==",         // 1e 0f 00 00 00 00 00 00 00 01 01 03 02
	"Hg8AAAAAAAAAAQADAQ==",         // 1e 0f 00 00 00 00 00 00 00 01 00 03 01
	"Hg8AAAAAAAAAAQIDAw==",         // 1e 0f 00 00 00 00 00 00 00 01 02 03 03
	"MQ==",                         // '1'
	"Mg==",                         // '2'
	"Mw==",                         // '3'
}

var PipelineTestOutputs = [][]byte{
	{52},
	{53},
	{1, 2, 3},
	{49},
	{50},
	{51},
}

func TestPipeline(t *testing.T) {
	chunks := make(chan Chunk)
	go func() {
		for _, testChunk := range PipelineTestInputs {
			data, _ := base64.StdEncoding.DecodeString(testChunk)
			chunks <- data
		}
		close(chunks)
	}()
	decodedMsgs := Extract(Assemble(chunks, 1024, time.Second*2), 1024)

	result := make([][]byte, 0)
	for msg := range decodedMsgs {
		result = append(result, msg)
	}
	if !reflect.DeepEqual(result, PipelineTestOutputs) {
		t.Fatalf("Expected %v got %v", PipelineTestOutputs, result)
	}
}
