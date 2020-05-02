package chanmgr

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
)

// ExitType is the type of the exit flag to the manager
type ExitType bool

// ExitChannel is the channel type to which to send the exit flag
type ExitChannel chan ExitType

// Exit is the flag to pass to the ExitChannel returned by New(), in order to stop listening on the channels
var Exit ExitType = true

// Config modifies the behaviour of the manager
type Config struct {
	Log           *log.Logger // Messages are logged to the supplied logger instance
	RequestBuffer int         // The size of the manager's request buffer
}

// defaultConfig is used if New() receives a nil config
var defaultConfig = &Config{
	Log:           log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile),
	RequestBuffer: 1,
}

// New creates a manager instance that monitors for work in the supplied channels and processes accordingly.
// To terminate, send Exit to the returned ExitChannel.  This will be created if exitChannel=nil
func New(channels []*InOut, exitChannel ExitChannel, config *Config) (ExitChannel, error) {
	if len(channels) == 0 {
		return nil, errors.New("channels must not be empty")
	}
	for i, c := range channels {
		if c.m != nil {
			return nil, errors.New("channels must only be associated with one manager")
		}
		if c.Processor == nil {
			return nil, fmt.Errorf("channel[%d].Processor must not be nil", i)
		}
	}
	if exitChannel == nil {
		exitChannel = make(ExitChannel)
	}
	if config == nil {
		config = defaultConfig
	}
	if config.RequestBuffer < 1 || config.RequestBuffer > 10000 {
		return nil, errors.New("config.RequestBuffer must be greater than 1 and less than 10,000")
	}

	// Launch a manager, and return it's exit channel so it can be terminated
	mgr := &manager{config: config, c: make(chan *request, config.RequestBuffer), exit: exitChannel}
	go mgr.run()

	// Ensure InOut can issue requests to manager
	for _, c := range channels {
		c.m = mgr
	}

	return mgr.exit, nil
}

// manager abstracts the complexity of channel processing
type manager struct {
	config *Config
	c      chan *request
	exit   chan ExitType
}

// run handles the dispatching of In messages to respective functions and then sends replies via the Out channel
func (m *manager) run() {
	defer func() {
		if r := recover(); r != nil {
			m.config.Log.Printf("DEBUG Manager.run() panicked: %v", r)
		}
		m.config.Log.Print("DEBUG Manager.run() is exiting")
	}()
	m.config.Log.Print("DEBUG Manager.run() has started")

	for {
		select {
		case _, ok := <-m.exit:
			if !ok {
				m.config.Log.Print("WARNING Manager received bad exit message - ignoring")
				continue
			}
			return
		case request, ok := <-m.c:
			{
				if !ok {
					m.config.Log.Print("WARNING Manager received bad request - ignoring")
					continue
				}
				m.process(request)
			}
		}
	}
}

// process allows the manager to catch a panic
// request.c should never block, allowing the manager to move to the next request
func (m *manager) process(request *request) {
	defer func() {
		if r := recover(); r != nil {
			m.config.Log.Printf("ERROR Processing request generated a panic: %s\n", r)
			if request.c != nil {
				request.c <- &Response{Err: fmt.Errorf("%s", r)}
			}
		}
	}()

	var resp Response
	resp.Data, resp.Err = request.p(request.data)

	if request.c != nil {
		request.c <- &resp
	}
}
