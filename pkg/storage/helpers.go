package storage

type SizeWriter struct {
	written int64
}

func (s *SizeWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	s.written += int64(n)
	return
}

func (s *SizeWriter) Size() int64 {
	return s.written
}
