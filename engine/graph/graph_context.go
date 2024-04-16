package graph

import (
	"context"
	"fmt"
	"log"
	"sync"

	"xxxx/util/safe"
)

// Context graph context
type Context struct {
	ClusterContext     *ClusterContext
	Graph              *Graph
	VertexContextTable map[*Vertex]*VertexContext
	AllInputIDs        map[DIObjectKey]bool
	AllOutputIDs       map[DIObjectKey]bool
	ExternDataContext  *DataContext
}

// Reset graph context reset
func (c *Context) Reset() {
	c.ExternDataContext = nil
	for _, v := range c.VertexContextTable {
		v.Reset()
	}
}

// NewContext new graph context
func NewContext(cc *ClusterContext, g *Graph) (*Context, error) {
	c := &Context{ClusterContext: cc,
		Graph:              g,
		VertexContextTable: make(map[*Vertex]*VertexContext),
		AllInputIDs:        make(map[DIObjectKey]bool),
		AllOutputIDs:       make(map[DIObjectKey]bool)}
	for _, v := range g.vertexMap {
		vc, err := NewVertexContext(c, v)
		if err != nil {
			return nil, fmt.Errorf("new vertex:%v context err:%w", v.getDotLabel(), err)
		}
		c.VertexContextTable[v] = vc
		if vc.ProcessorDI != nil {
			for _, dikey := range vc.ProcessorDI.InputIDs {
				// if _, ok := c.AllInputIDs[*dikey]; ok {
				//	return nil, fmt.Errorf("Duplicate input name:%v in graph:%v", name, g.Name)
				// }
				c.AllInputIDs[*dikey] = true
			}
			for name, dikey := range vc.ProcessorDI.OutputIDs {
				if _, ok := c.AllOutputIDs[*dikey]; ok {
					return nil, fmt.Errorf("Duplicate output name:%v in graph:%v", name, g.Name)
				}
				c.AllOutputIDs[*dikey] = true
			}
		}
	}
	return c, nil
}

// Execute execute graph context with datacontext
func (c *Context) Execute(ctx context.Context, dataContext *DataContext) error {
	c.ExternDataContext = dataContext
	return c.execute(ctx)
}

// Execute execute graph context
func (c *Context) execute(ctx context.Context) error {
	for id := range c.AllInputIDs {
		c.ExternDataContext.RegisterData(id)
	}
	for id := range c.AllOutputIDs {
		c.ExternDataContext.RegisterData(id)
	}
	var readySuccessors []*VertexContext
	for _, v := range c.VertexContextTable {
		if v.Ready() {
			readySuccessors = append(readySuccessors, v)
		}
	}
	return c.ExecuteReadyVertexes(ctx, readySuccessors)
}

// ExecuteReadyVertexes execute ready vertexes
func (c *Context) ExecuteReadyVertexes(ctx context.Context, vertexes []*VertexContext) error {
	if len(vertexes) == 0 {
		return nil
	}
	var wg sync.WaitGroup
	wg.Add(len(vertexes))
	for i := range vertexes {
		i := i
		safe.Go(func() {
			defer wg.Done()
			if err := vertexes[i].Execute(ctx); err != nil {
				log.Printf("vertex execute err:%v", err)
			}
			_ = c.OnVertexDone(ctx, vertexes[i])
		})
	}
	wg.Wait()
	return nil
}

// OnVertexDone do something after vertex execute
func (c *Context) OnVertexDone(ctx context.Context, vertexContext *VertexContext) error {
	var readySuccessors []*VertexContext
	for _, successor := range vertexContext.Vertex.successorVertex {
		successorCtx, ok := c.VertexContextTable[successor]
		if !ok {
			panic("not found successor vertex")
		}
		waitNum :=
			successorCtx.SetDependencyResult(vertexContext.Vertex, vertexContext.result)
		// last dependency
		if waitNum == 0 {
			readySuccessors = append(readySuccessors, successorCtx)
		}
	}
	return c.ExecuteReadyVertexes(ctx, readySuccessors)
}
