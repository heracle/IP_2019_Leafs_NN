package infogen

import (
	"github.com/pkg/errors"
)

const noClasses = 32

// AllClassesInfo the json format returned by all classes handler.
type AllClassesInfo struct {
	Data [noClasses]QueryInfo
}

// GenerateAllClassesInfo returns a json with info for all classes stored.
func GenerateAllClassesInfo() (*AllClassesInfo, error) {
	var ret AllClassesInfo
	for classID := 0; classID < noClasses; classID ++ {
		act, err := GetInfoForClass(classID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get YAML bytes for classID %v", classID)
		}
		ret.Data[classID] = *act
	}
	return &ret, nil
}
