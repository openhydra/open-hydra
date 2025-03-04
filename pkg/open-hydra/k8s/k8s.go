package k8s

import (
	"context"
	"fmt"
	"log/slog"
	"open-hydra/pkg/open-hydra/apis"
	"strconv"
	"strings"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	coreV1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
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
	OpenHydraSandboxKey          = "openhydra-sandbox"
)

type DefaultHelper struct {
	clientSet         *kubernetes.Clientset
	configMapInformer cache.SharedIndexInformer
	podInformer       cache.SharedIndexInformer
	svcInformer       cache.SharedIndexInformer
	nodeInformer      cache.SharedIndexInformer
	podCache          coreV1listers.PodLister
	svcCache          coreV1listers.ServiceLister
	nodeCache         coreV1listers.NodeLister
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
	selector, err := metaV1.ParseToLabelSelector(label)
	if err != nil {
		return nil, fmt.Errorf("failed to parse label selector: %v", err)
	}
	parsedSelector, err := metaV1.LabelSelectorAsSelector(selector)
	if err != nil {
		return nil, fmt.Errorf("failed to convert label selector: %v", err)
	}

	result, err := help.podCache.Pods(namespace).List(parsedSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	noRefPods := make([]coreV1.Pod, 0)
	for _, pod := range result {
		noRefPods = append(noRefPods, *pod)
	}
	return noRefPods, nil
}

func (help *DefaultHelper) ListPod(namespace string, client *kubernetes.Clientset) ([]coreV1.Pod, error) {
	result, err := help.podCache.Pods(namespace).List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	noRefPods := make([]coreV1.Pod, 0)
	for _, pod := range result {
		noRefPods = append(noRefPods, *pod)
	}
	return noRefPods, nil
}

func (help *DefaultHelper) GetUserPods(label, namespace string, client *kubernetes.Clientset) ([]coreV1.Pod, error) {
	selector, err := metaV1.ParseToLabelSelector(label)
	if err != nil {
		return nil, fmt.Errorf("failed to parse label selector: %v", err)
	}
	parsedSelector, err := metaV1.LabelSelectorAsSelector(selector)
	if err != nil {
		return nil, fmt.Errorf("failed to convert label selector: %v", err)
	}

	result, err := help.podCache.Pods(namespace).List(parsedSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	noRefPods := make([]coreV1.Pod, 0)
	for _, pod := range result {
		noRefPods = append(noRefPods, *pod)
	}
	return noRefPods, nil
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
	label := fmt.Sprintf("%s=%s", OpenHydraWorkloadLabelKey, OpenHydraWorkloadLabelValue)
	selector, err := metaV1.ParseToLabelSelector(label)
	if err != nil {
		return nil, fmt.Errorf("failed to parse label selector: %v", err)
	}
	parsedSelector, err := metaV1.LabelSelectorAsSelector(selector)
	if err != nil {
		return nil, fmt.Errorf("failed to convert label selector: %v", err)
	}

	result, err := help.svcCache.Services(namespace).List(parsedSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	noRefSvc := make([]coreV1.Service, 0)
	for _, svc := range result {
		noRefSvc = append(noRefSvc, *svc)
	}
	return noRefSvc, nil
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

func createDeployment(deployParameter *DeploymentParameters) *appsV1.Deployment {
	baseName := fmt.Sprintf(OpenHydraDeployNameTemplate, deployParameter.Username)
	replicas := int32(1)
	resourceReq, resourceLim := createResource(deployParameter.CpuMemorySet, deployParameter.GpuSet)
	ideTypeLabelValue := OpenHydraIDELabelUnset
	deployment := &appsV1.Deployment{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      baseName,
			Namespace: deployParameter.Namespace,
			Labels: map[string]string{
				OpenHydraUserLabelKey:     deployParameter.Username,
				OpenHydraWorkloadLabelKey: OpenHydraWorkloadLabelValue,
				OpenHydraIDELabelKey:      ideTypeLabelValue,
				OpenHydraSandboxKey:       deployParameter.SandboxName,
			},
		},
		Spec: appsV1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metaV1.LabelSelector{
				MatchLabels: map[string]string{
					OpenHydraUserLabelKey: deployParameter.Username,
				},
			},
			Template: coreV1.PodTemplateSpec{
				ObjectMeta: metaV1.ObjectMeta{
					Labels: map[string]string{
						OpenHydraUserLabelKey:     deployParameter.Username,
						OpenHydraWorkloadLabelKey: OpenHydraWorkloadLabelValue,
						OpenHydraIDELabelKey:      ideTypeLabelValue,
						OpenHydraSandboxKey:       deployParameter.SandboxName,
					},
				},
				Spec: coreV1.PodSpec{
					Volumes:    createVolume(deployParameter.Volumes),
					Containers: createContainers(baseName, deployParameter.Image, deployParameter.Username, deployParameter.VolumeMounts, resourceReq, resourceLim, deployParameter.Command, deployParameter.Args, deployParameter.Ports, deployParameter.CustomLabels),
				},
			},
		},
	}

	for key, value := range deployParameter.CustomLabels {
		// add label to deployment if key not found
		if _, ok := deployment.Labels[key]; !ok {
			deployment.Labels[key] = value
		}
		// add label to pod if key not found
		if _, ok := deployment.Spec.Template.Labels[key]; !ok {
			deployment.Spec.Template.Labels[key] = value
		}
	}

	deployment.Spec.Template.Spec.Affinity = deployParameter.Affinity

	return deployment
}

func (help *DefaultHelper) CreateDeployment(deployParameter *DeploymentParameters) error {
	if deployParameter.Client == nil {
		return fmt.Errorf("client is nil")
	}

	deployment := createDeployment(deployParameter)

	_, err := deployParameter.Client.AppsV1().Deployments(deployParameter.Namespace).Create(context.TODO(), deployment, metaV1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func createResource(cpuMemorySet CpuMemorySet, gpuSet apis.GpuSet) (coreV1.ResourceList, coreV1.ResourceList) {
	resourceReq := coreV1.ResourceList{
		coreV1.ResourceCPU:    resource.MustParse(cpuMemorySet.CpuRequest),
		coreV1.ResourceMemory: resource.MustParse(cpuMemorySet.MemoryRequest),
	}
	resourceLim := coreV1.ResourceList{
		coreV1.ResourceCPU:    resource.MustParse(cpuMemorySet.CpuLimit),
		coreV1.ResourceMemory: resource.MustParse(cpuMemorySet.MemoryLimit),
	}
	if gpuSet.Gpu > 0 {
		resourceReq[coreV1.ResourceName(gpuSet.GpuDriverName)] = resource.MustParse(strconv.Itoa(int(gpuSet.Gpu)))
		resourceLim[coreV1.ResourceName(gpuSet.GpuDriverName)] = resource.MustParse(strconv.Itoa(int(gpuSet.Gpu)))
	}
	return resourceReq, resourceLim
}

func createContainers(baseName, image, username string, volumes []apis.VolumeMount, resourceReq, resourceLimit coreV1.ResourceList, command, args []string, ports map[string]int, additionalLabels map[string]string) []coreV1.Container {
	container := coreV1.Container{
		Name:            baseName + "-container",
		Image:           image,
		ImagePullPolicy: coreV1.PullPolicy("IfNotPresent"),
		Resources: coreV1.ResourceRequirements{
			Limits:   resourceLimit,
			Requests: resourceReq,
		},
	}

	var envs []coreV1.EnvVar
	var portsExported []coreV1.ContainerPort
	for name, port := range ports {
		portsExported = append(portsExported, coreV1.ContainerPort{
			Name:          name,
			ContainerPort: int32(port),
		})
		envs = append(envs, coreV1.EnvVar{
			Name:  fmt.Sprintf("OPENHYDRA_%s", strings.ReplaceAll(strings.ToUpper(name), "-", "_")),
			Value: fmt.Sprintf("%s-%s", username, name),
		})
	}

	if _, ok := additionalLabels["openhydra-jupyter-lab-token"]; ok {
		envs = append(envs, coreV1.EnvVar{
			Name:  "OPENHYDRA_JUPYTER_LAB_TOKEN",
			Value: additionalLabels["openhydra-jupyter-lab-token"],
		})
	}

	if _, ok := additionalLabels["openhydra-service-account-token"]; ok {
		envs = append(envs, coreV1.EnvVar{
			Name:  "OPENHYDRA_SERVICE_ACCOUNT",
			Value: additionalLabels["openhydra-service-account-token"],
		})
	}

	if _, ok := additionalLabels["openhydra-server-address"]; ok {
		envs = append(envs, coreV1.EnvVar{
			Name:  "OPENHYDRA_SERVER_ADDRESS",
			Value: additionalLabels["openhydra-server-address"],
		})
	}

	if _, ok := additionalLabels["openhydra-heartbeat-lost-interval"]; ok {
		envs = append(envs, coreV1.EnvVar{
			Name:  "OPENHYDRA_HEARTBEAT_LOST_INTERVAL",
			Value: additionalLabels["openhydra-heartbeat-lost-interval"],
		})
	}

	envs = append(envs, coreV1.EnvVar{
		Name:  "OPENHYDRA_USER",
		Value: username,
	})

	container.Ports = portsExported
	container.Env = envs

	if len(command) > 0 {
		container.Command = command
	}

	if len(args) > 0 {
		container.Args = args
	}

	for _, volume := range volumes {
		container.VolumeMounts = append(container.VolumeMounts, coreV1.VolumeMount{
			Name:      volume.Name,
			MountPath: volume.MountPath,
			ReadOnly:  volume.ReadOnly,
		})
	}

	return []coreV1.Container{
		container,
	}
}

func createVolume(volumes []apis.Volume) []coreV1.Volume {
	// Todo: also move host path to volume
	var volumeMounts []coreV1.Volume

	for _, volume := range volumes {
		if volume.EmptyDir != nil {
			sizeLimit := resource.MustParse(fmt.Sprintf("%dMi", volume.EmptyDir.SizeLimit))
			volumeMounts = append(volumeMounts, coreV1.Volume{
				Name: volume.EmptyDir.Name,
				VolumeSource: coreV1.VolumeSource{
					EmptyDir: &coreV1.EmptyDirVolumeSource{
						Medium:    coreV1.StorageMedium(volume.EmptyDir.Medium),
						SizeLimit: &sizeLimit,
					},
				},
			})
		}
		if volume.HostPath != nil {
			volumeMounts = append(volumeMounts, coreV1.Volume{
				Name: volume.HostPath.Name,
				VolumeSource: coreV1.VolumeSource{
					HostPath: &coreV1.HostPathVolumeSource{
						Path: volume.HostPath.Path,
						Type: (*coreV1.HostPathType)(&volume.HostPath.Type),
					},
				},
			})
		}
	}

	return volumeMounts
}

func (help *DefaultHelper) CreateService(namespace, studentID, ideType string, client *kubernetes.Clientset, ports map[string]int) error {

	var portsExported []coreV1.ServicePort
	for name, port := range ports {
		portsExported = append(portsExported, coreV1.ServicePort{
			Name: name,
			Port: int32(port),
		})
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
			Ports: portsExported,
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
	selector, err := metaV1.ParseToLabelSelector(label)
	if err != nil {
		return nil, fmt.Errorf("failed to parse label selector: %v", err)
	}
	parsedSelector, err := metaV1.LabelSelectorAsSelector(selector)
	if err != nil {
		return nil, fmt.Errorf("failed to convert label selector: %v", err)
	}

	result, err := help.svcCache.Services(namespace).List(parsedSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("service not found")
	}

	return result[0], nil
}

func (help *DefaultHelper) GetAllNode(client *kubernetes.Clientset) ([]coreV1.Node, error) {

	nodes, err := help.nodeCache.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %v", err)
	}

	noRefNodes := make([]coreV1.Node, 0)
	for _, node := range nodes {
		noRefNodes = append(noRefNodes, *node)
	}

	return noRefNodes, nil
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

func (help *DefaultHelper) GetConfigMap(name, namespace string) (*coreV1.ConfigMap, error) {
	if help.configMapInformer == nil {
		return nil, fmt.Errorf("informer is not initialized")
	}

	key := fmt.Sprintf("%s/%s", namespace, name)

	cm, exists, err := help.configMapInformer.GetStore().GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("config map %s not found", key)
	}

	// Assert the object to *corev1.ConfigMap
	cmAsserted, ok := cm.(*coreV1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("object is not a ConfigMap")
	}

	return cmAsserted, nil
}

func (help *DefaultHelper) InitInformer() {
	factory := informers.NewSharedInformerFactory(help.clientSet, 0)
	if help.configMapInformer == nil {
		help.configMapInformer = factory.Core().V1().ConfigMaps().Informer()
	}
	if help.podInformer == nil {
		help.podInformer = factory.Core().V1().Pods().Informer()
	}

	if help.svcInformer == nil {
		help.svcInformer = factory.Core().V1().Services().Informer()
	}

	if help.nodeInformer == nil {
		help.nodeInformer = factory.Core().V1().Nodes().Informer()
	}
}

func (help *DefaultHelper) RunInformers(stopChan <-chan struct{}) {
	go help.configMapInformer.Run(stopChan)
	go help.podInformer.Run(stopChan)
	go help.svcInformer.Run(stopChan)
	go help.nodeInformer.Run(stopChan)
	if !cache.WaitForCacheSync(stopChan, help.configMapInformer.HasSynced, help.podInformer.HasSynced) {
		slog.Error("failed to sync all informers informer")
	}
	slog.Info("all informer synced")
	help.podCache = coreV1listers.NewPodLister(help.podInformer.GetIndexer())
	help.svcCache = coreV1listers.NewServiceLister(help.svcInformer.GetIndexer())
	help.nodeCache = coreV1listers.NewNodeLister(help.nodeInformer.GetIndexer())
}

func (help *DefaultHelper) UpdateConfigMap(name, namespace string, data map[string]string) error {
	if help.clientSet == nil {
		return fmt.Errorf("client is nil")
	}
	cm, err := help.clientSet.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metaV1.GetOptions{})
	if err != nil {
		return err
	}
	cm.Data = data
	_, err = help.clientSet.CoreV1().ConfigMaps(namespace).Update(context.TODO(), cm, metaV1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}
