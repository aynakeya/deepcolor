package filters

import (
	"github.com/aynakeya/deepcolor/transform"
)

func init() {
	transform.RegisterFilter(RegExp(nil, false))
	transform.RegisterFilter(Or(nil), And(nil), Not(nil))
	transform.RegisterFilter(In([]int{}), Equal(1))
}
