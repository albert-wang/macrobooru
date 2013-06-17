package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/textproto"
	"os"
	"strings"

	"mime"
	"mime/multipart"
)

type RequestWrapper struct {
	Operation string `json:"operation"`
	AuthToken string `json:"token,omitempty"`

	RawData json.RawMessage `json:"data"`
	Data    Operation       `json:"-"`

	/* XXX: This has to be reconstructed from a map[string][]multipart.File, which by
	 * default is constructed by peeking at the Content-Disposition header. We need to
	 * also peek at the Content-ID of each part if the Content-Disposition doesn't have
	 * a 'name' parameter. 
	 */
	Attachments map[string]multipart.File `json:"-"`
	MimeTypes   map[string]string         `json:"-"`
}

type ResponseWrapper struct {
	StatusCode    int64           `json:"statusCode"`
	StatusMessage string          `json:"statusMsg"`
	Data          json.RawMessage `json:"data,omitempty"`
}

func (wrapper *RequestWrapper) Close() {
	for _, file := range wrapper.Attachments {
		if osfile, ok := file.(*os.File); ok {
			osfile.Close()
			os.Remove(osfile.Name())
		}
	}

	wrapper.Attachments = nil
}

func (wrapper *RequestWrapper) MarshalJSON() ([]byte, error) {
	obj := map[string]interface{}{
		"operation": wrapper.Operation,
		"data":      wrapper.Data,
	}

	if wrapper.AuthToken != "" {
		obj["token"] = wrapper.AuthToken
	}

	return json.Marshal(obj)
}

func UnwrapHttpRequest(r *http.Request) (req *RequestWrapper, er error) {
	contentType := r.Header.Get("Content-Type")
	defer r.Body.Close()

	if strings.Index(contentType, "multipart/") == 0 {
		req, er = unwrapMultipartHttpRequest(r)

	} else {
		req, er = unwrapJsonHttpRequest(r)
	}

	if er != nil {
		return nil, er
	}

	if er := req.Parse(); er != nil {
		return nil, er
	}

	return req, nil
}

func UnwrapHttpResponse(r io.Reader) (*ResponseWrapper, error) {
	responseBytes, er := ioutil.ReadAll(r)
	if er != nil {
		return nil, er
	}

	wrapper := ResponseWrapper{}

	if er := json.Unmarshal(responseBytes, &wrapper); er != nil {
		log.Printf("Response contents:\n%s\n", string(responseBytes))
		return nil, er
	}

	return &wrapper, nil
}

func WrapHttpRequest(endpoint string, req *RequestWrapper) (*http.Request, error) {
	/* XXX: THIS WORKS COMPLETELY IN-MEMORY BECAUSE mime/multipart.Writer IS STUPID */
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	/* Generate the Attachments body */
	for name, attachment := range req.Attachments {
		if _, er := attachment.Seek(0, 0); er != nil {
			return nil, er
		}

		mime := req.MimeTypes[name]

		header := textproto.MIMEHeader{}
		header.Set("Content-Type", mime)
		header.Set("Content-ID", name)

		partWriter, er := writer.CreatePart(header)
		if er != nil {
			return nil, er
		}

		if _, er := io.Copy(partWriter, attachment); er != nil {
			return nil, er
		}
	}

	/* Serialize data payload to JSON and jam in into the request data */
	data, er := json.Marshal(req)
	if er != nil {
		return nil, er
	}

	header := textproto.MIMEHeader{}
	header.Set("Content-Type", "application/json")
	header.Set("Content-ID", "data")

	partWriter, er := writer.CreatePart(header)
	if er != nil {
		return nil, er
	}

	if _, er := io.Copy(partWriter, bytes.NewBuffer(data)); er != nil {
		return nil, er
	}

	if er := writer.Close(); er != nil {
		return nil, er
	}

	/* Package up a http request */
	httpReq, er := http.NewRequest("POST", endpoint, buf)
	if er != nil {
		return nil, er
	}

	contentType := fmt.Sprintf("multipart/mixed;boundary=\"%s\"", writer.Boundary())
	httpReq.Header.Set("Content-Type", contentType)
	httpReq.ContentLength = int64(buf.Len())

	return httpReq, nil
}

func unwrapJsonHttpRequest(r *http.Request) (*RequestWrapper, error) {
	jsonBytes, er := ioutil.ReadAll(r.Body)
	if er != nil {
		return nil, er
	}

	wrapper := RequestWrapper{}

	if er := json.Unmarshal(jsonBytes, &wrapper); er != nil {
		log.Printf("Unmarshal error\n%s\n", string(jsonBytes))
		return nil, er
	}

	wrapper.Attachments = make(map[string]multipart.File)

	return &wrapper, nil
}

func unwrapMultipartHttpRequest(r *http.Request) (*RequestWrapper, error) {
	contentType := r.Header.Get("Content-Type")
	_, params, er := mime.ParseMediaType(contentType)
	if er != nil {
		return nil, er
	}

	boundary, ok := params["boundary"]
	if !ok {
		return nil, fmt.Errorf("No boundary parameter in multipart request")
	}

	reader := multipart.NewReader(r.Body, boundary)

	attachments := map[string]multipart.File{}
	mimeTypes := map[string]string{}

	/* We're returning open file handles here, so make sure to kill them 
	 * all off if we happen to return an error prematurely. */
	filesToDestroy := []*os.File{}
	defer func() {
		for _, f := range filesToDestroy {
			f.Close()
			os.Remove(f.Name())
		}
	}()

	var jsonBytes []byte = nil

	for {
		part, er := reader.NextPart()

		if er != nil {
			if er == io.EOF {
				break
			}

			return nil, er
		}

		contentId := part.Header.Get("Content-ID")
		contentType := part.Header.Get("Content-Type")

		if contentId == "" {
			contentId = part.FormName()
		}

		if contentId == "" {
			contentId = part.FileName()
		}

		if contentType == "" {
			contentType = "application/octect-stream"
		}

		if contentId == "" {
			/* XXX: Potentially emit an error here */
			continue
		}

		/* Shim to pull out the data segment to avoid another file
		 * allocation */
		if contentId == "data" {
			if jsonBytes, er = ioutil.ReadAll(part); er != nil {
				return nil, er
			}

			continue
		}

		/* Dump the rest of the data into separate files rather than keeping
		 * them in-memory. This is a bit slow, but ideally the server is configured
		 * to use a swap-backed tmpfs. We'll be discarding these files in either case.
		 */
		tmpFile, er := ioutil.TempFile("", "macrobooru-")
		if er != nil {
			return nil, er
		}
		filesToDestroy = append(filesToDestroy, tmpFile)

		if _, er := io.Copy(tmpFile, part); er != nil {
			return nil, er
		}

		if _, er := tmpFile.Seek(0, os.SEEK_SET); er != nil {
			return nil, er
		}

		attachments[contentId] = tmpFile
		mimeTypes[contentId] = contentType
	}

	if jsonBytes == nil {
		return nil, fmt.Errorf("Missing 'data' section")
	}

	wrapper := RequestWrapper{}

	if er := json.Unmarshal(jsonBytes, &wrapper); er != nil {
		log.Printf("Unmarshal error\n%s\n", string(jsonBytes))
		return nil, er
	}

	wrapper.Attachments = attachments
	wrapper.MimeTypes = mimeTypes

	filesToDestroy = nil
	return &wrapper, nil
}
