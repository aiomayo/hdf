package finder

import (
	"path/filepath"
	"strings"

	"github.com/aiomayo/hdf/internal/detect"
	"github.com/aiomayo/hdf/internal/process"
)

type nameStrategy struct{}

func (s *nameStrategy) Find(provider process.Provider, query detect.Query) ([]process.Info, error) {
	all, err := provider.List()
	if err != nil {
		return nil, err
	}

	var result []process.Info
	for _, info := range all {
		if query.Type == detect.TypeGlob {
			if matchGlob(info.Name, query.Name) || matchGlob(info.Cmdline, query.Name) {
				result = append(result, info)
			}
		} else {
			if matchName(info, query.Name) {
				result = append(result, info)
			}
		}
	}
	return result, nil
}

func matchName(info process.Info, name string) bool {
	lower := strings.ToLower(name)
	if strings.ToLower(info.Name) == lower {
		return true
	}
	if strings.Contains(strings.ToLower(info.Cmdline), lower) {
		return true
	}
	return false
}

func matchGlob(value, pattern string) bool {
	matched, err := filepath.Match(strings.ToLower(pattern), strings.ToLower(value))
	if err != nil {
		return false
	}
	return matched
}
