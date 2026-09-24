package storage

import "time"

// DefaultFavorites are baked-in stations for a first launch (no favorites.json yet).
func DefaultFavorites() []FavoriteStation {
	added := time.Now().UTC().Format(time.RFC3339)
	return []FavoriteStation{
		{
			UUID:    "e6fa9a8a-02a8-11e9-a1be-52543be04c81",
			Name:    "SomaFM Groove Salad Classic",
			URL:     "https://ice6.somafm.com/gsclassic-128-mp3",
			Tags:    "ambient,downtempo,electronica",
			Country: "US",
			Bitrate: 128,
			Codec:   "MP3",
			AddedAt: added,
		},
		{
			UUID:    "960eb2e9-0601-11e8-ae97-52543be04c81",
			Name:    "SomaFM Drone Zone",
			URL:     "https://ice2.somafm.com/dronezone-128-mp3",
			Tags:    "ambient,drone",
			Country: "US",
			Bitrate: 128,
			Codec:   "MP3",
			AddedAt: added,
		},
		{
			UUID:    "960c36e6-0601-11e8-ae97-52543be04c81",
			Name:    "Echoes of Bluemars",
			URL:     "http://streams.echoesofbluemars.org:8000/bluemars",
			Tags:    "ambient,space",
			Country: "US",
			Bitrate: 128,
			Codec:   "MP3",
			AddedAt: added,
		},
		{
			UUID:    "db9f38eb-9f35-4bdf-bdc9-00df1228de2f",
			Name:    "Echoes of Bluemars - Cryosleep",
			URL:     "http://streams.echoesofbluemars.org:8000/cryosleep",
			Tags:    "ambient,space,drone",
			Country: "US",
			Bitrate: 128,
			Codec:   "MP3",
			AddedAt: added,
		},
		{
			UUID:    "5886456b-6ed3-40c8-a2c4-4167793d225b",
			Name:    "Ambient Sleeping Pill",
			URL:     "https://radio.stereoscenic.com/asp-h",
			Tags:    "ambient,sleep",
			Country: "US",
			Bitrate: 256,
			Codec:   "MP3",
			AddedAt: added,
		},
	}
}
