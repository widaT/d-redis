package server

type Config struct {
	net     string
	laddr   string
	snapCount uint64
}

func DefaultConfig() *Config {
	return &Config{
		net:    "tcp",
		laddr:    ":6380",
		snapCount :100000,
	}
}

func (c *Config) Net(p string) *Config {
	c.net = p
	return c
}

func (c *Config) Laddr(h string) *Config {
	c.laddr = h
	return c
}

func (c *Config) SnapCount(n uint64) *Config {
	c.snapCount = n
	return c
}
func (c *Config)Gaddr() string{
	return c.laddr
}