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
	tests := []struct {
		name     string
		args     args
		want     ExitChannel
		wantErr  bool
		sendReq  bool
		runChan  int
		runParam interface{}
		runWant  interface{}
		runErr   bool
	}{
		{
			name: "No channels",
			args: args{
				channels: nil,
				config:   nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "No In chan",
			args: args{
				channels: []InOut{
					{
						In:  nil,
						Fn:  nil,
						Out: nil,
					},
				},
				config: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Wrong type for In chan",
			args: args{
				channels: []InOut{
					{
						In:  []byte{},
						Fn:  nil,
						Out: nil,
					},
				},
				config: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "No Fn specified",
			args: args{
				channels: []InOut{
					{
						In:  make(chan bool),
						Fn:  nil,
						Out: nil,
					},
				},
				config: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "No Out should be fine",
			args: args{
				channels: []InOut{
					{
						In:  make(chan bool),
						Fn:  func(interface{}) (interface{}, error) { return nil, nil },
						Out: nil,
					},
				},
				exitChannel: exitChannel,
				config:      nil,
			},
			want:    exitChannel,
			wantErr: false,
		},
		{
			name: "Reflection",
			args: args{
				channels: []InOut{
					{
						In:  make(chan bool),
						Fn:  func(i interface{}) (interface{}, error) { return i, nil },
						Out: make(chan *Response),
					},
				},
				exitChannel: exitChannel,
				config:      nil,
			},
			want:     exitChannel,
			wantErr:  false,
			sendReq:  true,
			runChan:  0,
			runParam: true,
			runWant:  true,
			runErr:   false,
		},
		{
			name: "Panic in function",
			args: args{
				channels: []InOut{
					{
						In:  make(chan bool),
						Fn:  func(i interface{}) (interface{}, error) { panic("Boom") },
						Out: make(chan *Response),
					},
				},
				exitChannel: exitChannel,
				config:      nil,
			},
			want:     exitChannel,
			wantErr:  false,
			sendReq:  true,
			runChan:  0,
			runParam: true,
			runWant:  nil,
			runErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.channels, tt.args.exitChannel, tt.args.config)
			defer func() {
				if got != nil {
					got <- Exit
				}
			}()
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
				return
			}
			if got != nil && tt.sendReq {
				tt.args.channels[tt.runChan].Send(tt.runParam)
				if tt.args.channels[tt.runChan].Out != nil {
					response := <-tt.args.channels[tt.runChan].Out
					if (response.Err != nil) != tt.runErr {
						t.Errorf("Channel submission error = %v, runErr %v", response.Err, tt.runErr)
						return
					}
					if !reflect.DeepEqual(response.Data, tt.runWant) {
						t.Errorf("Channel submission error, got = %v, want %v", response.Data, tt.runWant)
						return
					}
				}
			}
		})
	}
}
