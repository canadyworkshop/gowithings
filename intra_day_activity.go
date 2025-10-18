package gowithings

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var IntraActivityDataFields = map[string]string{
	"steps":       "steps",
	"elevation":   "elevation",
	"calories":    "calories",
	"distance":    "distance",
	"stroke":      "stroke",
	"pool_lap":    "pool_lap",
	"duration":    "duration",
	"heart_rate":  "heart_rate",
	"spo2_auto":   "spo2_auto",
	"rmssd":       "rmssd",
	"sdnn1":       "sdnn1",
	"hrv_quality": "hrv_quality",
}

type GetIntraDayActivityWrapper struct {
	Status int                      `json:"status"`
	Body   IntraDayActivityResponse `json:"body"`
}

type IntraDayActivityResponse struct {
	Series map[int]IntraDayActivitySeries `json:"series"`
}

type IntraDayActivitySeries struct {
	DeviceID   int     `json:"deviceid"`
	Model      string  `json:"model"`
	ModelID    int     `json:"model_id"`
	Steps      int     `json:"steps"`
	Elevation  float64 `json:"elevation"`
	Calories   float64 `json:"calories"`
	Distance   float64 `json:"distance"`
	Stroke     int     `json:"stroke"`
	PoolLap    int     `json:"pool_lap"`
	Duration   int     `json:"duration"`
	HeartRate  float64 `json:"heart_rate"`
	RMSSD      float64 `json:"rms_sd"`
	SDNN1      float64 `json:"sdnn1"`
	HRVQuality int     `json:"hrv_quality"`
}

type GetIntraDayActivityParam struct {
	StartDate  time.Time
	EndDate    time.Time
	DataFields []string
}

func (param *GetIntraDayActivityParam) URLEncode() (string, error) {
	v := url.Values{}
	v.Add("action", "getintradayactivity")

	if !param.StartDate.IsZero() {
		v.Add("startdate", strconv.FormatInt(param.StartDate.Unix(), 10))
	}
	if !param.EndDate.IsZero() {
		v.Add("enddate", strconv.FormatInt(param.EndDate.Unix(), 10))
	}

	if len(param.DataFields) > 0 {
		v.Add("data_fields", strings.Join(param.DataFields, ","))
	}

	return v.Encode(), nil
}

// GetIntraDayActivity will return the intra day activity as specified by the request param up to the API limit per
// response. If the return set is larger the offset will be provided to perform a second request for additional
// measrues.
func (c *UserClient) GetIntraDayActivity(ctx context.Context, param GetIntraDayActivityParam) (map[int]IntraDayActivitySeries, error) {
	apiURL := "https://wbsapi.withings.net/v2/measure"

	response := GetIntraDayActivityWrapper{}

	paramValues, err := param.URLEncode()
	if err != nil {
		return response.Body.Series, fmt.Errorf("failed to generate values: %w", err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, apiURL, strings.NewReader(paramValues))
	if err != nil {
		return response.Body.Series, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return response.Body.Series, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response.Body.Series, err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return response.Body.Series, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Status != 0 {
		return response.Body.Series, fmt.Errorf("failed with status %d", response.Status)
	}

	return response.Body.Series, nil
}
