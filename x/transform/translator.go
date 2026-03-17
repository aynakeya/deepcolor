package transform

import (
	"github.com/aynakeya/deepcolor/internal/dynmarshaller"
	"reflect"
)

type Translator interface {
	GetType() string
	Apply(value interface{}) (interface{}, error)
	MustApply(value interface{}) interface{}
}

type BaseTranslator struct {
	Type string
}

func (t BaseTranslator) GetType() string {
	return t.Type
}

type translatorWrapper struct {
	typ        string
	translator func(value interface{}) (interface{}, error)
}

func (t *translatorWrapper) GetType() string {
	return t.typ
}

func (t *translatorWrapper) Apply(value interface{}) (interface{}, error) {
	return t.translator(value)
}

func (t *translatorWrapper) MustApply(value interface{}) interface{} {
	v, _ := t.translator(value)
	return v
}

func WrapTranslator(typ string, trans func(value interface{}) (interface{}, error)) Translator {
	return &translatorWrapper{
		typ:        typ,
		translator: trans,
	}
}

func applyTranslator(src, dest Value, translator Translator) error {
	value, err := translator.Apply(src.Interface())
	if err != nil {
		return err
	}
	dest.Set(reflect.ValueOf(value))
	return nil
}

func Transform(value interface{}, translator Translator, src Field, dest Field) error {
	return applyTranslator(src.GetValue(value), dest.GetValue(value), translator)
}

var translatorRegistry = dynmarshaller.NewDynamicUnmarshaller[Translator](make(map[string]Translator))

func RegisterTranslator(translators ...Translator) {
	for _, translator := range translators {
		translatorRegistry.Register(translator.GetType(), translator)
	}
}

func UnmarshalTranslator(data []byte) (Translator, error) {
	return translatorRegistry.Unmarshal(data)
}
