package apis

type VolumeMount struct {
	Name       string `json:"name"`
	MountPath  string `json:"mount_path"`
	SourcePath string `json:"source_path"`
}

type GpuSet struct {
	GpuDriverName string `json:"gpu_driver_name"`
	Gpu           uint8  `json:"gpu"`
}
