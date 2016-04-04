package keys

import (
	"log"
	"os"
	"sort"
)

type ConfigFn func(*Keys) error

type Config interface {
	Order() int
	Configure(*Keys) error
}

type config struct {
	order int
	fn    ConfigFn
}

func DefaultConfig(fn ConfigFn) Config {
	return config{50, fn}
}

func NewConfig(order int, fn ConfigFn) Config {
	return config{order, fn}
}

func (c config) Order() int {
	return c.order
}

func (c config) Configure(k *Keys) error {
	return c.fn(k)
}

type configList []Config

func (c configList) Len() int {
	return len(c)
}

func (c configList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c configList) Less(i, j int) bool {
	return c[i].Order() < c[j].Order()
}

type Configuration interface {
	Add(...Config)
	AddFn(...ConfigFn)
	Configure(*Keys) error
	Configured() bool
}

type configuration struct {
	configured bool
	list       configList
}

func newConfiguration(conf ...Config) *configuration {
	c := &configuration{
		list: builtIns,
	}
	c.Add(conf...)
	return c
}

func (c *configuration) Add(conf ...Config) {
	c.list = append(c.list, conf...)
}

func (c *configuration) AddFn(fns ...ConfigFn) {
	for _, fn := range fns {
		c.list = append(c.list, DefaultConfig(fn))
	}
}

func configure(k *Keys, conf ...Config) error {
	for _, c := range conf {
		err := c.Configure(k)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *configuration) Configure(k *Keys) error {
	sort.Sort(c.list)

	err := configure(k, c.list...)
	if err != nil {
		k.Fatalf("%s", err.Error())
	}
	c.configured = true

	return err
}

func (c *configuration) Configured() bool {
	return c.configured
}

var builtIns = []Config{
	config{1000, DefaultLogger},
	config{1001, DefaultSettings},
	config{1002, DefaultLoader},
	config{1003, DefaultHandle},
	config{1004, loadAndConfigureChains},
}

func DefaultLogger(k *Keys) error {
	if k.Logger == nil {
		k.Logger = log.New(os.Stdout, "keter: ", log.Lmicroseconds|log.Llongfile)
	}
	return nil
}

func Logger(l *log.Logger) Config {
	return config{
		50,
		func(k *Keys) error {
			k.Logger = l
			return nil
		},
	}
}

func loadAndConfigureChains(k *Keys) error {
	chains, lErr := k.Load(k.LoadPath())
	if lErr != nil {
		return lErr //k.Fatalf("chain loading error: %s", lErr.Error())
	}

	ccErr := k.ConfigureChains(chains)
	if ccErr != nil {
		return ccErr //k.Fatalf("key chain configuration error: %s", ccErr.Error())
	}
	return nil
}
