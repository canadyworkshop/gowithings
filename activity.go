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

var ActivityDataFields = map[string]string{
	"steps":         "steps",
	"distance":      "distance",
	"elevation":     "elevation",
	"soft":          "soft",
	"moderate":      "moderate",
	"intense":       "intense",
	"active":        "active",
	"calories":      "calories",
	"totalcalories": "totalcalories",
	"hr_average":    "hr_average",
	"hr_min":        "hr_min",
	"hr_max":        "hr_max",
	"hr_zone_0":     "hr_zone_0",
	"hr_zone_1":     "hr_zone_1",
	"hr_zone_2":     "hr_zone_2",
	"hr_zone_3":     "hr_zone_3",
}

type GetActivityResponseWrapper struct {
	Status int              `json:"status"`
	Body   ActivityResponse `json:"body"`
}

// ActivityResponse is the raw response from the Getactivit API request.
type ActivityResponse struct {
	Activities []Activity

	More   bool `json:"more"`
	Offset int  `json:"offset"`
}

// Activity defines the results of a single activity.
type Activity struct {
	Date          string  `json:"date"`
	Timezone      string  `json:"timezone"`
	DeviceID      string  `json:"deviceid"`
	HashDeviceID  string  `json:"hash_deviceid"`
	Brand         int     `json:"brand"`
	IsTracker     bool    `json:"is_tracker"`
	Steps         int     `json:"steps"`
	Distance      float64 `json:"distance"`
	Elevation     float64 `json:"elevation"`
	Soft          int     `json:"soft"`
	Moderator     int     `json:"moderator"`
	Intense       int     `json:"intense"`
	Active        int     `json:"active"`
	Calories      float64 `json:"calories"`
	TotalCalories float64 `json:"totalcalories"`
	HRAverage     int     `json:"hr_average"`
	HRMin         int     `json:"hr_min"`
	HRMax         int     `json:"hr_max"`
	HRZone0       int     `json:"hr_zone_0"`
	HRZone1       int     `json:"hr_zone_1"`
	HRZone2       int     `json:"hr_zone_2"`
	HRZone3       int     `json:"hr_zone_3"`
}

type GetActivityParam struct {
	StartDateYMD time.Time `json:"startdateymd"`
	EndDateYMD   time.Time `json:"enddateymd"`
	LastUpdate   time.Time `json:"lastupdate"`
	Offset       int       `json:"offset"`
	DataFields   []string  `json:"data_fields"`
}

// URLEncode encodes the parameter values into a URL encoded form and validates the parameters.
func (param *GetActivityParam) URLEncode() (string, error) {
	v := url.Values{}
	v.Add("action", "getactivity")

	if !param.StartDateYMD.IsZero() && !param.LastUpdate.IsZero() {
		return "", fmt.Errorf("LastUpdate cannot be set if StartDateYMD is set")
	}

	if param.EndDateYMD.IsZero() && param.LastUpdate.IsZero() {
		return "", fmt.Errorf("LastUpdate cannot be set if EndDateYMD is set")
	}

	if !param.LastUpdate.IsZero() {
		v.Add("lastupdate", strconv.FormatInt(param.LastUpdate.Unix(), 10))
	}

	if !param.StartDateYMD.IsZero() {
		v.Add("startdateymd", param.StartDateYMD.Format("2006-01-02"))
	}

	if !param.EndDateYMD.IsZero() {
		v.Add("enddateymd", param.EndDateYMD.Format("2006-01-02"))
	}

	if param.Offset > 0 {
		v.Add("offset", strconv.Itoa(param.Offset))
	}

	if len(param.DataFields) > 0 {
		v.Add("data_fields", strings.Join(param.DataFields, ","))
	}

	return v.Encode(), nil
}

// GetActivity will return the activites as specified by the request param up to the API limit per response. If the
// return set is larger the offset will be provided to perform a second request for additional measrues.
func (c *UserClient) GetActivity(ctx context.Context, param GetActivityParam) (ActivityResponse, error) {
	apiURL := "https://wbsapi.withings.net/v2/measure"

	response := GetActivityResponseWrapper{}

	paramValues, err := param.URLEncode()
	if err != nil {
		return response.Body, fmt.Errorf("failed to generate values: %w", err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, apiURL, strings.NewReader(paramValues))
	if err != nil {
		return response.Body, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return response.Body, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response.Body, err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return response.Body, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Status != 0 {
		return response.Body, fmt.Errorf("failed with status %d", response.Status)
	}

	return response.Body, nil
}

// GetAllActivities will return activities as specified by the request and iterate over the offset until all measures are
// retrieved.
func (c *UserClient) GetAllActivities(ctx context.Context, param GetActivityParam) ([]Activity, error) {
	activities := make([]Activity, 0)
	for {
		resp, err := c.GetActivity(ctx, param)
		if err != nil {
			return nil, fmt.Errorf("failed to get all measures at offset %d: %w", param.Offset, err)
		}
		activities = append(activities, resp.Activities...)

		if resp.Offset == 0 {
			break
		}
		param.Offset = resp.Offset
	}
	return activities, nil
}
