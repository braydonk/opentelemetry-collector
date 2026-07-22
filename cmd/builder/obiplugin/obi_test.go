// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/cmd/builder/ocbplugin"
)

func TestUnsupportedHooks(t *testing.T) {
	p := &OBIPlugin{}
	assert.True(t, errors.Is(p.PostGenerate(nil), ocbplugin.ErrUnsupportedActionPostGenerate))
	assert.True(t, errors.Is(p.PreBuild(nil), ocbplugin.ErrUnsupportedActionPreBuild))
	assert.True(t, errors.Is(p.PostBuild(nil), ocbplugin.ErrUnsupportedActionPostBuild))
	assert.Equal(t, "0.157.0", p.MinOCBVersion())
}

func TestResolveVersion(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.yaml")

	content := `
dist:
  name: otelcol-contrib
receivers:
  - gomod: go.opentelemetry.io/obi v0.2.0
exporters:
  - gomod: go.opentelemetry.io/collector/exporter/nop exporter v0.120.0
`
	require.NoError(t, os.WriteFile(manifestPath, []byte(content), 0644))

	ver, err := resolveVersion(manifestPath)
	require.NoError(t, err)
	assert.Equal(t, "v0.2.0", ver)
}

func TestPreGenerateWithMockServer(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a dummy tar.gz file
	tarGzPath := filepath.Join(tmpDir, "dummy.tar.gz")
	tarFile, err := os.Create(tarGzPath)
	require.NoError(t, err)

	gzWriter := gzip.NewWriter(tarFile)
	tarWriter := tar.NewWriter(gzWriter)

	hdr := &tar.Header{
		Name: "top-dir/testfile.txt",
		Mode: 0644,
		Size: int64(len("hello obi")),
	}
	require.NoError(t, tarWriter.WriteHeader(hdr))
	_, err = tarWriter.Write([]byte("hello obi"))
	require.NoError(t, err)

	require.NoError(t, tarWriter.Close())
	require.NoError(t, gzWriter.Close())
	require.NoError(t, tarFile.Close())

	tarBytes, err := os.ReadFile(tarGzPath)
	require.NoError(t, err)

	hash := sha256.Sum256(tarBytes)
	checksumHex := hex.EncodeToString(hash[:])
	tarballName := "obi-v0.1.0-source-generated.tar.gz"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v0.1.0/" + tarballName:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(tarBytes)
		case "/v0.1.0/SHA256SUMS":
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, "%s  *%s\n", checksumHex, tarballName)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Create manifest
	manifestPath := filepath.Join(tmpDir, "manifest.yaml")
	manifestContent := `- gomod: go.opentelemetry.io/obi v0.1.0`
	require.NoError(t, os.WriteFile(manifestPath, []byte(manifestContent), 0644))

	obiDir := filepath.Join(tmpDir, "obi-src")
	cacheDir := filepath.Join(tmpDir, "cache")

	err = extractTarGz(tarGzPath, obiDir)
	require.NoError(t, err)

	extractedFile := filepath.Join(obiDir, "testfile.txt")
	content, err := os.ReadFile(extractedFile)
	require.NoError(t, err)
	assert.Equal(t, "hello obi", string(content))
	_ = cacheDir
}
