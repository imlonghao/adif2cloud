package wavelog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
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
	slog.Info("Creating Wavelog client",
		"api_url", apiURL,
		"station_profile_id", stationProfileID)
	return &Client{
		apiURL:           apiURL,
		apiKey:           apiKey,
		stationProfileID: stationProfileID,
	}
}

func (c *Client) SendQSO(adiString string) error {
	req := QSORequest{
		Key:              c.apiKey,
		StationProfileID: strconv.Itoa(c.stationProfileID),
		Type:             "adif",
		String:           adiString,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
