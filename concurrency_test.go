package chanmgr

import (
	"io/ioutil"
	"log"
	"reflect"
	"testing"
)

func TestConcurrency(t *testing.T) {
	channels := []*InOut{
		&InOut{
			Processor: func(i interface{}) (interface{}, error) {
				v := i.(int)
				return v * v, nil
			},
			WantResponse: WantResponse},
	}

	// Start a manager instance, with request buffer of length 1,000
	exit, _ := New(
		channels,
		nil,
		&Config{
			Log:           log.New(ioutil.Discard, "", 0),
			RequestBuffer: 1000,
		})

	responses := []*Response{}

	// Use the Context feature of Send() to store expected result
	for i := 0; i < 100000; i++ {
		responses = append(responses, channels[0].Send(i, i*i))
	}

	for i, r := range responses {
		if err := r.Get(); err != nil {
			t.Errorf("Received an error during Get() for %d: %v", i, err)
		}
		if r.Err != nil {
			t.Errorf("Received an error unexpectedly for %d: %v", i, r.Err)
			return
		}
		if !reflect.DeepEqual(r.Data, r.Context) {
			t.Errorf("Processing error for %d: expected %v, got %v", i, r.Context, r.Data)
		}
	}

	exit <- Exit
}
