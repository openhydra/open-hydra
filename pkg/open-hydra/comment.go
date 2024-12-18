package openhydra

import (
	"encoding/json"
	stdErr "errors"
	"fmt"
	"log/slog"
	"strings"

	"open-hydra/cmd/open-hydra-server/app/config"
	xDeviceV1 "open-hydra/pkg/apis/open-hydra-api/device/core/v1"
	xUserV1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"
	"open-hydra/pkg/open-hydra/apis"
	"open-hydra/pkg/open-hydra/k8s"
	"open-hydra/pkg/util"

	"github.com/emicklei/go-restful/v3"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const OpenhydraNamespace = "open-hydra"

type HttpErrMsg struct {
	Error string `json:"errMsg"`
}

func writeHttpResponseAndLogError(response *restful.Response, httpStatusCode int, err string) {
	slog.Error(err)
	response.WriteHeader(httpStatusCode)
	response.WriteAsJson(HttpErrMsg{Error: err})
}

func reasonAndCodeForError(err error) (metav1.StatusReason, int32) {
	if status, ok := err.(errors.APIStatus); ok || stdErr.As(err, &status) {
		return status.Status().Reason, status.Status().Code
	}
	return metav1.StatusReasonUnknown, 500
}

func writeAPIStatusError(response *restful.Response, err error) {
	_, code := reasonAndCodeForError(err)
	writeHttpResponseAndLogError(response, int(code), err.Error())
}

func combineDeviceList(pods []coreV1.Pod, services []coreV1.Service, users xUserV1.OpenHydraUserList, config *config.OpenHydraServerConfig) []xDeviceV1.Device {
	podFlat := make(map[string]coreV1.Pod)

	// now wo are going combine user and pod
	// first we put all pod into a map with label app as key
	for _, pod := range pods {
		if _, found := pod.Labels[k8s.OpenHydraUserLabelKey]; !found {
			continue
		}
		podFlat[pod.Labels[k8s.OpenHydraUserLabelKey]] = pod
	}

	serviceFlat := make(map[string]coreV1.Service)
	for _, service := range services {
		if _, found := service.Labels[k8s.OpenHydraUserLabelKey]; !found {
			continue
		}
		serviceFlat[service.Labels[k8s.OpenHydraUserLabelKey]] = service
	}

	var result []xDeviceV1.Device

	for _, user := range users.Items {

		device := xDeviceV1.Device{}
		util.FillKindAndApiVersion(&device.TypeMeta, "Device")
		device.Name = user.Name
		device.Namespace = user.Namespace
		device.Spec.Role = user.Spec.Role
		device.Spec.ChineseName = user.Spec.ChineseName
		if _, found := podFlat[user.Name]; found {
			// only fill up device if we found a pod
			device.Labels = podFlat[user.Name].Labels
			device.Spec.DeviceCpu = podFlat[user.Name].Spec.Containers[0].Resources.Requests.Cpu().String()
			device.Spec.DeviceRam = podFlat[user.Name].Spec.Containers[0].Resources.Requests.Memory().String()
			device.Spec.DeviceIP = podFlat[user.Name].Status.PodIP
			device.Spec.DeviceName = podFlat[user.Name].Name
			device.Spec.DeviceNamespace = podFlat[user.Name].Namespace
			if _, foundGpuDriver := podFlat[user.Name].Spec.Containers[0].Resources.Requests[coreV1.ResourceName(config.DefaultGpuDriver)]; foundGpuDriver {
				device.Spec.DeviceType = "gpu"
				device.Spec.GpuDriver = config.DefaultGpuDriver
				device.Spec.DeviceGpu = uint8(podFlat[user.Name].Spec.Containers[0].Resources.Requests.Name(coreV1.ResourceName(config.DefaultGpuDriver), resource.DecimalSI).Value())
			} else {
				device.Spec.DeviceType = "cpu"
			}
			device.Spec.OpenHydraUsername = user.Name
			// Todo: add line no
			device.Spec.LineNo = "0"
			device.CreationTimestamp = podFlat[user.Name].CreationTimestamp
			device.Spec.DeviceStatus = string(podFlat[user.Name].Status.Phase)
			if podFlat[user.Name].DeletionTimestamp != nil {
				device.Spec.DeviceStatus = "Terminating"
			}
			if _, found := podFlat[user.Name].Labels[k8s.OpenHydraSandboxKey]; found {
				device.Spec.SandboxName = podFlat[user.Name].Labels[k8s.OpenHydraSandboxKey]
			}
		}

		if _, found := serviceFlat[user.Name]; found {
			var portURLs []string
			for _, port := range serviceFlat[user.Name].Spec.Ports {
				portURLs = append(portURLs, combineUrl(config.ServerIP, user.Name, port.Name, port.NodePort, config.EnableJupyterLabBaseURL, config))
			}
			device.Spec.SandboxURLs = strings.Join(portURLs, ",")
		}
		result = append(result, device)
	}
	return result
}

func combineUrl(serverAddress, username, portName string, port int32, enableJupyterLabBaseURL bool, config *config.OpenHydraServerConfig) string {
	addressSet := strings.Split(serverAddress, ",")
	if len(addressSet) <= 1 {
		if enableJupyterLabBaseURL {
			if _, ok := config.ApplyPortNameForIngress[portName]; ok {
				return fmt.Sprintf("http://%s:%d/%s/%s", serverAddress, config.IngressPort, fmt.Sprintf("%s-%s", username, portName), config.ApplyPortNameForIngress[portName])
			} else {
				return fmt.Sprintf("http://%s:%d/%s", serverAddress, config.IngressPort, username)
			}
		} else {
			return fmt.Sprintf("http://%s:%d", serverAddress, port)
		}
	}
	var result []string
	for _, address := range addressSet {
		result = append(result, fmt.Sprintf("http://%s:%d", address, port))
	}
	return strings.Join(result, ",")
}

func ParseJsonToPluginList(jsonData string) (apis.PluginList, error) {
	var plugins apis.PluginList
	err := json.Unmarshal([]byte(jsonData), &plugins)
	if err != nil {
		return plugins, err
	}
	return plugins, nil
}

func preCreateUserDir(volumes []apis.Volume, username string, config *config.OpenHydraServerConfig) error {
	for index, volume := range volumes {
		if volume.HostPath == nil {
			continue
		}
		dirToCreate := volume.HostPath.Path
		if strings.Contains(volume.HostPath.Path, "{username}") || strings.Contains(volume.HostPath.Path, "{workspace}") {
			// only private dir needs to be create on pod booting
			dirToCreate = strings.Replace(dirToCreate, "{username}", username, -1)
			dirToCreate = strings.Replace(dirToCreate, "{workspace}", config.WorkspacePath, -1)
			err := util.CreateDirIfNotExists(dirToCreate)
			if err != nil {
				return err
			}
			volumes[index].HostPath.Path = dirToCreate

			// copy a file read me to the workspace
			err = util.CopyFile(fmt.Sprintf("%s/%s", config.WorkspacePath, "README.md"), fmt.Sprintf("%s/%s", dirToCreate, "README.md"))
			if err != nil {
				slog.Error("copy readme failed", "msg", err)
			}

			continue
		}
		if strings.Contains(volume.HostPath.Path, "{dataset-public}") {
			dirToCreate = strings.Replace(dirToCreate, "{dataset-public}", config.PublicDatasetBasePath, -1)
		}
		if strings.Contains(volume.HostPath.Path, "{course-public}") {
			dirToCreate = strings.Replace(dirToCreate, "{course-public}", config.PublicCourseBasePath, -1)
		}

		volumes[index].HostPath.Path = dirToCreate
	}
	return nil
}
