package engine

import (
	"testing"

	"xxxx/dagengine/engine/graph"
	"xxxx/dagengine/engine/processor"
)

func TestDAGConfig_loadTomlScriptFile(t *testing.T) {
	type fields struct {
		opMeta     []processor.OperatorMeta
		graph      graph.Cluster
		scriptPath string
	}
	type args struct {
		tomlScript string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "t1", fields: fields{opMeta: []processor.OperatorMeta{
			{Name: "phase3"},
		}}, args: args{
			tomlScript: "../cmd/example8.toml",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &DAGConfig{
				opMeta:     tt.fields.opMeta,
				graph:      tt.fields.graph,
				scriptPath: tt.fields.scriptPath,
			}
			if err := p.loadTomlScriptFile(tt.args.tomlScript); (err != nil) != tt.wantErr {
				t.Errorf("loadTomlScriptFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
