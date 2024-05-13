package openhydra

import (
	"encoding/json"
	"fmt"
	"net/http"
	xSetting "open-hydra/pkg/apis/open-hydra-api/setting/core/v1"
	"open-hydra/pkg/open-hydra/apis"

	"open-hydra/pkg/util"

	"github.com/emicklei/go-restful/v3"
)

func (builder *OpenHydraRouteBuilder) AddGetSettingRoute() {
	path := "/" + SettingPath + "/{name}"
	builder.addPathAuthorization(path, http.MethodGet, 1)
	builder.RootWS.Route(builder.RootWS.GET(path).Operation("getSetting").To(builder.GetSettingRouteHandler).
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusBadRequest, "bad request", "").
		Returns(http.StatusOK, "OK", xSetting.Setting{}))
}

func (builder *OpenHydraRouteBuilder) GetSettingRouteHandler(request *restful.Request, response *restful.Response) {
	result := xSetting.Setting{}
	util.FillKindAndApiVersion(&result.TypeMeta, SettingKind)
	result.Name = request.PathParameter("name")
	result.Spec = xSetting.SettingSpec{}
	result.Spec.DefaultGpuPerDevice = builder.Config.DefaultGpuPerDevice
	// now get all plugins from configmap
	cm, err := builder.k8sHelper.GetMap("openhydra-plugin", builder.Config.OpenHydraNamespace, builder.kubeClient)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get configmap: %v", err))
		return
	}
	jsonData := cm.Data["plugins"]
	var plugins apis.PluginList
	err = json.Unmarshal([]byte(jsonData), &plugins)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to unmarshal json: %v", err))
		return
	}
	result.Spec.PluginList = plugins
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (builder *OpenHydraRouteBuilder) AddUpdateSettingRoute() {
	path := "/" + SettingPath + "/{name}"
	builder.addPathAuthorization(path, http.MethodPut, 1)
	builder.RootWS.Route(builder.RootWS.PUT(path).Operation("createSetting").To(builder.UpdateSettingRouteHandler).
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusBadRequest, "bad request", "").
		Returns(http.StatusOK, "OK", xSetting.Setting{}))
}

func (builder *OpenHydraRouteBuilder) UpdateSettingRouteHandler(request *restful.Request, response *restful.Response) {
	setting := xSetting.Setting{}
	err := request.ReadEntity(&setting)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, fmt.Sprintf("Failed to read request entity: %v", err))
		return
	}
	setting.Name = request.PathParameter("name")
	util.FillKindAndApiVersion(&setting.TypeMeta, SettingKind)
	builder.Config.DefaultGpuPerDevice = setting.Spec.DefaultGpuPerDevice
	response.WriteHeaderAndEntity(http.StatusOK, setting)
}
