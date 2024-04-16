package param

import (
	"reflect"
	"testing"
)

func TestParams_GetString(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		p    Params
		args args
		want string
	}{
		{name: "t1", p: map[string]interface{}{"key": "value"}, args: args{
			"key",
		}, want: "value"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.GetString(tt.args.key); got != tt.want {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParams_GetInt64(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		p    Params
		args args
		want int64
	}{
		{name: "t1", p: map[string]interface{}{"k1": int64(1)}, args: args{
			key: "k1",
		}, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.GetInt64(tt.args.key); got != tt.want {
				t.Errorf("GetInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParams_GetFloat64(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		p    Params
		args args
		want float64
	}{
		{name: "t1", p: map[string]interface{}{"k1": float64(1)}, args: args{
			key: "k1",
		}, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.GetFloat64(tt.args.key); got != tt.want {
				t.Errorf("GetFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParams_GetParams(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		p    Params
		args args
		want float64
	}{
		{name: "t1", p: map[string]interface{}{
			"k1": map[string]interface{}{"k2": float64(1)},
		}, args: args{
			key: "k1",
		}, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Get(tt.args.key).GetFloat64("k2"); got != tt.want {
				t.Errorf("GetFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildExpParams(t *testing.T) {
	type args struct {
		m map[string]int64
	}
	tests := []struct {
		name string
		args args
		want *Params
	}{
		{name: "TestBuildExpParams1",
			args: args{
				m: map[string]int64{"k1": 100, "k2": 200},
			},
			want: &Params{"k1": int64(100), "k2": int64(200)}},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if got := BuildMapStrInt64(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildMapStrInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildExpInfo(t *testing.T) {
	type args struct {
		m map[string]int64
	}
	tests := []struct {
		name string
		args args
		want *Params
	}{
		{name: "TestBuildExpParams1",
			args: args{
				m: map[string]int64{"k1": 100, "k2": 200},
			},
			want: &Params{"EXP": Params{"k1": int64(100), "k2": int64(200)}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildExpInfo(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildExpInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
