package wadman

import (
	"fmt"
	"github.com/csmith/wadman/curse"
	"github.com/csmith/wadman/wow"
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

func (c *CurseForgeAddon) Update(w *wow.Install, force, verbose bool) error {
	if verbose {
		fmt.Println()
		fmt.Printf("================================================================================\n")
		fmt.Printf("Checking for updates to addon %d (%s)\n\n", c.Id, c.Name)
	}
	details, err := curse.GetAddon(c.Id)
	if err != nil {
		return err
	}

	c.Name = details.Name

	latest := curse.LatestFile(details, verbose)
	if latest == nil {
		return fmt.Errorf("no releases found for addon %d (%s)", c.Id, c.Name)
	}

	if force {
		fmt.Printf("'%s': force updating to version %s\n", c.Name, latest.DisplayName)
	} else if c.FileId == 0 {
		fmt.Printf("'%s': installing version %s\n", c.Name, latest.DisplayName)
	} else if latest.FileId != c.FileId {
		fmt.Printf("'%s': updating to version %s\n", c.Name, latest.DisplayName)
	} else {
		if verbose {
			fmt.Printf(
				"No update found for '%s'. Installed file ID: %d, latest file ID: %d (version: %s)\n",
				c.Name,
				c.FileId,
				latest.FileId,
				latest.DisplayName,
			)
		}

		return nil
	}

	// Remove all the existing directories associated with the addon
	if err := w.RemoveAddons(c.Directories); err != nil {
		return err
	}

	// Deploy the new version
	dirs, err := w.InstallAddonFromUrl(latest.Url)
	if err != nil {
		return err
	}

	// Update our metadata
	c.FileId = latest.FileId
	c.Directories = dirs
	return nil
}
