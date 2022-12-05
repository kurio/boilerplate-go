package goboilerplate

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

/*
Create new event by following this step:
1. Add the event name const
2. Create struct for the event body and event
	- event body should implement SystemEventBody interface
	- event should implement SystemEvent and json.Marshaller interface
3. Create a JSONUnmarshalFunc that JSONUnmarshal the system event
4. Add the unmarshalFunc to eventNameToJsonUnmarshalFunc
More importantly, add the test in system_event_test.go

*/

type (
	// ContextKey is key of context.
	ContextKey string

	// EventName is name of event.
	EventName string

	// EventPublisher is the interface that wraps the basic Publsih method.
	EventPublisher interface {
		Publish(e SystemEvent)
	}

	// SystemEventBody is the interface for system event body.
	SystemEventBody interface {
		// Name returns the name of the event.
		Name() EventName
		// TopicKey returns the context key to get publisher for the event.
		TopicKey() ContextKey
		// GenerateEvent returns the system event.
		GenerateEvent() SystemEvent
	}
	// SystemEvent is the interface for system event.
	SystemEvent interface {
		// GetSystemEvent returns the system event body struct.
		GetSystemEventBody() SystemEventBody
		// GetOccuredTime returns the event occured time.
		GetOccuredTime() time.Time
	}

	// JSONUnmarshalFunc is a function adapter that unmarshal JSON data to a system event.
	JSONUnmarshalFunc func(data []byte) (event SystemEvent, err error)
)

// String returns string representation of the context key
func (c ContextKey) String() string {
	return string(c)
}

// String representation of the event name
func (n EventName) String() string {
	return string(n)
}

const (
	// TODO: Add context key for publisher here.
	ContextKeyFoo = ContextKey("foo")

	// List of event names.
	// TODO: update
	EventFooCreated = EventName("foo.created")
	EventFooUpdated = EventName("foo.updated")
	EventFooDeleted = EventName("foo.deleted")
)

var eventNameToJSONUnmarshalFunc = map[EventName]JSONUnmarshalFunc{
	// TODO: update
	EventFooCreated: unmarshalFooCreatedEvent,
	// EventFooUpdated: unmarshalFooUpdatedEvent,
	// EventFooDeleted: unmarshalFooDeletedEvent,
}

// publisherFromContext get Publisher from ctx.
func publisherFromContext(ctx context.Context, contextKeyPublisher ContextKey) EventPublisher {
	pub, ok := ctx.Value(contextKeyPublisher).(EventPublisher)
	if !ok {
		return nil
	}
	return pub
}

// PublishSystemEvent publish the system event to the respective topic by using the associated publisher
func PublishSystemEvent(ctx context.Context, eb SystemEventBody) {
	publisher := publisherFromContext(ctx, eb.TopicKey())
	if publisher == nil {
		logrus.Debugf("No publisher for event '%s'", eb.Name())
		return
	}

	e := eb.GenerateEvent()
	publisher.Publish(e)
}

// SystemEventFromJSON will unmarshall the JSON bytes to the correct system event struct.
func SystemEventFromJSON(data []byte) (e SystemEvent, err error) {
	var h struct {
		Name EventName `json:"name"`
	}
	err = json.Unmarshal(data, &h)
	if err != nil {
		return
	}

	unmarshal, ok := eventNameToJSONUnmarshalFunc[h.Name]
	if !ok {
		err = errors.Errorf("no unmarshaller for event %s", h.Name)
		return
	}

	e, err = unmarshal(data)
	return
}

// Implementation of events
type FooCreated struct {
	Foo interface{} `json:"foo"`
}

func (eb FooCreated) Name() EventName {
	return EventFooCreated
}

func (eb FooCreated) TopicKey() ContextKey {
	return ContextKeyFoo
}

func (eb FooCreated) GenerateEvent() SystemEvent {
	return FooCreatedEvent{
		Body:        eb,
		OccuredTime: time.Now().Round(0),
	}
}

type FooCreatedEvent struct {
	Body        FooCreated `json:"body"`
	OccuredTime time.Time  `json:"occured_time"`
}

func (e FooCreatedEvent) GetSystemEventBody() SystemEventBody {
	return e.Body
}
func (e FooCreatedEvent) GetOccuredTime() time.Time {
	return e.OccuredTime
}

func (e FooCreatedEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":         e.Body.Name(),
		"body":         e.Body,
		"occured_time": e.OccuredTime.Format(time.RFC3339Nano),
	})
}

func unmarshalFooCreatedEvent(data []byte) (res SystemEvent, err error) {
	event := FooCreatedEvent{}
	err = json.Unmarshal(data, &event)
	if err != nil {
		return
	}
	res = event
	return
}
