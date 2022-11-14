package goboilerplate_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	goboilerplate "github.com/kurio/boilerplate-go"
	"github.com/kurio/boilerplate-go/internal/ebus"
	"github.com/stretchr/testify/require"
)

func TestNewSystemEvent(t *testing.T) {
	foo := map[string]interface{}{
		"id": "some-id",
	}

	tests := map[string]goboilerplate.SystemEventBody{
		"foo created": goboilerplate.FooCreated{
			Foo: foo,
		},
	}

	for testName, eventBody := range tests {
		t.Run(testName, func(t *testing.T) {
			e := eventBody.GenerateEvent()

			eventJSON, err := json.Marshal(e)
			require.NoError(t, err)

			var res interface{}
			res, err = goboilerplate.SystemEventFromJSON(eventJSON)
			require.NoError(t, err)
			require.Equal(t, e, res)
		})
	}
}

func ExamplePublishSystemEvent() {
	eventBus := new(ebus.Bus)
	eventBus.Subscribe(ebus.HandlerFunc(func(e goboilerplate.SystemEvent) {
		eb := e.GetSystemEventBody()
		j, err := json.Marshal(eb)
		if err != nil {
			panic(err)
		}

		fmt.Printf("name: %s\n", eb.Name())
		fmt.Printf("body: %s\n", j)
	}))

	ctx := context.WithValue(context.Background(), goboilerplate.ContextKeyFoo, eventBus)

	foo := map[string]interface{}{
		"id": "my-id",
	}

	goboilerplate.PublishSystemEvent(ctx, goboilerplate.FooCreated{
		Foo: foo,
	})

	// Output:
	// name: foo.created
	// body: {"foo":{"id":"my-id"}}
}
