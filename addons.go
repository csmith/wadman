package wadman

import (
	"fmt"
	"github.com/csmith/wadman/wow"
	"io"
	"time"
)

type AddonType string

const (
	TypeUnspecified  AddonType = ""
	TypeCurseForge   AddonType = "curse"
	TypeWowInterface AddonType = "wowi"
)

func (t AddonType) NewInstance() (Addon, error) {
	switch t {
	case TypeUnspecified:
		// For compatibility with old configs, if the type field is missing default to curseforge
		return NewCurseForgeAddon(0), nil
	case TypeCurseForge:
		return NewCurseForgeAddon(0), nil
	case TypeWowInterface:
		return NewWowInterfaceAddon(0), nil
	default:
		return nil, fmt.Errorf("unknown addon type: %s", t)
	}
}

type Addon interface {
	ShortName() string
	DisplayName() string
	Dirs() []string
	CurrentVersion() string
	LastUpdated() time.Time

	Update(w *wow.Install, debug io.Writer, force bool) (updated bool, err error)
}

type BaseAddon struct {
	Type        AddonType `json:"type"`
	Directories []string  `json:"directories"`
	Version     string    `json:"version"`
	LastUpdate  time.Time `json:"last_update"`
}

func (a *BaseAddon) Dirs() []string {
	return a.Directories
}

func (a *BaseAddon) CurrentVersion() string {
	return a.Version
}

func (a *BaseAddon) LastUpdated() time.Time {
	return a.LastUpdate
}
