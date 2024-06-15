package k8s

import (
	"open-hydra/pkg/open-hydra/apis"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type DeploymentParameters struct {
	CpuMemorySet CpuMemorySet
	Image        string
	Namespace    string
	Username     string
	SandboxName  string
	VolumeMounts []apis.VolumeMount
	GpuSet       apis.GpuSet
	Client       *kubernetes.Clientset
	Command      []string
	Args         []string
	Ports        map[string]int
	Volumes      []apis.Volume
	Affinity     *coreV1.Affinity
}

type IOpenHydraK8sHelper interface {
	ListDeploymentWithLabel(label, namespace string, client *kubernetes.Clientset) ([]appsV1.Deployment, error)
	ListPodWithLabel(label, namespace string, client *kubernetes.Clientset) ([]coreV1.Pod, error)
	ListPod(namespace string, client *kubernetes.Clientset) ([]coreV1.Pod, error)
	GetUserPods(label, namespace string, client *kubernetes.Clientset) ([]coreV1.Pod, error)
	ListDeployment(namespace string, client *kubernetes.Clientset) ([]appsV1.Deployment, error)
	ListService(namespace string, client *kubernetes.Clientset) ([]coreV1.Service, error)
	DeleteUserDeployment(label, namespace string, client *kubernetes.Clientset) error
	CreateDeployment(deployParameter *DeploymentParameters) error
	CreateService(namespace, userName, ideType string, client *kubernetes.Clientset, ports map[string]int) error
	DeleteUserService(label, namespace string, client *kubernetes.Clientset) error
	GetUserService(label, namespace string, client *kubernetes.Clientset) (*coreV1.Service, error)
	DeleteUserReplicaSet(label, namespace string, client *kubernetes.Clientset) error
	DeleteUserPod(label, namespace string, client *kubernetes.Clientset) error
	GetMap(name, namespace string, client *kubernetes.Clientset) (*coreV1.ConfigMap, error)
}

func NewDefaultK8sHelper() IOpenHydraK8sHelper {
	return &DefaultHelper{}
}

func NewDefaultK8sHelperWithFake() IOpenHydraK8sHelper {
	fake := &Fake{}
	fake.Init()
	return fake
}

type CpuMemorySet struct {
	CpuRequest    string
	CpuLimit      string
	MemoryRequest string
	MemoryLimit   string
}
