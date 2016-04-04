package keys

import "github.com/wSCP/xandle/x"

type Handle interface {
	x.Handle
	KeyStore
}

type handle struct {
	x.Handle
	KeyStore
}

func NewHandle(display string) (Handle, error) {
	xh, err := x.New(display)
	kb, err := NewKeyboard(xh.Conn(), xh.Setup())
	if err != nil {
		return nil, err
	}

	h := &handle{}
	h.Handle = xh

	ks := NewKeyStore(kb)

	var calls = []x.Xevent{
		x.KeyPressEvent,
		x.KeyReleaseEvent,
		x.ButtonPressEvent,
		x.ButtonReleaseEvent,
	}
	for _, c := range calls {
		xh.SetCaller(c, ks.Caller())
	}

	h.KeyStore = ks
	return h, nil
}

func DefaultHandle(k *Keys) error {
	if k.Handle == nil {
		h, err := NewHandle("")
		if err != nil {
			return err
		}
		k.Handle = h
	}
	return nil
}
