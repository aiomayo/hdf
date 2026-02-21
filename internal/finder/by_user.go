package finder

import (
	"strings"

	"github.com/aiomayo/hdf/internal/detect"
	"github.com/aiomayo/hdf/internal/process"
)

type userStrategy struct{}

func (s *userStrategy) Find(provider process.Provider, query detect.Query) ([]process.Info, error) {
	all, err := provider.List()
	if err != nil {
		return nil, err
	}

	var result []process.Info
	for _, info := range all {
		if strings.EqualFold(info.User, query.Name) {
			result = append(result, info)
		}
	}
	return result, nil
}
