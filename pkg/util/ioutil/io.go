package ioutil

import "io"

func Pipe(src func(w io.Writer) error, dest func(r io.Reader) error) error {
	errChan := make(chan error)
	defer close(errChan)

	reader, writer := io.Pipe()
	go func() {
		defer reader.Close()

		errChan <- dest(io.NopCloser(reader))
	}()

	if err := src(writer); err != nil {
		return err
	}
	_ = writer.Close()
	return <-errChan
}
