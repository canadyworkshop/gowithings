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

var WorkoutDataFields = map[string]string{
	"calories":            "calories",
	"intensity":           "intensity",
	"manual_distance":     "manual_distance",
	"manual_calories":     "manual_calories",
	"hr_average":          "hr_average",
	"hr_min":              "hr_min",
	"hr_max":              "hr_max",
	"hr_zone_0":           "hr_zone_0",
	"hr_zone_1":           "hr_zone_1",
	"hr_zone_2":           "hr_zone_2",
	"hr_zone_3":           "hr_zone_3",
	"pause_duration":      "pause_duration",
	"algo_pause_duration": "algo_pause_duration",
	"spo2_average":        "spo2_average",
	"steps":               "steps",
	"distance":            "distance",
	"elevation":           "elevation",
	"pool_laps":           "pool_laps",
	"strokes":             "strokes",
	"pool_length":         "pool_length",
}

var WorkoutCategoryByID = map[int]string{
	1:   "Walk",
	2:   "Run",
	3:   "Hiking",
	4:   "Skating",
	5:   "BMX",
	6:   "Bicycling",
	7:   "Swimming",
	8:   "Surfing",
	9:   "Kitesurfing",
	10:  "Windsurfing",
	11:  "Bodyboard",
	12:  "Tennis",
	13:  "Table tennis",
	14:  "Squash",
	15:  "Badminton",
	16:  "Lift weights",
	17:  "Fitness",
	18:  "Elliptical",
	19:  "Pilates",
	20:  "Basket-ball",
	21:  "Soccer",
	22:  "Football",
	23:  "Rugby",
	24:  "Volley-ball",
	25:  "Waterpolo",
	26:  "Horse riding",
	27:  "Golf",
	28:  "Yoga",
	29:  "Dancing",
	30:  "Boxing",
	31:  "Fencing",
	32:  "Wrestling",
	33:  "Martial arts",
	34:  "Skiing",
	35:  "Snowboarding",
	36:  "Other",
	128: "No activity",
	187: "Rowing",
	188: "Zumba",
	191: "Baseball",
	192: "Handball",
	193: "Hockey",
	194: "Ice hockey",
	195: "Climbing",
	196: "Ice skating",
	272: "Multi-sport",
	306: "Indoor walk",
	307: "Indoor running",
	308: "Indoor cycling",
}

type GetWorkoutResponseWrapper struct {
	Status int             `json:"status"`
	Body   WorkoutResponse `json:"body"`
}

type WorkoutResponse struct {
	Series []Workout `json:"series"`

	More   bool `json:"more"`
	Offset int  `json:"offset"`
}

type Workout struct {
	Category  int            `json:"category"`
	Timezone  string         `json:"timezone"`
	Model     int            `json:"model"`
	Attrib    int            `json:"attrib"`
	StartDate int            `json:"start_date"`
	EndDate   int            `json:"end_date"`
	Date      string         `json:"date"`
	Modified  int            `json:"modified"`
	DeviceID  string         `json:"deviceid"`
	Data      WorkoutDetails `json:"data"`
}

type WorkoutDetails struct {
	AlgoPauseDuration int     `json:"algo_pause_duration"`
	Calories          float64 `json:"calories"`
	Distance          float64 `json:"distance"`
	Elevation         float64 `json:"elevation"`
	HRAverage         int     `json:"hr_average"`
	HRMax             int     `json:"hr_max"`
	HRMin             int     `json:"hr_min"`
	HRZone0           int     `json:"hr_zone_0"`
	HRZone1           int     `json:"hr_zone_1"`
	HRZone2           int     `json:"hr_zone_2"`
	HRZone3           int     `json:"hr_zone_3"`
	Intensity         int     `json:"intensity"`
	ManualCalories    float64 `json:"manual_calories"`
	ManualDistance    float64 `json:"manual_distance"`
	PauseDuration     int     `json:"pause_duration"`
	PoolLaps          int     `json:"pool_laps"`
	PoolLength        int     `json:"pool_length"`
	SPO2Average       int     `json:"spo2_average"`
	Steps             int     `json:"steps"`
	Strokes           int     `json:"strokes"`
}

type GetWorkoutParam struct {
	StartDateYMD time.Time `json:"startdateymd"`
	EndDateYMD   time.Time `json:"enddateymd"`
	LastUpdate   time.Time `json:"lastupdate"`
	Offset       int       `json:"offset"`
	DataFields   []string  `json:"data_fields"`
}

// URLEncode encodes the parameter values into a URL encoded form and validates the parameters.
func (param *GetWorkoutParam) URLEncode() (string, error) {
	v := url.Values{}
	v.Add("action", "getworkouts")

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

// GetWorkout will return the workouts as specified by the request param up to the API limit per response. If the
// return set is larger the offset will be provided to perform a second request for additional measrues.
func (c *UserClient) GetWorkout(ctx context.Context, param GetWorkoutParam) (WorkoutResponse, error) {
	apiURL := "https://wbsapi.withings.net/v2/measure"

	response := GetWorkoutResponseWrapper{}

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

// GetAllWorkouts will return workouts as specified by the request and iterate over the offset until all measures are
// retrieved.
func (c *UserClient) GetAllWorkouts(ctx context.Context, param GetWorkoutParam) ([]Workout, error) {
	workouts := make([]Workout, 0)
	for {
		resp, err := c.GetWorkout(ctx, param)
		if err != nil {
			return nil, fmt.Errorf("failed to get all measures at offset %d: %w", param.Offset, err)
		}
		workouts = append(workouts, resp.Series...)

		if resp.Offset == 0 {
			break
		}
		param.Offset = resp.Offset
	}
	return workouts, nil
}
