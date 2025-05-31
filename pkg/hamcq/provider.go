package hamcq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"git.esd.cc/imlonghao/adif2cloud/internal/consts"

	"github.com/projectdiscovery/retryablehttp-go"
)

// HamCQConfig 定义了 HamCQ 配置
type HamCQConfig struct {
	Key string `mapstructure:"key"`
}

// HamCQProvider 实现了 Provider 接口，用于 HamCQ 服务
type HamCQProvider struct {
	config HamCQConfig
}

type QSORequest struct {
	Key  string `json:"key"`
	ADIF string `json:"adif"`
	App  string `json:"app"`
}

// NewHamCQProvider 创建一个新的 HamCQProvider 实例
func NewHamCQProvider(cfg HamCQConfig) *HamCQProvider {
	slog.Debug("Creating HamCQ provider", "key", cfg.Key)
	return &HamCQProvider{
		config: cfg,
	}
}

// GetSize 获取 HamCQ 上 ADIF 文件的大小
func (p *HamCQProvider) GetSize() (int64, error) {
	// HamCQ 不直接提供文件大小，返回 0
	return 0, nil
}

// Download 从 HamCQ 下载 ADIF 文件
func (p *HamCQProvider) Download(w io.Writer) error {
	// HamCQ 不直接提供下载功能，返回错误
	return fmt.Errorf("club log does not support direct file download")
}

// Upload 上传 QSO 记录到 HamCQ
func (p *HamCQProvider) Upload(_ string, line string) error {
	qsoReq := QSORequest{
		Key:  p.config.Key,
		ADIF: line,
		App:  "ADIF2Cloud",
	}

	jsonData, err := json.Marshal(qsoReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	client := retryablehttp.NewClient(retryablehttp.DefaultOptionsSingle)
	req, err := retryablehttp.NewRequest(http.MethodPost, "https://api.hamcq.cn/v1/logbook?from=gridtracker", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("adif2cloud/%s (+https://git.esd.cc/imlonghao/adif2cloud)", consts.Version))
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetName 获取提供商的名称
func (p *HamCQProvider) GetName() string {
	return fmt.Sprintf("HamCQ->%s", p.config.Key)
}
