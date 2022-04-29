package storage

import (
	"encoding/json"
	"errors"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/content/local"
	"github.com/opencontainers/go-digest"
	"os"
	"path/filepath"
)

var (
	LabelNotExists = errors.New("label not exists")
)

type Store = content.Store

func NewStore(root string) (Store, error) {
	return local.NewLabeledStore(root, &localLabelStore{root: root})
}

type localLabelStore struct {
	root string
}

func (l *localLabelStore) Set(digest digest.Digest, labels map[string]string) error {
	manifestPath := filepath.Join(l.root, "manifests", digest.Algorithm().String(), digest.Hex())

	f, err := CreateOrOpen(manifestPath)
	if err != nil {
		if err != nil {
			return err
		}
	}
	defer f.Close()

	ref, ok := labels["$ref"]
	if ok {
		delete(labels, "$ref")

		if mt, ok := labels["mediaType"]; ok {
			switch mt {
			case MediaTypeImageJPEG:
				blobPath := filepath.Join(l.root, "blobs", digest.Algorithm().String(), digest.Hex())
				imagePath := filepath.Join(l.root, "_links", "images", ref+".jpg")
				if err := ForceSymlink(blobPath, imagePath); err != nil {
					return err
				}
			case MediaTypeVideoMP4:
				blobPath := filepath.Join(l.root, "blobs", digest.Algorithm().String(), digest.Hex())
				imagePath := filepath.Join(l.root, "_links", "videos", ref+".mp4")
				if err := ForceSymlink(blobPath, imagePath); err != nil {
					return err
				}
			}
		}
	}

	return json.NewEncoder(f).Encode(labels)
}

func (l *localLabelStore) Update(digest digest.Digest, labels map[string]string) (map[string]string, error) {
	currentLabels, err := l.Get(digest)
	if err != nil {
		if err == LabelNotExists {
			return labels, l.Set(digest, labels)
		}
		return nil, err
	}
	for k, v := range labels {
		currentLabels[k] = v
	}
	return currentLabels, l.Set(digest, currentLabels)
}

func (l *localLabelStore) Get(digest digest.Digest) (ret map[string]string, err error) {
	p := filepath.Join(l.root, "manifests", digest.Algorithm().String(), digest.Hex())
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, LabelNotExists
		}
		return nil, err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&ret); err != nil {
		return nil, err
	}
	return
}

func CreateOrOpen(name string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(name), os.ModePerm); err != nil {
		return nil, err
	}
	f, err := os.Create(name)
	if err != nil {
		if os.IsExist(err) {
			return os.Open(name)
		}
		return nil, err
	}
	return f, nil
}

func ForceSymlink(from string, to string) error {
	if err := os.MkdirAll(filepath.Dir(to), os.ModePerm); err != nil {
		return err
	}
	_ = os.RemoveAll(to)
	return os.Symlink(from, to)
}
