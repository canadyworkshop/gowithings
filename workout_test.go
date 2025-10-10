package gowithings_test

import (
	"context"
	"testing"
	"time"

	"github.com/canadyworkshop/gowithings"
)

func TestUserClient_GetWorkout(t *testing.T) {

	c := testClient()

	u, err := c.DemoUser(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	dataFields := []string{}
	for _, v := range gowithings.WorkoutDataFields {
		dataFields = append(dataFields, v)
	}

	param := gowithings.GetWorkoutParam{
		//DataFields: dataFields,
		LastUpdate: time.Now().AddDate(-10, 0, 0),
		Offset:     0,
	}

	r, err := u.GetWorkout(context.Background(), param)
	if err != nil {
		t.Fatal(err)
	}

	if len(r.Series) <= 0 {
		t.Errorf("expected at least one workouts, got %d", len(r.Series))
	}
}

func TestUserClient_GetAllWorkouts(t *testing.T) {
	c := testClient()

	u, err := c.DemoUser(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	dataFields := []string{}
	for _, v := range gowithings.WorkoutDataFields {
		dataFields = append(dataFields, v)
	}

	param := gowithings.GetWorkoutParam{
		DataFields: dataFields,
		LastUpdate: time.Now().AddDate(-10, 0, 0),
		Offset:     0,
	}

	r, err := u.GetAllWorkouts(context.Background(), param)
	if err != nil {
		t.Fatal(err)
	}

	if len(r) < 400 {
		t.Errorf("expected at least 2000 workouts, got %d", len(r))
	}
}
