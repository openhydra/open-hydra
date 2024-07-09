package apis

// +k8s:openapi-gen=true
type VolumeMount struct {
	Name       string `json:"name"`
	MountPath  string `json:"mount_path"`
	SourcePath string `json:"source_path"`
	ReadOnly   bool   `json:"read_only"`
}

// +k8s:openapi-gen=true
type Volume struct {
	EmptyDir *EmptyDir `json:"empty_dir,omitempty"`
	HostPath *HostPath `json:"host_path,omitempty"`
}

// +k8s:openapi-gen=true
type EmptyDir struct {
	Medium    string `json:"medium"`
	SizeLimit uint64 `json:"size_limit"`
	Name      string `json:"name"`
}

// +k8s:openapi-gen=true
type HostPath struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
}

type GpuSet struct {
	GpuDriverName string `json:"gpu_driver_name"`
	Gpu           uint8  `json:"gpu"`
}

// +k8s:openapi-gen=true
type Sandbox struct {
	CPUImageName    string            `json:"cpuImageName,omitempty"`
	GPUImageSet     map[string]string `json:"gpuImageSet,omitempty"`
	Command         []string          `json:"command,omitempty"`
	Args            []string          `json:"args,omitempty"`
	DisplayTitle    string            `json:"display_title,omitempty"`
	Description     string            `json:"description,omitempty"`
	DevelopmentInfo []string          `json:"developmentInfo,omitempty"`
	Status          string            `json:"status,omitempty"`
	Ports           []uint16          `json:"ports,omitempty"`
	VolumeMounts    []VolumeMount     `json:"volume_mounts,omitempty"`
	Volumes         []Volume          `json:"volumes,omitempty"`
	IconName        string            `json:"icon_name,omitempty"`
}

// +k8s:openapi-gen=true
type PluginList struct {
	DefaultSandbox string             `json:"defaultSandbox"`
	Sandboxes      map[string]Sandbox `json:"sandboxes"`
}
