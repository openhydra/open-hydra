package apis

// +k8s:openapi-gen=true
type VolumeMount struct {
	Name       string `json:"name"`
	MountPath  string `json:"mount_path"`
	SourcePath string `json:"source_path"`
	ReadOnly   bool   `json:"read_only"`
}

type GpuSet struct {
	GpuDriverName string `json:"gpu_driver_name"`
	Gpu           uint8  `json:"gpu"`
}

// +k8s:openapi-gen=true
type Sandbox struct {
	CPUImageName    string        `json:"cpuImageName,omitempty"`
	GPUImageName    string        `json:"gpuImageName,omitempty"`
	Command         []string      `json:"command,omitempty"`
	Args            []string      `json:"args,omitempty"`
	Description     string        `json:"description,omitempty"`
	DevelopmentInfo []string      `json:"developmentInfo,omitempty"`
	Status          string        `json:"status,omitempty"`
	Ports           []uint16      `json:"ports,omitempty"`
	VolumeMounts    []VolumeMount `json:"volume_mounts,omitempty"`
	// TODO: more deployment properties here such as env, resources, etc.
}

// +k8s:openapi-gen=true
type PluginList struct {
	Sandboxes map[string]Sandbox `json:"sandboxes"`
}
