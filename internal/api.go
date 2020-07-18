package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Type int

const (
	Release Type = 1
	Beta         = 2
	Alpha        = 3
)

type AddonFile struct {
	FileId      int       `json:"id"`
	Flavour     string    `json:"gameVersionFlavor"`
	Type        Type      `json:"releaseType"`
	Url         string    `json:"downloadUrl"`
	Date        time.Time `json:"fileDate"`
	Alternate   bool      `json:"isAlternate"`
	DisplayName string    `json:"displayName"`
}

type AddonResponse struct {
	Id    int         `json:"id"`
	Name  string      `json:"name"`
	Files []AddonFile `json:"latestFiles"`
}

func GetAddon(id int) (*AddonResponse, error) {
	res, err := http.Get(fmt.Sprintf("https://addons-ecs.forgesvc.net/api/v2/addon/%d", id))
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	addon := &AddonResponse{}
	err = json.NewDecoder(res.Body).Decode(addon)
	return addon, err
}

func SearchAddons(query string) ([]*AddonResponse, error) {
	res, err := http.Get(fmt.Sprintf("https://addons-ecs.forgesvc.net/api/v2/addon/search?gameId=1&searchFilter=%s", url.QueryEscape(query)))
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var addons []*AddonResponse
	err = json.NewDecoder(res.Body).Decode(&addons)
	return addons, err
}

func LatestFile(details *AddonResponse, verbose bool) *AddonFile {
	var (
		latestTime time.Time
		latestFile *AddonFile
	)

	for i := range details.Files {
		f := details.Files[i]

		if verbose {
			fmt.Printf(
				"Found file %d (%s)\n\tFlavour: %s (valid: %t)\n\tType: %d (valid: %t)\n\tAlternative: %t (valid: %t)\n\tTime: %s (latest: %s; valid: %t)\n\n",
				f.FileId,
				f.DisplayName,
				f.Flavour,
				f.Flavour == "wow_retail",
				f.Type,
				f.Type <= Beta,
				f.Alternate,
				!f.Alternate,
				f.Date,
				latestTime,
				f.Date.After(latestTime),
			)
		}

		if f.Flavour == "wow_retail" && f.Type <= Beta && !f.Alternate && f.Date.After(latestTime) {
			latestTime = f.Date
			latestFile = &f
		}
	}

	return latestFile
}
