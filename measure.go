package gowithings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// MeasureTypes is a convenience map of all the currently known measure types.
var MeasureTypes = map[string]string{
	"Weight":                         "1",
	"Height":                         "4",
	"FatFreeMassKG":                  "5",
	"FatRatio":                       "6",
	"FatMassWeight":                  "8",
	"DiastolicBloodPressure":         "9",
	"SystolicBloodPressure":          "10",
	"HeartPulse":                     "11",
	"Temperature":                    "12",
	"SP02":                           "54",
	"BodyTemperature":                "71",
	"SkinTemperature":                "73",
	"MuscleMass":                     "76",
	"Hydration":                      "77",
	"BoneMass":                       "88",
	"PulseWaveVelocity":              "91",
	"VO2":                            "123",
	"AtrialFibrillation":             "130",
	"QRS":                            "135",
	"PR":                             "136",
	"QT":                             "137",
	"CorrectedQT":                    "138",
	"AtrialFPPG":                     "139",
	"Vascular":                       "155",
	"NerveHealthScoreConductance":    "167",
	"ExtracellularWater":             "168",
	"IntracellularWater":             "169",
	"VisceralFat":                    "170",
	"FatFreeMass":                    "173",
	"FatMass":                        "174",
	"MuscleMassSegments":             "175",
	"ElectrodermalActivityFeet":      "196",
	"BasalMetabolicRate":             "226",
	"MetabolicAge":                   "227",
	"ElectrochemicalSkinConductance": "229",
}

var MeasureTypesByKey = map[int]string{
	1:   "Weight",
	4:   "Height",
	5:   "FatFreeMassKG",
	6:   "FatRatio",
	8:   "FatMassWeight",
	9:   "DiastolicBloodPressure",
	10:  "SystolicBloodPressure",
	11:  "HeartPulse",
	12:  "Temperature",
	54:  "SP02",
	71:  "BodyTemperature",
	73:  "SkinTemperature",
	76:  "MuscleMass",
	77:  "Hydration",
	88:  "BoneMass",
	91:  "PulseWaveVelocity",
	123: "VO2",
	130: "AtrialFibrillation",
	135: "QRS",
	136: "PR",
	137: "QT",
	138: "CorrectedQT",
	139: "AtrialFPPG",
	155: "Vascular",
	167: "NerveHealthScoreConductance",
	168: "ExtracellularWater",
	169: "IntracellularWater",
	170: "VisceralFat",
	173: "FatFreeMass",
	174: "FatMass",
	175: "MuscleMassSegments",
	196: "ElectrodermalActivityFeet",
	226: "BasalMetabolicRate",
	227: "MetabolicAge",
	229: "ElectrochemicalSkinConductance",
}

const (
	MeasureCategoryRealMeasures   = "1"
	MeasureCateogryUserObjectives = "2"
)

// GetMeasureParam is the parameter needed to specify what measures to retreive.
type GetMeasureParam struct {
	MeasureTypes []string
	Category     string
	StartDate    time.Time
	EndDate      time.Time
	LastUpdate   time.Time
	Offset       int
}

type GetMeasureResponseWrapper struct {
	Status int             `json:"status"`
	Body   MeasureResponse `json:"body"`
}

// MeasureResponse is the raw response of a get measure API request.
type MeasureResponse struct {
	UpdateTime    int            `json:"updatetime"`
	Timezone      string         `json:"timezone"`
	More          int            `json:"more"`
	Offset        int            `json:"offset"`
	MeasureGroups []MeasureGroup `json:"measuregrps"`
}

// MeasureGroup is a group of measurements as returned by the API. Each group of measurements were recoreded at the
// exact same time by the same device.
type MeasureGroup struct {
	GroupID      int       `json:"grpid"`
	Attrib       int       `json:"attrib"`
	Date         int       `json:"date"`
	Created      int       `json:"created"`
	Modified     int       `json:"modified"`
	Category     int       `json:"category"`
	DeviceID     string    `json:"deviceid"`
	HashDeviceID int       `json:"has_deviceid"`
	Timezone     string    `json:"timezone"`
	Measures     []Measure `json:"measures"`
}

// Measure is a specific measurement as returned by the API.
type Measure struct {
	Value    int     `json:"value"`
	Type     int     `json:"type"`
	Unit     int     `json:"unit"`
	Algo     int     `json:"algo"`
	FM       float64 `json:"fm"`
	Position int     `json:"position"`
}

func (m Measure) Float64() float64 {
	return float64(m.Value) * math.Pow10(m.Unit)
}

// URLEncode encodes the parameter values into a URL encoded from.
func (m GetMeasureParam) URLEncode() (string, error) {
	v := url.Values{}
	v.Add("action", "getmeas")

	switch len(m.MeasureTypes) {
	case 0:
		return "", errors.New("no measure types provided")
	case 1:
		v.Add("meastype", m.MeasureTypes[0])
	default:
		v.Add("meastypes", strings.Join(m.MeasureTypes, ","))
	}

	if m.Category == "0" || m.Category == "1" {
		v.Add("category", "1")
	} else {
		v.Add("category", (m.Category))
	}

	if !m.LastUpdate.IsZero() {
		v.Add("lastupdate", strconv.FormatInt(m.LastUpdate.Unix(), 10))
	}

	if !m.StartDate.IsZero() {
		v.Add("startdate", strconv.FormatInt(m.StartDate.Unix(), 10))
	}

	if !m.EndDate.IsZero() {
		v.Add("enddate", strconv.FormatInt(m.EndDate.Unix(), 10))
	}

	if m.Offset > 0 {
		v.Add("offset", strconv.Itoa(m.Offset))
	}

	return v.Encode(), nil

}

// GetMeasure will return the measures as specified by the request param up to the API limit per response. If the
// return set is larger the offset will be provided to perform a second request for additional measrues.
func (c *UserClient) GetMeasure(ctx context.Context, param GetMeasureParam) (MeasureResponse, error) {
	apiURL := "https://wbsapi.withings.net/measure"

	response := GetMeasureResponseWrapper{}

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

// GetAllMeasures will return measures as specified by the request and iterate over the offset until all measures are
// retrieved.
func (c *UserClient) GetAllMeasures(ctx context.Context, param GetMeasureParam) ([]MeasureGroup, error) {
	measures := make([]MeasureGroup, 0)
	for {
		resp, err := c.GetMeasure(ctx, param)
		if err != nil {
			return nil, fmt.Errorf("failed to get all measures at offset %d: %w", param.Offset, err)
		}
		measures = append(measures, resp.MeasureGroups...)

		if resp.Offset == 0 {
			break
		}
		param.Offset = resp.Offset
	}
	return measures, nil
}
