package gwf

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSaveUploadedFile(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("ip", "127.0.0.1")
	_ = writer.WriteField("name", "babytree")
	err := writer.Close()
	if err != nil {
		t.Fatalf("Multipart writer close err: %v", err)
	}
	ts := createPostServer(t)
	req, _ := http.NewRequest("POST", ts.URL+"/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Response err: %v", err)
	}
	defer resp.Body.Close()
	respBody := &bytes.Buffer{}
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		t.Fatalf("Response body read err: %v", err)
	}
	expectBody := "ip=127.0.0.1,name=babytree"
	if respBody.String() != expectBody {
		t.Errorf("Expect %s, but got %s", expectBody, respBody.String())
	}
}

func createPostServer(t *testing.T) *httptest.Server {
	ts := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if r.URL.Path == "/upload" {
				multipartReader, err := r.MultipartReader()
				if err != nil {
					_, _ = w.Write([]byte("This is not multipart."))
					return
				}
				var readBody string
				for {
					b := make([]byte, 100)
					part, err := multipartReader.NextPart()
					if err == io.EOF {
						break
					}
					_, _ = part.Read(b)
					readBody += fmt.Sprintf("%s=%s,", part.FormName(), string(b))
				}
				writeBody := strings.ReplaceAll(readBody, "\x00", "")
				writeBody = strings.TrimRight(writeBody, ",")
				_, _ = w.Write([]byte(writeBody))
			}
		}
	})
	return ts
}
