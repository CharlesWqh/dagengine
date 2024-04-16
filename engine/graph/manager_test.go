package graph

import (
	"context"
	"log"
	"reflect"
	"sort"
	"testing"

	"xxxx/dagengine/engine/param"
	"xxxx/dagengine/engine/processor"
	"xxxx/innererror"
)

type Mid struct {
	Name string
}

type phase0 struct {
	REQ *testReq `graph:"extern_input"`
	Mid []Mid    `graph:"output"`
}

func (p *phase0) OnInit() {
}

func (p *phase0) OnExecute(_ context.Context, params *param.Params) error {
	p.Mid = append(p.Mid, Mid{})
	p.Mid[0].Name = params.GetString("name")
	p.REQ.name = p.Mid[0].Name
	p.REQ.strs[0] = p.Mid[0].Name
	p.REQ.id[0] = int(params.GetInt64("id"))
	return innererror.Error(int32(params.GetInt64("id")))
}

type s struct {
	i int
}

type phase1 struct {
	REQ *testReq `graph:"extern_input"`
	Mid []Mid    `graph:"input"`
	ID  *s       `graph:"output"`
}

func (p *phase1) OnInit() {
}

func (p *phase1) OnExecute(_ context.Context, params *param.Params) error {
	if len(p.Mid) > 0 {
		p.REQ.name = p.Mid[0].Name
		p.REQ.strs[0] = p.Mid[0].Name
		p.REQ.id[0] = int(params.GetInt64("id"))
	} else {
		p.REQ.name = params.GetString("name")
		p.REQ.strs[0] = params.GetString("name")
		p.REQ.id[0] = int(params.GetInt64("id"))
	}
	p.ID.i = int(params.GetInt64("id"))
	return nil
}

type phase2 struct {
	REQ *testReq      `graph:"extern_input"`
	IDs map[string]*s `graph:"multi_input"`
	OK  bool          `graph:"output"`
	i   int
}

func (p *phase2) OnInit() {
	p.i = 1
}

func (p *phase2) OnExecute(_ context.Context, params *param.Params) error {
	p.REQ.name = ""
	p.REQ.strs = nil
	p.REQ.id = nil
	for _, v := range p.IDs {
		p.REQ.id = append(p.REQ.id, v.i)
	}
	sort.Slice(p.REQ.id, func(i, j int) bool {
		return p.REQ.id[i] < p.REQ.id[j]
	})
	return nil
}

type testReq struct {
	name string
	id   []int
	strs []string
}

type phase3 struct {
	REQ *testReq `graph:"extern_input"`
	OK  bool     `graph:"output"`
}

func (p *phase3) OnInit() {
}

func (p *phase3) OnExecute(_ context.Context, params *param.Params) error {
	p.REQ.name = "p3"
	p.REQ.id[1] = 4
	p.REQ.strs[0] = "s1"
	return nil
}

type phase4 struct {
	REQ *testReq `graph:"extern_input"`
	Mid []Mid    `graph:"output"`
	OK  bool     `graph:"input"`
}

func (p *phase4) OnInit() {
}

func (p *phase4) OnExecute(_ context.Context, params *param.Params) error {
	return nil
}

type phase5 struct {
	REQ *testReq `graph:"extern_input"`
	ID  *s       `graph:"input"`
	OK  bool     `graph:"input"`
}

func (p *phase5) OnInit() {
}

func (p *phase5) OnExecute(_ context.Context, params *param.Params) error {
	return nil
}

func TestManager_Execute(t *testing.T) {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	type fields struct {
		clusters map[string]*Cluster
	}
	type args struct {
		ctx         context.Context
		dataContext *DataContext
		clusterName string
		graphName   string
		params      *param.Params
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want1   testReq
	}{
		{name: "test_extern_input",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "extern_input_test.toml",
				graphName:   "enter",
				params:      nil,
			},
			wantErr: false,
			want1:   testReq{name: "p3", id: []int{1, 4, 3}, strs: []string{"s1", "s1", "s2"}}},
		{name: "test_expect&expect_config",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "expect_expect_config_test.toml",
				graphName:   "enter",
				params:      &param.Params{"EXP": 101},
			},
			wantErr: false,
			want1:   testReq{name: "v12", id: []int{12, 2, 3}, strs: []string{"v12", "s1", "s2"}}},
		{name: "test_expect&expect_config1",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "expect_expect_config_test.toml",
				graphName:   "enter",
				params:      &param.Params{"EXP": 102},
			},
			wantErr: false,
			want1:   testReq{name: "v1", id: []int{11, 2, 3}, strs: []string{"v1", "s1", "s2"}}},
		{name: "test_expect&expect_config2",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "expect_expect_config_test.toml",
				graphName:   "enter",
				params:      &param.Params{"EXP": 103},
			},
			wantErr: false,
			want1:   testReq{name: "v13", id: []int{13, 2, 3}, strs: []string{"v13", "s1", "s2"}}},
		{name: "test_expect&expect_config3",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "expect_expect_config_test.toml",
				graphName:   "enter",
				params:      nil,
			},
			wantErr: false,
			want1:   testReq{name: "v13", id: []int{13, 2, 3}, strs: []string{"v13", "s1", "s2"}}},
		{name: "test_dep1",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "dep_test.toml",
				graphName:   "enter",
				params:      &param.Params{"EXP": 101},
			},
			wantErr: false,
			want1:   testReq{name: "v01", id: []int{11, 2, 3}, strs: []string{"v01", "s1", "s2"}}},
		{name: "test_dep2",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "dep_test.toml",
				graphName:   "enter",
				params:      &param.Params{"EXP": 102},
			},
			wantErr: false,
			want1:   testReq{name: "v00", id: []int{10, 2, 3}, strs: []string{"v00", "s1", "s2"}}},
		{name: "test_dep3",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "dep_test.toml",
				graphName:   "enter",
				params:      &param.Params{"EXP": 103},
			},
			wantErr: false,
			want1:   testReq{name: "v02", id: []int{12, 2, 3}, strs: []string{"v02", "s1", "s2"}}},
		{name: "test_select_args1",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "select_args_test.toml",
				graphName:   "enter",
				params:      &param.Params{"EXP": 101},
			},
			wantErr: false,
			want1:   testReq{name: "v2", id: []int{2, 2, 3}, strs: []string{"v2", "s1", "s2"}}},
		{name: "test_select_args2",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "select_args_test.toml",
				graphName:   "enter",
				params:      &param.Params{"EXP": 100},
			},
			wantErr: false,
			want1:   testReq{name: "v1", id: []int{1, 2, 3}, strs: []string{"v1", "s1", "s2"}}},
		{name: "test_select_args3",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "select_args_test.toml",
				graphName:   "enter",
				params:      &param.Params{"EXP": 103},
			},
			wantErr: false,
			want1:   testReq{name: "v0", id: []int{0, 2, 3}, strs: []string{"v0", "s1", "s2"}}},
		{name: "test_subgraph1",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "subgraph_test.toml",
				graphName:   "enter",
				params:      &param.Params{"EXP": 10000},
			},
			wantErr: false,
			want1:   testReq{name: "p1", id: []int{100, 2, 3}, strs: []string{"p1", "s1", "s2"}}},
		{name: "test_json_format1",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "json_format_test.json",
				graphName:   "enter",
				params:      nil,
			},
			wantErr: false,
			want1:   testReq{name: "", id: []int{0, 2, 3}, strs: []string{"", "s1", "s2"}}},
	}
	processor.Register("phase0", func() processor.Processor { return &phase0{} })
	processor.Register("phase1", func() processor.Processor { return &phase1{} })
	processor.Register("phase2", func() processor.Processor { return &phase2{} })
	processor.Register("phase3", func() processor.Processor { return &phase3{} })
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filepath := "../../cmd/" + tt.args.clusterName
			if err := LoadFile(filepath); (err != nil) != tt.wantErr {
				t.Fatalf("Manager.LoadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			ts := &testReq{name: "ts", id: []int{1, 2, 3}, strs: []string{"s0", "s1", "s2"}}
			if tt.args.dataContext != nil {
				var midi interface{} = ts
				tt.args.dataContext.Set(NewDIObjectKey("REQ", reflect.TypeOf(ts)), reflect.ValueOf(midi))
			}
			if err := Execute(tt.args.ctx, tt.args.clusterName, tt.args.graphName,
				tt.args.dataContext, tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("Manager.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(*ts, tt.want1) {
				t.Errorf("Manager.Execute() extern input = %v, want1 %v", *ts, tt.want1)
			}
		})
	}
}

func TestManager_Execute1(t *testing.T) {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	type fields struct {
		clusters map[string]*Cluster
	}
	type args struct {
		ctx         context.Context
		dataContext *DataContext
		clusterName string
		graphName   string
		params      *param.Params
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want1   testReq
	}{
		{name: "test_json_format1",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "json_format_test.json",
				graphName:   "enter",
				params:      nil,
			},
			wantErr: false,
			want1:   testReq{name: "", id: []int{0, 2, 3}, strs: []string{"", "s1", "s2"}}},
		{name: "dep_ret_code_test1",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "dep_ret_code_test.toml",
				graphName:   "enter",
				params:      &param.Params{"EXP": 100},
			},
			wantErr: false,
			want1:   testReq{name: "v1", id: []int{10, 2, 3}, strs: []string{"v1", "s1", "s2"}}},
		{name: "dep_ret_code_test2",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "dep_ret_code_test.toml",
				graphName:   "enter",
				params:      &param.Params{"EXP": 101},
			},
			wantErr: false,
			want1:   testReq{name: "v2", id: []int{11, 2, 3}, strs: []string{"v2", "s1", "s2"}}},
		{name: "dep_ret_code_test3",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "dep_ret_code_test.toml",
				graphName:   "enter",
				params:      &param.Params{"EXP": 102},
			},
			wantErr: false,
			want1:   testReq{name: "v00", id: []int{12, 2, 3}, strs: []string{"v00", "s1", "s2"}}},
		{name: "aggregate_test",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "aggregate_test.toml",
				graphName:   "enter",
				params:      &param.Params{"EXP": 102},
			},
			wantErr: false,
			want1:   testReq{name: "", id: []int{0, 11}, strs: nil}},
		{name: "circle_test",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "circle.toml",
				graphName:   "enter",
				params:      &param.Params{"EXP": 102},
			},
			wantErr: true,
			want1:   testReq{name: "", id: []int{0, 11}, strs: nil}},
		{name: "optional_input_test",
			fields: fields{clusters: make(map[string]*Cluster)},
			args: args{
				ctx:         context.Background(),
				dataContext: NewDataContext(),
				clusterName: "optional_input_test.toml",
				graphName:   "enter",
				params:      &param.Params{"EXP": 102},
			},
			wantErr: false,
			want1:   testReq{name: "p3", id: []int{1, 4, 3}, strs: []string{"s1", "s1", "s2"}}},
	}
	processor.Register("phase0", func() processor.Processor { return &phase0{} })
	processor.Register("phase1", func() processor.Processor { return &phase1{} })
	processor.Register("phase2", func() processor.Processor { return &phase2{} })
	processor.Register("phase3", func() processor.Processor { return &phase3{} })
	processor.Register("phase4", func() processor.Processor { return &phase4{} })
	processor.Register("phase5", func() processor.Processor { return &phase5{} })
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filepath := "../../cmd/" + tt.args.clusterName
			err := LoadFile(filepath)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Manager.LoadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			ts := &testReq{name: "ts", id: []int{1, 2, 3}, strs: []string{"s0", "s1", "s2"}}
			if tt.args.dataContext != nil {
				var midi interface{} = ts
				tt.args.dataContext.Set(NewDIObjectKey("REQ", reflect.TypeOf(ts)), reflect.ValueOf(midi))
			}
			if err := Execute(tt.args.ctx, tt.args.clusterName, tt.args.graphName,
				tt.args.dataContext, tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("Manager.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(*ts, tt.want1) {
				t.Errorf("Manager.Execute() extern input = %v, want1 %v", *ts, tt.want1)
			}
		})
	}
}
