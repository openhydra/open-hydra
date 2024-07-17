package apiserver

import (
	"fmt"
	"log/slog"
	"net"

	"open-hydra/cmd/open-hydra-server/app/config"
	"open-hydra/cmd/open-hydra-server/app/option"
	openHydraDeviceV1 "open-hydra/pkg/apis/open-hydra-api/device/core/v1"
	openHydraUserV1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"
	openHydraOpenApi "open-hydra/pkg/generated/apis/openapi"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilRuntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/endpoints/openapi"
	genericApiServer "k8s.io/apiserver/pkg/server"
)

func init() {
	utilRuntime.Must(openHydraUserV1.AddToScheme(option.Scheme))
	utilRuntime.Must(openHydraDeviceV1.AddToScheme(option.Scheme))

	// Setting VersionPriority is critical in the InstallAPIGroup call (done in New())
	utilRuntime.Must(option.Scheme.SetVersionPriority(option.GroupVersion))
	metaV1.AddToGroupVersion(option.Scheme, option.GroupVersion)

	// add a comment for ci test
	// TODO(devdattakulkarni) -- Following comments coming from sample-apiserver.
	// Leaving them for now.
	// TODO: keep the generic API server from wanting this
	unVersioned := schema.GroupVersion{Group: "", Version: "v1"}
	option.Scheme.AddUnversionedTypes(unVersioned,
		&metaV1.Status{},
		&metaV1.APIVersions{},
		&metaV1.APIGroupList{},
		&metaV1.APIGroup{},
		&metaV1.APIResourceList{},
	)

	// Start collecting provenance
	//go provenance.CollectProvenance()
}

func RunApiServer(recommendOption *option.Options, config *config.OpenHydraServerConfig, stopChan <-chan struct{}) error {
	if err := recommendOption.ApiServerOption.SecureServing.MaybeDefaultWithSelfSignedCerts("0.0.0.0", nil, []net.IP{net.ParseIP("0.0.0.0")}); err != nil {
		return err
	}
	recommendedConfig := genericApiServer.NewRecommendedConfig(option.Codecs)

	if err := recommendOption.ApiServerOption.ApplyTo(recommendedConfig); err != nil {
		return err
	}

	completedConfig := recommendedConfig.Complete()
	completedConfig.EnableDiscovery = false
	completedConfig.OpenAPIConfig = genericApiServer.DefaultOpenAPIConfig(openHydraOpenApi.GetOpenAPIDefinitions, openapi.NewDefinitionNamer(option.Scheme))
	completedConfig.OpenAPIConfig.Info.Title = "open-hydra-server.openhydra.io"
	completedConfig.OpenAPIConfig.Info.Version = "2"
	completedConfig.OpenAPIV3Config = genericApiServer.DefaultOpenAPIV3Config(openHydraOpenApi.GetOpenAPIDefinitions, openapi.NewDefinitionNamer(option.Scheme))
	completedConfig.OpenAPIV3Config.Info.Title = "open-hydra-server.openhydra.io"
	completedConfig.OpenAPIV3Config.Info.Version = "3"

	gApiServer, err := completedConfig.New("open-hydra api-server", genericApiServer.NewEmptyDelegate())
	if err != nil {
		slog.Error("Failed to create genericApiServer", "error", err)
		return err

	}

	installApiGroup := genericApiServer.NewDefaultAPIGroupInfo(option.GroupVersion.Group, option.Scheme, metaV1.ParameterCodec, option.Codecs)

	err = gApiServer.InstallAPIGroup(&installApiGroup)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to install API group: %v", err))
		return err
	}

	// register the discovery service
	registerDiscoveryService(gApiServer)
	// register the api resource
	err = registerApiResource(gApiServer, config, stopChan)
	if err != nil {
		slog.Error("Failed to register api resource", "error", err)
		return err
	}

	return gApiServer.PrepareRun().Run(stopChan)
}
