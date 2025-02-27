package openhydra

import (
	"fmt"
	"log/slog"
	"net/http"
	"open-hydra/cmd/open-hydra-server/app/config"
	xDeviceV1 "open-hydra/pkg/apis/open-hydra-api/device/core/v1"
	v1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"
	envApi "open-hydra/pkg/open-hydra/apis"
	"open-hydra/pkg/open-hydra/k8s"
	"open-hydra/pkg/util"
	"os"
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

	filter := request.QueryParameters("group")
	fmt.Println(filter)

	serverConfig, err := builder.GetServerConfigFromConfigMap()
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError,
			fmt.Sprintf("Failed to get server config: %v", err))
		return
	}

	result := xDeviceV1.DeviceList{}
	result.Kind = "List"
	result.APIVersion = "v1"

	users, err := builder.Database.ListUsers()
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, err.Error())
		return
	}

	allUserDevice, err := builder.k8sHelper.ListPod("open-hydra", builder.kubeClient)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, err.Error())
		return
	}

	allUserService, err := builder.k8sHelper.ListService(OpenhydraNamespace, builder.kubeClient)
	if err != nil {
		slog.Warn("Failed to list service", err)
	}

	result.Items = combineDeviceList(allUserDevice, allUserService, users, serverConfig)

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

	serverConfig, err := builder.GetServerConfigFromConfigMap()
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError,
			fmt.Sprintf("Failed to get server config: %v", err))
		return
	}

	username := request.PathParameter("username")
	if !serverConfig.DisableAuth {
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

	device, err := builder.k8sHelper.GetUserPods(userLabel, OpenhydraNamespace, builder.kubeClient)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, err.Error())
		return
	}

	service, err := builder.k8sHelper.GetUserService(userLabel, OpenhydraNamespace, builder.kubeClient)
	if err != nil {
		slog.Warn("Failed to get user service", err)
	}

	services := []coreV1.Service{}

	if service != nil {
		services = append(services, *service)
	}

	result := combineDeviceList(device, services, v1.OpenHydraUserList{Items: []v1.OpenHydraUser{*user}}, serverConfig)
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

	slog.Info("Received request to create device", "device", reqDevice)

	if reqDevice.Spec.OpenHydraUsername == "" {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, "openHydraUsername is empty")
		return
	}

	if reqDevice.Spec.SandboxName == "" {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, "SandboxName is empty")
		return
	}

	serverConfig, err := builder.GetServerConfigFromConfigMap()
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError,
			fmt.Sprintf("Failed to get server config: %v", err))
		return
	}

	if !serverConfig.DisableAuth {
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
	pod, err := builder.k8sHelper.ListPodWithLabel(fmt.Sprintf("%s=%s", k8s.OpenHydraUserLabelKey, reqDevice.Spec.OpenHydraUsername), OpenhydraNamespace, builder.kubeClient)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, err.Error())
		return
	}
	// we consider pod as user device if we found it already exists then we return error
	if len(pod) > 0 {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, fmt.Sprintf("device with student name %s already exists", reqDevice.Spec.OpenHydraUsername))
		return
	}

	gpuSet := builder.BuildGpu(reqDevice, serverConfig)

	// we need to get config map openhydra-plugin first
	// TODO: we should use informer to cache config map instead of query api-server directly for performance
	pluginConfigMap, err := builder.k8sHelper.GetConfigMap("openhydra-plugin", OpenhydraNamespace)
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
		if len(plugins.Sandboxes[reqDevice.Spec.SandboxName].Ports) > int(serverConfig.MaximumPortsPerSandbox) {
			// if sandbox exceed maximum ports limit then return error
			writeHttpResponseAndLogError(response, http.StatusBadRequest, fmt.Sprintf("sandbox %s exceed maximum ports limit", reqDevice.Spec.SandboxName))
			return
		}
		// TODO: we need consider security issue for certain volume mount
		volumeMounts = plugins.Sandboxes[reqDevice.Spec.SandboxName].VolumeMounts
		volumes = plugins.Sandboxes[reqDevice.Spec.SandboxName].Volumes
		// handle private dir creation
		err = preCreateUserDir(volumes, reqDevice.Spec.OpenHydraUsername, serverConfig)
		if err != nil {
			writeHttpResponseAndLogError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to create user dir: %v", err))
			return
		}

		if builder.cfg.AddProjectResource && reqDevice.Spec.OpenHydraProjectId != "" {
			// check if dir exists
			projectDatasetFullPath := fmt.Sprintf("%s/%s", builder.cfg.ProjectDatasetBasePath, reqDevice.Spec.OpenHydraProjectId)
			_, err := os.Stat(projectDatasetFullPath)
			if err != nil {
				slog.Warn(fmt.Sprintf("project dataset base path %s not found will not mount project dataset dir in container", builder.cfg.ProjectDatasetBasePath))
			} else {
				volumes = append(volumes, envApi.Volume{
					HostPath: &envApi.HostPath{
						Name: "project-dataset",
						Path: projectDatasetFullPath,
					},
				})
				volumeMounts = append(volumeMounts, envApi.VolumeMount{
					Name:      "project-dataset",
					MountPath: builder.cfg.ProjectDatasetStudentMountPath,
					ReadOnly:  true,
				})
			}

			projectCourseFullPath := fmt.Sprintf("%s/%s", builder.cfg.ProjectCourseBasePath, reqDevice.Spec.OpenHydraProjectId)
			_, err = os.Stat(projectCourseFullPath)
			if err != nil {
				slog.Warn(fmt.Sprintf("project course base path %s not found will not mount project course dir in container", builder.cfg.ProjectCourseBasePath))
			} else {
				volumes = append(volumes, envApi.Volume{
					HostPath: &envApi.HostPath{
						Name: "project-course",
						Path: projectCourseFullPath,
					},
				})
				volumeMounts = append(volumeMounts, envApi.VolumeMount{
					Name:      "project-course",
					MountPath: builder.cfg.ProjectCourseStudentMountPath,
					ReadOnly:  true,
				})
			}
		}

		command = plugins.Sandboxes[reqDevice.Spec.SandboxName].Command
		args = plugins.Sandboxes[reqDevice.Spec.SandboxName].Args
		ports = make(map[string]int)
		for _, port := range plugins.Sandboxes[reqDevice.Spec.SandboxName].Ports {

			ports[port.Name] = int(port.Port)
		}
		// set image with different hardware type if match
		if gpuSet.Gpu > 0 {
			// go with gpu image
			if reqDevice.Spec.GpuDriver == "" {
				if serverConfig.DefaultGpuDriver == "" {
					writeHttpResponseAndLogError(response, http.StatusBadRequest, "both gpu driver and DefaultGpuDriver are empty")
					return
				}
				reqDevice.Spec.GpuDriver = serverConfig.DefaultGpuDriver
			}

			// ensure gpu is allowed
			// should be in config
			gpuIsAllowed := false
			for _, gpuAllowed := range serverConfig.GpuResourceKeys {
				if gpuAllowed == reqDevice.Spec.GpuDriver {
					gpuIsAllowed = true
					break
				}
			}
			if !gpuIsAllowed {
				writeHttpResponseAndLogError(response, http.StatusBadRequest, fmt.Sprintf("gpu driver %s is not allowed", reqDevice.Spec.GpuDriver))
				return
			}

			// ensure key is found in GPUImageSet
			// we do not put any default fall back option here which is on purpose
			// because different gpu must go with different image especially for none cuda compatible gpu
			if _, found := plugins.Sandboxes[reqDevice.Spec.SandboxName].GPUImageSet[reqDevice.Spec.GpuDriver]; !found {
				writeHttpResponseAndLogError(response, http.StatusBadRequest, fmt.Sprintf("gpu image %s not found in sandbox %s", reqDevice.Spec.GpuDriver, reqDevice.Spec.SandboxName))
				return
			}

			if plugins.Sandboxes[reqDevice.Spec.SandboxName].GPUImageSet[reqDevice.Spec.GpuDriver] == "" {
				writeHttpResponseAndLogError(response, http.StatusBadRequest, fmt.Sprintf("gpu image %s is empty in sandbox %s", reqDevice.Spec.GpuDriver, reqDevice.Spec.SandboxName))
				return
			}

			image = plugins.Sandboxes[reqDevice.Spec.SandboxName].GPUImageSet[reqDevice.Spec.GpuDriver]
			slog.Debug(fmt.Sprintf("set image to gpu image '%s' with driver name '%s'", image, reqDevice.Spec.GpuDriver))
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
		CpuMemorySet: builder.CombineReqLimit(reqDevice, serverConfig),
		Image:        image,
		Namespace:    OpenhydraNamespace,
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
		CustomLabels: reqDevice.Labels,
	}

	err = builder.k8sHelper.CreateDeployment(deployParameter)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, err.Error())
		return
	}

	err = builder.k8sHelper.CreateService(OpenhydraNamespace, reqDevice.Spec.OpenHydraUsername, reqDevice.Spec.SandboxName, builder.kubeClient, ports)
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

	serverConfig, err := builder.GetServerConfigFromConfigMap()
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError,
			fmt.Sprintf("Failed to get server config: %v", err))
		return
	}

	username := request.PathParameter("username")
	if !serverConfig.DisableAuth {
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

	err = builder.k8sHelper.DeleteUserDeployment(fmt.Sprintf("%s=%s", k8s.OpenHydraUserLabelKey, username), OpenhydraNamespace, builder.kubeClient)
	if err != nil {
		slog.Error("Failed to delete user deployment will proceed to delete service any way", err)
	}

	if serverConfig.PatchResourceNotRelease {
		// if with certain calico version, we may encounter bug like delete deploy but rs and pod will not be deleted
		// so we have to manually delete rs and pod
		err = builder.k8sHelper.DeleteUserReplicaSet(fmt.Sprintf("%s=%s", k8s.OpenHydraUserLabelKey, username), OpenhydraNamespace, builder.kubeClient)
		if err != nil {
			slog.Error("patch:PatchResourceNotRelease -> Failed to delete user replica set will proceed anyway", err)
		}

		err = builder.k8sHelper.DeleteUserPod(fmt.Sprintf("%s=%s", k8s.OpenHydraUserLabelKey, username), OpenhydraNamespace, builder.kubeClient)
		if err != nil {
			slog.Error("patch:PatchResourceNotRelease -> Failed to delete user pod will proceed anyway", err)
		}
	}

	err = builder.k8sHelper.DeleteUserService(fmt.Sprintf("%s=%s", k8s.OpenHydraUserLabelKey, username), OpenhydraNamespace, builder.kubeClient)
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

func (builder *OpenHydraRouteBuilder) GetCpu(postDevice xDeviceV1.Device, serverConfig *config.OpenHydraServerConfig) (string, string) {
	cpuReq := serverConfig.DefaultCpuPerDevice
	cpuLimit := serverConfig.DefaultCpuPerDevice
	if postDevice.Spec.DeviceCpu != "" {
		i, err := strconv.ParseUint(postDevice.Spec.DeviceCpu, 10, 64)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed convert %s to uint64 fall back to default cpu setting", postDevice.Spec.DeviceCpu))
		} else {
			cpuReq = i
			cpuLimit = i
		}
	}

	if serverConfig.CpuOverCommitRate > 1 {
		// apply cpu over commit rate
		cpuReq = cpuReq / uint64(serverConfig.CpuOverCommitRate)
	}

	return fmt.Sprintf("%dm", cpuReq), fmt.Sprintf("%dm", cpuLimit)
}

func (builder *OpenHydraRouteBuilder) GetRam(postDevice xDeviceV1.Device, serverConfig *config.OpenHydraServerConfig) (string, string) {
	memoryReq := serverConfig.DefaultRamPerDevice
	memoryLimit := serverConfig.DefaultRamPerDevice
	if postDevice.Spec.DeviceRam != "" {
		i, err := strconv.ParseUint(postDevice.Spec.DeviceRam, 10, 64)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed convert %s to uint64 fall back to default memory setting", postDevice.Spec.DeviceRam))
		} else {
			memoryReq = i
			memoryLimit = i
		}
	}

	if serverConfig.MemoryOverCommitRate > 1 {
		// apply memory over commit rate
		memoryReq = memoryReq / uint64(serverConfig.MemoryOverCommitRate)
	}

	return fmt.Sprintf("%dMi", memoryReq), fmt.Sprintf("%dMi", memoryLimit)
}

func (builder *OpenHydraRouteBuilder) CombineReqLimit(postDevice xDeviceV1.Device, serverConfig *config.OpenHydraServerConfig) k8s.CpuMemorySet {
	cpuReq, cpuLimit := builder.GetCpu(postDevice, serverConfig)
	memoryReq, memoryLimit := builder.GetRam(postDevice, serverConfig)
	return k8s.CpuMemorySet{
		CpuRequest:    cpuReq,
		CpuLimit:      cpuLimit,
		MemoryRequest: memoryReq,
		MemoryLimit:   memoryLimit,
	}
}

func (builder *OpenHydraRouteBuilder) BuildGpu(postDevice xDeviceV1.Device, serverConfig *config.OpenHydraServerConfig) envApi.GpuSet {
	result := envApi.GpuSet{
		GpuDriverName: serverConfig.DefaultGpuDriver,
		Gpu:           postDevice.Spec.DeviceGpu,
	}

	if result.Gpu == 0 && serverConfig.UseDefaultGpuConfigWhenZeroIsGiven {
		result.Gpu = serverConfig.DefaultGpuPerDevice
	}

	if postDevice.Spec.GpuDriver != "" {
		result.GpuDriverName = postDevice.Spec.GpuDriver
	}

	return result
}
