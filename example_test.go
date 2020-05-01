package chanmgr

import (
	"fmt"

	"github.com/golang/example/stringutil"
)

func initialise() ([]InOut, ExitChannel) {

	// Declare the channels to be processed, and the functions to be called
	// CreateInOut will panic if incorrect values are used
	channels := []InOut{
		CreateInOut(
			make(chan string), // This must be a channel able to send messages
			func(i interface{}) (interface{}, error) {
				s := i.(string)
				return stringutil.Reverse(s), nil
			}, // This function handles sent messages, optionally returning values
			WantResponse, // WantResponse|IgnoreResponse indicates whether function return values should be forwarded
		),
	}

	// Start the manager
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
