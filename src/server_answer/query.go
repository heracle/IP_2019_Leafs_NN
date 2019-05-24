package infogen

import (
	"strconv"
)

var example = `
ID: 0
Name: pubescent bamboo
Details:
  Wikipedia: https://en.wikipedia.org/wiki/Phyllostachys_edulis
`

// QueryInfo denotes the format of the returned message. JSON fields must be public in order for unmarshal to correctly populate the datamlObja.
type QueryInfo struct {
	ID            string
	Common_name   string
	Specific_name string
	Details       struct {
		Wikipedia string
	}
}

// http://flavia.sourceforge.net/

var commonNames = []string{"Pubescent bamboo", "Chinese horse chestnut", "Anhui Barberry",
	"Chinese redbud", "True indigo", "Japanese maple", "Nanmu", "Castor aralia",
	"Chinese cinnamon", "Goldenrain tree", "Big-fruited Holly", "Japanese cheesewood",
	"Wintersweet", "Camphortree", "Japan Arrowwood", "Sweet osmanthus", "Deodar", "Ginkgo, maidenhair tree",
	"Crape myrtle, Crepe myrtle", "Oleander", "Yew plum pine", "Japanese Flowering Cherry", "Glossy Privet",
	"Chinese Toon", "Peach", "Ford Woodlotus", "Trident maple", "Beales barberry", "Southern magnolia",
	"Canadian poplar", "Chinese tulip tree", "Tangerine",
}

var specificNames = []string{"Phyllostachys edulis (Carr.) Houz", "Aesculus chinensis", "Berberis anhweiensis Ahrendt",
	"Cercis chinensis", "Indigofera tinctoria L.", "Acer Palmatum", "Phoebe nanmu (Oliv.) Gamble",
	"Kalopanax septemlobus (Thunb. ex A.Murr.) Koidz.", "Cinnamomum japonicum Sieb.", "Koelreuteria paniculata Laxm.",
	"Ilex macrocarpa Oliv.", "Pittosporum tobira (Thunb.) Ait. f.", "Chimonanthus praecox L.",
	"Cinnamomum camphora (L.) J. Presl", "Viburnum awabuki K.Koch", "Osmanthus fragrans Lour.",
	"Cedrus deodara (Roxb.) G. Don", "Ginkgo biloba L.", "Lagerstroemia indica (L.) Pers.",
	"Nerium oleander L.", "Podocarpus macrophyllus (Thunb.) Sweet",
	"Prunus serrulata Lindl. var. lannesiana auct.", "Ligustrum lucidum Ait. f.",
	"Tonna sinensis M. Roem.", "Prunus persica (L.) Batsch", "Manglietia fordiana Oliv.",
	"Acer buergerianum Miq.", "Mahonia bealei (Fortune) Carr.", "Magnolia grandiflora L.",
	"Populus ×canadensis Moench", "Liriodendron chinense (Hemsl.) Sarg.", "Citrus reticulata Blanco",
}

var wikiLinks = []string{
	"https://en.wikipedia.org/wiki/Phyllostachys_edulis",
	"https://en.wikipedia.org/wiki/Aesculus",
	"https://en.wikipedia.org/wiki/Berberis",
	"https://en.wikipedia.org/wiki/Cercis",
	"https://en.wikipedia.org/wiki/Indigofera_tinctoria",
	"https://en.wikipedia.org/wiki/Acer_palmatum",
	"https://en.wikipedia.org/wiki/Phoebe_nanmu",
	"https://en.wikipedia.org/wiki/Kalopanax",
	"https://en.wikipedia.org/wiki/Cinnamomum_pedunculatum",
	"https://en.wikipedia.org/wiki/Koelreuteria_paniculata",
	"https://en.wikipedia.org/wiki/Holly",
	"https://en.wikipedia.org/wiki/Pittosporum_tobira",
	"https://en.wikipedia.org/wiki/Chimonanthus",
	"https://en.wikipedia.org/wiki/Cinnamomum_camphora",
	"https://en.wikipedia.org/wiki/Viburnum_odoratissimum",
	"https://en.wikipedia.org/wiki/Osmanthus_fragrans",
	"https://en.wikipedia.org/wiki/Cedrus_deodara",
	"https://en.wikipedia.org/wiki/Ginkgo_biloba",
	"https://en.wikipedia.org/wiki/Lagerstroemia_indica",
	"https://en.wikipedia.org/wiki/Nerium",
	"https://en.wikipedia.org/wiki/Podocarpus_macrophyllus",
	"https://en.wikipedia.org/wiki/Prunus_serrulata",
	"https://en.wikipedia.org/wiki/Ligustrum_lucidum",
	"https://en.wikipedia.org/wiki/Toona_sinensis",
	"https://en.wikipedia.org/wiki/Peach",
	"https://en.wikipedia.org/wiki/Manglietia",
	"https://en.wikipedia.org/wiki/Maple",
	"https://en.wikipedia.org/wiki/Mahonia_bealei",
	"https://en.wikipedia.org/wiki/Magnolia_grandiflora",
	"https://en.wikipedia.org/wiki/Populus_%C3%97_canadensis",
	"https://en.wikipedia.org/wiki/Liriodendron_chinense",
	"https://en.wikipedia.org/wiki/Mandarin_orange",
}

// GetInfoForClass returns a json with details about the plant with "classID".
func GetInfoForClass(classID int) (*QueryInfo, error) {
	jsonObj := new(QueryInfo)
	jsonObj.ID = strconv.Itoa(classID)

	if classID < 0 || classID > len(commonNames) {
		jsonObj.Common_name = "Could not find any leaf in the image"
		return jsonObj, nil
	}

	jsonObj.Common_name = commonNames[classID]
	jsonObj.Specific_name = specificNames[classID]
	jsonObj.Details.Wikipedia = wikiLinks[classID]
	return jsonObj, nil
}
