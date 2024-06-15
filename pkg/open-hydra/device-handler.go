package openhydra

import (
	"fmt"
	"log/slog"
	"net/http"
	xDeviceV1 "open-hydra/pkg/apis/open-hydra-api/device/core/v1"
	v1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"
	envApi "open-hydra/pkg/open-hydra/apis"
	"open-hydra/pkg/open-hydra/k8s"
	"open-hydra/pkg/util"
	"strconv"

	"github.com/emicklei/go-restful/v3"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (builder *OpenHydraRouteBuilder) AddDeviceListRoute() {
	path := "/" + DevicePath
	builder.addPathAuthorization(path, http.MethodGet, 1)
	builder.RootWS.Route(builder.RootWS.GET(path).Operation("listDevice").To(builder.DeviceListRouteHandler).
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", "").
		Returns(http.StatusOK, "OK", xDeviceV1.DeviceList{}))
}

func (builder *OpenHydraRouteBuilder) DeviceListRouteHandler(request *restful.Request, response *restful.Response) {
	result := xDeviceV1.DeviceList{}
	result.Kind = "List"
	result.APIVersion = "v1"

	users, err := builder.Database.ListUsers()
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, err.Error())
		return
	}

	allUserDevice, err := builder.k8sHelper.ListPod(builder.Config.OpenHydraNamespace, builder.kubeClient)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, err.Error())
		return
	}

	allUserService, err := builder.k8sHelper.ListService(builder.Config.OpenHydraNamespace, builder.kubeClient)
	if err != nil {
		slog.Warn("Failed to list service", err)
	}

	result.Items = combineDeviceList(allUserDevice, allUserService, users, builder.Config)

	response.WriteEntity(result)
}

func (builder *OpenHydraRouteBuilder) AddDeviceGetRoute() {
	path := "/" + DevicePath + "/{username}"
	builder.addPathAuthorization(path, http.MethodGet, 3)
	builder.RootWS.Route(builder.RootWS.GET(path).Operation("getDevice").To(builder.DeviceGetRouteHandler).
		Returns(http.StatusNotFound, "not found", "").
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", "").
		Returns(http.StatusOK, "OK", xDeviceV1.Device{}))
}

func (builder *OpenHydraRouteBuilder) DeviceGetRouteHandler(request *restful.Request, response *restful.Response) {
	if request.PathParameter("username") == "" {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, "username is empty")
		return
	}

	username := request.PathParameter("username")
	if !builder.Config.DisableAuth {
		reqUser := request.HeaderParameter(openHydraHeaderUser)
		reqRole := request.HeaderParameter(openHydraHeaderRole)
		if reqUser == "" || reqRole == "" {
			writeHttpResponseAndLogError(response, http.StatusUnauthorized, "no user or role found in request header")
			return
		}

		if reqRole != "1" {
			// only teacher can get other user info
			if username != reqUser {
				writeHttpResponseAndLogError(response, http.StatusForbidden, fmt.Sprintf("user: %s do not have the right to get device for user: %s", reqUser, username))
				return
			}
		}
	}

	user, err := builder.Database.GetUser(username)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, err.Error())
		return
	}

	userLabel := fmt.Sprintf("%s=%s", k8s.OpenHydraUserLabelKey, username)

	device, err := builder.k8sHelper.GetUserPods(userLabel, builder.Config.OpenHydraNamespace, builder.kubeClient)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, err.Error())
		return
	}

	service, err := builder.k8sHelper.GetUserService(userLabel, builder.Config.OpenHydraNamespace, builder.kubeClient)
	if err != nil {
		slog.Warn("Failed to get user service", err)
	}

	services := []coreV1.Service{}

	if service != nil {
		services = append(services, *service)
	}

	result := combineDeviceList(device, services, v1.OpenHydraUserList{Items: []v1.OpenHydraUser{*user}}, builder.Config)
	if len(result) == 0 {
		writeHttpResponseAndLogError(response, http.StatusNotFound, "not found")
		return
	}

	response.WriteAsJson(result[0])
}

func (builder *OpenHydraRouteBuilder) AddDeviceCreateRoute() {
	path := "/" + DevicePath
	builder.addPathAuthorization(path, http.MethodPost, 3)
	builder.RootWS.Route(builder.RootWS.POST(path).Operation("createDevice").To(builder.DeviceCreateRouteHandler).
		Returns(http.StatusCreated, "created", "").
		Returns(http.StatusBadRequest, "bad request", "").
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", ""))
}

func (builder *OpenHydraRouteBuilder) DeviceCreateRouteHandler(request *restful.Request, response *restful.Response) {
	reqDevice := xDeviceV1.Device{}
	err := request.ReadEntity(&reqDevice)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, fmt.Sprintf("Failed to read request entity due to %s", err.Error()))
		return
	}

	if reqDevice.Spec.OpenHydraUsername == "" {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, "openHydraUsername is empty")
		return
	}

	if reqDevice.Spec.SandboxName == "" {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, "SandboxName is empty")
		return
	}

	if !builder.Config.DisableAuth {
		reqUser := request.HeaderParameter(openHydraHeaderUser)
		reqRole := request.HeaderParameter(openHydraHeaderRole)
		if reqUser == "" || reqRole == "" {
			writeHttpResponseAndLogError(response, http.StatusUnauthorized, "no user or role found in request header")
			return
		}

		if reqRole != "1" {
			// only teacher can get other user info
			if reqDevice.Spec.OpenHydraUsername != reqUser {
				writeHttpResponseAndLogError(response, http.StatusForbidden, fmt.Sprintf("user: %s do not have the right to create device for user: %s", reqUser, reqDevice.Spec.OpenHydraUsername))
				return
			}

			if reqDevice.Spec.DeviceGpu != 0 {
				writeHttpResponseAndLogError(response, http.StatusForbidden, "user do not have the right to create gpu device")
				return
			}
		}
	}

	// check if user exists
	_, err = builder.Database.GetUser(reqDevice.Spec.OpenHydraUsername)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, "user not found")
		return
	}

	// check if device already exists
	pod, err := builder.k8sHelper.ListPodWithLabel(fmt.Sprintf("%s=%s", k8s.OpenHydraUserLabelKey, reqDevice.Spec.OpenHydraUsername), builder.Config.OpenHydraNamespace, builder.kubeClient)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, err.Error())
		return
	}
	// we consider pod as user device if we found it already exists then we return error
	if len(pod) > 0 {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, fmt.Sprintf("device with student name %s already exists", reqDevice.Spec.OpenHydraUsername))
		return
	}

	gpuSet := builder.BuildGpu(reqDevice)

	// we need to get config map openhydra-plugin first
	// TODO: we should use informer to cache config map instead of query api-server directly for performance
	pluginConfigMap, err := builder.k8sHelper.GetMap("openhydra-plugin", builder.Config.OpenHydraNamespace, builder.kubeClient)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get configmap: %v", err))
		return
	}

	// parse to plugin list
	plugins, err := ParseJsonToPluginList(pluginConfigMap.Data["plugins"])
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to unmarshal json: %v", err))
		return
	}

	var image string
	var ports map[string]int
	var command, args []string
	var volumeMounts []envApi.VolumeMount
	var volumes []envApi.Volume
	if _, found := plugins.Sandboxes[reqDevice.Spec.SandboxName]; !found {
		// if sandbox name did not match any sandbox name then return error
		writeHttpResponseAndLogError(response, http.StatusBadRequest, fmt.Sprintf("sandbox %s not found, please ensure sandbox is proper config", reqDevice.Spec.SandboxName))
		return
	} else {
		if len(plugins.Sandboxes[reqDevice.Spec.SandboxName].Ports) > int(builder.Config.MaximumPortsPerSandbox) {
			// if sandbox exceed maximum ports limit then return error
			writeHttpResponseAndLogError(response, http.StatusBadRequest, fmt.Sprintf("sandbox %s exceed maximum ports limit", reqDevice.Spec.SandboxName))
			return
		}
		// TODO: we need consider security issue for certain volume mount
		volumeMounts = plugins.Sandboxes[reqDevice.Spec.SandboxName].VolumeMounts
		volumes = plugins.Sandboxes[reqDevice.Spec.SandboxName].Volumes
		// handle private dir creation
		err = preCreateUserDir(volumes, reqDevice.Spec.OpenHydraUsername, builder.Config)
		if err != nil {
			writeHttpResponseAndLogError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to create user dir: %v", err))
			return
		}
		command = plugins.Sandboxes[reqDevice.Spec.SandboxName].Command
		args = plugins.Sandboxes[reqDevice.Spec.SandboxName].Args
		ports = make(map[string]int)
		for index, port := range plugins.Sandboxes[reqDevice.Spec.SandboxName].Ports {
			name := fmt.Sprintf("port-%d", index)
			ports[name] = int(port)
		}
		// set image with different hardware type if match
		if gpuSet.Gpu > 0 {
			// go with gpu image
			image = plugins.Sandboxes[reqDevice.Spec.SandboxName].GPUImageName
			slog.Debug(fmt.Sprintf("set image to gpu image '%s'", image))
		} else {
			// go with cpu image
			image = plugins.Sandboxes[reqDevice.Spec.SandboxName].CPUImageName
			slog.Debug(fmt.Sprintf("set image to cpu image '%s'", image))
		}
	}

	// if image is empty then return error
	if image == "" {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, fmt.Sprintf("no image found for sandbox %s", reqDevice.Spec.SandboxName))
		return
	}

	// if no ports found then return error
	if len(ports) == 0 {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, fmt.Sprintf("no ports found for sandbox %s", reqDevice.Spec.SandboxName))
		return
	}

	deployParameter := &k8s.DeploymentParameters{
		CpuMemorySet: builder.CombineReqLimit(reqDevice),
		Image:        image,
		Namespace:    builder.Config.OpenHydraNamespace,
		Username:     reqDevice.Spec.OpenHydraUsername,
		SandboxName:  reqDevice.Spec.SandboxName,
		VolumeMounts: volumeMounts,
		GpuSet:       gpuSet,
		Client:       builder.kubeClient,
		Command:      command,
		Args:         args,
		Ports:        ports,
		Volumes:      volumes,
		Affinity:     reqDevice.Spec.Affinity,
	}

	err = builder.k8sHelper.CreateDeployment(deployParameter)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, err.Error())
		return
	}

	err = builder.k8sHelper.CreateService(builder.Config.OpenHydraNamespace, reqDevice.Spec.OpenHydraUsername, reqDevice.Spec.SandboxName, builder.kubeClient, ports)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, err.Error())
		return
	}

	reqDevice.Spec.DeviceType = "cpu"
	if gpuSet.Gpu > 0 {
		reqDevice.Spec.DeviceType = "gpu"
	}

	reqDevice.Spec.DeviceStatus = "Creating"
	response.WriteEntity(&reqDevice)
}

func (builder *OpenHydraRouteBuilder) AddDeviceUpdateRoute() {
	path := "/" + DevicePath + "/{username}"
	builder.addPathAuthorization(path, http.MethodPut, 1)
	builder.RootWS.Route(builder.RootWS.PUT(path).Operation("getUpdateDevice").To(builder.DeviceUpdateRouteHandler).
		Returns(http.StatusNotFound, "not found", "").
		Returns(http.StatusBadRequest, "bad request", "").
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", ""))
}

func (builder *OpenHydraRouteBuilder) DeviceUpdateRouteHandler(request *restful.Request, response *restful.Response) {
	if request.PathParameter("username") == "" {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, "username is empty")
		return
	}

	_, found := builder.CacheDevices[request.PathParameter("username")]
	if !found {
		writeHttpResponseAndLogError(response, http.StatusNotFound, "not found")
		return
	}

	response.WriteAsJson(builder.CacheDevices[request.PathParameter("username")])
}

func (builder *OpenHydraRouteBuilder) AddDeviceDeleteRoute() {
	path := "/" + DevicePath + "/{username}"
	builder.addPathAuthorization(path, http.MethodDelete, 3)
	builder.RootWS.Route(builder.RootWS.DELETE(path).Operation("deleteDevice").To(builder.DeviceDeleteRouteHandler).
		Returns(http.StatusNotFound, "not found", "").
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", ""))
}

func (builder *OpenHydraRouteBuilder) DeviceDeleteRouteHandler(request *restful.Request, response *restful.Response) {
	if request.PathParameter("username") == "" {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, "username is empty")
		return
	}

	username := request.PathParameter("username")
	if !builder.Config.DisableAuth {
		reqUser := request.HeaderParameter(openHydraHeaderUser)
		reqRole := request.HeaderParameter(openHydraHeaderRole)
		if reqUser == "" || reqRole == "" {
			writeHttpResponseAndLogError(response, http.StatusUnauthorized, "no user or role found in request header")
			return
		}

		if reqRole != "1" {
			// only teacher can get other user info
			if username != reqUser {
				writeHttpResponseAndLogError(response, http.StatusForbidden, fmt.Sprintf("user: %s do not have the right to delete device for user: %s", reqUser, username))
				return
			}
		}
	}

	err := builder.k8sHelper.DeleteUserDeployment(fmt.Sprintf("%s=%s", k8s.OpenHydraUserLabelKey, username), builder.Config.OpenHydraNamespace, builder.kubeClient)
	if err != nil {
		slog.Error("Failed to delete user deployment will proceed to delete service any way", err)
	}

	if builder.Config.PatchResourceNotRelease {
		// if with certain calico version, we may encounter bug like delete deploy but rs and pod will not be deleted
		// so we have to manually delete rs and pod
		err = builder.k8sHelper.DeleteUserReplicaSet(fmt.Sprintf("%s=%s", k8s.OpenHydraUserLabelKey, username), builder.Config.OpenHydraNamespace, builder.kubeClient)
		if err != nil {
			slog.Error("patch:PatchResourceNotRelease -> Failed to delete user replica set will proceed anyway", err)
		}

		err = builder.k8sHelper.DeleteUserPod(fmt.Sprintf("%s=%s", k8s.OpenHydraUserLabelKey, username), builder.Config.OpenHydraNamespace, builder.kubeClient)
		if err != nil {
			slog.Error("patch:PatchResourceNotRelease -> Failed to delete user pod will proceed anyway", err)
		}
	}

	err = builder.k8sHelper.DeleteUserService(fmt.Sprintf("%s=%s", k8s.OpenHydraUserLabelKey, username), builder.Config.OpenHydraNamespace, builder.kubeClient)
	if err != nil {
		slog.Error("Failed to delete user service", err)
	}

	result := xDeviceV1.Device{
		ObjectMeta: metaV1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", username, "device"),
		},
		Spec: xDeviceV1.DeviceSpec{
			OpenHydraUsername: username,
			DeviceStatus:      "Terminating",
		},
	}

	util.FillKindAndApiVersion(&result.TypeMeta, "Device")
	response.WriteEntity(&result)
}

func (builder *OpenHydraRouteBuilder) GetCpu(postDevice xDeviceV1.Device) (string, string) {
	cpuReq := builder.Config.DefaultCpuPerDevice
	cpuLimit := builder.Config.DefaultCpuPerDevice
	if postDevice.Spec.DeviceCpu != "" {
		i, err := strconv.ParseUint(postDevice.Spec.DeviceCpu, 10, 64)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed convert %s to uint64 fall back to default cpu setting", postDevice.Spec.DeviceCpu))
		} else {
			cpuReq = i
			cpuLimit = i
		}
	}

	if builder.Config.CpuOverCommitRate > 1 {
		// apply cpu over commit rate
		cpuReq = cpuReq / uint64(builder.Config.CpuOverCommitRate)
	}

	return fmt.Sprintf("%dm", cpuReq), fmt.Sprintf("%dm", cpuLimit)
}

func (builder *OpenHydraRouteBuilder) GetRam(postDevice xDeviceV1.Device) (string, string) {
	memoryReq := builder.Config.DefaultRamPerDevice
	memoryLimit := builder.Config.DefaultRamPerDevice
	if postDevice.Spec.DeviceRam != "" {
		i, err := strconv.ParseUint(postDevice.Spec.DeviceRam, 10, 64)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed convert %s to uint64 fall back to default memory setting", postDevice.Spec.DeviceRam))
		} else {
			memoryReq = i
			memoryLimit = i
		}
	}

	if builder.Config.MemoryOverCommitRate > 1 {
		// apply memory over commit rate
		memoryReq = memoryReq / uint64(builder.Config.MemoryOverCommitRate)
	}

	return fmt.Sprintf("%dMi", memoryReq), fmt.Sprintf("%dMi", memoryLimit)
}

func (builder *OpenHydraRouteBuilder) CombineReqLimit(postDevice xDeviceV1.Device) k8s.CpuMemorySet {
	cpuReq, cpuLimit := builder.GetCpu(postDevice)
	memoryReq, memoryLimit := builder.GetRam(postDevice)
	return k8s.CpuMemorySet{
		CpuRequest:    cpuReq,
		CpuLimit:      cpuLimit,
		MemoryRequest: memoryReq,
		MemoryLimit:   memoryLimit,
	}
}

func (builder *OpenHydraRouteBuilder) BuildGpu(postDevice xDeviceV1.Device) envApi.GpuSet {
	result := envApi.GpuSet{
		GpuDriverName: builder.Config.DefaultGpuDriver,
		Gpu:           builder.Config.DefaultGpuPerDevice,
	}
	if postDevice.Spec.DeviceGpu > 0 {
		result.Gpu = postDevice.Spec.DeviceGpu
	}

	if postDevice.Spec.GpuDriver != "" {
		result.GpuDriverName = postDevice.Spec.GpuDriver
	}

	return result
}
