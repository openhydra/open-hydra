package apis

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

type Sandbox struct {
	CPUImageName    string   `json:"cpuImageName"`
	GPUImageName    string   `json:"gpuImageName"`
	Command         []string `json:"command"`
	Description     string   `json:"description"`
	DevelopmentInfo []string `json:"developmentInfo"`
	Status          string   `json:"status"`
}

type SandboxList struct {
	Sandboxes map[string]Sandbox `json:"sandboxes"`
}
