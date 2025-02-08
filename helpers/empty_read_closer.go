package helpers

import "io"

type EmptyReadCloser struct{}

func (e EmptyReadCloser) Close() error { return nil }
func (e EmptyReadCloser) Read(p []byte) (int, error) {
	return 0, io.EOF
}
