package graph

import "testing"

func TestAddEvent(t *testing.T) {
	type args struct {
		e *Event
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "TestAddEvent1",
			args: args{
				e: &Event{
					Processor: "pp1",
					Duration:  100,
				},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddEvent(tt.args.e)
		})
	}
}
