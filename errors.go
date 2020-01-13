package gwf

type ErrorType uint8

const (
	// ErrorTypeInternal 内部错误类型
	ErrorTypeInternal ErrorType = iota
)

type Error struct {
	Err   interface{}
	Type  ErrorType
	Stack string
}
