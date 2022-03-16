package template

type kv struct {
	Key   string
	Value interface{}
}

func KS(kvs ...kv) Params {
	p := Params{}
	for _, kv := range kvs {
		p[kv.Key] = kv.Value
	}
	return p
}

func KV(key string, value interface{}) *kv {
	return &kv{Key: key, Value: value}
}

type Params map[string]interface{}

func (p Params) Set(name string, val interface{}) string {
	if _, ok := p[name]; ok {
		panic("")
	}
	p[name] = val
	return ""
}
