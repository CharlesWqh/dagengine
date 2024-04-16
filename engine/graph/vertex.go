package graph

import (
	"fmt"
	"strings"

	"xxxx/dagengine/engine/param"
	"xxxx/dagengine/engine/processor"
	"xxxx/innererror"
)

// CondParams condition param
type CondParams struct {
	Match string       `toml:"match"`
	Args  param.Params `toml:"args"`
}

// Unit min unit for vertex input/output
type Unit struct {
	ID         string   `toml:"id" json:"id"`
	Field      string   `toml:"field" json:"field"`
	Aggregate  []string `toml:"aggregate" json:"aggregate"`
	Cond       string   `toml:"cond" json:"cond"`
	Required   bool     `toml:"required" json:"required"`
	Optional   bool     `toml:"optional" json:"optional"`
	Move       bool     `toml:"move" json:"move"`
	IsExtern   bool     `toml:"extern" json:"extern"`
	IsInOut    bool
	IsMapInput bool
}

// Vertex vertex detail struct
type Vertex struct {
	ID           string       `toml:"id" json:"id"`
	Processor    string       `toml:"processor" json:"processor"`
	Cond         string       `toml:"cond" json:"cond"`
	Expect       string       `toml:"expect" json:"expect"`
	ExpectConfig string       `toml:"expect_config" json:"expect_config"`
	SelectArgs   []CondParams `toml:"select_args" json:"select_args"`
	Params       param.Params `toml:"args" json:"args"`

	Cluster        string   `toml:"cluster" json:"cluster"`
	Graph          string   `toml:"graph" json:"graph"`
	Successor      []string `toml:"successor" json:"successor"`
	SuccessorOnOk  []string `toml:"if" json:"if"`
	SuccessorOnErr []string `toml:"else" json:"else"`
	Deps           []string `toml:"deps" json:"deps"`
	DepsOnOk       []string `toml:"deps_on_ok" json:"deps_on_ok"`
	DepsOnErr      []string `toml:"deps_on_err" json:"deps_on_err"`

	Input  []Unit `toml:"input" json:"input"`
	Output []Unit `toml:"output" json:"output"`
	Start  bool   `toml:"start" json:"start"`

	successorVertex map[string]*Vertex
	depsResults     map[string]int
	isIDGenerated   bool
	isGenerated     bool
	g               *Graph
}

func (v *Vertex) dumpDotDefine(s *strings.Builder) {
	s.WriteString("    ")
	s.WriteString(v.getDotID())
	s.WriteString(" [label=\"")
	s.WriteString(v.getDotLabel())
	s.WriteString("\"")
	if len(v.Cond) > 0 {
		s.WriteString(" shape=diamond color=black fillcolor=aquamarine style=filled")
	} else if len(v.Graph) > 0 {
		s.WriteString(" shape=box3d, color=blue fillcolor=aquamarine style=filled")
	} else {
		s.WriteString(" color=black fillcolor=linen style=filled")
	}
	s.WriteString("];\n")
}

func (v *Vertex) dumpDotEdge(s *strings.Builder) {
	if len(v.ExpectConfig) > 0 {
		expectConfigID := v.g.Name + "_" + v.ExpectConfig
		expectConfigID = strings.ReplaceAll(expectConfigID, "!", "")
		s.WriteString("    ")
		s.WriteString(expectConfigID)
		s.WriteString(" -> ")
		s.WriteString(v.getDotID())
		if v.ExpectConfig[0] == '!' {
			s.WriteString(" [style=dashed color=red label=\"err\"];\n")
		} else {
			s.WriteString(" [style=bold label=\"ok\"];\n")
		}

		s.WriteString("    ")
		s.WriteString(v.g.Name + "__START__")
		s.WriteString(" -> ")
		s.WriteString(expectConfigID + ";\n")
	}
	if len(v.Expect) > 0 {
		expect := v.g.Name + "_" + v.Expect
		expect = strings.ReplaceAll(expect, "\"", "")
		expect = strings.ReplaceAll(expect, "'", "")
		expect = strings.ReplaceAll(expect, "[", "")
		expect = strings.ReplaceAll(expect, "]", "")
		expect = strings.ReplaceAll(expect, ",", "")
		expect = strings.ReplaceAll(expect, "!", "")
		expect = strings.ReplaceAll(expect, "=", "")
		expect = strings.ReplaceAll(expect, ">", "")
		expect = strings.ReplaceAll(expect, "<", "")
		expect = strings.ReplaceAll(expect, " ", "")
		// build expect node
		s.WriteString("    ")
		s.WriteString(expect)
		s.WriteString(" [label=\"")
		s.WriteString(v.Expect)
		s.WriteString("\"")
		s.WriteString(" shape=diamond color=black fillcolor=aquamarine style=filled];\n")
		// build edge
		s.WriteString("    ")
		s.WriteString(expect)
		s.WriteString(" -> ")
		s.WriteString(v.getDotID())
		if v.Expect[0] == '!' {
			s.WriteString(" [style=dashed color=red label=\"err\"];\n")
		} else {
			s.WriteString(" [style=bold label=\"ok\"];\n")
		}

		s.WriteString("    ")
		s.WriteString(v.g.Name + "__START__")
		s.WriteString(" -> ")
		s.WriteString(expect + ";\n")
	}
	if v.isSuccessorsEmpty() {
		s.WriteString("    " + v.getDotID() + " -> " + v.g.Name + "__STOP__;\n")
	}
	if v.isDepsEmpty() {
		s.WriteString("    " + v.g.Name + "__START__ -> " + v.getDotID() + ";\n")
	}
	v.dumpDepsResult(s)
}

func (v *Vertex) dumpDepsResult(s *strings.Builder) {
	if v.depsResults != nil && len(v.depsResults) > 0 {
		for id, expect := range v.depsResults {
			dep := v.g.getVertexByID(id)
			s.WriteString("    " + dep.getDotID() + " -> " + v.getDotID())
			switch expect {
			case innererror.VResultOk:
				s.WriteString(" [style=dashed label=\"ok\"];\n")
			case innererror.VResultErr:
				s.WriteString(" [style=dashed color=red label=\"err\"];\n")
			default:
				s.WriteString(" [style=bold label=\"all\"];\n")
			}
		}
	}
}

func (v *Vertex) findVertexInSuccessors(current *Vertex, visited map[string]bool) bool {
	visited[v.ID] = true
	if v.successorVertex != nil {
		_, exist := v.successorVertex[current.ID]
		if exist {
			return true
		}
		for _, successor := range v.successorVertex {
			if _, exist := visited[successor.ID]; !exist {
				if successor.findVertexInSuccessors(current, visited) {
					return true
				}
			}
		}
	}
	return false
}

func (v *Vertex) isSuccessorsEmpty() bool {
	return len(v.successorVertex) == 0
}
func (v *Vertex) isDepsEmpty() bool {
	return len(v.depsResults) == 0
}

func (v *Vertex) verify() error {
	if !v.Start {
		if v.isDepsEmpty() && v.isSuccessorsEmpty() {
			return fmt.Errorf("Vertex:%s/%s has no deps and successors", v.g.Name, v.getDotLabel())
		}
	} else {
		if !v.isDepsEmpty() {
			return fmt.Errorf("Vertex:%s/%s is start vertex, but has non empty deps", v.g.Name, v.getDotLabel())
		}
	}
	return nil
}

func (v *Vertex) getDotID() string {
	return v.g.Name + "_" + v.ID
}
func (v *Vertex) getDotLabel() string {
	if len(v.Cond) > 0 {
		return strings.ReplaceAll(v.Cond, "\"", "\\\"")
	}
	if len(v.Processor) > 0 {
		if !v.isIDGenerated {
			return v.ID
		}
		return v.Processor
	}
	if len(v.Graph) > 0 {
		return fmt.Sprintf("%s::%s", v.Cluster, v.Graph)
	}
	return "unknown"
}

func (v *Vertex) buildInputOutput() error {
	meta := v.g.cluster.getOpMeta(v.Processor)
	if meta == nil && v.Cluster == "" && v.Cond == "" {
		return fmt.Errorf("ID:%v No Processor found", v.ID)
	}
	if meta == nil {
		return nil
	}
	v.buildInput(meta)
	v.buildOutput(meta)
	return nil
}

func (v *Vertex) depend(prev *Vertex, expected int) {
	if v.depsResults == nil {
		v.depsResults = make(map[string]int)
	}
	v.depsResults[prev.ID] = expected
	if prev.successorVertex == nil {
		prev.successorVertex = make(map[string]*Vertex)
	}
	prev.successorVertex[v.ID] = v
}
func (v *Vertex) buildDeps(deps []string, expectedResult int) error {
	for _, id := range deps {
		dep := v.g.getVertexByID(id)
		if dep == nil {
			return fmt.Errorf("[%s/%s]No dep vertex id:%s", v.g.Name, v.getDotLabel(), id)
		}
		v.depend(dep, expectedResult)
	}
	return nil
}

func (v *Vertex) buildSuccessor(sucessors []string, expectedResult int) error {
	for _, id := range sucessors {
		successor := v.g.getVertexByID(id)
		if successor == nil {
			return fmt.Errorf("[%s]No successor id:%s", v.getDotLabel(), id)
		}
		successor.depend(v, expectedResult)
	}
	return nil
}

func (v *Vertex) buildInputDeps(dep *Vertex, data Unit) error {
	if dep == nil {
		if !data.IsExtern && !data.Optional {
			return fmt.Errorf("[%s/%s]No dep input id:%s", v.g.Name, v.getDotLabel(), data.ID)
		}
		return nil
	}
	v.depend(dep, innererror.VResultAll)
	return nil
}

func (v *Vertex) buildDataDeps() error {
	for _, data := range v.Input {
		if len(data.Aggregate) == 0 && !data.IsMapInput {
			dep := v.g.getVertexByData(data.ID)
			if data.IsInOut && dep == v {
				continue
			}
			if err := v.buildInputDeps(dep, data); err != nil {
				return err
			}
		} else {
			for _, id := range data.Aggregate {
				dep := v.g.getVertexByData(id)
				if err := v.buildInputDeps(dep, data); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (v *Vertex) build() error {
	for _, cond := range v.SelectArgs {
		if !v.g.cluster.ContainsConfigSetting(cond.Match) {
			return fmt.Errorf("No config_setting with name:%s defined", cond.Match)
		}
	}
	if err := v.buildDataDeps(); err != nil {
		return err
	}
	if err := v.buildDeps(v.DepsOnErr, innererror.VResultErr); err != nil {
		return err
	}
	if err := v.buildDeps(v.DepsOnOk, innererror.VResultOk); err != nil {
		return err
	}
	if err := v.buildDeps(v.Deps, innererror.VResultAll); err != nil {
		return err
	}
	if err := v.buildSuccessor(v.SuccessorOnErr, innererror.VResultErr); err != nil {
		return err
	}
	if err := v.buildSuccessor(v.SuccessorOnOk, innererror.VResultOk); err != nil {
		return err
	}
	if err := v.buildSuccessor(v.Successor, innererror.VResultAll); err != nil {
		return err
	}
	return nil
}

func (v *Vertex) checkAndFitUnit(units []Unit) error {
	for idx := range units {
		data := &units[idx]
		if len(data.Field) == 0 {
			return fmt.Errorf("Empty data field for node:%s", v.ID)
		}
		if len(data.ID) == 0 {
			data.ID = data.Field
		}
	}
	return nil
}

func (v *Vertex) buildInput(meta *processor.OperatorMeta) {
	for _, opInput := range meta.Input {
		match := false
		for _, localInput := range v.Input {
			if localInput.Field == opInput.Name {
				match = true
				break
			}
		}
		if !match {
			field := Unit{
				ID:         opInput.Name,
				Field:      opInput.Name,
				IsExtern:   opInput.Flags.Extern > 0,
				IsInOut:    opInput.Flags.InOut > 0,
				IsMapInput: opInput.Flags.Aggregate > 0,
			}
			v.Input = append(v.Input, field)
		}
	}
}

func (v *Vertex) buildOutput(meta *processor.OperatorMeta) {
	for _, opOutput := range meta.Output {
		match := false
		for _, localOutput := range v.Output {
			if localOutput.Field == opOutput.Name {
				match = true
				break
			}
		}
		if !match {
			field := Unit{
				ID:    opOutput.Name,
				Field: opOutput.Name,
			}
			v.Output = append(v.Output, field)
		}
	}
}
