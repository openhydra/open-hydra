package k8s

import (
	"context"
	"fmt"
	"open-hydra/pkg/open-hydra/apis"
	"strconv"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	OpenHydraWorkloadLabelKey    = "openhydra"
	OpenHydraWorkloadLabelValue  = "openhydra-workload"
	OpenHydraUserLabelKey        = "openhydra-user"
	OpenHydraDeployNameTemplate  = "openhydra-deploy-%s"
	OpenHydraServiceNameTemplate = "openhydra-service-%s"
	OpenHydraDeployHookKey       = "openhydra-hook"
	OpenHydraIDELabelKey         = "openhydra-ide-type"
	OpenHydraIDELabelJuptyerLab  = "jupyterlab"
	OpenHydraIDELabelVSCode      = "vscode"
	OpenHydraIDELabelUnset       = "unset"
)

type DefaultHelper struct {
}

func (help *DefaultHelper) ListDeploymentWithLabel(label, namespace string, client *kubernetes.Clientset) ([]appsV1.Deployment, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}

	deployments, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metaV1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		return nil, err
	}
	return deployments.Items, nil
}

func (help *DefaultHelper) ListPodWithLabel(label, namespace string, client *kubernetes.Clientset) ([]coreV1.Pod, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}

	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metaV1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

func (help *DefaultHelper) ListPod(namespace string, client *kubernetes.Clientset) ([]coreV1.Pod, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metaV1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", OpenHydraWorkloadLabelKey, OpenHydraWorkloadLabelValue),
	})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

func (help *DefaultHelper) GetUserPods(label, namespace string, client *kubernetes.Clientset) ([]coreV1.Pod, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metaV1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

func (help *DefaultHelper) ListDeployment(namespace string, client *kubernetes.Clientset) ([]appsV1.Deployment, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	deployments, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metaV1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", OpenHydraWorkloadLabelKey, OpenHydraWorkloadLabelValue),
	})
	if err != nil {
		return nil, err
	}
	return deployments.Items, nil
}

func (help *DefaultHelper) ListService(namespace string, client *kubernetes.Clientset) ([]coreV1.Service, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	services, err := client.CoreV1().Services(namespace).List(context.TODO(), metaV1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", OpenHydraWorkloadLabelKey, OpenHydraWorkloadLabelValue),
	})
	if err != nil {
		return nil, err
	}
	return services.Items, nil
}

func (help *DefaultHelper) DeleteUserDeployment(label, namespace string, client *kubernetes.Clientset) error {
	if client == nil {
		return fmt.Errorf("client is nil")
	}

	deployments, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metaV1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		return err
	}
	for _, deployment := range deployments.Items {
		err := client.AppsV1().Deployments(namespace).Delete(context.TODO(), deployment.Name, metaV1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (help *DefaultHelper) CreateDeployment(cpu, memory, image, namespace, studentID, ideType string, volumes []apis.VolumeMount, gpuSet apis.GpuSet, client *kubernetes.Clientset) error {
	if client == nil {
		return fmt.Errorf("client is nil")
	}
	baseName := fmt.Sprintf(OpenHydraDeployNameTemplate, studentID)
	replicas := int32(1)
	resourceReqLim := createResource(cpu, memory, gpuSet)
	ideTypeLabelValue := OpenHydraIDELabelUnset
	if ideType != "" {
		ideTypeLabelValue = ideType
	}
	deployment := &appsV1.Deployment{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      baseName,
			Namespace: namespace,
			Labels: map[string]string{
				OpenHydraUserLabelKey:     studentID,
				OpenHydraWorkloadLabelKey: OpenHydraWorkloadLabelValue,
				OpenHydraIDELabelKey:      ideTypeLabelValue,
			},
		},
		Spec: appsV1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metaV1.LabelSelector{
				MatchLabels: map[string]string{
					OpenHydraUserLabelKey: studentID,
				},
			},
			Template: coreV1.PodTemplateSpec{
				ObjectMeta: metaV1.ObjectMeta{
					Labels: map[string]string{
						OpenHydraUserLabelKey:     studentID,
						OpenHydraWorkloadLabelKey: OpenHydraWorkloadLabelValue,
						OpenHydraIDELabelKey:      ideTypeLabelValue,
					},
				},
				Spec: coreV1.PodSpec{
					Volumes:    createVolume(volumes),
					Containers: createContainers(baseName, image, volumes, resourceReqLim),
				},
			},
		},
	}

	_, err := client.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metaV1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func createResource(cpu, memory string, gpuSet apis.GpuSet) coreV1.ResourceList {
	resourceReqLim := coreV1.ResourceList{
		coreV1.ResourceCPU:    resource.MustParse(cpu),
		coreV1.ResourceMemory: resource.MustParse(memory),
	}
	if gpuSet.Gpu > 0 {
		resourceReqLim[coreV1.ResourceName(gpuSet.GpuDriverName)] = resource.MustParse(strconv.Itoa(int(gpuSet.Gpu)))
	}
	return resourceReqLim
}

func createContainers(baseName, image string, volumes []apis.VolumeMount, resourceReqLim coreV1.ResourceList) []coreV1.Container {
	container := coreV1.Container{
		Name:            baseName + "-container",
		Image:           image,
		ImagePullPolicy: coreV1.PullPolicy("IfNotPresent"),
		Ports: []coreV1.ContainerPort{
			{
				ContainerPort: 5000,
				Protocol:      coreV1.ProtocolTCP,
			},
			{
				ContainerPort: 8888,
				Protocol:      coreV1.ProtocolTCP,
			},
		},
		// now we keep all pod in qos group guaranteed
		// do not break it for the god sake
		Resources: coreV1.ResourceRequirements{
			Limits:   resourceReqLim,
			Requests: resourceReqLim,
		},
	}

	for _, volume := range volumes {
		container.VolumeMounts = append(container.VolumeMounts, coreV1.VolumeMount{
			Name:      volume.Name,
			MountPath: volume.MountPath,
		})
	}

	return []coreV1.Container{
		container,
	}
}

func createVolume(volumes []apis.VolumeMount) []coreV1.Volume {
	var volumeMounts []coreV1.Volume
	hpType := coreV1.HostPathDirectoryOrCreate
	for _, volume := range volumes {
		volumeMounts = append(volumeMounts, coreV1.Volume{
			Name: volume.Name,
			VolumeSource: coreV1.VolumeSource{
				HostPath: &coreV1.HostPathVolumeSource{
					Path: volume.SourcePath,
					Type: &hpType,
				},
			},
		})
	}
	return volumeMounts
}

func (help *DefaultHelper) CreateService(namespace, studentID, ideType string, client *kubernetes.Clientset) error {
	ports := []coreV1.ServicePort{
		{
			Name: "easy-train",
			Port: 5000,
		},
		{
			Name: "lab",
			Port: 8888,
		},
	}
	if ideType == OpenHydraIDELabelVSCode {
		ports = []coreV1.ServicePort{
			{
				Name: "vscode",
				Port: 3000,
			},
		}
	}
	service := &coreV1.Service{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      fmt.Sprintf(OpenHydraServiceNameTemplate, studentID),
			Namespace: namespace,
			Labels: map[string]string{
				OpenHydraWorkloadLabelKey: OpenHydraWorkloadLabelValue,
				OpenHydraUserLabelKey:     studentID,
			},
		},
		Spec: coreV1.ServiceSpec{
			Type: coreV1.ServiceTypeNodePort,
			Selector: map[string]string{
				OpenHydraUserLabelKey: studentID,
			},
			Ports: ports,
		},
	}

	_, err := client.CoreV1().Services(namespace).Create(context.Background(), service, metaV1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (help *DefaultHelper) DeleteUserService(label, namespace string, client *kubernetes.Clientset) error {
	if client == nil {
		return fmt.Errorf("client is nil")
	}
	services, err := client.CoreV1().Services(namespace).List(context.TODO(), metaV1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		return err
	}
	for _, service := range services.Items {
		err := client.CoreV1().Services(namespace).Delete(context.Background(), service.Name, metaV1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (help *DefaultHelper) GetUserService(label, namespace string, client *kubernetes.Clientset) (*coreV1.Service, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	services, err := client.CoreV1().Services(namespace).List(context.TODO(), metaV1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		return nil, err
	}
	if len(services.Items) > 0 {
		return &services.Items[0], nil
	}
	return nil, fmt.Errorf("no service found")
}

func (help *DefaultHelper) GetAllNode(client *kubernetes.Clientset) ([]coreV1.Node, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	nodes, err := client.CoreV1().Nodes().List(context.Background(), metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return nodes.Items, nil
}

func (help *DefaultHelper) DeleteUserReplicaSet(label, namespace string, client *kubernetes.Clientset) error {
	if client == nil {
		return fmt.Errorf("client is nil")
	}
	replicaSets, err := client.AppsV1().ReplicaSets(namespace).List(context.TODO(), metaV1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		return err
	}
	for _, replicaSet := range replicaSets.Items {
		err := client.AppsV1().ReplicaSets(namespace).Delete(context.Background(), replicaSet.Name, metaV1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (help *DefaultHelper) DeleteUserPod(label, namespace string, client *kubernetes.Clientset) error {
	if client == nil {
		return fmt.Errorf("client is nil")
	}
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metaV1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		err := client.CoreV1().Pods(namespace).Delete(context.Background(), pod.Name, metaV1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
