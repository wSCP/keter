package keys

import (
	"sync"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/wSCP/xandle/x"
)

type KeyStore interface {
	Keyboard() *Keyboard
	Caller() x.Call
	Put(Call, Keyable)
	Get(x.Event) ([]Keyable, string, error)
}

type keyStore struct {
	keyboard *Keyboard
	keys     map[x.Xevent]map[xproto.Window]map[Call][]Keyable
	lck      *sync.RWMutex
}

func NewKeyStore(kb *Keyboard) KeyStore {
	return &keyStore{
		kb,
		make(map[x.Xevent]map[xproto.Window]map[Call][]Keyable),
		&sync.RWMutex{},
	}
}

func (k *keyStore) Keyboard() *Keyboard {
	return k.keyboard
}

type Call struct {
	evt x.Xevent
	win xproto.Window
	mod uint16
	cod byte
	but byte
}

func mkCall(e x.Xevent, w xproto.Window, m uint16, c byte, b byte) Call {
	return Call{e, w, m, c, b}
}

func (k *keyStore) Caller() x.Call {
	return func(e x.Event) error {
		keyables, params, err := k.Get(e)
		if err == nil {
			for _, kb := range keyables {
				runKeyable(kb, params)
			}
		}
		return nil
	}
}

func (k *keyStore) Put(c Call, kb Keyable) {
	k.lck.Lock()
	defer k.lck.Unlock()

	if _, ok := k.keys[c.evt]; !ok {
		k.keys[c.evt] = make(map[xproto.Window]map[Call][]Keyable)
	}
	if _, ok := k.keys[c.evt][c.win]; !ok {
		k.keys[c.evt][c.win] = make(map[Call][]Keyable)
	}

	k.keys[c.evt][c.win][c] = append(k.keys[c.evt][c.win][c], kb)
}

func parseXprotoEvent(e interface{}) (xproto.Window, uint16, byte, byte) {
	var w xproto.Window
	var s uint16
	var d, b byte
	switch evt := e.(type) {
	case xproto.KeyPressEvent:
		w, s, d = evt.Root, evt.State, byte(evt.Detail)
	case xproto.KeyReleaseEvent:
		w, s, d = evt.Root, evt.State, byte(evt.Detail)
	case xproto.ButtonPressEvent:
		w, s, b = evt.Root, evt.State, byte(evt.Detail)
	case xproto.ButtonReleaseEvent:
		w, s, b = evt.Root, evt.State, byte(evt.Detail)
	}
	return w, s, d, b
}

var NoKey = Krror("there is no corresponding key for %+v").Out

func runKeyable(k Keyable, param string) {
	go k.Run(param)
}

func (k *keyStore) Get(e x.Event) ([]Keyable, string, error) {
	k.lck.RLock()
	defer k.lck.RUnlock()

	win, mods, key, button := parseXprotoEvent(e.Evt)

	if ky, ok := k.keys[e.Tag][win][mkCall(e.Tag, win, mods, key, button)]; ok {
		return ky, ByteToString(k.Keyboard(), key, button), nil
	}

	return nil, "", NoKey(e)
}
