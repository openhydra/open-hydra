package k8s

import (
	"fmt"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Fake struct {
	namespacedPod     map[string][]coreV1.Pod
	namespacedDeploy  map[string][]appsV1.Deployment
	namespacedService map[string][]coreV1.Service
	labelPod          map[string][]coreV1.Pod
	labelDeploy       map[string][]appsV1.Deployment
	labelService      map[string][]coreV1.Service
}

func (f *Fake) Init() {
	f.namespacedPod = make(map[string][]coreV1.Pod)
	f.namespacedDeploy = make(map[string][]appsV1.Deployment)
	f.namespacedService = make(map[string][]coreV1.Service)
	f.labelPod = make(map[string][]coreV1.Pod)
	f.labelDeploy = make(map[string][]appsV1.Deployment)
	f.labelService = make(map[string][]coreV1.Service)
}

func (f *Fake) ListDeploymentWithLabel(label, namespace string, client *kubernetes.Clientset) ([]appsV1.Deployment, error) {
	var result []appsV1.Deployment
	if _, ok := f.labelDeploy[label]; ok {
		result = f.labelDeploy[label]
	}
	return result, nil
}
func (f *Fake) ListPodWithLabel(label, namespace string, client *kubernetes.Clientset) ([]coreV1.Pod, error) {
	var result []coreV1.Pod
	if _, ok := f.labelPod[label]; ok {
		result = f.labelPod[label]
	}
	return result, nil
}
func (f *Fake) ListPod(namespace string, client *kubernetes.Clientset) ([]coreV1.Pod, error) {
	var result []coreV1.Pod
	if _, ok := f.namespacedPod[namespace]; ok {
		result = f.namespacedPod[namespace]
	}
	return result, nil
}
func (f *Fake) GetUserPods(label, namespace string, client *kubernetes.Clientset) ([]coreV1.Pod, error) {
	var result []coreV1.Pod
	if _, ok := f.labelPod[label]; ok {
		result = f.labelPod[label]
	}
	return result, nil
}
func (f *Fake) ListDeployment(namespace string, client *kubernetes.Clientset) ([]appsV1.Deployment, error) {
	var result []appsV1.Deployment
	if _, ok := f.namespacedDeploy[namespace]; ok {
		result = f.namespacedDeploy[namespace]
	}
	return result, nil
}
func (f *Fake) ListService(namespace string, client *kubernetes.Clientset) ([]coreV1.Service, error) {
	var result []coreV1.Service
	if _, ok := f.namespacedService[namespace]; ok {
		result = f.namespacedService[namespace]
	}
	return result, nil
}
func (f *Fake) DeleteUserDeployment(label, namespace string, client *kubernetes.Clientset) error {
	delete(f.labelDeploy, label)
	return nil
}
func (f *Fake) CreateDeployment(deployParameter *DeploymentParameters) error {
	label := fmt.Sprintf("%s=%s", OpenHydraUserLabelKey, deployParameter.Username)
	f.labelDeploy[label] = append(f.labelDeploy[label], appsV1.Deployment{})
	f.namespacedDeploy[deployParameter.Namespace] = append(f.namespacedDeploy[deployParameter.Namespace], appsV1.Deployment{})
	f.labelPod[label] = append(f.labelPod[label], coreV1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				OpenHydraUserLabelKey: deployParameter.Username,
				OpenHydraSandboxKey:   deployParameter.SandboxName,
			},
		},
	})
	return nil
}
func (f *Fake) CreateService(namespace, studentID, ideType string, client *kubernetes.Clientset, ports map[string]int) error {
	label := fmt.Sprintf("%s=%s", OpenHydraUserLabelKey, studentID)
	f.labelService[label] = append(f.labelService[label], coreV1.Service{})
	f.namespacedService[namespace] = append(f.namespacedService[namespace], coreV1.Service{})
	return nil
}
func (f *Fake) DeleteUserService(label, namespace string, client *kubernetes.Clientset) error {
	delete(f.labelService, label)
	return nil
}
func (f *Fake) GetUserService(label, namespace string, client *kubernetes.Clientset) (*coreV1.Service, error) {
	var result *coreV1.Service
	if _, ok := f.labelService[label]; ok {
		result = &f.labelService[label][0]
	} else {
		return nil, fmt.Errorf("service not found")
	}
	return result, nil
}
func (f *Fake) DeleteUserReplicaSet(label, namespace string, client *kubernetes.Clientset) error {
	return nil
}
func (f *Fake) DeleteUserPod(label, namespace string, client *kubernetes.Clientset) error {
	delete(f.labelPod, label)
	return nil
}

func (f *Fake) GetMap(name, namespace string, client *kubernetes.Clientset) (*coreV1.ConfigMap, error) {
	return &coreV1.ConfigMap{
		Data: map[string]string{
			"plugins": `{
			"defaultSandbox": "test",
			"sandboxes":{
				"test": {
					"display_title": "test",
					"cpuImageName": "test",
					"gpuImageSet": {
						"nvidia.com/gpu": "nvidia-gpu-image",
						"amd.com/gpu": ""
					},
					"icon_name": "test1.png",
					"command": ["test"],
					"description": "test",
					"developmentInfo": ["test"],
					"status": "test",
					"ports": [
						8888
					],
					"volume_mounts": [
						{
							"name": "jupyter-lab",
							"mount_path": "/root/notebook",
							"source_path": "/mnt/jupyter-lab"
						},
						{
							"name": "public-dataset",
							"mount_path": "/root/notebook/dataset-public",
							"source_path": "/mnt/public-dataset"
						},
						{
							"name": "public-course",
							"mount_path": "/mnt/public-course",
							"source_path": "/mnt/public-course"
						}
					]
				},
				"jupyter-lab": {
					"display_title": "jupyter-lab",
					"cpuImageName": "jupyter-lab-test",
					"gpuImageSet": {
						"nvidia.com/gpu": "nvidia-gpu-image",
						"amd.com/gpu": ""
					},
					"icon_name": "test2.png",
					"command": ["jupyter-lab-test"],
					"description": "jupyter-lab-test",
					"developmentInfo": ["jupyter-lab-test"],
					"status": "running",
					"ports": [
						8888
					],
					"volume_mounts": [
						{
							"name": "jupyter-lab",
							"mount_path": "/root/notebook",
							"source_path": "/mnt/jupyter-lab"
						},
						{
							"name": "public-dataset",
							"mount_path": "/root/notebook/dataset-public",
							"source_path": "/mnt/public-dataset"
						},
						{
							"name": "public-course",
							"mount_path": "/mnt/public-course",
							"source_path": "/mnt/public-course"
						}
					]
				},
				"jupyter-lab-lot-ports": {
					"display_title": "jupyter-lab-lot-ports",
					"cpuImageName": "jupyter-lab-test",
					"gpuImageSet": {
						"nvidia.com/gpu": "nvidia-gpu-image",
						"amd.com/gpu": ""
					},
					"icon_name": "test3.png",
					"command": ["jupyter-lab-test"],
					"description": "jupyter-lab-test",
					"developmentInfo": ["jupyter-lab-test"],
					"status": "running",
					"ports": [
						8888,
						8889,
						8890,
						8891
					],
					"volume_mounts": [
						{
							"name": "jupyter-lab",
							"mount_path": "/root/notebook",
							"source_path": "/mnt/jupyter-lab"
						},
						{
							"name": "public-dataset",
							"mount_path": "/root/notebook/dataset-public",
							"source_path": "/mnt/public-dataset"
						},
						{
							"name": "public-course",
							"mount_path": "/mnt/public-course",
							"source_path": "/mnt/public-course"
						}
					]
				},
				"jupyter-lab-not-ports": {
					"display_title": "jupyter-lab-test",
					"cpuImageName": "jupyter-lab-test",
					"gpuImageSet": {
						"nvidia.com/gpu": "nvidia-gpu-image",
						"amd.com/gpu": ""
					},
					"icon_name": "test4.png",
					"command": ["jupyter-lab-test"],
					"description": "jupyter-lab-test",
					"developmentInfo": ["jupyter-lab-test"],
					"status": "running"
				}
			}}`,
		},
	}, nil
}
