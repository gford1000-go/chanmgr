package chanmgr

import (
	"errors"
	"fmt"
)

// Response instances contain the outcome of Processor calls
type Response struct {
	Context interface{}    // Optional additional information: these are not passed to the Processor
	Input   interface{}    // Input information passed to the Processor to get this Response
	Data    interface{}    // Outcome of the call to the Processor, if processing completed successfully
	Err     error          // Any error during the call to the Processor
	c       chan *Response // The channel Get() uses to retrieve the processing outcomes
}

// IsAvailable will be true when the Processor has returned results
func (r *Response) IsAvailable() bool {
	return len(r.c) != 0
}

// Get provides access to Processor results.  Get will block until results are available
func (r *Response) Get() (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("Get() panicked: %v", e)
		}
	}()

	resp, ok := <-r.c // Blocks until a response if available
	if !ok {
		return errors.New("Get() error reading response")
	}

	close(r.c) // Ensure no further attempts to read from the channel succeed

	r.Data, r.Err = resp.Data, resp.Err
	return nil
}

// newResponse returns an empty Response instance, which can wait and handle Responses
// from manager when these are available
func newResponse(context, inputs interface{}, requestChannel <-chan *Response) *Response {

	var c = make(chan *Response)
	go func(responseChannel chan *Response, requestChannel <-chan *Response) {
		defer func() {
			if r := recover(); r != nil {
				responseChannel <- &Response{Err: fmt.Errorf("Panicked whilst receiving response: %s", r)}
			}
		}()

		// Blocks waiting for manager to provide Response from processing, which
		// ensures that manager.run() will not block
		resp := <-requestChannel

		// Blocks until Get() is called
		responseChannel <- resp
	}(c, requestChannel)

	return &Response{c: c, Input: inputs, Context: context}
}
