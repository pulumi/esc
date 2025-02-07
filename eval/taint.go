package eval

type taintFlags uint8

const (
	taintInvalid taintFlags = 1 << iota
	taintUnknown
	taintSecret
	taintRotateOnly
)

func (t taintFlags) has(mask taintFlags) bool {
	return t&mask == mask
}
