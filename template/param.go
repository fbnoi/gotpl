package template

type kv struct {
	Key   string
	Value any
}

func KS(kvs ...kv) Params {
	p := Params{}
	for _, kv := range kvs {
		p[kv.Key] = kv.Value
	}
	return p
}

func KV(key string, value any) *kv {
	return &kv{Key: key, Value: value}
}

type Params map[string]any

func (p Params) Set(name string, val any) string {
	if _, ok := p[name]; ok {
		panic("")
	}
	p[name] = val
	return ""
}
