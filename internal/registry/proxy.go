package registry

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/graperegistry/grape/internal/config"
	"github.com/graperegistry/grape/internal/logger"
)

const (
	maxMetadataSize = 50 * 1024 * 1024 // 50MB - 元数据最大大小（大型包如 vite/typescript 可能很大）
	maxTarballSize  = 500 * 1024 * 1024 // 500MB - tarball 最大大小
)

// Upstream 上游配置
type Upstream struct {
	Name    string
	URL     string
	Scope   string
	Timeout time.Duration
	Enabled bool
	client  *http.Client
}

// Proxy 多上游代理
type Proxy struct {
	upstreams []*Upstream
	defaultUp *Upstream
	scopeMap  map[string]*Upstream // scope -> upstream
	mu        sync.RWMutex
}

func NewProxy(cfg *config.RegistryConfig) *Proxy {
	p := &Proxy{
		upstreams: make([]*Upstream, 0),
		scopeMap:  make(map[string]*Upstream),
	}

	// 处理多上游配置
	for _, uc := range cfg.Upstreams {
		if !uc.Enabled {
			continue
		}

		timeout := uc.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}

		up := &Upstream{
			Name:    uc.Name,
			URL:     strings.TrimSuffix(uc.URL, "/"),
			Scope:   uc.Scope,
			Timeout: timeout,
			Enabled: uc.Enabled,
			client: &http.Client{
				Timeout: timeout,
			},
		}

		p.upstreams = append(p.upstreams, up)

		// 建立 scope 映射
		if uc.Scope != "" {
			p.scopeMap[uc.Scope] = up
		} else {
			p.defaultUp = up
		}
	}

	// 向后兼容：如果没有配置多上游，使用单一上游
	if len(p.upstreams) == 0 && cfg.Upstream != "" {
		up := &Upstream{
			Name:    "default",
			URL:     strings.TrimSuffix(cfg.Upstream, "/"),
			Scope:   "",
			Timeout: 60 * time.Second,
			Enabled: true,
			client: &http.Client{
				Timeout: 60 * time.Second,
			},
		}
		p.upstreams = append(p.upstreams, up)
		p.defaultUp = up
	}

	return p
}

// selectUpstream 根据包名选择上游
func (p *Proxy) selectUpstream(packageName string) *Upstream {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 检查是否为 scoped 包
	if strings.HasPrefix(packageName, "@") {
		// 提取 scope，如 @company/package -> @company
		idx := strings.Index(packageName, "/")
		if idx > 0 {
			scope := packageName[:idx]
			if up, ok := p.scopeMap[scope]; ok {
				return up
			}
		}
	}

	// 返回默认上游
	return p.defaultUp
}

// buildUpstreamURL 构建上游 URL，对包名进行 URL 编码
func buildUpstreamURL(base, packageName string) string {
	segments := strings.Split(packageName, "/")
	encoded := make([]string, len(segments))
	for i, seg := range segments {
		encoded[i] = url.PathEscape(seg)
	}
	return base + "/" + strings.Join(encoded, "/")
}

// GetMetadata 从上游获取包元数据
func (p *Proxy) GetMetadata(packageName string) ([]byte, error) {
	up := p.selectUpstream(packageName)
	if up == nil {
		return nil, fmt.Errorf("no upstream configured for package: %s", packageName)
	}

	urlStr := buildUpstreamURL(up.URL, packageName)
	logger.Debugf("Fetching metadata from upstream [%s]: %s", up.Name, urlStr)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// 设置 Accept-Encoding 支持 gzip
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := up.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from upstream [%s]: %w", up.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrPackageNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream [%s] returned status %d", up.Name, resp.StatusCode)
	}

	// 处理响应体（可能是 gzip 压缩）
	var reader io.Reader = resp.Body
	
	// 调试：记录响应头信息
	logger.Debugf("Response headers for %s: Content-Length=%s, Content-Encoding=%s", 
		packageName, 
		resp.Header.Get("Content-Length"),
		resp.Header.Get("Content-Encoding"))
	
	if resp.Header.Get("Content-Encoding") == "gzip" {
		logger.Debugf("Response is gzip compressed, decompressing...")
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	// 使用 Buffer 读取，支持流式读取大文件
	var buf bytes.Buffer
	limitedReader := io.LimitReader(reader, maxMetadataSize)
	
	// 复制到 buffer
	if _, err := io.Copy(&buf, limitedReader); err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	data := buf.Bytes()
	
	logger.Debugf("Read %d bytes for package %s", len(data), packageName)
	
	// 检查是否读取到数据
	if len(data) == 0 {
		return nil, fmt.Errorf("empty response from upstream [%s]", up.Name)
	}
	
	// 检查是否达到限制（可能被截断）
	if len(data) >= maxMetadataSize {
		logger.Warnf("Metadata size reached limit (%d bytes), data might be truncated!", len(data))
	}

	// 验证 JSON 完整性
	if err := validateJSON(data); err != nil {
		logger.Warnf("Invalid JSON from upstream [%s] for package %s: %v", up.Name, packageName, err)
		logger.Warnf("Response size: %d bytes", len(data))
		
		// 保存原始响应到文件以便调试
		debugFile := fmt.Sprintf("/tmp/grape-debug-%s-%d.json", packageName, time.Now().Unix())
		if writeErr := os.WriteFile(debugFile, data, 0644); writeErr == nil {
			logger.Warnf("Debug: Saved raw response to %s", debugFile)
		}
		
		// 记录响应片段
		if len(data) > 200 {
			logger.Debugf("Response start: %s...", string(data[:100]))
			logger.Debugf("Response end: ...%s", string(data[len(data)-100:]))
		}
		return nil, fmt.Errorf("invalid JSON from upstream: %w", err)
	}

	logger.Debugf("Successfully fetched metadata for %s: %d bytes", packageName, len(data))
	return data, nil
}

// validateJSON 验证 JSON 是否完整有效
func validateJSON(data []byte) error {
	// 尝试解析 JSON 以验证其完整性
	var raw json.RawMessage
	return json.Unmarshal(data, &raw)
}

// GetTarball 从上游获取 tarball
func (p *Proxy) GetTarball(packageName, filename string) ([]byte, error) {
	up := p.selectUpstream(packageName)
	if up == nil {
		return nil, fmt.Errorf("no upstream configured for package: %s", packageName)
	}

	// 对包名和文件名进行 URL 编码
	encodedPackage := buildUpstreamURL(up.URL, packageName)
	encodedFilename := url.PathEscape(filename)
	urlStr := encodedPackage + "/-/" + encodedFilename
	
	logger.Debugf("Fetching tarball from upstream [%s]: %s", up.Name, urlStr)

	resp, err := up.client.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tarball from upstream [%s]: %w", up.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrTarballNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream [%s] returned status %d for tarball", up.Name, resp.StatusCode)
	}

	// 限制读取大小
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxTarballSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read tarball body: %w", err)
	}

	return data, nil
}

// Upstream 返回默认上游 URL（向后兼容）
func (p *Proxy) Upstream() string {
	if p.defaultUp != nil {
		return p.defaultUp.URL
	}
	return ""
}

// Upstreams 返回所有上游配置
func (p *Proxy) Upstreams() []UpstreamInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]UpstreamInfo, 0, len(p.upstreams))
	for _, up := range p.upstreams {
		result = append(result, UpstreamInfo{
			Name:    up.Name,
			URL:     up.URL,
			Scope:   up.Scope,
			Enabled: up.Enabled,
		})
	}
	return result
}

// SetUpstreams 动态更新上游配置（热加载）
func (p *Proxy) SetUpstreams(cfgUpstreams []config.UpstreamConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.upstreams = make([]*Upstream, 0, len(cfgUpstreams))
	p.scopeMap = make(map[string]*Upstream)
	p.defaultUp = nil

	for _, uc := range cfgUpstreams {
		if !uc.Enabled {
			continue
		}

		timeout := uc.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}

		up := &Upstream{
			Name:    uc.Name,
			URL:     strings.TrimSuffix(uc.URL, "/"),
			Scope:   uc.Scope,
			Timeout: timeout,
			Enabled: uc.Enabled,
			client: &http.Client{
				Timeout: timeout,
			},
		}

		p.upstreams = append(p.upstreams, up)

		if uc.Scope != "" {
			p.scopeMap[uc.Scope] = up
		} else {
			p.defaultUp = up
		}
	}
}

// UpstreamInfo 上游信息
type UpstreamInfo struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Scope   string `json:"scope,omitempty"`
	Enabled bool   `json:"enabled"`
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