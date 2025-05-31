package webhook

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"git.esd.cc/imlonghao/adif2cloud/internal/consts"
	"git.esd.cc/imlonghao/adif2cloud/pkg/adif"
	"github.com/projectdiscovery/retryablehttp-go"
)

// WebhookConfig 定义了 Webhook 配置
type WebhookConfig struct {
	URL     string            `mapstructure:"url"`
	Method  string            `mapstructure:"method"`
	Headers map[string]string `mapstructure:"headers"`
	Body    string            `mapstructure:"body"`
}

// WebhookProvider 实现了 Provider 接口，用于 Webhook 服务
type WebhookProvider struct {
	config WebhookConfig
}

// NewWebhookProvider 创建一个新的 WebhookProvider 实例
func NewWebhookProvider(cfg WebhookConfig) *WebhookProvider {
	slog.Debug("Creating Webhook provider",
		"url", cfg.URL,
		"method", cfg.Method)
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
func (p *WebhookProvider) Upload(_, line string) error {
	client := retryablehttp.NewClient(retryablehttp.DefaultOptionsSingle)

	// 准备请求体
	var bodyReader io.Reader
	if p.config.Body != "" {
		body, err := adif.FillTemplate(p.config.Body, adif.Parse(line))
		if err != nil {
			return fmt.Errorf("failed to fill template: %w", err)
		}
		bodyReader = strings.NewReader(body)
	}

	// 创建请求
	method := p.config.Method
	if method == "" {
		method = http.MethodGet
	}
	req, err := retryablehttp.NewRequest(method, p.config.URL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", fmt.Sprintf("adif2cloud/%s (+https://git.esd.cc/imlonghao/adif2cloud)", consts.Version))
	for h, v := range p.config.Headers {
		req.Header.Set(h, v)
	}

	// 发送请求
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
