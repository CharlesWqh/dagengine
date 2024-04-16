package processor

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"sync"

	"xxxx/dagengine/engine/param"
)

// Processor Execute the unified abstraction of operators.
// Operators need to implement this interface,
// and explicitly call the Register interface to register
type Processor interface {
	OnInit()
	OnExecute(ctx context.Context, params *param.Params) error
}

// Creator 创建processor的函数签名
type Creator func() Processor

var (
	processors = make(map[string]Creator)
	lock       sync.RWMutex
)

// Register register Processor implement
func Register(name string, p Creator) {
	lock.Lock()
	processors[name] = p
	lock.Unlock()
}

// Get get processor。
func Get(name string) Processor {
	lock.RLock()
	p := processors[name]
	lock.RUnlock()
	return p()
}

// FieldFlags FieldFlags
type FieldFlags struct {
	Extern    int `json:"is_extern"`
	InOut     int `json:"is_in_out"`
	Aggregate int `json:"is_aggregate"`
}

// FieldMeta FieldMeta
type FieldMeta struct {
	Name  string     `json:"name"`
	Flags FieldFlags `json:"flags"`
}

// OperatorMeta processor.OperatorMeta
type OperatorMeta struct {
	Name   string      `json:"name"`
	Input  []FieldMeta `json:"input"`
	Output []FieldMeta `json:"output"`
}

// GenerateMetas generate all processor input output meta
func GenerateMetas() []OperatorMeta {
	lock.RLock()
	defer lock.RUnlock()
	ops := make([]OperatorMeta, 0, len(processors))
	for name, p := range processors {
		ops = append(ops, GenerateMeta(name, p()))
	}
	return ops
}

// GenerateMeta generate one processor input output meta
func GenerateMeta(name string, p Processor) OperatorMeta {
	var input, output []FieldMeta
	rType := reflect.TypeOf(p)
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
	}
	for i := 0; i < rType.NumField(); i++ {
		t := rType.Field(i)
		tag := t.Tag.Get("graph")
		if tag == "input" {
			input = append(input, FieldMeta{Name: t.Name})
		} else if tag == "multi_input" {
			input = append(input, FieldMeta{Name: t.Name, Flags: FieldFlags{Aggregate: 1}})
		} else if tag == "output" {
			output = append(output, FieldMeta{Name: t.Name})
		}
	}
	return OperatorMeta{Name: name, Input: input, Output: output}
}

// DumpMetaFile dump meta to file
func DumpMetaFile(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	content, err := json.Marshal(GenerateMetas())
	if err != nil {
		return err
	}
	_, err = f.Write(content)
	return err
}
