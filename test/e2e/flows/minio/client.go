package minio

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"
)

type Client struct {
	endpoint string
	client   *http.Client
}

func New(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
		client:   &http.Client{},
	}
}

func (c *Client) UploadImage(t *testing.T, token string, filename string, content []byte) *http.Response {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(http.Header)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	h.Set("Content-Type", "image/jpeg")
	part, err := writer.CreatePart(textproto.MIMEHeader(h))
	if err != nil {
		t.Fatal(err)
	}
	part.Write(content)
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/files/upload/image", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func (c *Client) UploadImages(t *testing.T, token string, files map[string][]byte) *http.Response {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for filename, content := range files {
		h := make(http.Header)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="files"; filename="%s"`, filename))
		h.Set("Content-Type", "image/jpeg")
		part, err := writer.CreatePart(textproto.MIMEHeader(h))
		if err != nil {
			t.Fatal(err)
		}
		part.Write(content)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/files/upload/images", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func (c *Client) UploadDoc(t *testing.T, token string, filename string, content []byte) *http.Response {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(http.Header)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	h.Set("Content-Type", "application/pdf")
	part, err := writer.CreatePart(textproto.MIMEHeader(h))
	if err != nil {
		t.Fatal(err)
	}
	part.Write(content)
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/files/upload/doc", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func (c *Client) Download(t *testing.T, token string, filePath string) *http.Response {
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/files/download?file-path="+filePath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}
