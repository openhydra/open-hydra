package openhydra

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/emicklei/go-restful/v3"

	xUserV1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"
	"open-hydra/pkg/open-hydra/k8s"
)

/*
 * OpenHydraRouteBuilder is a helper class to build routes for OpenHydra api
 * we build all user related route and handle route action here
 */

func (builder *OpenHydraRouteBuilder) AddXUserLoginRoute() {
	builder.RootWS.Route(builder.RootWS.POST("/"+OpenHydraUserPath+"/login/{name}").Operation("createLogin").To(builder.XUserLoginRouteHandler).
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusBadRequest, "bad request", "").
		Returns(http.StatusOK, "OK", xUserV1.OpenHydraUser{}))
}

func (builder *OpenHydraRouteBuilder) XUserLoginRouteHandler(request *restful.Request, response *restful.Response) {
	xUser := xUserV1.OpenHydraUser{}
	err := request.ReadEntity(&xUser)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, fmt.Sprintf("Failed to read request entity: %v", err))
		return
	}

	user, err := builder.Database.LoginUser(xUser.Name, xUser.Spec.Password)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to login user: %v", err))
		return
	}
	response.WriteEntity(user)
}

func (builder *OpenHydraRouteBuilder) AddXUserListRoute() {
	// only teacher can list all users
	path := "/" + OpenHydraUserPath
	builder.addPathAuthorization(path, http.MethodGet, 1)
	builder.RootWS.Route(builder.RootWS.GET(path).Operation("listUser").To(builder.XUserListRouteHandler).
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", "").
		Returns(http.StatusOK, "OK", xUserV1.OpenHydraUserList{}))
}

func (builder *OpenHydraRouteBuilder) XUserListRouteHandler(request *restful.Request, response *restful.Response) {
	xUserList, err := builder.Database.ListUsers()
	if err != nil {
		// do not return database related error to client
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, "Failed to list users")
		return
	}
	xUserList.Kind = "List"
	xUserList.APIVersion = "v1"
	response.WriteEntity(xUserList)
}

func (builder *OpenHydraRouteBuilder) AddXUserCreateRoute() {
	path := "/" + OpenHydraUserPath
	builder.addPathAuthorization(path, http.MethodPost, 1)
	builder.RootWS.Route(builder.RootWS.POST(path).Operation("createUser").To(builder.XUserCreateRouteHandler).
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", "").
		Returns(http.StatusCreated, "OK", xUserV1.OpenHydraUser{}))
}

func (builder *OpenHydraRouteBuilder) XUserCreateRouteHandler(request *restful.Request, response *restful.Response) {
	xUser := xUserV1.OpenHydraUser{}
	err := request.ReadEntity(&xUser)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, fmt.Sprintf("Failed to read request entity: %v", err))
		return
	}
	err = builder.Database.CreateUser(&xUser)
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, xUser)
}

func (builder *OpenHydraRouteBuilder) AddXUserGetRoute() {
	// all user can get their own info
	// 1 | 2 = 3
	path := "/" + OpenHydraUserPath + "/{name}"
	builder.addPathAuthorization(path, http.MethodGet, 3)
	builder.RootWS.Route(builder.RootWS.GET(path).Operation("getUser").To(builder.XUserGetRouteHandler).
		Returns(http.StatusNotFound, "not found", "").
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", "").
		Returns(http.StatusOK, "OK", xUserV1.OpenHydraUser{}))
}

func (builder *OpenHydraRouteBuilder) XUserGetRouteHandler(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	if !builder.Config.DisableAuth {
		reqUser := request.HeaderParameter(openHydraHeaderUser)
		reqRole := request.HeaderParameter(openHydraHeaderRole)
		if reqUser == "" || reqRole == "" {
			writeHttpResponseAndLogError(response, http.StatusUnauthorized, "no user or role found in request header")
			return
		}
		if reqRole != "1" {
			// only teacher can get other user info
			if name != reqUser {
				writeHttpResponseAndLogError(response, http.StatusForbidden, fmt.Sprintf("user: %s do not have the right to access user: %s", reqUser, name))
				return
			}
		}
	}

	xUser, err := builder.Database.GetUser(name)
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}
	response.WriteEntity(xUser)
}

// i don't know it is joke or not that 'updateUser' is not working here
// so i have to put something funny here to make it work 'getUpdateUser'
func (builder *OpenHydraRouteBuilder) AddXUserUpdateRoute() {
	path := "/" + OpenHydraUserPath + "/{name}"
	builder.addPathAuthorization(path, http.MethodPut, 1)
	builder.RootWS.Route(builder.RootWS.PUT(path).Operation("getUpdateUser").To(builder.XUserUpdateRouteHandler).
		Returns(http.StatusNotFound, "not found", "").
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", "").
		Returns(http.StatusOK, "OK", xUserV1.OpenHydraUser{}))
}

func (builder *OpenHydraRouteBuilder) AddXUserPatchRoute() {
	path := "/" + OpenHydraUserPath + "/{name}"
	builder.addPathAuthorization(path, http.MethodPatch, 1)
	// for kubectl apply we need to add application/merge-patch+json as acceptable content type
	builder.RootWS.Route(builder.RootWS.PATCH(path).Operation("getPatchUser").
		Consumes(restful.MIME_JSON, restful.MIME_XML, "application/merge-patch+json").
		To(builder.XUserUpdateRouteHandler).
		Returns(http.StatusNotFound, "not found", "").
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", "").
		Returns(http.StatusOK, "OK", xUserV1.OpenHydraUser{}))
}

func (builder *OpenHydraRouteBuilder) XUserUpdateRouteHandler(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	xUser := xUserV1.OpenHydraUser{}
	var err error
	// for patch from kubectl we may go with application/merge-patch+json
	if request.Request.Header.Get("Content-Type") == "application/merge-patch+json" {
		mergePatchJsonEntityReader := restful.NewEntityAccessorJSON("application/merge-patch+json")
		err = mergePatchJsonEntityReader.Read(request, &xUser)
	} else {
		// for header with application/json or application/xml
		err = request.ReadEntity(&xUser)
	}
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, fmt.Sprintf("Failed to read request entity: %v", err))
		return
	}
	oldUser, err := builder.Database.GetUser(name)
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}
	if oldUser.Spec == xUser.Spec {
		response.WriteHeader(http.StatusOK)
		return
	}
	err = builder.Database.UpdateUser(&xUser)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, "Failed to update user")
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (builder *OpenHydraRouteBuilder) AddXUserDeleteRoute() {
	path := "/" + OpenHydraUserPath + "/{name}"
	builder.addPathAuthorization(path, http.MethodDelete, 1)
	builder.RootWS.Route(builder.RootWS.DELETE(path).Operation("deleteUser").To(builder.XUserDeleteRouteHandler).
		Returns(http.StatusNotFound, "not found", "").
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", "").
		Returns(http.StatusOK, "OK", xUserV1.OpenHydraUser{}))
}

func (builder *OpenHydraRouteBuilder) XUserDeleteRouteHandler(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("name")
	oldUser, err := builder.Database.GetUser(username)
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}
	err = builder.Database.DeleteUser(username)
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}

	slog.Info(fmt.Sprintf("one shot attempting to delete related k8s resource for user: %s", username))
	_ = builder.k8sHelper.DeleteUserDeployment(fmt.Sprintf("%s=%s", k8s.OpenHydraUserLabelKey, username), builder.Config.OpenHydraNamespace, builder.kubeClient)
	_ = builder.k8sHelper.DeleteUserService(fmt.Sprintf("%s=%s", k8s.OpenHydraUserLabelKey, username), builder.Config.OpenHydraNamespace, builder.kubeClient)
	if builder.Config.PatchResourceNotRelease {
		_ = builder.k8sHelper.DeleteUserReplicaSet(fmt.Sprintf("%s=%s", k8s.OpenHydraUserLabelKey, username), builder.Config.OpenHydraNamespace, builder.kubeClient)
		_ = builder.k8sHelper.DeleteUserPod(fmt.Sprintf("%s=%s", k8s.OpenHydraUserLabelKey, username), builder.Config.OpenHydraNamespace, builder.kubeClient)
	}

	response.WriteEntity(oldUser)
}
