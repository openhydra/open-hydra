package openhydra

import (
	stdErr "errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xCourseV1 "open-hydra/pkg/apis/open-hydra-api/course/core/v1"

	"github.com/emicklei/go-restful/v3"
)

func (builder *OpenHydraRouteBuilder) AddCourseListRoute() {
	path := "/" + CoursePath
	builder.addPathAuthorization(path, http.MethodGet, 1)
	builder.RootWS.Route(builder.RootWS.GET(path).Operation("listCourse").To(builder.CourseListRouteHandler).
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", "").
		Returns(http.StatusOK, "OK", xCourseV1.CourseList{}))
}

func (builder *OpenHydraRouteBuilder) CourseListRouteHandler(request *restful.Request, response *restful.Response) {
	courseList, err := builder.Database.ListCourses()
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}
	courseList.Kind = "List"
	courseList.APIVersion = "v1"
	response.WriteEntity(courseList)
}

func (builder *OpenHydraRouteBuilder) AddCourseGetRoute() {
	path := "/" + CoursePath + "/{course-name}"
	builder.addPathAuthorization(path, http.MethodGet, 1)
	builder.RootWS.Route(builder.RootWS.GET(path).Operation("getCourse").To(builder.CourseGetRouteHandler).
		Returns(http.StatusNotFound, "not found", "").
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusUnauthorized, "unauthorized", "").
		Returns(http.StatusOK, "OK", xCourseV1.Course{}))
}

func (builder *OpenHydraRouteBuilder) CourseGetRouteHandler(request *restful.Request, response *restful.Response) {
	courseName := request.PathParameter("course-name")
	course, err := builder.Database.GetCourse(courseName)
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}
	response.WriteEntity(course)
}

func (builder *OpenHydraRouteBuilder) AddCourseCreateRoute() {
	path := "/" + CoursePath
	builder.addPathAuthorization(path, http.MethodPost, 1)
	builder.RootWS.Route(builder.RootWS.POST(path).Operation("createCourse").To(builder.CourseCreateRouteHandler).
		Returns(http.StatusCreated, "created", "").
		Returns(http.StatusBadRequest, "bad request", "").
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", "").
		Consumes("multipart/form-data"))
}

func (builder *OpenHydraRouteBuilder) CourseCreateRouteHandler(request *restful.Request, response *restful.Response) {

	serverConfig, err := builder.GetServerConfigFromConfigMap()
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError,
			fmt.Sprintf("Failed to get server config: %v", err))
		return
	}

	err = request.Request.ParseMultipartForm(serverConfig.PublicCourseMaxSize)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusBadRequest,
			fmt.Sprintf("Failed to parse multipart form: %v", err))
		return
	}
	description := request.Request.PostFormValue("description")
	createdBy := request.Request.PostFormValue("createdBy")
	name := request.Request.PostFormValue("name")
	levelRaw := request.Request.PostFormValue("level")
	sandboxName := request.Request.PostFormValue("sandboxName")
	if name == "" {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, "Course name is empty")
		return
	}

	// we need to get config map openhydra-plugin first
	// TODO: we should use informer to cache config map instead of query api-server directly for performance
	pluginConfigMap, err := builder.k8sHelper.GetConfigMap("openhydra-plugin", OpenhydraNamespace)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get config map: %v", err))
		return
	}

	// parse to plugin list
	plugins, err := ParseJsonToPluginList(pluginConfigMap.Data["plugins"])
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to unmarshal json: %v", err))
		return
	}

	if sandboxName == "" {
		sandboxName = plugins.DefaultSandbox
	} else {
		if _, found := plugins.Sandboxes[sandboxName]; !found {
			writeHttpResponseAndLogError(response, http.StatusBadRequest, fmt.Sprintf("Sandbox %s not found", sandboxName))
			return
		}
	}

	level := 0
	if levelRaw != "" {
		// try parse to int
		level, err = strconv.Atoi(levelRaw)
		if err != nil {
			writeHttpResponseAndLogError(response, http.StatusBadRequest,
				fmt.Sprintf("Failed to parse level: %v", err.Error()))
			return
		}
	}

	file, fileHeader, err := request.Request.FormFile("file")
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusBadRequest,
			err.Error())
		return
	}
	defer file.Close()
	if strings.ToLower(filepath.Ext(fileHeader.Filename)) != ".zip" {
		writeHttpResponseAndLogError(response, http.StatusBadRequest,
			"Only zip file is supported")
		return
	}
	_, err = builder.Database.GetCourse(name)
	if err == nil {
		writeHttpResponseAndLogError(response, http.StatusConflict,
			fmt.Sprintf("Course %s already exists", name))
		return
	}
	if !errors.IsNotFound(err) {
		writeAPIStatusError(response, err)
		return
	}

	coursePath, err := filepath.Abs(filepath.Join(serverConfig.PublicCourseBasePath, name))
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusBadRequest,
			fmt.Sprintf("Failed to get course path: %v", err.Error()))
		return
	}
	err = unzipTo(file, fileHeader.Size, coursePath)
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}
	course := &xCourseV1.Course{
		ObjectMeta: metaV1.ObjectMeta{
			Name: name,
		},
		Spec: xCourseV1.CourseSpec{
			Description: description,
			CreatedBy:   createdBy,
			Level:       level,
			SandboxName: sandboxName,
		},
	}

	err = builder.Database.CreateCourse(course)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError,
			fmt.Sprintf("Failed create course: %v", err.Error()))
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, course)
}

func (builder *OpenHydraRouteBuilder) AddCourseUpdateRoute() {
	path := "/" + CoursePath + "/{course-name}"
	builder.addPathAuthorization(path, http.MethodPut, 1)
	builder.RootWS.Route(builder.RootWS.PUT(path).Operation("getUpdateCourse").To(builder.CourseUpdateRouteHandler).
		Returns(http.StatusNotFound, "not found", "").
		Returns(http.StatusBadRequest, "bad request", "").
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", "").
		Consumes("multipart/form-data"))
}

func (builder *OpenHydraRouteBuilder) CourseUpdateRouteHandler(request *restful.Request, response *restful.Response) {

	serverConfig, err := builder.GetServerConfigFromConfigMap()
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError,
			fmt.Sprintf("Failed to get server config: %v", err))
		return
	}

	courseName := request.PathParameter("course-name")
	err = request.Request.ParseMultipartForm(serverConfig.PublicCourseMaxSize)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusBadRequest,
			fmt.Sprintf("Failed to parse multipart form: %v", err))
		return
	}
	file, fileHeader, err := request.Request.FormFile("file")
	if err != nil {
		if stdErr.Is(err, http.ErrMissingFile) {
			response.WriteHeader(http.StatusOK)
			return
		}
		writeHttpResponseAndLogError(response, http.StatusBadRequest,
			err.Error())
		return
	}
	defer file.Close()
	if strings.ToLower(filepath.Ext(fileHeader.Filename)) != ".zip" {
		writeHttpResponseAndLogError(response, http.StatusBadRequest,
			"Only zip file is supported")
		return
	}

	description := request.Request.PostFormValue("description")
	course, err := builder.Database.GetCourse(courseName)
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}
	course.Spec.Description = description
	err = builder.Database.UpdateCourse(course)
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}

	// update course file

	coursePath, err := filepath.Abs(filepath.Join(serverConfig.PublicCourseBasePath, course.Name))
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusBadRequest,
			fmt.Sprintf("Failed to get course path: %v", err.Error()))
		return
	}
	err = unzipTo(file, fileHeader.Size, coursePath)
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (builder *OpenHydraRouteBuilder) AddCourseDeleteRoute() {
	path := "/" + CoursePath + "/{course-name}"
	builder.addPathAuthorization(path, http.MethodDelete, 1)
	builder.RootWS.Route(builder.RootWS.DELETE(path).Operation("deleteCourse").To(builder.CourseDeleteRouteHandler).
		Returns(http.StatusNotFound, "not found", "").
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", ""))
}

func (builder *OpenHydraRouteBuilder) CourseDeleteRouteHandler(request *restful.Request, response *restful.Response) {

	serverConfig, err := builder.GetServerConfigFromConfigMap()
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError,
			fmt.Sprintf("Failed to get server config: %v", err))
		return
	}

	courseName := request.PathParameter("course-name")
	err = builder.Database.DeleteCourse(courseName)
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}
	coursePath, err := filepath.Abs(filepath.Join(serverConfig.PublicCourseBasePath, courseName))
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError,
			fmt.Sprintf("Failed to get course path: %v", err.Error()))
	}
	err = os.RemoveAll(coursePath)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError,
			fmt.Sprintf("Failed to remove course file: %v", err.Error()))
		return
	}
	response.WriteHeader(http.StatusOK)
}
