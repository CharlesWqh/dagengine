package param

// Params 图运行的配置参数
type Params map[string]interface{}

// BuildExpInfo BuildExpInfo
func BuildExpInfo(m map[string]int64) *Params {
	return &Params{"EXP": *BuildMapStrInt64(m)}
}

// BuildMapStrInt64 build params with map string int64
func BuildMapStrInt64(m map[string]int64) *Params {
	p := make(Params)
	for k, v := range m {
		p[k] = v
	}
	return &p
}

// BuildExpStrStr build params with map string
func BuildExpStrStr(m map[string]string) *Params {
	p := make(Params)
	for k, v := range m {
		p[k] = v
	}
	return &p
}

// GetString get string value
func (p Params) GetString(key string) string {
	if i, ok := p[key]; ok {
		str, ok := i.(string)
		if ok {
			return str
		}
	}
	return ""
}

// GetStringList get string list value
func (p Params) GetStringList(key string) []string {
	var result []string
	if i, ok := p[key]; ok {
		is, ok := i.([]interface{})
		if ok {
			for _, v := range is {
				if str, ok := v.(string); ok {
					result = append(result, str)
				} else {
					result = append(result, "")
				}
			}
		}
	}
	return result
}

// GetFloat64 get float64 value
func (p Params) GetFloat64(key string) float64 {
	if i, ok := p[key]; ok {
		ft, ok := i.(float64)
		if ok {
			return ft
		}
	}
	return 0
}

// GetInt64 get int64 value
func (p Params) GetInt64(key string) int64 {
	if i, ok := p[key]; ok {
		i64, ok := i.(int64)
		if ok {
			return i64
		}
	}
	return 0
}

// GetBool get bool value
func (p Params) GetBool(key string) bool {
	if i, ok := p[key]; ok {
		b, ok := i.(bool)
		if ok {
			return b
		}
	}
	return false
}

// Get param by key
func (p Params) Get(key string) Params {
	if i, ok := p[key]; ok {
		if p, ok := i.(map[string]interface{}); ok {
			return p
		} else if p, ok := i.(Params); ok {
			return p
		}
	}
	return Params{}
}

// Set param by key no thread safety
func (p *Params) Set(key string, params Params) {
	if p == nil {
		return
	}
	if *p == nil {
		*p = make(map[string]interface{})
	}
	(*p)[key] = params
}

// SetInt64 set int64 by key no thread safety
func (p *Params) SetInt64(key string, i int64) {
	if p == nil {
		return
	}
	if *p == nil {
		*p = make(map[string]interface{})
	}
	(*p)[key] = i
}

// SetString string by key no thread safety
func (p *Params) SetString(key, i string) {
	if p == nil {
		return
	}
	if *p == nil {
		*p = make(map[string]interface{})
	}
	(*p)[key] = i
}

// SetBool set bool by key no thread safety
func (p *Params) SetBool(key string, i bool) {
	if p == nil {
		return
	}
	if *p == nil {
		*p = make(map[string]interface{})
	}
	(*p)[key] = i
}

// Clone clone a param
func (p *Params) Clone() *Params {
	if p == nil {
		return nil
	}
	newp := make(Params)
	for k, v := range *p {
		newp[k] = v
	}
	return &newp
}
