package sign

type Provider interface {
	Sign(cmd string, seq uint32, data []byte) (*Response, error)
}
