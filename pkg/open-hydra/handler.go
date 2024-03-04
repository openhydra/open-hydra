package openhydra

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"open-hydra/cmd/open-hydra-server/app/config"
	"open-hydra/cmd/open-hydra-server/app/option"
	xDeviceV1 "open-hydra/pkg/apis/open-hydra-api/device/core/v1"
	xUserV1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"
	"open-hydra/pkg/database"
	openHydraK8s "open-hydra/pkg/open-hydra/k8s"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/client-go/kubernetes"
)

const (
	openHydraHeaderUser       = "Open-Hydra-User"
	openHydraHeaderRole       = "Open-Hydra-Role"
	openHydraAuthStringHeader = "Open-Hydra-Auth"
)

type CacheDevices map[string]*xDeviceV1.Device

type OpenHydraRouteBuilder struct {
	Database database.IDataBase
	Config   *config.OpenHydraServerConfig
	RootWS   *restful.WebService
	// this is local cache relation between user and device
	CacheDevices     CacheDevices
	kubeClient       *kubernetes.Clientset
	k8sHelper        openHydraK8s.IOpenHydraK8sHelper
	authorizationMap map[string]map[string]int
}

func NewOpenHydraRouteBuilder(db database.IDataBase, config *config.OpenHydraServerConfig, rootWS *restful.WebService, client *kubernetes.Clientset, k8sHelper openHydraK8s.IOpenHydraK8sHelper) *OpenHydraRouteBuilder {
	return &OpenHydraRouteBuilder{
		Database:         db,
		Config:           config,
		RootWS:           rootWS,
		CacheDevices:     map[string]*xDeviceV1.Device{},
		kubeClient:       client,
		authorizationMap: make(map[string]map[string]int),
		k8sHelper:        k8sHelper,
	}
}

func (builder *OpenHydraRouteBuilder) Filter(r1 *restful.Request, r2 *restful.Response, fc *restful.FilterChain) {
	// here you can put your authentication and authorization logic
	if builder.AuthAndAuthorization(r1, r2) {
		fc.ProcessFilter(r1, r2)
	}
}

func (builder *OpenHydraRouteBuilder) AuthAndAuthorization(r1 *restful.Request, r2 *restful.Response) bool {
	if strings.HasPrefix(r1.Request.URL.Path, fmt.Sprintf("/apis/%s/v1/%s/login/", option.GroupVersion.Group, OpenHydraUserPath)) {
		// login does not need authentication and authorization
		return true
	}

	switch r1.Request.URL.Path {
	case "/apis", "/apis/", fmt.Sprintf("/apis/%s", option.GroupVersion.Group), fmt.Sprintf("/apis/%s/", option.GroupVersion.Group), fmt.Sprintf("/apis/%s/v1", option.GroupVersion.Group), fmt.Sprintf("/apis/%s/v1/", option.GroupVersion.Group):
		slog.Info(fmt.Sprintf("skip authentication and authorization for path: %s", r1.Request.URL.Path))
		return true
	}

	basicAuth := r1.Request.Header.Get(openHydraAuthStringHeader)
	if basicAuth == "" {
		writeHttpResponseAndLogError(r2, http.StatusUnauthorized, fmt.Sprintf("no auth header found for path: %s", r1.Request.URL.Path))
		return false
	}

	authTypeAndValue := strings.Split(basicAuth, " ")
	if len(authTypeAndValue) != 2 {
		writeHttpResponseAndLogError(r2, http.StatusUnauthorized, "format is not recognized")
		return false
	}

	if authTypeAndValue[0] != "Bearer" {
		writeHttpResponseAndLogError(r2, http.StatusUnauthorized, "only support Bearer")
		return false
	}

	credSet, err := base64.StdEncoding.DecodeString(authTypeAndValue[1])
	if err != nil {
		writeHttpResponseAndLogError(r2, http.StatusUnauthorized, "decode base64 failed")
		return false
	}

	userAndPass := strings.Split(string(credSet), ":")
	if len(userAndPass) != 2 {
		writeHttpResponseAndLogError(r2, http.StatusUnauthorized, "auth format is not recognized")
		return false
	}

	user, err := builder.Database.LoginUser(userAndPass[0], userAndPass[1])
	if err != nil {
		writeHttpResponseAndLogError(r2, http.StatusInternalServerError, "login failed")
		return false
	}

	if !builder.authorization(r1, user) {
		writeHttpResponseAndLogError(r2, http.StatusForbidden, fmt.Sprintf("user: %s do not have the right to access path: %s", user.Name, r1.Request.URL.Path))
		return false
	} else {
		r1.Request.Header.Set(openHydraHeaderUser, user.Name)
		r1.Request.Header.Set(openHydraHeaderRole, fmt.Sprintf("%d", user.Spec.Role))
	}

	return true
}

func (builder *OpenHydraRouteBuilder) authorization(r1 *restful.Request, user *xUserV1.OpenHydraUser) bool {
	relPath := strings.ReplaceAll(r1.SelectedRoutePath(), fmt.Sprintf("/apis/%s/v1", option.GroupVersion.Group), "")
	if _, found := builder.authorizationMap[relPath]; !found {
		slog.Warn(fmt.Sprintf("no authorization found for path: %s", relPath))
		return false
	}

	if _, found := builder.authorizationMap[relPath][r1.Request.Method]; !found {
		slog.Warn(fmt.Sprintf("no authorization found for path: %s with http method: %s", relPath, r1.Request.Method))
		return false
	}

	// simply xor it
	// user.role=1 | route.required=3 | 1 == user.role that saids it gets right to access the route
	return user.Spec.Role == builder.authorizationMap[relPath][r1.Request.Method]&user.Spec.Role
}

func (builder *OpenHydraRouteBuilder) addPathAuthorization(relPath, httpMethod string, requiredRole int) {
	if _, found := builder.authorizationMap[relPath]; !found {
		builder.authorizationMap[relPath] = make(map[string]int)
		builder.authorizationMap[relPath][httpMethod] = requiredRole
	} else {
		builder.authorizationMap[relPath][httpMethod] = requiredRole
	}
}
