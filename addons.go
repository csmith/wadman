package wadman

import "fmt"

type AddonType string

const (
	TypeUnspecified AddonType = ""
	TypeCurseForge  AddonType = "curse"
)

type BaseAddon struct {
	Type        AddonType `json:"type"`
	Directories []string  `json:"directories"`
}

type CurseForgeAddon struct {
	BaseAddon
	Id     int    `json:"id"`
	Name   string `json:"name"`
	FileId int    `json:"file_id"`
}

func (t AddonType) NewInstance() (*CurseForgeAddon, error) {
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
