package lakehouse

import (
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
)

func containerdImportTar(ctx context.Context, path string) (string, func(), error) {
	noop := func() {}
	if err := ctx.Err(); err != nil {
		return "", noop, err
	}
	source, err := os.Open(path)
	if err != nil {
		return "", noop, err
	}
	defer source.Close()
	var magic [2]byte
	if _, err := io.ReadFull(source, magic[:]); err != nil {
		return "", noop, err
	}
	if magic != [2]byte{0x1f, 0x8b} {
		return path, noop, nil
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return "", noop, err
	}
	reader, err := gzip.NewReader(source)
	if err != nil {
		return "", noop, err
	}
	defer reader.Close()
	target, err := os.CreateTemp(filepath.Dir(path), ".import-*.tar")
	if err != nil {
		return "", noop, err
	}
	cleanup := func() { _ = os.Remove(target.Name()) }
	ok := false
	defer func() {
		_ = target.Close()
		if !ok {
			cleanup()
		}
	}()
	// Bound expansion even for a checksum-pinned archive. Callers already bound
	// the download size; this independently limits temporary disk consumption.
	const limit = int64(64 << 30)
	n, err := io.CopyBuffer(target, &imageContextReader{ctx: ctx, reader: io.LimitReader(reader, limit+1)}, make([]byte, 1<<20))
	if err != nil {
		return "", noop, err
	}
	if n > limit {
		return "", noop, errors.New("image archive exceeds the 64 GiB decompression limit")
	}
	if err := target.Close(); err != nil {
		return "", noop, err
	}
	ok = true
	return target.Name(), cleanup, nil
}

type imageContextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *imageContextReader) Read(data []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(data)
}
