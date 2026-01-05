package provider

type DiscoveryPasswordPayload struct {
	Partition *Partition `json:"partition"`
}

type Partition struct {
	Name            string   `yaml:"name,omitempty" mapstructure:"name" json:"name,omitempty"`
	FilesystemLabel string   `yaml:"label,omitempty" mapstructure:"label" json:"label,omitempty"`
	Size            uint     `yaml:"size,omitempty" mapstructure:"size" json:"size,omitempty"`
	FS              string   `yaml:"fs,omitempty" mapstrcuture:"fs" json:"fs,omitempty"`
	Flags           []string `yaml:"flags,omitempty" mapstrcuture:"flags" json:"flags,omitempty"`
	UUID            string   `yaml:"uuid,omitempty" mapstructure:"uuid" json:"uuid,omitempty"`
}
