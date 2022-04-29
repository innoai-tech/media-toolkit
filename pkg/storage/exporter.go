package storage

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/opencontainers/go-digest"
	"io"
	"path"
)

func ExportDataset(ctx context.Context, store Store, w io.Writer, blobs []blob.Info) (dgst digest.Digest, err error) {
	digester := digest.Canonical.Digester()
	mw := io.MultiWriter(w, digester.Hash())

	gw := gzip.NewWriter(mw)
	defer func() {
		_ = gw.Close()
		// should calc digest when gz close
		dgst = digester.Digest()
	}()

	tw := tar.NewWriter(gw)
	defer func() {
		_ = tw.Close()
	}()

	labels := bytes.NewBuffer(nil)

	for i := range blobs {
		b := blobs[i]

		r, err := store.ReaderAt(ctx, b.Ref)
		if err != nil {
			return "", err
		}

		_, _ = fmt.Fprintln(labels, b.String())

		if err := copyToTar(tw, io.NewSectionReader(r, 0, r.Size()), tar.Header{
			Name: path.Join(b.Ref.BlobPath("")),
			Size: r.Size(),
		}); err != nil {
			return "", err
		}
	}

	if err := copyToTar(tw, labels, tar.Header{
		Name: "labels",
		Size: int64(labels.Len()),
	}); err != nil {
		return "", err
	}

	return "", nil
}

func copyToTar(tw *tar.Writer, r io.Reader, header tar.Header) error {
	header.Mode = 0644
	if err := tw.WriteHeader(&header); err != nil {
		return err
	}
	if _, err := io.Copy(tw, r); err != nil {
		return err
	}
	return nil
}
