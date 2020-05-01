package chanmgr

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
)

// InOut provides the details a single function that is to be triggered via a channel and responds via a second
type InOut struct {
	In  interface{}                            // Entries from In will be passed to the Fn
	Out chan *Response                         // If Out is nil the return from Fn is discarded, otherwise the response is put on the Out chan
	Fn  func(interface{}) (interface{}, error) // The function to be invoked with messages from In
}

// Send places the specified data onto the In channel
func (i InOut) Send(data interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	reflect.ValueOf(i.In).Send(reflect.ValueOf(data))
	return err
}

// Response wraps the returned values from a function call
type Response struct {
	Data interface{} // Response from InOut.Fn()
	Err  error       // Error raised by InOut.Fn(), or details of panic if Fn() panicked whilst processing
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
	if config == nil {
		config = defaultConfig
	}
	if exitChannel == nil {
		exitChannel = make(ExitChannel)
	}

	// Take a copy so that the manager's processing is immutable
	chans := make([]InOut, len(channels))
	for i := range channels {
		chans[i].In, chans[i].Fn, chans[i].Out = channels[i].In, channels[i].Fn, channels[i].Out
	}

	// Launch a manager, and return it's exit channel so it can be terminated
	mgr := &manager{config: config, chans: chans, exit: exitChannel}
	if err := mgr.validate(); err != nil {
		return nil, err
	}
	go mgr.run()
	return mgr.exit, nil
}

// manager abstracts the complexity of channel processing
type manager struct {
	config *Config
	chans  []InOut
	exit   chan ExitType
}

// validate checks for invalid InOut combinations
func (m *manager) validate() error {
	if len(m.chans) == 0 {
		return errors.New("channels must contain at least one InOut instance")
	}

	for i, v := range m.chans {
		if v.In == nil {
			return fmt.Errorf("Invalid channels (%d): In channel must not be nil", i)
		}
		if reflect.ValueOf(v.In).Kind() != reflect.Chan {
			return fmt.Errorf("Invalid channels (%d): In channel must be of type chan, got %v", i, reflect.ValueOf(v.In).Kind())
		}
		if v.Fn == nil {
			return fmt.Errorf("Invalid channels (%d): Fn must not be nil", i)
		}
	}
	return nil
}

// run handles the dispatching of In messages to respective functions and then sends replies via the Out channel
func (m *manager) run() {
	cases := make([]reflect.SelectCase, len(m.chans)+1)
	for i, ch := range m.chans {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch.In)}
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
			if channel.Out != nil {
				m.config.Log.Printf("ERROR In channel[%d], processing generated a panic: '%s'\n", chosen, r)
				channel.Out <- &Response{Err: fmt.Errorf("%s", r)}
			}
		}
	}()

	var response Response
	response.Data, response.Err = channel.Fn(value)

	if channel.Out != nil {
		channel.Out <- &response
	}
}
