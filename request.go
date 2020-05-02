package chanmgr

// request instances are processed by manager
type request struct {
	c    chan *Response // The channel on which to send Response to processing
	data interface{}    // The input for processing
	p    Processor      // The Processor to be used
}

// newRequest constructs a new request
func newRequest(data interface{}, i *InOut) *request {
	var c chan *Response
	if i.WantResponse == WantResponse {
		c = make(chan *Response)
	}
	return &request{
		c:    c,
		data: data,
		p:    i.Processor,
	}
}
