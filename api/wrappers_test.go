package api

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"testing"
)

type BufCloser bytes.Buffer

func (bc *BufCloser) Read(bs []byte) (int, error) {
	buf := (*bytes.Buffer)(bc)
	return buf.Read(bs)
}

func (bc *BufCloser) Close() error {
	return nil
}

func TestParseJsonWrapper(t *testing.T) {
	body := bytes.NewBuffer([]byte(`
		{ "operation" : "query"
		, "token" : "asdf"
		, "data" : 
			{ "foo" :
				{ "#model" : "Book"
				, "where" :
					{ "title" : "pants"
					}
				}
			}
		}
	`))

	req := &http.Request{
		Method: "POST",
		Header: map[string][]string{
			"Content-Type": []string{"application/json"},
		},
		Body: (*BufCloser)(body),
	}

	wrapper, er := UnwrapHttpRequest(req)
	if er != nil {
		t.Fatal(er)
	}
	defer wrapper.Close()
}

func TestParseMultipartWrapper(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	jsonData := `
		{ "operation" : "modify"
		, "token" : "asdf"
		, "data" : 
			[	{ "#model" : "Static"
				, "#primary" : "4"
				, "path " : "static1"
				}
			]
		}
	`

	static1Data := "gif86a                       "

	if er := writer.WriteField("data", jsonData); er != nil {
		t.Fatal(er)
	}

	if er := writer.WriteField("static1", static1Data); er != nil {
		t.Fatal(er)
	}

	if er := writer.Close(); er != nil {
		t.Fatal(er)
	}

	req := &http.Request{
		Method: "POST",
		Header: map[string][]string{
			"Content-Type": []string{writer.FormDataContentType()},
		},
		Body: (*BufCloser)(buf),
	}

	wrapper, er := UnwrapHttpRequest(req)
	if er != nil {
		t.Fatal(er)
	}
	defer wrapper.Close()
}
