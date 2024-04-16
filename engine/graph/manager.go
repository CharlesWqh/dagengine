package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"xxxx/dagengine/engine/param"
	"xxxx/dagengine/engine/processor"

	"github.com/BurntSushi/toml"
)

// LoadFile LoadFile
func LoadFile(filepath string) error {
	if path.Ext(filepath) == ".toml" {
		return DefaultManager.loadFile(filepath, &TomlCodec{})
	}
	return DefaultManager.loadFile(filepath, &JSONCodec{})
}

// Load Load
func Load(name string, content []byte, c Codec) error {
	return DefaultManager.load(name, content, c)
}

// Execute  execute one graph on cluster
func Execute(ctx context.Context, clusterName string, graphName string,
	dataContext *DataContext, params *param.Params) error {
	return DefaultManager.execute(ctx, clusterName, graphName, dataContext, params)
}

// Manager manager of cluster
type Manager struct {
	clusters map[string]*Cluster
	lock     sync.RWMutex
}

// Codec json toml unmarshal
type Codec interface {
	Name() string
	Unmarshal([]byte, interface{}) error
}

// JSONCodec JSON codec
type JSONCodec struct{}

// Name JSON codec
func (*JSONCodec) Name() string {
	return "json"
}

// Unmarshal JSON decode
func (c *JSONCodec) Unmarshal(in []byte, out interface{}) error {
	return json.Unmarshal(in, out)
}

// TomlCodec toml codec
type TomlCodec struct{}

// Name toml codec
func (*TomlCodec) Name() string {
	return "toml"
}

// Unmarshal toml decode
func (c *TomlCodec) Unmarshal(in []byte, out interface{}) error {
	return toml.Unmarshal(in, out)
}

// LoadFile load cluster from file
func (m *Manager) loadFile(filepath string, c Codec) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	name := path.Base(filepath)
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	return m.load(name, content, c)
}

// LoadFile load cluster from content
func (m *Manager) load(name string, content []byte, c Codec) error {
	cluster := &Cluster{}
	if err := c.Unmarshal(content, cluster); err != nil {
		return err
	}
	if cluster.DefaultContextPoolSize == 0 {
		cluster.DefaultContextPoolSize = defaultContextPoolSize
	}
	if err := cluster.Build(processor.GenerateMetas()); err != nil {
		return err
	}
	m.lock.Lock()
	m.clusters[name] = cluster
	m.lock.Unlock()
	return nil
}

// Execute cluster by clusterName and graphName
func (m *Manager) execute(ctx context.Context, clusterName string, graphName string,
	dataContext *DataContext, params *param.Params) error {
	if dataContext == nil {
		dataContext = NewDataContext()
	}
	m.lock.RLock()
	cluster, ok := m.clusters[clusterName]
	m.lock.RUnlock()
	if !ok {
		return fmt.Errorf("not find cluter:%v", clusterName)
	}
	clusterContext, err := cluster.ClusterContextPool.Get(cluster)
	if err != nil {
		return err
	}
	defer func() {
		clusterContext.Reset()
		cluster.ClusterContextPool.Put(clusterContext)
	}()
	return clusterContext.Execute(ctx, graphName, dataContext, params)
}

// DefaultManager default cluster manager
var DefaultManager = New()

// New manager
func New() *Manager {
	return &Manager{clusters: make(map[string]*Cluster)}
}
