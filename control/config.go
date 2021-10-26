package control

import "net/url"

type Config struct {
	MeshCentralURL         *url.URL
	MeshCentralGroupPrefix string `split_words:"true" default:"plexus"`
	MeshCentralUser        string `split_words:"true" required:"true"`
	MeshCentralPass        string `split_words:"true" required:"true"`
	MeshCentralDomain      string `split_words:"true" default:"control"`
}

func (c *Config) MeshCentralControlURL() string {
	return c.MeshCentralURL.String() + "/" + c.MeshCentralDomain + "/control.ashx"
}
func (c *Config) MeshCentralAgentURL() string {
	return c.MeshCentralURL.String() + "/" + c.MeshCentralDomain + "/agent.ashx"
}
func (c *Config) MeshRelayURL() string {
	return c.MeshCentralURL.String() + "/" + c.MeshCentralDomain + "/meshrelay.ashx"
}
