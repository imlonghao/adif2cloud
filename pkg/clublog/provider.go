package clublog

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

const API_KEY = ""

// ClubLogConfig 定义了 Club Log 配置
type ClubLogConfig struct {
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
	Callsign string `mapstructure:"callsign"`
}

// ClubLogProvider 实现了 Provider 接口，用于 Club Log 服务
type ClubLogProvider struct {
	config ClubLogConfig
}

// NewClubLogProvider 创建一个新的 ClubLogProvider 实例
func NewClubLogProvider(cfg ClubLogConfig) *ClubLogProvider {
	if API_KEY == "" {
		slog.Error("API_KEY is not set")
		return nil
	}
	slog.Info("Creating Club Log provider",
		"email", cfg.Email,
		"callsign", cfg.Callsign)
	return &ClubLogProvider{
		config: cfg,
	}
}

// GetSize 获取 Club Log 上 ADIF 文件的大小
func (p *ClubLogProvider) GetSize() (int64, error) {
	// Club Log 不直接提供文件大小，返回 0
	return 0, nil
}

// Download 从 Club Log 下载 ADIF 文件
func (p *ClubLogProvider) Download(w io.Writer) error {
	// Club Log 不直接提供下载功能，返回错误
	return fmt.Errorf("club log does not support direct file download")
}

// Upload 上传 QSO 记录到 Club Log
func (p *ClubLogProvider) Upload(_ string, line string) error {
	// 准备表单数据
	formData := url.Values{}
	formData.Set("email", p.config.Email)
	formData.Set("password", p.config.Password)
	formData.Set("callsign", p.config.Callsign)
	formData.Set("adif", line)
	formData.Set("api", API_KEY)

	// 发送 POST 请求
	resp, err := http.PostForm("https://clublog.org/realtime.php", formData)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

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
func (p *ClubLogProvider) GetName() string {
	return fmt.Sprintf("ClubLog->%s", p.config.Callsign)
}
