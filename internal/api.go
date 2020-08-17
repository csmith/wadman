package internal

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
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
	Versions    []string  `json:"gameVersion"`
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
	var matches []AddonFile

	for i := range details.Files {
		f := details.Files[i]

		if verbose {
			fmt.Printf(
				"Found file %d (%s)\n"+
					"\tFlavour: %s (valid: %t)\n"+
					"\tType: %d (valid: %t)\n"+
					"\tAlternative: %t (valid: %t)\n"+
					"\n",
				f.FileId,
				f.DisplayName,
				f.Flavour,
				f.Flavour == "wow_retail",
				f.Type,
				f.Type <= Beta,
				f.Alternate,
				!f.Alternate,
			)
		}

		if f.Flavour == "wow_retail" && f.Type <= Beta && !f.Alternate {
			matches = append(matches, f)
		}
	}

	if verbose {
		fmt.Printf("Found %d potential versions:\n", len(matches))
	}

	var bestFile *AddonFile
	bestAge := math.MaxFloat64
	bestValid := false
	for i := range matches {
		f := matches[i]
		age := time.Now().Sub(f.Date).Seconds()
		valid := validVersion(&f)
		if (valid == bestValid && age < bestAge) || (!bestValid && valid) {
			bestFile = &f
			bestAge = age
			bestValid = valid
			if verbose {
				fmt.Printf("\t[%d] Time: %s; Versions: %s << Best so far\n", f.FileId, f.Date, strings.Join(f.Versions, ","))
			}
		} else if verbose {
			fmt.Printf("\t[%d] Time: %s; Versions: %s << SKIPPED\n", f.FileId, f.Date, strings.Join(f.Versions, ","))
		}
	}

	if verbose {
		fmt.Println()
	}

	return bestFile
}

func validVersion(file *AddonFile) bool {
	var invalid = false
	for _, v := range file.Versions {
		// TODO: Make this configurable or dynamic based on the client version
		if strings.HasPrefix(v, "8.") || strings.HasPrefix(v, "7.") {
			return true
		} else {
			invalid = true
		}
	}
	return !invalid
}
