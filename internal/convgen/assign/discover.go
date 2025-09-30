package assign

import (
	"errors"
	"go/token"

	"github.com/emirpasic/gods/maps/hashbidimap"

	"github.com/sublee/convgen/internal/convgen/match"
	"github.com/sublee/convgen/internal/convgen/parse"
)

type (
	addFunc[T Object] = func(T, string)
	deleteFunc        = func(token.Pos)
)

type discovery[T Object] interface {
	DiscoverX(addFunc[T], deleteFunc) error
	DiscoverY(addFunc[T], deleteFunc) error
	ResolveX(parse.Path) (T, string, error)
	ResolveY(parse.Path) (T, string, error)
}

func discover[T Object](fac *factory, m *match.Matcher[T], d discovery[T]) error {
	explicit := hashbidimap.New()
	for _, pair := range fac.cfg.Match {
		// Index by last element's position
		explicit.Put(pair[0].Pos, pair[1].Pos)
	}

	errs := d.DiscoverX(func(obj T, key string) {
		if !obj.Exported() {
			if _, ok := explicit.Get(obj.Pos()); !ok {
				if !fac.cfg.DiscoverUnexportedEnabled || !fac.cfg.DiscoverUnexportedX {
					return
				}
			}
		}
		m.AddX(obj, key)
	}, m.DeleteX)

	errs = errors.Join(errs, d.DiscoverY(func(obj T, key string) {
		if !obj.Exported() {
			if _, ok := explicit.GetKey(obj.Pos()); !ok {
				if !fac.cfg.DiscoverUnexportedEnabled || !fac.cfg.DiscoverUnexportedY {
					return
				}
			}
		}
		m.AddY(obj, key)
	}, m.DeleteY))

	for _, pair := range fac.cfg.Match {
		pathX, pathY := pair[0], pair[1]

		objX, keyX, err := d.ResolveX(pathX)
		if err != nil {
			errs = errors.Join(errs, err)
		} else {
			m.AddX(objX, keyX)
		}

		objY, keyY, err := d.ResolveY(pathY)
		if err != nil {
			errs = errors.Join(errs, err)
		} else {
			m.AddY(objY, keyY)
		}
	}
	return errs
}
