package gowithings_test

import (
	"context"
	"testing"
	"time"

	"github.com/canadyworkshop/gowithings"
)

func TestUserClient_GetActivity(t *testing.T) {

	c := testClient()

	u, err := c.DemoUser(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	dataFields := []string{}
	for _, v := range gowithings.ActivityDataFields {
		dataFields = append(dataFields, v)
	}

	param := gowithings.GetActivityParam{
		//DataFields: dataFields,
		LastUpdate: time.Now().AddDate(-10, 0, 0),
		Offset:     0,
	}

	r, err := u.GetActivity(context.Background(), param)
	if err != nil {
		t.Fatal(err)
	}

	if len(r.Activities) < 0 {
		t.Errorf("expected at least one activity, got %d", len(r.Activities))
	}
}

func TestUserClient_GetAllActivities(t *testing.T) {
	c := testClient()

	u, err := c.DemoUser(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	dataFields := []string{}
	for _, v := range gowithings.ActivityDataFields {
		dataFields = append(dataFields, v)
	}

	param := gowithings.GetActivityParam{
		DataFields: dataFields,
		LastUpdate: time.Now().AddDate(-10, 0, 0),
		Offset:     0,
	}

	r, err := u.GetAllActivities(context.Background(), param)
	if err != nil {
		t.Fatal(err)
	}

	if len(r) < 2000 {
		t.Errorf("expected at least 2000 activities, got %d", len(r))
	}
}
