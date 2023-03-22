package main

type config struct {
	VmmType         string `json:"type"`
	UnikernelCmd    string `json:"cmdline,omitempty"`
	UnikernelBinary string `json:"binary"`
}

func (c *config) encode() {
	c.VmmType = encode(c.VmmType)
	c.UnikernelCmd = encode(c.UnikernelCmd)
	c.UnikernelBinary = encode(c.UnikernelBinary)
}
