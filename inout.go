package chanmgr

import "fmt"

// Processor is the type of function that handles supplied messages
type Processor func(interface{}) (interface{}, error)

// ResponseWanted indicated whether the outcome of Processor are of interest
type ResponseWanted bool

// WantResponse indicates that the results of processing should be returned (must then be consumed to avoid block)
var WantResponse ResponseWanted = true

// IgnoreResponse indicates that any results from processing should be discarded
var IgnoreResponse ResponseWanted = false

// InOut provides the details a single function that is to be triggered via a channel and responds via a second
type InOut struct {
	Processor    Processor      // The Processor to be called
	WantResponse ResponseWanted // Whether a Response is expected from the Processor
	m            *manager       // The manager instance handling this processing
}

// SendRecv invokes the Processor and blocks for its results
func (i InOut) SendRecv(data interface{}) (interface{}, error) {
	resp := i.Send(data, nil)
	if err := resp.Get(); err != nil {
		return nil, fmt.Errorf("SendRecv() received error during Get(): %v", err)
	}

	return resp.Data, resp.Err
}

// Send places the specified data onto the buffered In channel
func (i InOut) Send(data, context interface{}) *Response {
	request := newRequest(data, &i)

	var response *Response
	if i.WantResponse == WantResponse {
		response = newResponse(context, data, request.c)
	}

	i.m.c <- request

	return response
}
