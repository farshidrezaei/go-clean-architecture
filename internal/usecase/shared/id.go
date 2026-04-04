package shared

type IDGenerator interface {
	NewString() string
}
