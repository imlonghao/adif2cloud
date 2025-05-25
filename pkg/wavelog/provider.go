package wavelog

import (
	"fmt"
	"io"
)

// WavelogProvider 实现了 Provider 接口，用于 Wavelog 服务
type WavelogProvider struct {
	client *Client
}

// NewWavelogProvider 创建一个新的 WavelogProvider 实例
func NewWavelogProvider(apiURL, apiKey string, stationProfileID int) *WavelogProvider {
	return &WavelogProvider{
		client: NewClient(apiURL, apiKey, stationProfileID),
	}
}

// GetSize 获取 Wavelog 上 ADIF 文件的大小
func (p *WavelogProvider) GetSize() (int64, error) {
	// Wavelog 不直接提供文件大小，返回 0
	return 0, nil
}

// Download 从 Wavelog 下载 ADIF 文件
func (p *WavelogProvider) Download(w io.Writer) error {
	// Wavelog 不直接提供下载功能，返回错误
	return fmt.Errorf("wavelog does not support direct file download")
}

// Upload 上传 QSO 记录到 Wavelog
func (p *WavelogProvider) Upload(_ string, line string) error {
	// 将内容作为 QSO 记录发送
	return p.client.SendQSO(line)
}

// GetName 获取提供商的名称
func (p *WavelogProvider) GetName() string {
	return fmt.Sprintf("Wavelog->%s->%d", p.client.apiURL, p.client.stationProfileID)
}
