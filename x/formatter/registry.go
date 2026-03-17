package formatter

type IValueGetterCtor func(data string) IValueGetter

var registry map[string]IValueGetterCtor

func init() {
	registry = make(map[string]IValueGetterCtor)
	registry["json"] = NewJsonValueGetter
	registry["regex"] = NewRegexpValueGetter
	registry["xml"] = NewXmlValueGetter
}
