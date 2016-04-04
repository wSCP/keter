package keys

import "fmt"

type keysError struct {
	err  string
	vals []interface{}
}

func (k *keysError) Error() string {
	return fmt.Sprintf("[keys] %s", fmt.Sprintf(k.err, k.vals...))
}

func (k *keysError) Out(vals ...interface{}) *keysError {
	k.vals = vals
	return k
}

func Krror(err string) *keysError {
	return &keysError{err: err}
}
