package graph

import (
	"fmt"
	"log"
	"strings"
)

// Graph graph detail struct
type Graph struct {
	Name          string   `toml:"name" json:"name"`
	Vertex        []Vertex `toml:"vertex" json:"vertex"`
	ExpectVersion string   `toml:"expect_version" json:"expect_version"`
	Priority      int      `toml:"priority" json:"priority"`

	cluster     *Cluster
	vertexMap   map[string]*Vertex
	dataMapping map[string]*Vertex

	genIdx int
}

// DumpDot dump graph dot
func (g *Graph) DumpDot(buffer *strings.Builder) {
	buffer.WriteString("  subgraph cluster_")
	buffer.WriteString(g.Name)
	buffer.WriteString("{\n")
	buffer.WriteString("    style = rounded;\n")
	buffer.WriteString(fmt.Sprintf("    label = \"%s\";\n", g.Name))
	buffer.WriteString("    ")
	buffer.WriteString(g.Name + "__START__")
	buffer.WriteString("[color=black fillcolor=deepskyblue style=filled shape=Msquare label=\"START\"];\n")
	buffer.WriteString("    ")
	buffer.WriteString(g.Name + "__STOP__")
	buffer.WriteString("[color=black fillcolor=deepskyblue style=filled shape=Msquare label=\"STOP\"];\n")

	for _, v := range g.vertexMap {
		v.dumpDotDefine(buffer)
	}

	for _, c := range g.cluster.ConfigSetting {
		buffer.WriteString("    ")
		buffer.WriteString(g.Name + "_" + c.Name)
		buffer.WriteString(" [label=\"")
		buffer.WriteString(c.Name)
		buffer.WriteString("\"")
		buffer.WriteString(" shape=diamond color=black fillcolor=aquamarine style=filled];\n")
	}

	for _, v := range g.vertexMap {
		if v.isGenerated {
			continue
		}
		v.dumpDotEdge(buffer)
	}
	buffer.WriteString("};\n")
}

func (g *Graph) genVertexID() string {
	id := fmt.Sprintf("%s_%d", g.Name, g.genIdx)
	g.genIdx++
	return id
}
func (g *Graph) getVertexByData(data string) *Vertex {
	v, exist := g.dataMapping[data]
	if exist {
		return v
	}
	return nil
}
func (g *Graph) getVertexByID(id string) *Vertex {
	v, exist := g.vertexMap[id]
	if exist {
		return v
	}
	return nil
}

func (g *Graph) testCircle() bool {
	for _, v := range g.vertexMap {
		visited := make(map[string]bool)
		if v.findVertexInSuccessors(v, visited) {
			log.Printf("Graph:%s has a circle with vertex:%s", g.Name, v.ID)
			return true
		}
	}
	return false
}

func (g *Graph) fillIDAndCluster(v *Vertex) {
	if len(v.ID) == 0 {
		if len(v.Processor) > 0 {
			v.ID = v.Processor
		} else {
			v.ID = g.genVertexID()
			v.isIDGenerated = true
		}
	}
	if len(v.Graph) > 0 && (len(v.Cluster) == 0 || v.Cluster == ".") {
		v.Cluster = g.cluster.Name
	}
}

func (g *Graph) buildVertexMap() error {
	g.vertexMap = make(map[string]*Vertex)
	for i := range g.Vertex {
		v := &g.Vertex[i]
		g.fillIDAndCluster(v)
		if len(v.Expect) > 0 && len(v.ExpectConfig) > 0 {
			return fmt.Errorf("Vertex:%s can NOT both config 'expect' & 'expect_config'", v.ID)
		}
		if len(v.ExpectConfig) > 0 && !g.cluster.ContainsConfigSetting(v.ExpectConfig) {
			return fmt.Errorf("No config_setting with name:%s defined", v.ExpectConfig)
		}
		if _, exist := g.vertexMap[v.ID]; exist {
			return fmt.Errorf("Duplcate vertex id:%s", v.ID)
		}
		v.g = g
		g.vertexMap[v.ID] = v
	}
	return nil
}

func (g *Graph) buildInputOutput() error {
	g.dataMapping = make(map[string]*Vertex)
	for i := range g.Vertex {
		v := &g.Vertex[i]
		if err := v.buildInputOutput(); err != nil {
			return err
		}
		if err := v.checkAndFitUnit(v.Input); err != nil {
			return err
		}
		if err := v.checkAndFitUnit(v.Output); err != nil {
			return err
		}
		for idx := range v.Output {
			data := &v.Output[idx]
			if prev, exist := g.dataMapping[data.ID]; exist {
				return fmt.Errorf("Duplicate data name:%s in vertex:%s/%s, prev vertex:%s",
					data.ID, v.g.Name, v.getDotLabel(), prev.getDotLabel())
			}
			g.dataMapping[data.ID] = v
		}
	}
	return nil
}

// Build build graph
func (g *Graph) Build() error {
	if len(g.Vertex) == 0 {
		return fmt.Errorf("Graph:%s vertex empty", g.Name)
	}
	if err := g.buildVertexMap(); err != nil {
		return err
	}
	if err := g.buildInputOutput(); err != nil {
		return err
	}
	for _, v := range g.vertexMap {
		err := v.build()
		if err != nil {
			return err
		}
	}
	for _, v := range g.vertexMap {
		if len(v.Cond) > 0 {
			continue
		}
		err := v.verify()
		if err != nil {
			return err
		}
	}
	if g.testCircle() {
		return fmt.Errorf("Circle Exist")
	}
	return nil
}
