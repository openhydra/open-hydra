package apiserver

import (
	"fmt"
	"log/slog"
	"net/http"

	"open-hydra/cmd/open-hydra-server/app/config"
	"open-hydra/cmd/open-hydra-server/app/option"
	"open-hydra/pkg/database"
	openHydraHandler "open-hydra/pkg/open-hydra"
	openHydraK8s "open-hydra/pkg/open-hydra/k8s"

	"github.com/emicklei/go-restful/v3"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	genericApiServer "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/kubernetes"
)

const (
	httpStatusNotFoundMessage = "Not Found"
)

func registerDiscoveryService(apiServer *genericApiServer.GenericAPIServer) {
	ws := getWebService()
	ws.Path("/apis").Consumes(restful.MIME_JSON, restful.MIME_XML).Produces(restful.MIME_JSON, restful.MIME_XML)
	ws.Route(ws.GET("/").Operation("getApis").To(func(request *restful.Request, response *restful.Response) {
		apiGroup := metaV1.APIGroup{
			Name: option.GroupVersion.Group,
			PreferredVersion: metaV1.GroupVersionForDiscovery{
				GroupVersion: option.GroupVersion.Group + "/" + option.GroupVersion.Version,
				Version:      option.GroupVersion.Version,
			},
		}
		apiGroup.Versions = append(apiGroup.Versions, metaV1.GroupVersionForDiscovery{
			GroupVersion: option.GroupVersion.Group + "/" + option.GroupVersion.Version,
			Version:      option.GroupVersion.Version,
		})
		apiGroup.ServerAddressByClientCIDRs = append(apiGroup.ServerAddressByClientCIDRs, metaV1.ServerAddressByClientCIDR{
			ClientCIDR:    "0.0.0.0/0",
			ServerAddress: "",
		})
		apiGroup.Kind = "APIGroup"
		response.WriteAsJson(apiGroup)
	}).Returns(http.StatusOK, "OK", metaV1.APIGroup{}).Returns(http.StatusNotFound, httpStatusNotFoundMessage, ""))
	apiServer.Handler.GoRestfulContainer.Add(ws)
}

func registerApiResource(apiServer *genericApiServer.GenericAPIServer, config *config.OpenHydraServerConfig) error {
	ws := getWebService()
	resourceIndexPathTemplate := "/apis/%s/%s"
	resourceIndexPath := fmt.Sprintf(resourceIndexPathTemplate, option.GroupVersion.Group, option.GroupVersion.Version)
	ws.Path(resourceIndexPath).Consumes(restful.MIME_JSON, restful.MIME_XML).Produces(restful.MIME_JSON, restful.MIME_XML)
	ws.Route(ws.GET("/").Operation("getAllKinds").To(func(request *restful.Request, response *restful.Response) {
		list := &metaV1.APIResourceList{}

		list.Kind = "APIResourceList"
		list.GroupVersion = option.GroupVersion.Group + "/" + option.GroupVersion.Version
		list.APIVersion = option.GroupVersion.Version
		list.APIResources = openHydraHandler.ApiResources()
		response.WriteAsJson(list)
	}).Returns(http.StatusOK, "OK", metaV1.APIResource{}).Returns(http.StatusNotFound, httpStatusNotFoundMessage, ""))

	var db database.IDataBase
	switch config.DBType {
	case "mysql":
		db = database.NewMysql(config)
	case "etcd":
		db = &database.Etcd{Config: config}
	default:
		return fmt.Errorf("unknown db type %s", config.DBType)
	}

	err := db.InitDb()
	if err != nil {
		slog.Error("Failed to init db", err)
		return err
	}

	kubeClient, err := kubernetes.NewForConfig(config.KubeConfig)
	if err != nil {
		slog.Error("Failed to create kube client", err)
		return err
	}

	RBuilder := openHydraHandler.NewOpenHydraRouteBuilder(db, config, ws, kubeClient, openHydraK8s.NewDefaultK8sHelper())
	RBuilder.AddXUserListRoute()
	RBuilder.AddXUserCreateRoute()
	RBuilder.AddXUserGetRoute()
	RBuilder.AddXUserUpdateRoute()
	RBuilder.AddXUserDeleteRoute()
	RBuilder.AddDeviceListRoute()
	RBuilder.AddDeviceCreateRoute()
	RBuilder.AddDeviceGetRoute()
	RBuilder.AddDeviceUpdateRoute()
	RBuilder.AddDeviceDeleteRoute()
	RBuilder.AddSummaryGetRoute()
	RBuilder.AddDatasetListRoute()
	RBuilder.AddDatasetCreateRoute()
	RBuilder.AddDatasetGetRoute()
	RBuilder.AddDatasetUpdateRoute()
	RBuilder.AddDatasetDeleteRoute()
	RBuilder.AddXUserLoginRoute()
	RBuilder.AddGetSettingRoute()
	RBuilder.AddUpdateSettingRoute()
	RBuilder.AddCourseListRoute()
	RBuilder.AddCourseGetRoute()
	RBuilder.AddCourseCreateRoute()
	RBuilder.AddCourseUpdateRoute()
	RBuilder.AddCourseDeleteRoute()
	if !config.DisableAuth {
		ws.Filter(RBuilder.Filter)
	}
	apiServer.Handler.GoRestfulContainer.Add(ws)
	return nil
}

func getWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/apis")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON, restful.MIME_XML)
	ws.ApiVersion(option.GroupVersion.Group)
	return ws
}
