package processor

import (
	"context"
	"reflect"
	"testing"

	"xxxx/dagengine/engine/param"
)

type phase0 struct {
	Input  int `graph:"input"`
	Output int `graph:"output"`
}

func (p *phase0) OnInit() {
}

func (p *phase0) OnExecute(_ context.Context, params *param.Params) error {
	return nil
}

type phase1 struct {
	MultiInput1 int `graph:"multi_input"`
	Output1     int `graph:"output"`
}

func (p *phase1) OnInit() {
}

func (p *phase1) OnExecute(_ context.Context, params *param.Params) error {
	return nil
}

func TestGenerateMetas(t *testing.T) {
	tests := []struct {
		name string
		want []OperatorMeta
	}{
		{name: "t0", want: []OperatorMeta{
			{Name: "phase0",
				Input:  []FieldMeta{{Name: "Input"}},
				Output: []FieldMeta{{Name: "Output"}}},
			{Name: "phase1",
				Input:  []FieldMeta{{Name: "MultiInput1", Flags: FieldFlags{Aggregate: 1}}},
				Output: []FieldMeta{{Name: "Output1"}}},
		}},
	}
	Register("phase0", func() Processor { return &phase0{} })
	Register("phase1", func() Processor { return &phase1{} })
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateMetas(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateMetas() = %v, want %v", got, tt.want)
			}
		})
	}
}
