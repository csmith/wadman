package wadman

import (
	"encoding/json"
	"fmt"
	"github.com/csmith/wadman/wow"
	"io"
	"net/http"
)

type WowInterfaceAddon struct {
	BaseAddon
	Id           int    `json:"id"`
	Title        string `json:"title"`
	LastChecksum string `json:"last_checksum"`
}

func NewWowInterfaceAddon(id int) Addon {
	return &WowInterfaceAddon{BaseAddon: BaseAddon{Type: TypeWowInterface}, Id: id}
}

func (w *WowInterfaceAddon) DisplayName() string {
	return w.Title
}

func (w *WowInterfaceAddon) ShortName() string {
	return fmt.Sprintf("wowi:%d", w.Id)
}

func (w *WowInterfaceAddon) Update(install *wow.Install, _ io.Writer, force bool) (updated bool, err error) {
	var response []struct {
		Id       int    `json:"id"`
		Version  string `json:"version"`
		Checksum string `json:"checksum"`
		Url      string `json:"downloadUri"`
		Title    string `json:"title"`
	}

	url := fmt.Sprintf("https://api.mmoui.com/v4/game/WOW/filedetails/%d.json", w.Id)
	res, err := http.Get(url)
	if err != nil {
		return false, err
	}

	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return false, err
	}

	if len(response) != 1 {
		return false, fmt.Errorf("expected 1 result, got %d", len(response))
	}

	w.Title = response[0].Title
	if w.LastChecksum != response[0].Checksum || force {
		// New version to install
		dirs, err := install.InstallAddonFromUrl(response[0].Url)
		if err != nil {
			return false, err
		}

		w.LastChecksum = response[0].Checksum
		w.Directories = dirs
		w.Version = response[0].Version
		return true, nil
	} else {
		return false, nil
	}
}
