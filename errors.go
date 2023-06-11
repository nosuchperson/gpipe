package gpipe

import "fmt"

type GPWError struct {
	msg string
}

func (G GPWError) Error() string {
	return G.msg
}

func newGPWError(fmtStr string, args ...interface{}) GPWError {
	return GPWError{msg: fmt.Sprintf(fmtStr, args...)}
}

var (
	ErrEngineIsRunning = newGPWError("engine is running")
)
