package gardenrunc

import "io"

type StreamHandler struct {
}

func (c *StreamHandler) StreamIn(dstPath string, tarStream io.Reader) error {
	panic("not implemented: streamin")
	return nil
}

func (c *StreamHandler) StreamOut(srcPath string) (io.ReadCloser, error) {
	panic("not implemented: streamout")
	return nil, nil
}
