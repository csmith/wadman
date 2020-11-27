package wadman

import (
	"fmt"
	"github.com/csmith/wadman/wow"
	"io"
)

type AddonType string

const (
	TypeUnspecified AddonType = ""
	TypeCurseForge  AddonType = "curse"
)

func (t AddonType) NewInstance() (Addon, error) {
	switch t {
	case TypeUnspecified:
		// For compatibility with old configs, if the type field is missing default to curseforge
		return &CurseForgeAddon{BaseAddon: BaseAddon{Type: TypeCurseForge}}, nil
	case TypeCurseForge:
		return &CurseForgeAddon{BaseAddon: BaseAddon{Type: TypeCurseForge}}, nil
	default:
		return nil, fmt.Errorf("unknown addon type: %s", t)
	}
}

type Addon interface {
	ShortName() string
	DisplayName() string
	Dirs() []string
	Update(w *wow.Install, debug io.Writer, force bool) (updated bool, version string, err error)
}

type BaseAddon struct {
	Type        AddonType `json:"type"`
	Directories []string  `json:"directories"`
}

func (a *BaseAddon) Dirs() []string {
	return a.Directories
}
