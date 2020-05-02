package chanmgr

import (
	"io/ioutil"
	"log"
	"reflect"
	"testing"
)

func TestGetError(t *testing.T) {
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
			RequestBuffer: 1,
		})
	defer func() {
		exit <- Exit
	}()

	resp := channels[0].Send(5, 25)

	if err := resp.Get(); err != nil {
		t.Errorf("Received an unexpected error during Get(): %v", err)
		return
	}
	if resp.Err != nil {
		t.Errorf("Received an error unexpectedly in processing: %v", resp.Err)
		return
	}
	if !reflect.DeepEqual(resp.Data, resp.Context) {
		t.Errorf("Processing error: expected %v, got %v", resp.Context, resp.Data)
		return
	}

	// This is the actual test ...
	if err := resp.Get(); err == nil {
		t.Error("Should have returned an error on second call to Get()")
	} else {
		var expected = "Get() error reading response"
		if err.Error() != expected {
			t.Errorf("Incorrect error: expected %v, got %v", err, expected)
		}
	}
}
