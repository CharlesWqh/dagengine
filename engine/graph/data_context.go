package graph

import (
	"reflect"
	"sync"
)

// DataContext data for run cluster
type DataContext struct {
	Data sync.Map
}

// GlobalDataContext use as global di container
var GlobalDataContext = NewDataContext()

// NewDataContext new datacontext
func NewDataContext() *DataContext {
	return &DataContext{}
}

// Get get by key
func (d *DataContext) Get(key DIObjectKey) (interface{}, bool) {
	return d.Data.Load(key)
}

// Set set value
func (d *DataContext) Set(key DIObjectKey, value interface{}) {
	d.Data.Store(key, value)
}

// SetConfigSetting set configsetting
func (d *DataContext) SetConfigSetting(key string, value bool) {
	d.Data.Store(key, value)
}

// GetConfigSetting get configsetting
func (d *DataContext) GetConfigSetting(key string) bool {
	if configSetting, ok := d.Data.Load(key); ok {
		return configSetting.(bool)
	}
	return false
}

// RegisterData register key
func (d *DataContext) RegisterData(key DIObjectKey) {
	d.Set(key, nil)
}

// GetNoPtrType GetNoPtrType
func GetNoPtrType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}

// NewDIObjectKey new one
func NewDIObjectKey(name string, t reflect.Type) DIObjectKey {
	return DIObjectKey{
		Name:        name,
		ReflectType: GetNoPtrType(t),
	}
}

// DIObjectKey di object key
type DIObjectKey struct {
	Name        string
	ReflectType reflect.Type
}
