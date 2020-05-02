package chanmgr

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	exitChannel := make(ExitChannel)
	type args struct {
		channels    []*InOut
		exitChannel ExitChannel
		config      *Config
	}
	type argsGen func() *args
	tests := []struct {
		name          string
		args          argsGen
		want          ExitChannel
		wantPanic     bool
		wantErr       bool
		sendReq       bool
		runChan       int
		runParam      interface{}
		sendErr       bool
		respWant      interface{}
		respErr       bool
		processingErr bool
	}{
		{
			name: "No channels",
			args: func() *args {
				return &args{
					channels: nil,
					config:   nil,
				}
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "No Processor specified",
			args: func() *args {
				return &args{
					channels: []*InOut{&InOut{Processor: nil, WantResponse: WantResponse}},
					config:   nil,
				}
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "No Out should be fine",
			args: func() *args {
				return &args{
					channels: []*InOut{&InOut{
						Processor:    func(interface{}) (interface{}, error) { return nil, nil },
						WantResponse: IgnoreResponse}},
					exitChannel: exitChannel,
					config:      nil,
				}
			},
			want: exitChannel,
		},
		{
			name: "Reflection",
			args: func() *args {
				return &args{
					channels: []*InOut{&InOut{
						Processor:    func(i interface{}) (interface{}, error) { return i, nil },
						WantResponse: WantResponse}},
					exitChannel: exitChannel,
					config:      nil,
				}
			},
			want:     exitChannel,
			sendReq:  true,
			runChan:  0,
			runParam: true,
			respWant: true,
		},
		{
			name: "Wrong parameter type supplied",
			args: func() *args {
				return &args{
					channels: []*InOut{&InOut{
						Processor:    func(i interface{}) (interface{}, error) { return i.(int) + 1, nil },
						WantResponse: WantResponse}},
					exitChannel: exitChannel,
					config:      nil,
				}
			},
			want:          exitChannel,
			sendReq:       true,
			runChan:       0,
			runParam:      "Boom",
			processingErr: true,
		},
		{
			name: "Panic in function",
			args: func() *args {
				return &args{
					channels: []*InOut{&InOut{
						Processor:    func(interface{}) (interface{}, error) { panic("Boom") },
						WantResponse: WantResponse}},
					exitChannel: exitChannel,
					config:      nil,
				}
			},
			want:          exitChannel,
			sendReq:       true,
			runChan:       0,
			runParam:      true,
			respWant:      nil,
			processingErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argsCheck := func(f argsGen) (args *args, panicked bool) {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
				args = f()
				return args, panicked
			}
			// First test is to validate whether channels are created correctly
			args, panicked := argsCheck(tt.args)
			if panicked != tt.wantPanic {
				t.Errorf("Create args: returned %v, wantPanic = %v", panicked, tt.wantPanic)
				return
			}
			if tt.wantPanic {
				return // That was the test
			}

			// Second test is that the chanmgr is correctly created
			got, err := New(args.channels, args.exitChannel, args.config)
			defer func() {
				if got != nil {
					got <- Exit
				}
			}()
			if (err != nil) != tt.wantErr {
				t.Errorf("New() returned %v, wantErr = %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
				return
			}

			// Last test is to validate the channels are working as expected
			if got != nil && tt.sendReq {
				resp := args.channels[tt.runChan].Send(tt.runParam, nil)
				if (resp != nil && resp.Err != nil) != tt.sendErr {
					t.Errorf("Channel submission error = %v, sendErr %v", resp.Err, tt.sendErr)
					return
				}
				if tt.sendErr {
					return // End of test
				}

				if args.channels[tt.runChan].WantResponse == WantResponse {
					err := resp.Get()
					if (err != nil) != tt.respErr {
						t.Errorf("Response retrieval error = %v, respErr %v", err, tt.respErr)
						return
					}
					if (resp.Err != nil) != tt.processingErr {
						t.Errorf("Response processing error = %v, processingErr %v", err, tt.processingErr)
						return
					}
					if !reflect.DeepEqual(resp.Data, tt.respWant) {
						t.Errorf("Channel submission error, got = %v, want %v", resp, tt.respWant)
						return
					}
				}
			}
		})
	}
}
