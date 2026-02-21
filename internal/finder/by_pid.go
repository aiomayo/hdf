package finder

import (
	"github.com/aiomayo/hdf/internal/detect"
	"github.com/aiomayo/hdf/internal/process"
)

type pidStrategy struct{}

func (s *pidStrategy) Find(provider process.Provider, query detect.Query) ([]process.Info, error) {
	info, err := provider.FindByPID(query.PID)
	if err != nil {
		return nil, err
	}
	return []process.Info{*info}, nil
}
