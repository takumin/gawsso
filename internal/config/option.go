package config

type Option interface {
	Apply(*Config)
}

type LogLevel string

func (o LogLevel) Apply(c *Config) {
	c.LogLevel = string(o)
}

type IdentityStoreID string

func (o IdentityStoreID) Apply(c *Config) {
	c.IdentityStoreID = string(o)
}
