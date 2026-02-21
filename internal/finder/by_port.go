package finder

import (
	"github.com/aiomayo/hdf/internal/detect"
	"github.com/aiomayo/hdf/internal/process"
)

type portStrategy struct{}

func (s *portStrategy) Find(provider process.Provider, query detect.Query) ([]process.Info, error) {
	return provider.FindByPort(query.Port)
}
