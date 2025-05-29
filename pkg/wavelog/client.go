package wavelog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"git.esd.cc/imlonghao/adif2cloud/internal/consts"
	"github.com/projectdiscovery/retryablehttp-go"
)

type Client struct {
	apiURL           string
	apiKey           string
	stationProfileID int
}

type QSORequest struct {
	Key              string `json:"key"`
	StationProfileID string `json:"station_profile_id"`
	Type             string `json:"type"`
	String           string `json:"string"`
}

func NewClient(apiURL, apiKey string, stationProfileID int) *Client {
	slog.Debug("Creating Wavelog client",
		"api_url", apiURL,
		"station_profile_id", stationProfileID)
	return &Client{
		apiURL:           apiURL,
		apiKey:           apiKey,
		stationProfileID: stationProfileID,
	}
}

func (c *Client) SendQSO(adiString string) error {
	qsoReq := QSORequest{
		Key:              c.apiKey,
		StationProfileID: strconv.Itoa(c.stationProfileID),
		Type:             "adif",
		String:           adiString,
	}

	jsonData, err := json.Marshal(qsoReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	client := retryablehttp.NewClient(retryablehttp.DefaultOptionsSingle)
	req, err := retryablehttp.NewRequest(http.MethodPost, c.apiURL, bytes.NewBuffer(jsonData))
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

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
