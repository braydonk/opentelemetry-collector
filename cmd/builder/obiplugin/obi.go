// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"go.opentelemetry.io/collector/cmd/builder/ocbplugin"
)

var _ ocbplugin.OCBPlugin = (*OBIPlugin)(nil)

// Config defines the configuration parameters for OBIPlugin.
type Config struct {
	Manifest   string `mapstructure:"manifest"`
	OBIDir     string `mapstructure:"obi_dir"`
	OBIVersion string `mapstructure:"obi_version"`
	CacheDir   string `mapstructure:"cache_dir"`
}

// OBIPlugin implements ocbplugin.OCBPlugin to prepare OBI source files.
type OBIPlugin struct{}

func (o *OBIPlugin) PreGenerate(config map[string]any) error {
	var cfg Config
	if err := mapstructure.Decode(config, &cfg); err != nil {
		return fmt.Errorf("failed to decode OBI plugin configuration: %w", err)
	}

	if cfg.Manifest == "" {
		cfg.Manifest = "./distributions/otelcol-contrib/manifest.yaml"
	}
	if cfg.OBIDir == "" {
		cfg.OBIDir = "./internal/obi-src"
	}
	if cfg.CacheDir == "" {
		cfg.CacheDir = "./.local"
	}

	version := cfg.OBIVersion
	if version == "" {
		version = os.Getenv("OBI_VERSION")
	}
	if version == "" {
		v, err := resolveVersion(cfg.Manifest)
		if err != nil {
			return err
		}
		version = v
	}

	stampFile := filepath.Join(cfg.OBIDir, ".obi-"+version)
	if _, err := os.Stat(stampFile); err == nil {
		fmt.Printf("OBI %s already prepared at %s\n", version, cfg.OBIDir)
		return nil
	}

	tarballName := fmt.Sprintf("obi-%s-source-generated.tar.gz", version)
	baseURL := fmt.Sprintf("https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/releases/download/%s", version)
	tarballURL := fmt.Sprintf("%s/%s", baseURL, tarballName)
	checksumsURL := fmt.Sprintf("%s/SHA256SUMS", baseURL)

	if err := os.MkdirAll(cfg.CacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory %s: %w", cfg.CacheDir, err)
	}

	tarballCachePath := filepath.Join(cfg.CacheDir, tarballName)
	if _, err := os.Stat(tarballCachePath); os.IsNotExist(err) {
		fmt.Printf("Fetching OBI %s source-generated tarball...\n", version)
		if err := downloadFile(tarballURL, tarballCachePath); err != nil {
			return fmt.Errorf("failed to download %s: %w", tarballURL, err)
		}
	}

	fmt.Printf("Verifying OBI %s tarball checksum...\n", version)
	expectedChecksum, err := fetchExpectedChecksum(checksumsURL, tarballName)
	if err != nil {
		os.Remove(tarballCachePath)
		return err
	}

	actualChecksum, err := calculateSHA256(tarballCachePath)
	if err != nil {
		os.Remove(tarballCachePath)
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	if actualChecksum != expectedChecksum {
		os.Remove(tarballCachePath)
		return fmt.Errorf("checksum verification failed for %s", tarballName)
	}
	fmt.Printf("SHA256 verified: %s\n", tarballName)

	if err := os.RemoveAll(cfg.OBIDir); err != nil {
		return fmt.Errorf("failed to clean OBI directory %s: %w", cfg.OBIDir, err)
	}
	if err := os.MkdirAll(cfg.OBIDir, 0755); err != nil {
		return fmt.Errorf("failed to create OBI directory %s: %w", cfg.OBIDir, err)
	}

	if err := extractTarGz(tarballCachePath, cfg.OBIDir); err != nil {
		return fmt.Errorf("failed to extract %s: %w", tarballCachePath, err)
	}

	if err := os.WriteFile(stampFile, []byte{}, 0644); err != nil {
		return fmt.Errorf("failed to write stamp file %s: %w", stampFile, err)
	}

	fmt.Printf("OBI %s source prepared at %s\n", version, cfg.OBIDir)
	return nil
}

func (o *OBIPlugin) PostGenerate(config map[string]any) error {
	return ocbplugin.ErrUnsupportedActionPostGenerate
}

func (o *OBIPlugin) PreBuild(config map[string]any) error {
	return ocbplugin.ErrUnsupportedActionPreBuild
}

func (o *OBIPlugin) PostBuild(config map[string]any) error {
	return ocbplugin.ErrUnsupportedActionPostBuild
}

func (o *OBIPlugin) MinOCBVersion() string {
	return "0.157.0"
}

func resolveVersion(manifestPath string) (string, error) {
	file, err := os.Open(manifestPath)
	if err != nil {
		return "", fmt.Errorf("failed to open manifest at %s: %w", manifestPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "- gomod: go.opentelemetry.io/obi") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				return fields[len(fields)-1], nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading manifest %s: %w", manifestPath, err)
	}
	return "", fmt.Errorf("failed to resolve OBI version from %s", manifestPath)
}

func downloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d downloading %s", resp.StatusCode, url)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func fetchExpectedChecksum(checksumsURL, filename string) (string, error) {
	resp, err := http.Get(checksumsURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch checksums from %s: %w", checksumsURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch checksums: HTTP %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			checksum := fields[0]
			checksumFilename := strings.TrimPrefix(fields[1], "*")
			if checksumFilename == filename {
				return checksum, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading checksums: %w", err)
	}

	return "", fmt.Errorf("%s not found in SHA256SUMS", filename)
}

func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func extractTarGz(tarGzPath, destDir string) error {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		cleanPath := filepath.Clean(header.Name)
		parts := strings.Split(cleanPath, string(filepath.Separator))
		if len(parts) <= 1 {
			continue
		}

		subPath := filepath.Join(parts[1:]...)
		targetPath := filepath.Join(destDir, subPath)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			mode := header.Mode
			if mode == 0 {
				mode = 0644
			}
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		case tar.TypeSymlink:
			_ = os.Symlink(header.Linkname, targetPath)
		}
	}

	return nil
}
