package graph

import (
	"fmt"
	"reflect"

	"xxxx/dagengine/engine/processor"
)

const (
	cInput       = "input"
	cMultiInput  = "multi_input"
	cExternInput = "extern_input"
	cOutput      = "output"
)

// ProcessorDI processor di for execute
type ProcessorDI struct {
	Processor processor.Processor
	InputIDs  map[string]*DIObjectKey
	OutputIDs map[string]*DIObjectKey
}

// Reset reset after execute
func (p *ProcessorDI) Reset() {
	rType := reflect.TypeOf(p.Processor)
	rVal := reflect.ValueOf(p.Processor)
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
		rVal = rVal.Elem()
	}
	for i := 0; i < rType.NumField(); i++ {
		f := rVal.Field(i)
		if f.CanSet() {
			// reset init
			if f.Kind() == reflect.Ptr {
				f.Set(reflect.New(f.Type().Elem()))
			} else {
				f.Set(reflect.Zero(f.Type()))
			}
		}
	}
}

// InjectInput inject input for processor
func (p *ProcessorDI) InjectInput(dataContext *DataContext, metas []Unit) {
	rType := reflect.TypeOf(p.Processor)
	rVal := reflect.ValueOf(p.Processor)
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
		rVal = rVal.Elem()
	}
	mMetas := make(map[string]*Unit, len(metas))
	for i := range metas {
		mMetas[metas[i].Field] = &metas[i]
	}
	for i := 0; i < rType.NumField(); i++ {
		t := rType.Field(i)
		f := rVal.Field(i)
		tag := t.Tag.Get("graph")
		if tag == cInput {
			if v, ok := dataContext.Get(*p.InputIDs[t.Name]); ok {
				p.setInput(f, v)
			} else {
				p.resetInput(f)
			}
		} else if tag == cMultiInput {
			p.setMultiInput(dataContext, mMetas, f, t)
		} else if tag == cExternInput {
			externKey := NewDIObjectKey(t.Name, t.Type)
			if v, ok := dataContext.Get(externKey); ok {
				p.setInput(f, v)
			} else if v, ok := GlobalDataContext.Get(externKey); ok {
				p.setInput(f, v)
			} else {
				p.resetInput(f)
			}
		}
	}
}

// CollectOutput collect output for processor
func (p *ProcessorDI) CollectOutput(dataContext *DataContext) {
	rType := reflect.TypeOf(p.Processor)
	rVal := reflect.ValueOf(p.Processor)
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
		rVal = rVal.Elem()
	}
	for i := 0; i < rType.NumField(); i++ {
		t := rType.Field(i)
		if t.Tag.Get("graph") == cOutput {
			dataContext.Set(*p.OutputIDs[t.Name], rVal.Field(i))
		}
	}
}

// PrepareInput register input ids
func (p *ProcessorDI) PrepareInput(inputs []Unit) error {
	p.InputIDs = make(map[string]*DIObjectKey)
	return p.SetUpIDs("input", inputs, p.InputIDs)
}

// PrepareOutput register output ids
func (p *ProcessorDI) PrepareOutput(outputs []Unit) error {
	p.OutputIDs = make(map[string]*DIObjectKey)
	return p.SetUpIDs("output", outputs, p.OutputIDs)
}

// SetUpIDs set up ids
func (p *ProcessorDI) SetUpIDs(tag string, cfgs []Unit, ids map[string]*DIObjectKey) error {
	rType := reflect.TypeOf(p.Processor)
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
	}
	for i := 0; i < rType.NumField(); i++ {
		t := rType.Field(i)
		if t.Tag.Get("graph") == tag {
			dikey := NewDIObjectKey(t.Name, t.Type)
			ids[t.Name] = &dikey
		}
	}
	for _, cfg := range cfgs {
		id, ok := ids[cfg.Field]
		if !ok {
			if len(cfg.Aggregate) > 0 {
				continue
			}
			return fmt.Errorf("processor:%v not find field:%v", p.Processor, cfg.Field)
		}
		id.Name = cfg.ID
	}
	return nil
}

func (p *ProcessorDI) setInput(f reflect.Value, v interface{}) {
	rv, ok := v.(reflect.Value)
	if !ok {
		return
	}
	if rv.Kind() != f.Kind() {
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		if f.Kind() == reflect.Ptr {
			f = f.Elem()
		}
	}
	f.Set(rv)
}

func (p *ProcessorDI) resetInput(f reflect.Value) {
	// reset
	if f.Kind() == reflect.Ptr {
		f = f.Elem()
		f.Set(reflect.Zero(f.Type()))
	} else {
		f.Set(reflect.Zero(f.Type()))
	}
}

func (p *ProcessorDI) setMultiInput(dataContext *DataContext, mMetas map[string]*Unit,
	rVal reflect.Value, t reflect.StructField) {
	meta, ok := mMetas[t.Name]
	if !ok {
		return
	}
	f := reflect.MakeMap(t.Type)
	for _, agg := range meta.Aggregate {
		if v, ok := dataContext.Get(NewDIObjectKey(agg, t.Type.Elem())); ok {
			rv, ok := v.(reflect.Value)
			if !ok {
				continue
			}
			f.SetMapIndex(reflect.ValueOf(agg), rv)
		}
	}
	rVal.Set(f)
}
