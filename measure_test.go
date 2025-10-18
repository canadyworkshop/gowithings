package gowithings_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/canadyworkshop/gowithings"
)

// testClient builds a new test client from the test envs.
func testClient() *gowithings.Client {

	c := gowithings.NewClient(gowithings.Config{
		ClientID:     os.Getenv("GOWITHINGS_TEST_CLIENT_ID"),
		ClientSecret: os.Getenv("GOWITHINGS_TEST_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOWITHINGS_TEST_REDIRECT_URL"),
	})

	return c
}

func TestUserClient_GetMeasure(t *testing.T) {

	c := testClient()

	u, err := c.DemoUser(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	measureType := []string{}
	for _, v := range gowithings.MeasureTypes {
		measureType = append(measureType, v)
	}

	param := gowithings.GetMeasureParam{
		Category:     gowithings.MeasureCategoryRealMeasures,
		MeasureTypes: measureType,
		LastUpdate:   time.Now().AddDate(-10, 0, 0),
		Offset:       0,
	}

	r, err := u.GetMeasure(context.Background(), param)
	if err != nil {
		t.Fatal(err)
	}

	if len(r.MeasureGroups) < 0 {
		t.Errorf("expected at least one measure group, got %d", len(r.MeasureGroups))
	}
}

func TestUserClient_GetAllMeasures(t *testing.T) {
	c := testClient()

	u, err := c.DemoUser(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	measureType := []string{}
	for _, v := range gowithings.MeasureTypes {
		measureType = append(measureType, v)
	}

	param := gowithings.GetMeasureParam{
		Category:     gowithings.MeasureCategoryRealMeasures,
		MeasureTypes: measureType,
		LastUpdate:   time.Now().AddDate(-10, 0, 0),
		Offset:       0,
	}

	r, err := u.GetAllMeasures(context.Background(), param)
	if err != nil {
		t.Fatal(err)
	}

	if len(r) < 2000 {
		t.Errorf("expected at least 2000 measure groups, got %d", len(r))
	}
}
