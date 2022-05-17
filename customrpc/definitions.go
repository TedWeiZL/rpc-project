package customrpc

// error
type WrappedErr struct {
	Msg string `json:"msg"`
}

func (we *WrappedErr) Error() string {
	return we.Msg
}
