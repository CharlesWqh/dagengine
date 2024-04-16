package graph

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"xxxx/dagengine/engine/param"
	"xxxx/dagengine/engine/processor"
	"xxxx/innererror"

	"github.com/antonmedv/expr"
)

type vertexResult struct {
	conditionResult error
	processorResult error
}

// VertexContext vertex context
type VertexContext struct {
	GraphContext     *Context
	Vertex           *Vertex
	Processor        processor.Processor
	ProcessorDI      *ProcessorDI
	Params           *param.Params
	vertexDepResults sync.Map
	result           *vertexResult
	waitNum          int32
}

// Reset reset inner var
func (v *VertexContext) Reset() {
	v.waitNum = int32(len(v.Vertex.depsResults))
	v.result = new(vertexResult)
	v.vertexDepResults.Range(func(key interface{}, value interface{}) bool {
		v.vertexDepResults.Delete(key)
		return true
	})
}

// NewVertexContext new vertex context by graph context and vertex
func NewVertexContext(g *Context, v *Vertex) (*VertexContext, error) {
	vc := &VertexContext{GraphContext: g,
		Vertex: v,
		Params: &v.Params,
	}
	if vc.Vertex.Graph == "" && vc.Vertex.Processor != "" {
		vc.Processor = processor.Get(vc.Vertex.Processor)
	}
	if vc.Processor == nil && vc.Vertex.Processor != "" {
		return nil, fmt.Errorf("processor name:%v not find", vc.Vertex.Processor)
	}
	if vc.Processor != nil {
		vc.ProcessorDI = &ProcessorDI{Processor: vc.Processor}
		if err := vc.ProcessorDI.PrepareInput(vc.Vertex.Input); err != nil {
			return nil, err
		}
		if err := vc.ProcessorDI.PrepareOutput(vc.Vertex.Output); err != nil {
			return nil, err
		}
	}
	if vc.Processor != nil {
		vc.Processor.OnInit()
	}
	vc.waitNum = int32(len(vc.Vertex.depsResults))
	vc.result = new(vertexResult)
	return vc, nil
}

// Ready check vertex ready to run
func (v *VertexContext) Ready() bool {
	return atomic.LoadInt32(&v.waitNum) == 0
}

func (v *VertexContext) checkConditionResult() error {
	for id, expectR := range v.Vertex.depsResults {
		r, ok := v.vertexDepResults.Load(id)
		if !ok {
			return innererror.Errorf(innererror.VResultErr, "depend vertex:%v not find result", id)
		}
		vr, _ := r.(*vertexResult)
		err, ok := innererror.Upgrade(vr.conditionResult)
		if !ok {
			if expectR == innererror.VResultErr {
				return innererror.Errorf(innererror.VResultErr, "expect error but not return err")
			}
			continue
		}
		if (err.Code() & int32(expectR)) == 0 {
			return innererror.Errorf(innererror.VResultErr, "expect not match code:%v expect:%v", err.Code(), expectR)
		}
	}
	return nil
}

func (v *VertexContext) addDepProcessorResult(inputP *param.Params) *param.Params {
	var executeParams *param.Params
	if v.GraphContext.ClusterContext.ExecuteParams == nil {
		executeParams = &param.Params{}
	} else {
		executeParams = inputP.Clone()
	}
	for id := range v.Vertex.depsResults {
		r, ok := v.vertexDepResults.Load(id)
		if !ok {
			continue
		}
		vr, _ := r.(*vertexResult)
		code := innererror.Code(vr.processorResult)
		executeParams.SetInt64("RET_CODE_"+id, int64(code))
	}
	return executeParams
}

func (v *VertexContext) evalExpect() error {
	if len(v.Vertex.Expect) == 0 {
		return nil
	}
	executeParams := v.addDepProcessorResult(v.GraphContext.ClusterContext.ExecuteParams)
	output, err := expr.Eval(v.Vertex.Expect, *executeParams)
	if err != nil {
		log.Printf("expr name:%v expect:%v err:%v", v.Vertex.Expect, output, err)
	} else if expect, ok := output.(bool); ok && !expect {
		return innererror.Errorf(innererror.VResultErr, "expect:%v skip vertex:%s", v.Vertex.Expect, v.Vertex.ID)
	}
	return nil
}

func (v *VertexContext) evalCond() error {
	if len(v.Vertex.Cond) == 0 {
		return nil
	}
	if v.GraphContext.ClusterContext.ExecuteParams == nil {
		return innererror.Errorf(innererror.VResultErr,
			"expect:%v no execute param skip vertex:%s", v.Vertex.Expect, v.Vertex.ID)
	}
	output, err := expr.Eval(v.Vertex.Cond, *v.GraphContext.ClusterContext.ExecuteParams)
	if err != nil {
		log.Printf("expr name:%v cond:%v err:%v", v.Vertex.Cond, output, err)
	} else if expect, ok := output.(bool); ok && !expect {
		return innererror.Errorf(innererror.VResultErr, "cond:%v skip vertex", v.Vertex.Cond)
	}
	return nil
}

func (v *VertexContext) evalExpectConfig() error {
	if len(v.Vertex.ExpectConfig) > 0 {
		expect := v.GraphContext.ExternDataContext.GetConfigSetting(v.Vertex.ExpectConfig)
		if !expect {
			return innererror.Errorf(innererror.VResultErr, "expect config:%v skip vertex", v.Vertex.ExpectConfig)
		}
	}
	return nil
}

func (v *VertexContext) conditionCheck() error {
	if err := v.checkConditionResult(); err != nil {
		return err
	}
	if err := v.evalExpectConfig(); err != nil {
		return err
	}
	if err := v.evalExpect(); err != nil {
		return err
	}
	if err := v.evalCond(); err != nil {
		return err
	}
	return nil
}

func (v *VertexContext) paramCheck() error {
	if v.Vertex.Cluster == "" && v.Processor == nil && v.Vertex.Cond == "" {
		return fmt.Errorf("Vertex:%v has empty processor and empty subgraph context",
			v.Vertex.getDotLabel())
	}
	return nil
}

// Execute execute one vertex
func (v *VertexContext) Execute(ctx context.Context) error {
	if err := v.paramCheck(); err != nil {
		return err
	}
	if err := v.conditionCheck(); err != nil {
		v.result.conditionResult = err
		return err
	}
	if v.Processor != nil {
		return v.ExecuteProcessor(ctx)
	}
	if v.Vertex.Cluster != "" {
		return v.ExecuteSubGraph(ctx)
	}
	return nil
}

// ExecuteProcessor execute one processor
func (v *VertexContext) ExecuteProcessor(ctx context.Context) error {
	start := time.Now()
	defer func() {
		AddEvent(&Event{
			Processor: v.Vertex.Processor,
			Duration:  time.Since(start),
		})
	}()
	executeParams := v.GetExecuteParams()
	for _, condParams := range v.Vertex.SelectArgs {
		if expect := v.GraphContext.ExternDataContext.GetConfigSetting(condParams.Match); expect {
			executeParams = &(condParams.Args)
			break
		}
	}
	v.ProcessorDI.Reset()
	v.ProcessorDI.InjectInput(v.GraphContext.ExternDataContext, v.Vertex.Input)
	// set global params
	if v.GraphContext.ClusterContext.ExecuteParams != nil {
		executeParams = executeParams.Clone()
		executeParams.Set("GLOBAL", *v.GraphContext.ClusterContext.ExecuteParams)
	}
	v.result.processorResult = v.Processor.OnExecute(ctx, executeParams)
	v.ProcessorDI.CollectOutput(v.GraphContext.ExternDataContext)
	return nil
}

// ExecuteSubGraph execute sub graph
func (v *VertexContext) ExecuteSubGraph(ctx context.Context) error {
	return Execute(ctx, v.Vertex.Cluster, v.Vertex.Graph,
		v.GraphContext.ExternDataContext, v.GraphContext.ClusterContext.ExecuteParams)
}

// GetExecuteParams get vertex context execute params
func (v *VertexContext) GetExecuteParams() *param.Params {
	return v.Params
}

// SetDependencyResult set dependency vertex result
func (v *VertexContext) SetDependencyResult(vertex *Vertex, r interface{}) int32 {
	_, ok := v.vertexDepResults.Load(vertex.ID)
	v.vertexDepResults.Store(vertex.ID, r)
	if !ok {
		newWaitNum := atomic.AddInt32(&v.waitNum, -1)
		return newWaitNum
	}
	return atomic.LoadInt32(&v.waitNum)
}
