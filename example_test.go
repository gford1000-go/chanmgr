package chanmgr

import (
	"fmt"

	"github.com/golang/example/stringutil"
)

func initialise() ([]InOut, ExitChannel, error) {

	// Declare the channels to be processed, and the functions to be called
	channels := []InOut{
		{
			In: make(chan string),
			Fn: func(i interface{}) (interface{}, error) {
				s := i.(string)
				return stringutil.Reverse(s), nil
			},
			Out: make(chan *Response),
		},
	}

	// Start the manager
	exit, err := New(channels, nil, nil)

	return channels, exit, err
}

func Example() {

	// Initialise with a single function being handled
	channels, exit, _ := initialise()
	defer func() {
		exit <- Exit // Gracefully stop the chanmgr
	}()

	// Send messages on the channels, receive output from processing
	channels[0].Send("Hello")
	response := <-channels[0].Out

	fmt.Println(response.Data)
	// Output: olleH

	// Catches incorrect types being sent to channels
	if err := channels[0].Send(2); err != nil {
		// Do something
	}
}
