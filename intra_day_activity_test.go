package gowithings_test

import (
	"context"
	"testing"

	"github.com/canadyworkshop/gowithings"
)

func TestUserClient_GetIntraDayActivity(t *testing.T) {

	c := testClient()

	u, err := c.DemoUser(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	dataFields := []string{}
	for _, v := range gowithings.IntraActivityDataFields {
		dataFields = append(dataFields, v)
	}

	param := gowithings.GetIntraDayActivityParam{
		DataFields: dataFields,
		//StartDate: time.Now().AddDate(-10, 0, 0),
		//EndDate:   time.Now(),
	}

	_, err = u.GetIntraDayActivity(context.Background(), param)
	if err != nil {
		t.Fatal(err)
	}

}
