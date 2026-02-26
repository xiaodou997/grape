package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/graperegistry/grape/internal/config"
	"github.com/graperegistry/grape/internal/logger"
)

type Proxy struct {
	upstream string
	client   *http.Client
}

func NewProxy(cfg *config.RegistryConfig) *Proxy {
	return &Proxy{
		upstream: strings.TrimSuffix(cfg.Upstream, "/"),
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *Proxy) GetMetadata(packageName string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", p.upstream, packageName)
	logger.Debugf("Fetching metadata from upstream: %s", url)

	resp, err := p.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from upstream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrPackageNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

func (p *Proxy) GetTarball(packageName, filename string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s/-/%s", p.upstream, packageName, filename)
	logger.Debugf("Fetching tarball from upstream: %s", url)

	resp, err := p.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tarball from upstream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrTarballNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream returned status %d for tarball", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tarball body: %w", err)
	}

	return data, nil
}

func (p *Proxy) Upstream() string {
	return p.upstream
}

type PackageMetadata struct {
	ID          string                 `json:"_id"`
	Name        string                 `json:"name"`
	DistTags    map[string]string      `json:"dist-tags"`
	Versions    map[string]interface{} `json:"versions"`
	Time        map[string]string      `json:"time"`
	Users       map[string]string      `json:"users,omitempty"`
	Maintainers []interface{}          `json:"maintainers,omitempty"`
	Description string                 `json:"description,omitempty"`
	Keywords    []string               `json:"keywords,omitempty"`
	License     string                 `json:"license,omitempty"`
	Readme      string                 `json:"readme,omitempty"`
	README      string                 `json:"README,omitempty"`
}

func ParseMetadata(data []byte) (*PackageMetadata, error) {
	var meta PackageMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse package metadata: %w", err)
	}
	return &meta, nil
}
