package constants

type CtxKey string

const (
	CtxRequestID CtxKey = "RequestID"
	CtxTraceID   CtxKey = "TraceID"
)

func (key CtxKey) String() string {
	return string(key)
}
