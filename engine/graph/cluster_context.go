package graph

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"sync"

	"xxxx/dagengine/engine/param"

	"github.com/antonmedv/expr"
)

// ClusterContextPool cluster context pool
type ClusterContextPool struct {
	l    *list.List
	size int
	lock sync.Mutex
}

// NewClusterContextPool create cluster context pool
func NewClusterContextPool() *ClusterContextPool {
	return &ClusterContextPool{l: list.New()}
}

// Get get one cluster context from pool
func (cp *ClusterContextPool) Get(c *Cluster) (*ClusterContext, error) {
	cp.lock.Lock()
	f := cp.l.Front()
	if f != nil {
		cp.l.Remove(f)
		cp.size--
		cp.lock.Unlock()
		return f.Value.(*ClusterContext), nil
	}
	cp.lock.Unlock()
	return NewClusterContext(c)
}

// Put put cluster context into pool
func (cp *ClusterContextPool) Put(c *ClusterContext) {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	cp.l.PushBack(c)
	cp.size++
}

// NewClusterContext new cluster context
func NewClusterContext(c *Cluster) (*ClusterContext, error) {
	cp := &ClusterContext{
		GraphContextTable: make(map[string]*Context),
		ConfigSetting:     c.ConfigSetting,
	}
	cp.Cluster = c
	for i, cfg := range c.Graph {
		var err error
		cp.GraphContextTable[cfg.Name], err = NewContext(cp, &c.Graph[i])
		if err != nil {
			return nil, fmt.Errorf("new context err:%w", err)
		}
	}
	return cp, nil
}

// ClusterContext cluster execute context
type ClusterContext struct {
	Cluster           *Cluster
	GraphContextTable map[string]*Context
	ExternDataContext *DataContext
	ExecuteParams     *param.Params
	ConfigSetting     []ConfigSetting
}

// Execute cluster execute with datacontext and params
func (c *ClusterContext) Execute(ctx context.Context, graphName string,
	dataContext *DataContext, params *param.Params) error {
	c.ExternDataContext = dataContext
	c.ExecuteParams = params
	return c.execute(ctx, graphName)
}

// Execute cluster execute by graph name
func (c *ClusterContext) execute(ctx context.Context, graphName string) error {
	graphContext, ok := c.GraphContextTable[graphName]
	if !ok {
		return fmt.Errorf("not find graph:%v", graphName)
	}
	if c.ExecuteParams != nil {
		for _, cs := range c.ConfigSetting {
			output, err := expr.Eval(cs.Cond, *c.ExecuteParams)
			if err != nil {
				log.Printf("expr name:%v expect:%v err:%v", cs.Name, output, err)
			} else {
				if expect, ok := output.(bool); ok {
					c.ExternDataContext.SetConfigSetting(cs.Name, expect)
				}
			}
		}
	}
	return graphContext.Execute(ctx, c.ExternDataContext)
}

// Reset cluster context reset
func (c *ClusterContext) Reset() {
	c.ExternDataContext = nil
	c.ExecuteParams = nil
	for _, g := range c.GraphContextTable {
		g.Reset()
	}
}
