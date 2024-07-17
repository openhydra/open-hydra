package openhydra

import (
	"fmt"
	"net/http"
	xSetting "open-hydra/pkg/apis/open-hydra-api/setting/core/v1"

	"open-hydra/pkg/util"

	"github.com/emicklei/go-restful/v3"
	"gopkg.in/yaml.v2"
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
	serverConfig, err := builder.GetServerConfigFromConfigMap()
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError,
			fmt.Sprintf("Failed to get server config: %v", err))
		return
	}

	result := xSetting.Setting{}
	util.FillKindAndApiVersion(&result.TypeMeta, SettingKind)
	result.Name = request.PathParameter("name")
	result.Spec = xSetting.SettingSpec{}
	result.Spec.DefaultGpuPerDevice = serverConfig.DefaultGpuPerDevice
	// now get all plugins from configmap
	cm, err := builder.k8sHelper.GetConfigMap("openhydra-plugin", OpenhydraNamespace)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get configmap: %v", err))
		return
	}

	plugins, err := ParseJsonToPluginList(cm.Data["plugins"])
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

	serverConfig, err := builder.GetServerConfigFromConfigMap()
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError,
			fmt.Sprintf("Failed to get server config: %v", err))
		return
	}

	util.FillKindAndApiVersion(&setting.TypeMeta, SettingKind)
	serverConfig.DefaultGpuPerDevice = setting.Spec.DefaultGpuPerDevice

	configJson, err := yaml.Marshal(serverConfig)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to marshal yaml: %v", err))
		return
	}

	err = builder.k8sHelper.UpdateConfigMap("open-hydra-config", OpenhydraNamespace, map[string]string{"config.yaml": string(configJson)})
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to update configmap: %v", err))
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, setting)
}
