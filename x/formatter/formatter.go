package formatter

type valueCache struct {
	data  string
	cache map[string]IValueGetter
}

func (f *valueCache) getValue(info ValueGetter) interface{} {
	v, ok := f.cache[info.Type]
	if !ok {
		v = registry[info.Type](f.data)
	}
	return v.Value(info.Expression, info.indexes)
}

func (f *valueCache) getArray(info ValueGetter) []interface{} {
	v, ok := f.cache[info.Type]
	if !ok {
		v = registry[info.Type](f.data)
	}
	return v.Array(info.Expression, info.indexes)
}

func newValueCache(data string) *valueCache {
	return &valueCache{
		data:  data,
		cache: map[string]IValueGetter{},
	}
}
