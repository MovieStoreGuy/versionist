package filewalk

import (
	"io/fs"
	"os"

	"go.uber.org/multierr"

	"github.com/MovieStoreGuy/versionist/pkg/internal/generic"
)

// WalkedFS is used when walking
// a file system and only capturing a subset.
type WalkedFS map[string]struct{}

var (
	_ fs.FS = (*WalkedFS)(nil)
)

func NewWalkedFS(root string, glob string) (WalkedFS, error) {
	matched, err := fs.Glob(os.DirFS(root), glob)
	if err != nil {
		return nil, err
	}
	wfs := make(WalkedFS)
	for _, m := range matched {
		wfs[m] = struct{}{}
	}
	return wfs, nil
}

func (wfs WalkedFS) Open(name string) (fs.File, error) {
	if _, matched := wfs[name]; !fs.ValidPath(name) || !matched {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	return os.Open(name)
}

func (wfs WalkedFS) Range(fn func(name string, f *os.File) error) error {
	return generic.ParallelRangeMap(wfs, func(name string, _ struct{}) error {
		stat, err := os.Stat(name)
		if err != nil {
			return err
		}
		f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, stat.Mode())
		if err != nil {
			return err
		}
		return multierr.Combine(
			fn(name, f),
			f.Close(),
		)
	})
}
