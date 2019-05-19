package infogen

import (
	"gopkg.in/yaml.v2"
	"github.com/pkg/errors"
)

var example = `
ID: 0
Name: pubescent bamboo
Details:
  Wikipedia: https://en.wikipedia.org/wiki/Phyllostachys_edulis
`

// Note: YAML fields must be public in order for unmarshal to correctly populate the datamlObja.
type yamlClass struct {
	ID string
	Name string
	Details struct {
		Wikipedia string
	}
}

var commonNames = []string {"Pubescent bamboo", "Chinese horse chestnut", "Anhui Barberry",
"Chinese redbud", "True indigo", "Japanese maple", "Nanmu", "Castor aralia",
"Chinese cinnamon", "Goldenrain tree", "Big-fruited Holly", "Japanese cheesewood",
"Wintersweet", "Camphortree", "Japan Arrowwood", "Sweet osmanthus", "Deodar", "Ginkgo, maidenhair tree",
"Crape myrtle, Crepe myrtle", "Oleander", "Yew plum pine", "Japanese Flowering Cherry", "Glossy Privet",
"Chinese Toon", "Peach", "Ford Woodlotus", "Trident maple", "Beales barberry", "Southern magnolia",
"Canadian poplar", "Chinese tulip tree", "Tangerine"}

var wikiLinks = []string {
	"https://en.wikipedia.org/wiki/Phyllostachys_edulis",
}

// GetInfoForClass returns a yaml with details about the plant with "classID".
func GetInfoForClass(classID int) ([]byte, error){
	yamlObj := new(yamlClass)
	yamlObj.ID = string(classID)
	yamlObj.Name = commonNames[classID]
	yamlObj.Details.Wikipedia = wikiLinks[classID]

	ret, err := yaml.Marshal(&yamlObj)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to Marshal the yaml")
	}
	return ret, nil
}