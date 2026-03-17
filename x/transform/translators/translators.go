package translators

import (
	"github.com/aynakeya/deepcolor/transform"
)

func init() {
	transform.RegisterTranslator(&Pipeline{})
	transform.RegisterTranslator(&Switcher{})
	transform.RegisterTranslator(&Foreach{})
	transform.RegisterTranslator(&RegExpReplacer{})
	transform.RegisterTranslator(&RegExpFind{})
	transform.RegisterTranslator(&StrCase{})
	transform.RegisterTranslator(&Formatter{})
	transform.RegisterTranslator(&Cast{})
	transform.RegisterTranslator(&Value{})
}
