package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	defaultKubeConfigDir     = ".kube/config"
	DefaultLeaseDuration     = 30 * time.Second
	DefaultRenewDeadline     = 15 * time.Second
	DefaultRetryPeriod       = 5 * time.Second
	DefaultResourceName      = "open-hydra-api-leader-lock"
	DefaultResourceLock      = "endpointsleases"
	DefaultResourceNamespace = "default"
	defaultResyncPeriod      = 30 * time.Minute
	defaultNamespace         = "open-hydra"
)

type (
	LeaderElection struct {
		// leaderElect enables a leader election client to gain leadership
		// before executing the main loop. Enable this when running replicated
		// components for high availability.
		LeaderElect bool `json:"leader_elect" yaml:"leaderElect"`
		// leaseDuration is the duration that non-leader candidates will wait
		// after observing a leadership renewal until attempting to acquire
		// leadership of a led but unrenewed leader slot. This is effectively the
		// maximum duration that a leader can be stopped before it is replaced
		// by another candidate. This is only applicable if leader election is
		// enabled.
		LeaseDuration time.Duration `json:"lease_duration" yaml:"leaseDuration"`
		// renewDeadline is the interval between attempts by the acting master to
		// renew a leadership slot before it stops leading. This must be less
		// than or equal to the lease duration. This is only applicable if leader
		// election is enabled.
		RenewDeadline time.Duration `json:"renew_deadline" yaml:"renewDeadline"`
		// retryPeriod is the duration the clients should wait between attempting
		// acquisition and renewal of a leadership. This is only applicable if
		// leader election is enabled.
		RetryPeriod time.Duration `json:"retry_period" yaml:"retryPeriod"`
		// resourceLock indicates the resource object type that will be used to lock
		// during leader election cycles.
		ResourceLock string `json:"resource_lock" yaml:"resourceLock"`
		// resourceName indicates the name of resource object that will be used to lock
		// during leader election cycles.
		ResourceName string `json:"resource_name" yaml:"resourceName"`
		// resourceNamespace indicates the namespace of resource object that will be used to lock
		// during leader election cycles.
		ResourceNamespace string `json:"resource_namespace" yaml:"resourceNamespace"`
	}

	OpenHydraServerConfig struct {
		// in case of we want to control total pod number in cluster
		// -1 not limit pod allocatable will be count by resource wise
		PodAllocatableLimit int64 `json:"pod_allocatable_limit" yaml:"podAllocatableLimit"`
		// default = 2000
		// note unit is m
		DefaultCpuPerDevice uint64 `json:"default_cpu_per_device" yaml:"defaultCpuPerDevice"`
		// default = 8192
		// note unit is Mi
		DefaultRamPerDevice uint64 `json:"default_ram_per_device" yaml:"defaultRamPerDevice"`
		// default = 0
		DefaultGpuPerDevice uint8 `json:"default_gpu_per_device" yaml:"defaultGpuPerDevice"`
		// default = "/open-hydra/public-dataset"
		// all dataset upload by user will be store in this path
		PublicDatasetBasePath string `json:"dataset_base_path" yaml:"datasetBasePath"`
		PublicCourseBasePath  string `json:"course_base_path" yaml:"courseBasePath"`
		PublicDatasetMaxSize  int64  `json:"dataset_max_size" yaml:"datasetMaxSize"`
		PublicCourseMaxSize   int64  `json:"course_max_size" yaml:"courseMaxSize"`
		// default = "hostpath", hostpath or nfs
		// hostpath: open-hydra-server will use hostpath to mount dataset most likely for aio server or test
		// nfs: open-hydra-server will use nfs to mount dataset most likely for production
		PublicDatasetVolumeType string `json:"dataset_volume_type" yaml:"datasetVolumeType"`
		// default = "/root/public-dataset"
		PublicDatasetStudentMountPath string `json:"dataset_student_mount_path" yaml:"datasetStudentMountPath"`
		PublicCourseStudentMountPath  string `json:"course_student_mount_path" yaml:"courseStudentMountPath"`
		OpenHydraNamespace            string `json:"open-hydra_namespace" yaml:"open-hydraNamespace"`
		// should be no default value but fill it in installation script, because it is a runtime value
		// if not set we won't be able to start gpu pod at all
		DefaultGpuDriver        string `json:"default_gpu_driver" yaml:"defaultGpuDriver"`
		ServerIP                string `json:"server_ip" yaml:"serverIP"`
		KubeConfig              *rest.Config
		LeaderElection          *LeaderElection     `json:"leader_election" yaml:"leaderElection,omitempty"`
		MySqlConfig             *MySqlConfig        `json:"mysql_config" yaml:"mysqlConfig,omitempty"`
		EtcdConfig              *EtcdConfig         `json:"etcd_config" yaml:"etcdConfig,omitempty"`
		DBType                  string              `json:"db_type" yaml:"dbType"`
		DisableAuth             bool                `json:"disable_auth" yaml:"disableAuth"`
		PatchResourceNotRelease bool                `json:"patch_resource_not_release" yaml:"patchResourceNotRelease"`
		CpuOverCommitRate       uint8               `json:"cpu_over_commit_rate" yaml:"cpuOverCommitRate"`
		MemoryOverCommitRate    uint8               `json:"memory_over_commit_rate" yaml:"memoryOverCommitRate"`
		AuthDelegateConfig      *AuthDelegateConfig `json:"auth_delegate_config" yaml:"authDelegateConfig,omitempty"`
		MaximumPortsPerSandbox  uint8               `json:"maximum_ports_per_sandbox" yaml:"maximumPortsPerSandbox"`
		WorkspacePath           string              `json:"workspace_path" yaml:"workspacePath"`
	}
)

func DefaultLeaderElection() *LeaderElection {
	return &LeaderElection{
		LeaseDuration:     DefaultLeaseDuration,
		RenewDeadline:     DefaultRenewDeadline,
		RetryPeriod:       DefaultRetryPeriod,
		ResourceLock:      DefaultResourceLock,
		ResourceName:      DefaultResourceName,
		ResourceNamespace: DefaultResourceNamespace,
	}
}

func DefaultConfig() *OpenHydraServerConfig {
	return &OpenHydraServerConfig{
		PodAllocatableLimit:           -1,
		DefaultCpuPerDevice:           2000,
		DefaultRamPerDevice:           8192,
		DefaultGpuPerDevice:           0,
		PublicDatasetBasePath:         "/mnt/public-dataset",
		PublicCourseBasePath:          "/mnt/public-course",
		PublicDatasetMaxSize:          1 << 30, // 1 GiB
		PublicCourseMaxSize:           1 << 30, // 1 GiB
		PublicDatasetVolumeType:       "hostpath",
		PublicDatasetStudentMountPath: "/root/notebook/dataset-public",
		PublicCourseStudentMountPath:  "/root/notebook/course-public",
		MySqlConfig:                   DefaultMySqlConfig(),
		EtcdConfig:                    DefaultEtcdConfig(),
		OpenHydraNamespace:            defaultNamespace,
		LeaderElection:                DefaultLeaderElection(),
		DefaultGpuDriver:              "nvidia.com/gpu",
		ServerIP:                      "localhost",
		CpuOverCommitRate:             1, // no over commit for cpu by default,set to 2 cpu request will be divide by 2
		MemoryOverCommitRate:          1, // no over commit for memory by default,set to 2 meaning memory request will be divide by 2
		MaximumPortsPerSandbox:        3,
		WorkspacePath:                 "/mnt/workspace",
	}
}

type EtcdConfig struct {
	Endpoints []string `json:"endpoints" yaml:"endpoints"`
	CAFile    string   `json:"ca_file" yaml:"caFile"`
	CertFile  string   `json:"cert_file" yaml:"certFile"`
	KeyFile   string   `json:"key_file" yaml:"keyFile"`
}

type MySqlConfig struct {
	Address      string `json:"address" yaml:"address"`
	Port         uint16 `json:"port" yaml:"port"`
	Username     string `json:"username" yaml:"username"`
	Password     string `json:"password" yaml:"password"`
	DataBaseName string `json:"database_name" yaml:"databaseName"`
	Protocol     string `json:"protocol" yaml:"protocol"`
	Character    string `json:"character" yaml:"character"`
	Collation    string `json:"collation" yaml:"collation"`
}

type AuthDelegateConfig struct {
	// if KeystoneConfig is set to nil then auth plugin will fall backup to database auth
	KeystoneConfig *KeystoneConfig `json:"keystone_config" yaml:"keystoneConfig"`
}

type KeystoneConfig struct {
	Endpoint           string `json:"endpoint" yaml:"endpoint"`
	Username           string `json:"username" yaml:"username"`
	Password           string `json:"password" yaml:"password"`
	DomainId           string `json:"domain_id" yaml:"domainId"`
	ProjectId          string `json:"project_id" yaml:"projectId"`
	TokenKeyInResponse string `json:"token_key_in_response" yaml:"tokenKeyInResponse"`
	TokenKeyInRequest  string `json:"token_key_in_request" yaml:"tokenKeyInRequest"`
}

func DefaultEtcdConfig() *EtcdConfig {
	return &EtcdConfig{
		Endpoints: []string{"http://localhost:2379"},
		CAFile:    "",
		CertFile:  "",
		KeyFile:   "",
	}
}

func DefaultMySqlConfig() *MySqlConfig {
	return &MySqlConfig{
		Address:      "mysql.svc.cluster.local",
		Port:         3306,
		Username:     "root",
		Password:     "root",
		DataBaseName: "open-hydra",
		Protocol:     "tcp",
		Character:    "utf8mb3",
		Collation:    "utf8mb3_general_ci",
	}
}

func LoadConfig(configFilePath, kubeConfig string) (*OpenHydraServerConfig, error) {
	rawData, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	config := DefaultConfig()
	err = yaml.Unmarshal(rawData, config)
	if err != nil {
		return nil, err
	}

	if config.KubeConfig, err = rest.InClusterConfig(); err == nil {
		return config, err
	}

	if kubeConfig != "" {
		config.KubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	} else {
		config.KubeConfig, err = clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), defaultKubeConfigDir))
	}

	return config, err
}

func WriteConfig(configFilePath string, config *OpenHydraServerConfig) error {
	rawData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(configFilePath, rawData, 0644)
}
