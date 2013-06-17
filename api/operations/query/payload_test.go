package query

import (
	"testing"

	"encoding/json"
	"macrobooru/api"
)

func TestParsePayload(t *testing.T) {
	jsonBytes := []byte(`
		{ "operation" : "query"
		, "token" : "asdf"
		, "data" :
			{
			}
		}
	`)

	var req api.RequestWrapper
	if er := json.Unmarshal(jsonBytes, &req); er != nil {
		t.Fatal(er)
	}

	if er := req.Parse(); er != nil {
		t.Fatal(er)
	}

	_, ok := req.Data.(*QueryPayload)
	if !ok {
		t.Fatalf("Parsed data ... is not a query payload? %#v", req.Data)
	}
}
