package graph

import (
	"fmt"
	"strings"

	"xxxx/dagengine/engine/processor"
)

const defaultContextPoolSize = 10

// ConfigSetting some expr
type ConfigSetting struct {
	Name      string `toml:"name"`
	Cond      string `toml:"cond"`
	Processor string `toml:"processor"`
}

// Cluster multi graph cluster
type Cluster struct {
	Desc                   string          `toml:"desc" json:"desc"`
	StrictDsl              bool            `toml:"strict_dsl" json:"strict_dsl"`
	DefaultContextPoolSize int             `toml:"default_context_pool_size" json:"default_context_pool_size"`
	Graph                  []Graph         `toml:"graph" json:"graph"`
	ConfigSetting          []ConfigSetting `toml:"config_setting" json:"config_setting"`

	ClusterContextPool *ClusterContextPool
	GraphManager       *Manager
	Name               string

	graphMap map[string]*Graph
	opsMap   map[string]processor.OperatorMeta
}

// ContainsConfigSetting if cluster contains configsetting
func (c *Cluster) ContainsConfigSetting(name string) bool {
	for _, c := range c.ConfigSetting {
		if len(name) > 0 && name[0] == '!' {
			if c.Name == name[1:] {
				return true
			}
		} else {
			if c.Name == name {
				return true
			}
		}

	}
	return false
}

func (c *Cluster) getOpMeta(name string) *processor.OperatorMeta {
	v, exist := c.opsMap[name]
	if !exist {
		return nil
	}
	return &v
}

func (c *Cluster) initClusterContext() error {
	c.ClusterContextPool = NewClusterContextPool()
	for i := 0; i < c.DefaultContextPoolSize; i++ {
		cc, err := NewClusterContext(c)
		if err != nil {
			return err
		}
		c.ClusterContextPool.Put(cc)
	}
	return nil
}

// Build build graph cluster
func (c *Cluster) Build(ops []processor.OperatorMeta) error {
	if len(c.Graph) == 0 {
		return fmt.Errorf("Graph empty")
	}
	c.opsMap = make(map[string]processor.OperatorMeta)
	for _, op := range ops {
		c.opsMap[op.Name] = op
	}
	c.graphMap = make(map[string]*Graph)
	for i := range c.Graph {
		g := &c.Graph[i]
		g.cluster = c
		if existg, exist := c.graphMap[g.Name]; exist {
			if existg.ExpectVersion == g.ExpectVersion {
				return fmt.Errorf("Duplicate graph name:%v", g.Name)
			}
			if g.Priority <= existg.Priority {
				continue
			}
		}
		c.graphMap[g.Name] = g
	}
	for _, g := range c.graphMap {
		err := g.Build()
		if nil != err {
			return err
		}
	}
	return c.initClusterContext()
}

// DumpDot dump cluster dot file
func (c *Cluster) DumpDot(buffer *strings.Builder) {
	buffer.WriteString("digraph G {\n")
	buffer.WriteString("    rankdir=LR;\n")
	for i := len(c.Graph) - 1; i >= 0; i-- {
		c.Graph[i].DumpDot(buffer)
	}
	buffer.WriteString("}\n")
}
