package getter

// configure configures a client with options.
func (c *Client) configure() error {
	// Default decompressor values
	if c.Decompressors == nil {
		c.Decompressors = Decompressors
	}
	// Default detector values
	if c.Detectors == nil {
		c.Detectors = Detectors
	}
	// Default getter values
	if c.Getters == nil {
		c.Getters = Getters
	}

	return nil
}
