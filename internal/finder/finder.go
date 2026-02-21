package finder

import (
	"fmt"

	"github.com/aiomayo/hdf/internal/detect"
	"github.com/aiomayo/hdf/internal/process"
)

type strategy interface {
	Find(provider process.Provider, query detect.Query) ([]process.Info, error)
}

type Finder struct {
	provider   process.Provider
	strategies map[detect.QueryType]strategy
}

func New(provider process.Provider) *Finder {
	return &Finder{
		provider: provider,
		strategies: map[detect.QueryType]strategy{
			detect.TypePID:      &pidStrategy{},
			detect.TypePort:     &portStrategy{},
			detect.TypeHostPort: &portStrategy{},
			detect.TypeName:     &nameStrategy{},
			detect.TypeGlob:     &nameStrategy{},
		},
	}
}

func (f *Finder) Find(query detect.Query) ([]process.Info, error) {
	s, ok := f.strategies[query.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported query type: %s", query.Type)
	}
	return s.Find(f.provider, query)
}

func (f *Finder) FindByUser(username string) ([]process.Info, error) {
	s := &userStrategy{}
	return s.Find(f.provider, detect.Query{Name: username})
}
