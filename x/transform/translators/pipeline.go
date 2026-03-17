package translators

import (
	"encoding/json"
	"errors"
	"github.com/aynakeya/deepcolor/transform"
)

type Pipeline struct {
	transform.BaseTranslator
	Steps []transform.Translator
}

type Switcher struct {
	Pipeline
}

func NewPipeline(steps ...transform.Translator) transform.Translator {
	t := &Pipeline{
		BaseTranslator: transform.BaseTranslator{
			Type: "Pipeline",
		},
		Steps: steps}
	return t
}

func NewSwitcher(steps ...transform.Translator) transform.Translator {
	t := &Switcher{
		Pipeline: Pipeline{
			BaseTranslator: transform.BaseTranslator{
				Type: "Switcher",
			},
			Steps: steps,
		},
	}
	return t
}

func (p *Pipeline) UnmarshalJSON(data []byte) error {
	type Tmp Pipeline
	aux := &struct {
		*Tmp
		Steps []json.RawMessage
	}{
		Tmp: (*Tmp)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	p.Steps = make([]transform.Translator, len(aux.Steps))
	for index, t := range aux.Steps {
		translator, err := transform.UnmarshalTranslator(t)
		if err != nil {
			return err
		}
		p.Steps[index] = translator
	}
	return nil
}

func (p *Pipeline) Apply(value interface{}) (interface{}, error) {
	var err error
	for _, trans := range p.Steps {
		value, err = trans.Apply(value)
		if err != nil {
			return value, err
		}
	}
	return value, nil
}

func (p *Pipeline) MustApply(value interface{}) interface{} {
	v, _ := p.Apply(value)
	return v
}

func (s *Switcher) Apply(value interface{}) (interface{}, error) {
	for _, trans := range s.Steps {
		v1, err := trans.Apply(value)
		if err == nil {
			return v1, nil
		}
	}
	return value, errors.New("not valid switch case")
}

func (s *Switcher) MustApply(value interface{}) interface{} {
	v, _ := s.Apply(value)
	return v
}
