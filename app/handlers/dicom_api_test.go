package handlers

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tanema/dicom-microservice-api/app/storage"
	"github.com/tanema/dicom-microservice-api/config"
)

var (
	//go:embed testdata/example_eye.dcm
	testFile     []byte
	testFileHash = "ee1fcf71fecb6a8aee3b1219388cc7ded40a804056a7647674ba1a77b797ae8e"
)

func TestDICOMAPI(t *testing.T) {
	ts := httptest.NewServer(setupServer(t))
	defer ts.Close()

	// Create new doc
	res, err := http.Post(ts.URL, "", bytes.NewBuffer(testFile))
	require.NoError(t, err)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	doc := dicomDoc{}
	require.NoError(t, json.Unmarshal(body, &doc))
	assert.Equal(t, testFileHash, doc.Hash)

	// List the doc
	res, err = http.Get(fmt.Sprintf("%s/", ts.URL))
	require.NoError(t, err)
	body, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	list := docListResponse{}
	require.NoError(t, json.Unmarshal(body, &list))
	assert.Len(t, list.Docs, 1)

	// View all doc tags
	res, err = http.Get(fmt.Sprintf("%s/%s/info", ts.URL, testFileHash))
	require.NoError(t, err)
	body, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	alltags := map[string]any{}
	require.NoError(t, json.Unmarshal(body, &alltags))
	assert.Len(t, alltags, 53)

	// View single tag
	res, err = http.Get(fmt.Sprintf("%s/%s/info?tag=AccessionNumber", ts.URL, testFileHash))
	require.NoError(t, err)
	body, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	tags := map[string]any{}
	require.NoError(t, json.Unmarshal(body, &tags))
	assert.Len(t, tags, 1)
	assert.Equal(t, "[]", tags["AccessionNumber"])

	// View image
	res, err = http.Get(fmt.Sprintf("%s/%s?frame=0", ts.URL, testFileHash))
	require.NoError(t, err)
	body, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	assert.Equal(t, "image/jpeg", http.DetectContentType(body))
	assert.Len(t, body, 55016)
}

func setupServer(t *testing.T) http.Handler {
	cfg, err := config.Load("test")
	require.NoError(t, err)

	testStore, err := storage.New(cfg.Store)
	require.NoError(t, err)

	return NewDICOMAPI(cfg, testStore)
}
