package openhydra

import (
	stdErr "errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/emicklei/go-restful/v3"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"open-hydra/cmd/open-hydra-server/app/config"
	xDeviceV1 "open-hydra/pkg/apis/open-hydra-api/device/core/v1"
	xUserV1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"
	"open-hydra/pkg/open-hydra/k8s"
	"open-hydra/pkg/util"
)

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
		device.Spec.IDEType = k8s.OpenHydraIDELabelUnset
		if _, found := podFlat[user.Name]; found {
			// only fill up device if we found a pod
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
			if _, found := podFlat[user.Name].Labels[k8s.OpenHydraIDELabelKey]; found {
				device.Spec.IDEType = podFlat[user.Name].Labels[k8s.OpenHydraIDELabelKey]
			}
		}

		if _, found := serviceFlat[user.Name]; found {
			if podFlat[user.Name].Labels[k8s.OpenHydraIDELabelKey] == k8s.OpenHydraIDELabelVSCode {
				device.Spec.VSCodeURL = combineUrl(config.ServerIP, serviceFlat[user.Name].Spec.Ports[0].NodePort)
			} else {
				device.Spec.EasyTrainURL = combineUrl(config.ServerIP, serviceFlat[user.Name].Spec.Ports[0].NodePort)
				device.Spec.JupyterLabURL = combineUrl(config.ServerIP, serviceFlat[user.Name].Spec.Ports[1].NodePort)
			}

		}
		result = append(result, device)
	}
	return result
}

func combineUrl(serverAddress string, port int32) string {
	addressSet := strings.Split(serverAddress, ",")
	if len(addressSet) <= 1 {
		return fmt.Sprintf("http://%s:%d", serverAddress, port)
	}
	var result []string
	for _, address := range addressSet {
		result = append(result, fmt.Sprintf("http://%s:%d", address, port))
	}
	return strings.Join(result, ",")
}
