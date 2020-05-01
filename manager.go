package chanmgr

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
)

// Processor is the type of function that handles supplied messages
type Processor func(interface{}) (interface{}, error)

// response wraps the returned values from a function call
type response struct {
	data interface{} // Response from InOut.Fn()
	err  error       // Error raised by InOut.Fn(), or details of panic if Fn() panicked whilst processing
}

// Response indicated whether the outcome of Processor are of interest
type Response bool

// WantResponse indicates that the results of processing should be returned (must then be consumed to avoid block)
var WantResponse Response = true

// IgnoreResponse indicates that any results from processing should be discarded
var IgnoreResponse Response = false

// CreateInOut constructs an instance of InOut.  sendChan should be an instance of
// a channel which can be buffered or unbuffered.
func CreateInOut(sendChan interface{}, p Processor, wantOrIgnore Response, respBufferSize int) InOut {
	if sendChan == nil {
		panic("sendChan must not be nil")
	}

	v := reflect.ValueOf(sendChan)
	t := v.Type()
	if v.Kind() != reflect.Chan {
		panic(fmt.Sprintf("sendChan must be of type chan, got %v", t))
	}
	if t.ChanDir() == reflect.RecvDir {
		panic("sendChan must be able to send data")
	}

	if p == nil {
		panic("p must not be nil")
	}

	var retChan chan *response
	if wantOrIgnore == WantResponse {
		if respBufferSize < 1 || respBufferSize > 10000 {
			panic("respBufferSize must be between 1 and 10,000")
		}
		retChan = make(chan *response, respBufferSize)
	}

	return InOut{
		fn:         p,
		in:         v,
		inChanType: t.Elem(),
		ret:        retChan,
	}
}

// InOut provides the details a single function that is to be triggered via a channel and responds via a second
type InOut struct {
	in         reflect.Value  // Entries from In will be passed to the Fn
	inChanType reflect.Type   // The type of the channel
	inCap      int            // The size of the buffer of the channel
	ret        chan *response // If Out is nil the return from Fn is discarded, otherwise the response is put on the Out chan
	fn         Processor      // The function to be invoked with messages from In
}

// SendRecv places the specified data onto the In channel, and returns processing response
func (i InOut) SendRecv(data interface{}) (interface{}, error) {
	err := i.Send(data)
	if err != nil {
		return nil, err
	}
	return i.Recv()
}

// Send places the specified data onto the buffered In channel
func (i InOut) Send(data interface{}) error {
	v := reflect.ValueOf(data)
	if v.Type() != i.inChanType {
		return fmt.Errorf("sendChan expected data of type '%s', got '%s'", i.inChanType, v.Type())
	}

	i.in.Send(v) // Will block if cap == 1, or buffer exhausted, until reader has consumed a message
	return nil
}

// CanRecv indicates if responses can be received
func (i InOut) CanRecv() bool {
	return i.ret != nil
}

// Recv returns data that has been returned by the Processor
func (i InOut) Recv() (interface{}, error) {
	if i.ret == nil {
		return nil, fmt.Errorf("Attempting to receive when response not requested")
	}

	response := <-i.ret // Will block until message is added
	return response.data, response.err
}

// ExitType is the type of the exit flag to the manager
type ExitType bool

// ExitChannel is the channel type to which to send the exit flag
type ExitChannel chan ExitType

// Exit is the flag to pass to the ExitChannel returned by New(), in order to stop listening on the channels
var Exit ExitType = true

// Config modifies the behaviour of the manager
type Config struct {
	Log *log.Logger // Messages are logged to the supplied logger instance
}

// defaultConfig is used if New() receives a nil config
var defaultConfig = &Config{
	Log: log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile),
}

// New creates a manager instance that monitors for work in the supplied channels and processes accordingly.
// To terminate, send Exit to the returned ExitChannel.  This will be created if exitChannel=nil
func New(channels []InOut, exitChannel ExitChannel, config *Config) (ExitChannel, error) {
	if len(channels) == 0 {
		return nil, errors.New("channels must not be empty")
	}

	if config == nil {
		config = defaultConfig
	}

	if exitChannel == nil {
		exitChannel = make(ExitChannel)
	}

	// Take a copy so that the manager's processing is immutable
	chans := make([]InOut, len(channels))
	for i := range channels {
		chans[i].in, chans[i].fn, chans[i].ret = channels[i].in, channels[i].fn, channels[i].ret
	}

	// Launch a manager, and return it's exit channel so it can be terminated
	mgr := &manager{config: config, chans: chans, exit: exitChannel}
	go mgr.run()
	return mgr.exit, nil
}

// manager abstracts the complexity of channel processing
type manager struct {
	config *Config
	chans  []InOut
	exit   chan ExitType
}

// run handles the dispatching of In messages to respective functions and then sends replies via the Out channel
func (m *manager) run() {
	cases := make([]reflect.SelectCase, len(m.chans)+1)
	for i, ch := range m.chans {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: ch.in}
	}
	cases[len(m.chans)] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(m.exit)}

	for {
		chosen, value, ok := reflect.Select(cases)
		if !ok {
			m.config.Log.Printf("Channel[%d] received a bad value - ignored", chosen)
			continue
		}

		switch chosen {
		case len(cases) - 1:
			m.config.Log.Print("DEBUG Manager exiting")
			return // Exiting
		default:
			m.config.Log.Printf("DEBUG In channel[%d], received: '%s'\n", chosen, value)
			m.process(chosen, m.chans[chosen], value.Interface())
		}
	}
}

// process allows the manager to catch a panic
func (m *manager) process(chosen int, channel InOut, value interface{}) {
	defer func() {
		if r := recover(); r != nil {
			if channel.ret != nil {
				m.config.Log.Printf("ERROR In channel[%d], processing generated a panic: '%s'\n", chosen, r)
				channel.ret <- &response{err: fmt.Errorf("%s", r)}
			}
		}
	}()

	var resp response
	resp.data, resp.err = channel.fn(value)

	if channel.ret != nil {
		channel.ret <- &resp
	}
}
