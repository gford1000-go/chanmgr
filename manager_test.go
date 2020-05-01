package chanmgr

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	exitChannel := make(ExitChannel)
	type args struct {
		channels    []InOut
		exitChannel ExitChannel
		config      *Config
	}
	type argsGen func() *args
	tests := []struct {
		name      string
		args      argsGen
		want      ExitChannel
		wantPanic bool
		wantErr   bool
		sendReq   bool
		runChan   int
		runParam  interface{}
		sendErr   bool
		respWant  interface{}
		respErr   bool
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
			name: "No In chan",
			args: func() *args {
				return &args{
					channels: []InOut{CreateInOut(nil, nil, WantResponse, 1)},
					config:   nil,
				}
			},
			want:      nil,
			wantPanic: true,
		},
		{
			name: "Wrong type for In chan",
			args: func() *args {
				return &args{
					channels: []InOut{CreateInOut([]byte{}, nil, WantResponse, 1)},
					config:   nil,
				}
			},
			want:      nil,
			wantPanic: true,
		},
		{
			name: "No Fn specified",
			args: func() *args {
				return &args{
					channels: []InOut{CreateInOut(make(chan bool), nil, WantResponse, 1)},
					config:   nil,
				}
			},
			want:      nil,
			wantPanic: true,
		},
		{
			name: "Invalid channel direction",
			args: func() *args {
				return &args{
					channels: []InOut{CreateInOut(make(<-chan bool), nil, WantResponse, 1)},
					config:   nil,
				}
			},
			want:      nil,
			wantPanic: true,
		},
		{
			name: "No Out should be fine",
			args: func() *args {
				return &args{
					channels: []InOut{
						CreateInOut(
							make(chan bool),
							func(interface{}) (interface{}, error) { return nil, nil },
							IgnoreResponse, 1)},
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
					channels: []InOut{
						CreateInOut(
							make(chan bool),
							func(i interface{}) (interface{}, error) { return i, nil },
							WantResponse, 1)},
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
					channels: []InOut{
						CreateInOut(
							make(chan bool),
							func(i interface{}) (interface{}, error) { return i, nil },
							WantResponse, 1)},
					exitChannel: exitChannel,
					config:      nil,
				}
			},
			want:     exitChannel,
			sendReq:  true,
			runChan:  0,
			runParam: 2,
			sendErr:  true,
		},
		{
			name: "Panic in function",
			args: func() *args {
				return &args{
					channels: []InOut{
						CreateInOut(
							make(chan bool),
							func(interface{}) (interface{}, error) { panic("Boom") },
							WantResponse, 1),
					},
					exitChannel: exitChannel,
					config:      nil,
				}
			},
			want:     exitChannel,
			sendReq:  true,
			runChan:  0,
			runParam: true,
			respWant: nil,
			respErr:  true,
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
				sendErr := args.channels[tt.runChan].Send(tt.runParam)
				if (sendErr != nil) != tt.sendErr {
					t.Errorf("Channel submission error = %v, sendErr %v", sendErr, tt.sendErr)
					return
				}
				if tt.sendErr {
					return // End of test
				}

				if args.channels[tt.runChan].CanRecv() {
					response, err := args.channels[tt.runChan].Recv()
					if (err != nil) != tt.respErr {
						t.Errorf("Channel submission error = %v, respErr %v", err, tt.respErr)
						return
					}
					if !reflect.DeepEqual(response, tt.respWant) {
						t.Errorf("Channel submission error, got = %v, want %v", response, tt.respWant)
						return
					}
				}
			}
		})
	}
}
