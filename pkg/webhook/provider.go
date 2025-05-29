package webhook

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"git.esd.cc/imlonghao/adif2cloud/internal/consts"
	"github.com/projectdiscovery/retryablehttp-go"
)

// WebhookConfig 定义了 Webhook 配置
type WebhookConfig struct {
	URL string `mapstructure:"url"`
}

// WebhookProvider 实现了 Provider 接口，用于 Webhook 服务
type WebhookProvider struct {
	config WebhookConfig
}

// NewWebhookProvider 创建一个新的 WebhookProvider 实例
func NewWebhookProvider(cfg WebhookConfig) *WebhookProvider {
	slog.Debug("Creating Webhook provider",
		"url", cfg.URL)
	return &WebhookProvider{
		config: cfg,
	}
}

// GetSize 获取 Webhook 上 ADIF 文件的大小
func (p *WebhookProvider) GetSize() (int64, error) {
	// Webhook 不直接提供文件大小，返回 0
	return 0, nil
}

// Download 从 Webhook 下载 ADIF 文件
func (p *WebhookProvider) Download(w io.Writer) error {
	// Webhook 不直接提供下载功能，返回错误
	return fmt.Errorf("webhook does not support direct file download")
}

// Upload 上传 QSO 记录到 Webhook
func (p *WebhookProvider) Upload(_, _ string) error {
	client := retryablehttp.NewClient(retryablehttp.DefaultOptionsSingle)
	req, err := retryablehttp.NewRequest(http.MethodGet, p.config.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", fmt.Sprintf("adif2cloud/%s (+https://git.esd.cc/imlonghao/adif2cloud)", consts.Version))
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetName 获取提供商的名称
func (p *WebhookProvider) GetName() string {
	return fmt.Sprintf("Webhook->%s", p.config.URL)
}
