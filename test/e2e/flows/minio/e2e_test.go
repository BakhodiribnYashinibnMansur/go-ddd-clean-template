package minio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMinio_ComprehensiveFlow(t *testing.T) {
	cleanDB(t)
	server := startTestServer()
	defer server.Close()

	mClient := New(server.URL)

	// Unique user for this test run
	ts := time.Now().UnixNano()
	username := fmt.Sprintf("minio_user_%d", ts)
	phone := strconv.FormatInt(ts%1000000000000, 10)
	password := "password123"

	// 1. Sign Up
	signupBody := fmt.Sprintf(`{"username":"%s","phone":"%s","password":"%s"}`, username, phone, password)
	resp, err := http.Post(server.URL+"/api/v1/users/sign-up", "application/json", strings.NewReader(signupBody))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// 2. Sign In
	signinBody := fmt.Sprintf(`{"phone":"%s","password":"%s"}`, phone, password)
	resp, err = http.Post(server.URL+"/api/v1/users/sign-in", "application/json", strings.NewReader(signinBody))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var loginData struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&loginData)
	require.NoError(t, err)
	token := loginData.Data.AccessToken
	resp.Body.Close()

	t.Run("Upload Image", func(t *testing.T) {
		// Generate real 1x1 GIF
		buf := new(bytes.Buffer)
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		gif.Encode(buf, img, nil)
		content := buf.Bytes()
		filename := "test_image.gif"

		// Upload
		upResp := mClient.UploadImage(t, token, filename, content)
		defer upResp.Body.Close()
		require.Equal(t, http.StatusOK, upResp.StatusCode)

		var upResult struct {
			Data string `json:"data"`
		}
		err = json.NewDecoder(upResp.Body).Decode(&upResult)
		require.NoError(t, err)
		require.NotEmpty(t, upResult.Data)
		require.Contains(t, upResult.Data, ".jpeg") // Use case re-encodes as jpeg
	})

	t.Run("Download Local File", func(t *testing.T) {
		// The current API implementation of /files/download serves local files
		// We create a temp file to verify the endpoint works
		content := []byte("local-file-content-123")
		tmpFile, err := os.CreateTemp(t.TempDir(), "e2e-test-*.txt")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.Write(content)
		require.NoError(t, err)
		tmpFile.Close()

		dlResp := mClient.Download(t, token, tmpFile.Name())
		defer dlResp.Body.Close()
		require.Equal(t, http.StatusOK, dlResp.StatusCode)

		dlContent, err := io.ReadAll(dlResp.Body)
		require.NoError(t, err)
		require.Equal(t, content, dlContent)
	})

	t.Run("Upload Multiple Images", func(t *testing.T) {
		// Generate real GIFs
		buf1 := new(bytes.Buffer)
		gif.Encode(buf1, image.NewRGBA(image.Rect(0, 0, 1, 1)), nil)
		buf2 := new(bytes.Buffer)
		gif.Encode(buf2, image.NewRGBA(image.Rect(0, 0, 1, 1)), nil)

		files := map[string][]byte{
			"img1.gif": buf1.Bytes(),
			"img2.gif": buf2.Bytes(),
		}

		upResp := mClient.UploadImages(t, token, files)
		defer upResp.Body.Close()
		require.Equal(t, http.StatusOK, upResp.StatusCode)
	})

	t.Run("Upload Doc", func(t *testing.T) {
		content := []byte("fake-pdf-content")
		filename := "test.pdf"

		upResp := mClient.UploadDoc(t, token, filename, content)
		defer upResp.Body.Close()
		require.Equal(t, http.StatusOK, upResp.StatusCode)
	})

	t.Run("Unauthorized Upload", func(t *testing.T) {
		upResp := mClient.UploadImage(t, "", "no_token.jpg", []byte("data"))
		defer upResp.Body.Close()
		require.Equal(t, http.StatusUnauthorized, upResp.StatusCode)
	})

	t.Run("Invalid File Extension", func(t *testing.T) {
		upResp := mClient.UploadImage(t, token, "danger.exe", []byte("malicious"))
		defer upResp.Body.Close()
		require.Equal(t, http.StatusBadRequest, upResp.StatusCode)
	})
}
