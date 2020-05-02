package chanmgr

import (
	"fmt"

	"github.com/golang/example/stringutil"
)

func initialise() ([]*InOut, ExitChannel) {

	// Declare the channels to be processed, and the functions to be called
	// CreateInOut will panic if incorrect values are used
	channels := []*InOut{
		&InOut{
			Processor: func(i interface{}) (interface{}, error) {
				return stringutil.Reverse(i.(string)), nil
			}, // This function will receive messages, optionally returning values
			WantResponse: WantResponse, // WantResponse|IgnoreResponse indicates whether function return values should be forwarded
		},
	}

	// Start a manager instance
	exit, _ := New(channels, nil, nil)

	return channels, exit
}

func Example() {

	// Initialise with a single function being handled
	channels, exit := initialise()
	defer func() {
		exit <- Exit // Gracefully stop the chanmgr
	}()

	// Send messages on the channels, receive output from processing
	response, _ := channels[0].SendRecv("Hello")
	fmt.Println(response)
	// Output: olleH
}
