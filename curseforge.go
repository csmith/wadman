package wadman

import (
	"fmt"
	"github.com/csmith/wadman/curse"
	"github.com/csmith/wadman/wow"
	"io"
)

type CurseForgeAddon struct {
	BaseAddon
	Id     int    `json:"id"`
	Name   string `json:"name"`
	FileId int    `json:"file_id"`
}

func (c *CurseForgeAddon) DisplayName() string {
	return fmt.Sprintf("[curse:%d] %s", c.Id, c.Name)
}

func (c *CurseForgeAddon) ShortName() string {
	return fmt.Sprintf("curse:%d", c.Id)
}

func (c *CurseForgeAddon) Update(w *wow.Install, debug io.Writer, force bool) (updated bool, version string, err error) {
	fmt.Fprintf(debug, "\n================================================================================\n")
	fmt.Fprintf(debug, "Checking for updates to addon %d (%s)\n\n", c.Id, c.Name)

	details, err := curse.GetAddon(c.Id)
	if err != nil {
		return false, "", err
	}

	c.Name = details.Name

	latest := curse.LatestFile(details, debug)
	if latest == nil {
		return false, "", fmt.Errorf("no releases found for addon %d (%s)", c.Id, c.Name)
	}

	if !force && c.FileId == latest.FileId {
		fmt.Fprintf(
			debug,
			"No update found for '%s'. Installed file ID: %d, latest file ID: %d (version: %s)\n",
			c.Name,
			c.FileId,
			latest.FileId,
			latest.DisplayName,
		)
		return false, "", nil
	}

	// Remove all the existing directories associated with the addon
	if err := w.RemoveAddons(c.Directories); err != nil {
		return false, "", err
	}

	// Deploy the new version
	dirs, err := w.InstallAddonFromUrl(latest.Url)
	if err != nil {
		return false, "", err
	}

	// Update our metadata
	c.FileId = latest.FileId
	c.Directories = dirs
	return true, latest.DisplayName, nil
}
