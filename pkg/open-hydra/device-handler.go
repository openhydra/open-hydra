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
	"path"

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

	// create a dir on host for jupyter lab for this user if it does not exist
	err = util.CreateDirIfNotExists(path.Join(builder.Config.JupyterLabHostBaseDir, reqDevice.Spec.OpenHydraUsername))
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, err.Error())
		return
	}

	// create a dir on host for vscode for this user if it does not exist
	err = util.CreateDirIfNotExists(path.Join(builder.Config.PublicVSCodeBasePath, reqDevice.Spec.OpenHydraUsername))
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, err.Error())
		return
	}

	gpuSet := builder.BuildGpu(reqDevice)

	image := builder.Config.ImageRepo
	if reqDevice.Spec.IDEType == k8s.OpenHydraIDELabelVSCode {
		image = builder.Config.VSCodeImageRepo
	}

	err = builder.k8sHelper.CreateDeployment(builder.GetCpu(reqDevice), builder.GetRam(reqDevice), image, builder.Config.OpenHydraNamespace, reqDevice.Spec.OpenHydraUsername, reqDevice.Spec.IDEType, builder.BuildVolumes(reqDevice), gpuSet, builder.kubeClient)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, err.Error())
		return
	}

	err = builder.k8sHelper.CreateService(builder.Config.OpenHydraNamespace, reqDevice.Spec.OpenHydraUsername, reqDevice.Spec.IDEType, builder.kubeClient)
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

func (builder *OpenHydraRouteBuilder) GetCpu(postDevice xDeviceV1.Device) string {
	if postDevice.Spec.DeviceCpu != "" {
		return fmt.Sprintf("%sm", postDevice.Spec.DeviceCpu)
	}
	return fmt.Sprintf("%dm", builder.Config.DefaultCpuPerDevice)
}

func (builder *OpenHydraRouteBuilder) GetRam(postDevice xDeviceV1.Device) string {
	if postDevice.Spec.DeviceRam != "" {
		return fmt.Sprintf("%sMi", postDevice.Spec.DeviceRam)
	}
	return fmt.Sprintf("%dMi", builder.Config.DefaultRamPerDevice)
}

func (builder *OpenHydraRouteBuilder) BuildVolumes(postDevice xDeviceV1.Device) []envApi.VolumeMount {
	result := []envApi.VolumeMount{
		{
			Name:       "jupyter-lab",
			MountPath:  "/root/notebook", // so far image use /root/notebook as default path
			SourcePath: path.Join(builder.Config.JupyterLabHostBaseDir, postDevice.Spec.OpenHydraUsername),
		},
		{
			Name:      "public-dataset",
			MountPath: builder.Config.PublicDatasetStudentMountPath,
			// no bidirectional here so vfs will copy file from host to container, so no file concurrency issue here
			SourcePath: builder.Config.PublicDatasetBasePath,
		},
		{
			Name:      "public-course",
			MountPath: builder.Config.PublicCourseStudentMountPath,
			// no bidirectional here so vfs will copy file from host to container, so no file concurrency issue here
			SourcePath: builder.Config.PublicCourseBasePath,
		},
	}

	// if device is vscode device mount vscode workspace
	if postDevice.Spec.IDEType == k8s.OpenHydraIDELabelVSCode {
		result = append(result, envApi.VolumeMount{
			Name:       "public-vscode",
			MountPath:  builder.Config.VSCodeWorkspaceMountPath,
			SourcePath: path.Join(builder.Config.PublicVSCodeBasePath, postDevice.Spec.OpenHydraUsername),
		})
	}

	return result
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
