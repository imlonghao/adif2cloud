package hamqth

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"git.esd.cc/imlonghao/adif2cloud/internal/consts"

	"github.com/projectdiscovery/retryablehttp-go"
)

// HamQTHConfig 定义了 HamQTH 配置
type HamQTHConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Callsign string `mapstructure:"callsign"`
}

// HamQTHProvider 实现了 Provider 接口，用于 HamQTH 服务
type HamQTHProvider struct {
	config HamQTHConfig
}

// NewHamQTHProvider 创建一个新的 HamQTHProvider 实例
func NewHamQTHProvider(cfg HamQTHConfig) *HamQTHProvider {
	slog.Debug("Creating HamQTH provider", "username", cfg.Username, "callsign", cfg.Callsign)
	return &HamQTHProvider{
		config: cfg,
	}
}

// GetSize 获取 HamQTH 上 ADIF 文件的大小
func (p *HamQTHProvider) GetSize() (int64, error) {
	// HamQTH 不直接提供文件大小，返回 0
	return 0, nil
}

// Download 从 HamQTH 下载 ADIF 文件
func (p *HamQTHProvider) Download(w io.Writer) error {
	// HamQTH 不直接提供下载功能，返回错误
	return fmt.Errorf("hamqth does not support direct file download")
}

// Upload 上传 QSO 记录到 HamQTH
func (p *HamQTHProvider) Upload(_ string, line string) error {
	params := url.Values{}
	params.Set("u", p.config.Username)
	params.Set("p", p.config.Password)
	if p.config.Callsign != "" {
		params.Set("c", p.config.Callsign)
	}
	params.Set("adif", line)
	params.Set("prg", "adif2cloud")
	params.Set("cmd", "insert")

	client := retryablehttp.NewClient(retryablehttp.DefaultOptionsSingle)
	req, err := retryablehttp.NewRequest(http.MethodPost, "https://www.hamqth.com/qso_realtime.php", bytes.NewBufferString(params.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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

	// 根据状态码处理响应
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusForbidden:
		return fmt.Errorf("access denied: incorrect username or password")
	case http.StatusInternalServerError:
		return fmt.Errorf("server error: %s", string(body))
	case http.StatusBadRequest:
		return fmt.Errorf("qso rejected: %s", string(body))
	default:
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}
}

// GetName 获取提供商的名称
func (p *HamQTHProvider) GetName() string {
	return fmt.Sprintf("HamQTH->%s-%s", p.config.Username, p.config.Callsign)
}
