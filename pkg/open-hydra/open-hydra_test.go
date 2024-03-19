package openhydra

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"open-hydra/cmd/open-hydra-server/app/config"
	"open-hydra/cmd/open-hydra-server/app/option"
	xCourseV1 "open-hydra/pkg/apis/open-hydra-api/course/core/v1"
	xDataset "open-hydra/pkg/apis/open-hydra-api/dataset/core/v1"
	xDeviceV1 "open-hydra/pkg/apis/open-hydra-api/device/core/v1"
	xUserV1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"
	database "open-hydra/pkg/database"
	"open-hydra/pkg/open-hydra/k8s"
	"open-hydra/pkg/util"
	"os"
	"path"

	"net/http/httptest"

	xSetting "open-hydra/pkg/apis/open-hydra-api/setting/core/v1"

	"mime/multipart"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("open-hydra api-resource test", func() {
	list := []metaV1.APIResource{
		{
			Name:         OpenHydraUserPath,
			SingularName: "xuser",
			Namespaced:   false,
			Kind:         OpenHydraUserKind,
			Verbs:        metaV1.Verbs{"get", "list", "watch", "create", "update", "delete"},
		},
		{
			Name:         DevicePath,
			SingularName: "dev",
			Namespaced:   false,
			Kind:         DeviceKind,
			Verbs:        metaV1.Verbs{"get", "list", "watch", "create", "update", "delete"},
		},
		{
			Name:         SumUpPath,
			SingularName: "xsu",
			Namespaced:   false,
			Kind:         SumUpKind,
			Verbs:        metaV1.Verbs{"get"},
		},
		{
			Name:         DatasetPath,
			SingularName: "xds",
			Namespaced:   false,
			Kind:         DatasetKind,
			Verbs:        metaV1.Verbs{"get", "list", "watch", "create", "update", "delete"},
		},
		{
			Name:         SettingPath,
			SingularName: "setting",
			Namespaced:   false,
			Kind:         SettingKind,
			Verbs:        metaV1.Verbs{"get", "update"},
		},
		{
			Name:         CoursePath,
			SingularName: "course",
			Namespaced:   false,
			Kind:         CourseKind,
			Verbs:        metaV1.Verbs{"get", "list", "watch", "create", "update", "delete"},
		},
	}
	BeforeEach(func() {
	})

	Describe("api-resource result test", func() {
		It("should be expected", func() {
			Expect(ApiResources()).To(Equal(list))
		})

		AfterEach(func() {
		})
	})
})

var _ = Describe("open-hydra-server handler test", func() {
	var openHydraConfig *config.OpenHydraServerConfig
	var builder *OpenHydraRouteBuilder
	var device *xDeviceV1.Device
	BeforeEach(func() {
		openHydraConfig = config.DefaultConfig()
		openHydraConfig.DefaultGpuDriver = "nvidia.com/gpu"
		builder = NewOpenHydraRouteBuilder(nil, openHydraConfig, nil, nil, nil)
		device = &xDeviceV1.Device{
			TypeMeta: metaV1.TypeMeta{
				Kind: "Device",
			},
			ObjectMeta: metaV1.ObjectMeta{
				Name: "test",
			},
			Spec: xDeviceV1.DeviceSpec{
				DeviceName:      "test",
				DeviceNamespace: "test",
				DeviceType:      "test",
				DeviceIP:        "test",
				DeviceCpu:       "4000",
				DeviceRam:       "10240",
				GpuDriver:       "test",
				DeviceGpu:       0,
			},
		}
	})

	Describe("device handler test", func() {
		It("get cpu without over committing should be expected", func() {
			cpuReq, cpuLimit := builder.GetCpu(*device)
			Expect(cpuReq).To(Equal("4000m"))
			Expect(cpuLimit).To(Equal("4000m"))
			device.Spec.DeviceCpu = ""
			cpuReq, cpuLimit = builder.GetCpu(*device)
			Expect(cpuReq).To(Equal("2000m"))
			Expect(cpuLimit).To(Equal("2000m"))
		})

		It("get ram without over committing should be expected", func() {
			ramReq, ramLimit := builder.GetRam(*device)
			Expect(ramReq).To(Equal("10240Mi"))
			Expect(ramLimit).To(Equal("10240Mi"))
			device.Spec.DeviceRam = ""
			ramReq, ramLimit = builder.GetRam(*device)
			Expect(ramReq).To(Equal("8192Mi"))
			Expect(ramLimit).To(Equal("8192Mi"))
		})

		It("get cpu with over committing should be expected", func() {
			openHydraConfig.CpuOverCommitRate = 2
			cpuReq, cpuLimit := builder.GetCpu(*device)
			Expect(cpuReq).To(Equal("2000m"))
			Expect(cpuLimit).To(Equal("4000m"))
			device.Spec.DeviceCpu = ""
			cpuReq, cpuLimit = builder.GetCpu(*device)
			Expect(cpuReq).To(Equal("1000m"))
			Expect(cpuLimit).To(Equal("2000m"))
		})

		It("get ram with over committing should be expected", func() {
			openHydraConfig.MemoryOverCommitRate = 2
			ramReq, ramLimit := builder.GetRam(*device)
			Expect(ramReq).To(Equal("5120Mi"))
			Expect(ramLimit).To(Equal("10240Mi"))
			device.Spec.DeviceRam = ""
			ramReq, ramLimit = builder.GetRam(*device)
			Expect(ramReq).To(Equal("4096Mi"))
			Expect(ramLimit).To(Equal("8192Mi"))
		})

		It("combine cpu memory set with over commit should be expected", func() {
			openHydraConfig.CpuOverCommitRate = 2
			openHydraConfig.MemoryOverCommitRate = 2
			result := builder.CombineReqLimit(*device)
			Expect(result.CpuRequest).To(Equal("2000m"))
			Expect(result.CpuLimit).To(Equal("4000m"))
			Expect(result.MemoryRequest).To(Equal("5120Mi"))
			Expect(result.MemoryLimit).To(Equal("10240Mi"))
		})

		It("combine cpu memory set without over commit should be expected", func() {
			result := builder.CombineReqLimit(*device)
			Expect(result.CpuRequest).To(Equal("4000m"))
			Expect(result.CpuLimit).To(Equal("4000m"))
			Expect(result.MemoryRequest).To(Equal("10240Mi"))
			Expect(result.MemoryLimit).To(Equal("10240Mi"))
		})

		It("get volume should be expected", func() {
			volume := builder.BuildVolumes(*device)
			Expect(volume[0].Name).To(Equal("jupyter-lab"))
			Expect(volume[0].SourcePath).To(Equal(path.Join(openHydraConfig.JupyterLabHostBaseDir, device.Spec.OpenHydraUsername)))
			Expect(volume[0].MountPath).To(Equal("/root/notebook"))
			Expect(volume[1].Name).To(Equal("public-dataset"))
			Expect(volume[1].SourcePath).To(Equal(path.Join(openHydraConfig.PublicDatasetBasePath, device.Spec.OpenHydraUsername)))
			Expect(volume[1].MountPath).To(Equal(openHydraConfig.PublicDatasetStudentMountPath))
		})

		It("get gpu should be set 0", func() {
			gpu := builder.BuildGpu(*device)
			Expect(gpu.GpuDriverName).To(Equal("test"))
			Expect(gpu.Gpu).To(Equal(uint8(0)))
			device.Spec.GpuDriver = ""
			device.Spec.DeviceGpu = 1
			gpu = builder.BuildGpu(*device)
			Expect(gpu.GpuDriverName).To(Equal("nvidia.com/gpu"))
			Expect(gpu.Gpu).To(Equal(uint8(1)))
		})

		AfterEach(func() {
		})
	})
})

var _ = Describe("open-hydra-server combineDeviceList test", func() {
	var pods []coreV1.Pod
	var services []coreV1.Service
	var users xUserV1.OpenHydraUserList
	var openHydraConfig *config.OpenHydraServerConfig
	BeforeEach(func() {
		openHydraConfig = config.DefaultConfig()
		openHydraConfig.DefaultGpuDriver = "nvidia.com/gpu"
		pods = []coreV1.Pod{
			{
				ObjectMeta: metaV1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						k8s.OpenHydraWorkloadLabelKey: k8s.OpenHydraWorkloadLabelValue,
						k8s.OpenHydraUserLabelKey:     "user1",
					},
				},
				Spec: coreV1.PodSpec{
					Containers: []coreV1.Container{
						{
							Resources: coreV1.ResourceRequirements{
								Requests: coreV1.ResourceList{
									coreV1.ResourceCPU:    resource.MustParse("2000m"),
									coreV1.ResourceMemory: resource.MustParse("8192Mi"),
								},
							},
						},
					},
				},
			},
			{
				ObjectMeta: metaV1.ObjectMeta{
					Name: "test2",
					Labels: map[string]string{
						k8s.OpenHydraWorkloadLabelKey: k8s.OpenHydraWorkloadLabelValue,
						k8s.OpenHydraUserLabelKey:     "user2",
					},
				},
				Spec: coreV1.PodSpec{
					Containers: []coreV1.Container{
						{
							Resources: coreV1.ResourceRequirements{
								Requests: coreV1.ResourceList{
									coreV1.ResourceCPU:    resource.MustParse("2000m"),
									coreV1.ResourceMemory: resource.MustParse("8192Mi"),
									"nvidia.com/gpu":      resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
		}
		services = []coreV1.Service{
			{
				ObjectMeta: metaV1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						k8s.OpenHydraWorkloadLabelKey: k8s.OpenHydraWorkloadLabelValue,
						k8s.OpenHydraUserLabelKey:     "user1",
					},
				},
				Spec: coreV1.ServiceSpec{
					Ports: []coreV1.ServicePort{
						{
							Name:     "easy-train",
							NodePort: 5000,
						},
						{
							Name:     "lab",
							NodePort: 8888,
						},
					},
				},
			},
			{
				ObjectMeta: metaV1.ObjectMeta{
					Name: "test2",
					Labels: map[string]string{
						k8s.OpenHydraWorkloadLabelKey: k8s.OpenHydraWorkloadLabelValue,
						k8s.OpenHydraUserLabelKey:     "user2",
					},
				},
				Spec: coreV1.ServiceSpec{
					Ports: []coreV1.ServicePort{
						{
							Name:     "easy-train",
							NodePort: 5000,
						},
						{
							Name:     "lab",
							NodePort: 8888,
						},
					},
				},
			},
		}
		users = xUserV1.OpenHydraUserList{
			Items: []xUserV1.OpenHydraUser{
				{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "user1",
					},
				},
				{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "user2",
					},
				},
			},
		}
	})

	Describe("combineDeviceList result test", func() {
		It("should be expected", func() {
			devices := combineDeviceList(pods, services, users, openHydraConfig)
			Expect(devices[0].Spec.DeviceCpu).To(Equal("2"))
			Expect(devices[0].Spec.DeviceRam).To(Equal("8Gi"))
			Expect(devices[0].Spec.DeviceGpu).To(Equal(uint8(0)))
			Expect(devices[0].Spec.EasyTrainURL).To(Equal("http://localhost:5000"))
			Expect(devices[0].Spec.JupyterLabURL).To(Equal("http://localhost:8888"))
			Expect(devices[1].Spec.DeviceCpu).To(Equal("2"))
			Expect(devices[1].Spec.DeviceRam).To(Equal("8Gi"))
			Expect(devices[1].Spec.DeviceGpu).To(Equal(uint8(1)))
			Expect(devices[1].Spec.GpuDriver).To(Equal("nvidia.com/gpu"))
			Expect(devices[1].Spec.EasyTrainURL).To(Equal("http://localhost:5000"))
			Expect(devices[1].Spec.JupyterLabURL).To(Equal("http://localhost:8888"))
		})

		AfterEach(func() {
		})
	})
})

var _ = Describe("open-hydra-server authorization test", func() {
	var openHydraConfig *config.OpenHydraServerConfig
	var builder *OpenHydraRouteBuilder
	var teacher *xUserV1.OpenHydraUser
	var student *xUserV1.OpenHydraUser
	var newStudent, newTeacher *xUserV1.OpenHydraUser
	var fakeDb *database.Faker
	var container *restful.Container
	var req *restful.Request
	var device1, device2, device3 *xDeviceV1.Device
	var setting *xSetting.Setting
	var openHydraUsersURL = fmt.Sprintf("http://localhost/apis/%s/v1/%s", option.GroupVersion.Group, OpenHydraUserPath)
	var openHydraDevicesURL = fmt.Sprintf("http://localhost/apis/%s/v1/%s", option.GroupVersion.Group, DevicePath)
	var openHydraSettingsURL = fmt.Sprintf("http://localhost/apis/%s/v1/%s/default", option.GroupVersion.Group, SettingPath)
	var openHydraDatasetsURL = fmt.Sprintf("http://localhost/apis/%s/v1/%s", option.GroupVersion.Group, DatasetPath)
	var openHydraCoursesURL = fmt.Sprintf("http://localhost/apis/%s/v1/%s", option.GroupVersion.Group, CoursePath)
	var fakeService = func() *restful.WebService {
		ws := new(restful.WebService)
		ws.Path(fmt.Sprintf("/apis/%s/%s", option.GroupVersion.Group, option.GroupVersion.Version))
		ws.Consumes("*/*")
		ws.Produces(restful.MIME_JSON, restful.MIME_XML)
		ws.ApiVersion(option.GroupVersion.Group)
		return ws
	}
	var createFakeUser = func(name, password string, role int) *xUserV1.OpenHydraUser {
		user := &xUserV1.OpenHydraUser{
			ObjectMeta: metaV1.ObjectMeta{
				Name: name,
			},
			Spec: xUserV1.OpenHydraUserSpec{
				Role:     role,
				Password: password,
			},
		}
		return user
	}
	var createRequest = func(method, path string, header map[string][]string, body io.Reader) *restful.Request {
		iReq := httptest.NewRequest(method, path, body)
		iReq.Header = header
		iReq.Host = "localhost"
		r := restful.NewRequest(iReq)

		return r
	}
	var createResponse = func(writer http.ResponseWriter) *restful.Response {
		r := &restful.Response{
			ResponseWriter: writer,
		}
		return r
	}
	var createTokenValue = func(user *xUserV1.OpenHydraUser, additionalHeaders map[string]string) map[string][]string {
		result := fmt.Sprintf("%s:%s", user.Name, user.Spec.Password)
		// base64 encode it

		defaultHeader := map[string][]string{
			"Content-Type":            {"application/json"},
			openHydraAuthStringHeader: {fmt.Sprintf("Bearer %s", base64.StdEncoding.EncodeToString([]byte(result)))},
		}

		for k, v := range additionalHeaders {
			defaultHeader[k] = []string{v}
		}

		return defaultHeader
	}
	var createDevice = func(name string, gpu uint8) *xDeviceV1.Device {
		return &xDeviceV1.Device{
			ObjectMeta: metaV1.ObjectMeta{
				Name: name,
			},
			Spec: xDeviceV1.DeviceSpec{
				OpenHydraUsername: name,
				DeviceGpu:         gpu,
			},
		}
	}
	var callApi = func(method, path string, header map[string][]string, body io.Reader) (r1 *restful.Response, r2 *httptest.ResponseRecorder) {
		req = createRequest(method, path, header, body)
		httpResponse := httptest.NewRecorder()
		resp := createResponse(httpResponse)
		container.Dispatch(resp.ResponseWriter, req.Request)
		return resp, httpResponse
	}
	var createMultiPartBody = func(txtData map[string]string, filePath string) (io.Reader, string, error) {
		var (
			buf = new(bytes.Buffer)
			w   = multipart.NewWriter(buf)
		)

		for k, v := range txtData {
			_ = w.WriteField(k, v)
		}

		part, err := w.CreateFormFile("file", "test.zip")
		if err != nil {
			return nil, "", nil
		}

		data, err := util.ReadTxtFile(filePath)
		if err != nil {
			return nil, "", err
		}

		_, err = part.Write(data)
		if err != nil {
			return nil, "", err
		}

		w.Close()
		return buf, w.FormDataContentType(), nil
	}

	var initContainer = func() {
		container = restful.NewContainer()
		builder = NewOpenHydraRouteBuilder(fakeDb, openHydraConfig, fakeService(), nil, k8s.NewDefaultK8sHelperWithFake())
		builder.AddXUserListRoute()
		builder.AddXUserCreateRoute()
		builder.AddXUserGetRoute()
		builder.AddXUserUpdateRoute()
		builder.AddXUserDeleteRoute()
		builder.AddDeviceListRoute()
		builder.AddDeviceCreateRoute()
		builder.AddDeviceGetRoute()
		builder.AddDeviceUpdateRoute()
		builder.AddDeviceDeleteRoute()
		builder.AddSummaryGetRoute()
		builder.AddDatasetListRoute()
		builder.AddDatasetCreateRoute()
		builder.AddDatasetGetRoute()
		builder.AddDatasetUpdateRoute()
		builder.AddDatasetDeleteRoute()
		builder.AddXUserLoginRoute()
		builder.AddGetSettingRoute()
		builder.AddUpdateSettingRoute()
		builder.AddCourseListRoute()
		builder.AddCourseCreateRoute()
		builder.AddCourseGetRoute()
		builder.AddCourseUpdateRoute()
		builder.AddCourseDeleteRoute()
		if !openHydraConfig.DisableAuth {
			builder.RootWS.Filter(builder.Filter)
		}
		container.Add(builder.RootWS)
	}
	var uploadResource = func(url, testZipBaseDir string) {
		err := util.CreateDirIfNotExists("/tmp/test")
		Expect(err).To(BeNil())
		err = util.CreateDirIfNotExists("/tmp/test/test1")
		Expect(err).To(BeNil())
		err = util.WriteFileWithNosec("/tmp/test/test1/test.txt", []byte("test"))
		Expect(err).To(BeNil())
		err = util.WriteFileWithNosec("/tmp/test/test1/test2.txt", []byte("test"))
		Expect(err).To(BeNil())
		err = util.CreateDirIfNotExists("/tmp/test/sub1")
		Expect(err).To(BeNil())
		err = util.WriteFileWithNosec("/tmp/test/sub1/test.txt", []byte("test"))
		Expect(err).To(BeNil())
		err = util.ZipDir("/tmp/test", "/tmp/test.zip")
		Expect(err).To(BeNil())
		bodyTxt := map[string]string{}
		bodyTxt["name"] = "unit-test"
		bodyTxt["description"] = "unit-test"
		body, contentType, err := createMultiPartBody(bodyTxt, "/tmp/test.zip")
		//body, contentType, err := createMultiPartBody(bodyTxt, "/tmp/ds1.zip")
		Expect(err).To(BeNil())
		_, r2 := callApi(http.MethodPost, url, createTokenValue(teacher, map[string]string{"Content-Type": contentType}), body)
		Expect(r2.Code).To(Equal(http.StatusCreated))
		targetPath := path.Join(testZipBaseDir, "unit-test", "test", "test1", "test.txt")
		data, err := util.ReadTxtFile(targetPath)
		Expect(err).To(BeNil())
		Expect(string(data)).To(Equal("test"))
		targetPath = path.Join(testZipBaseDir, "unit-test", "test", "test1", "test2.txt")
		data, err = util.ReadTxtFile(targetPath)
		Expect(err).To(BeNil())
		Expect(string(data)).To(Equal("test"))
		targetPath = path.Join(testZipBaseDir, "unit-test", "test", "sub1", "test.txt")
		data, err = util.ReadTxtFile(targetPath)
		Expect(err).To(BeNil())
		Expect(string(data)).To(Equal("test"))
	}
	BeforeEach(func() {
		openHydraConfig = config.DefaultConfig()
		openHydraConfig.JupyterLabHostBaseDir = "/tmp/jupyter-lab"
		openHydraConfig.PublicDatasetBasePath = "/tmp/public-dataset"
		openHydraConfig.PublicCourseBasePath = "/tmp/public-course"
		openHydraConfig.PublicVSCodeBasePath = "/tmp/public-vscode"
		util.CreateDirIfNotExists(openHydraConfig.JupyterLabHostBaseDir)
		util.CreateDirIfNotExists(openHydraConfig.PublicDatasetBasePath)
		util.CreateDirIfNotExists(openHydraConfig.PublicCourseBasePath)
		openHydraConfig.DefaultGpuDriver = "nvidia.com/gpu"
		teacher = createFakeUser("teacher", "teacher", 1)
		student = createFakeUser("student", "student", 2)
		newTeacher = createFakeUser("newTeacher", "newTeacher", 1)
		newStudent = createFakeUser("newStudent", "newStudent", 2)
		device1 = createDevice("teacher", 1)
		device2 = createDevice("student", 0)
		device3 = createDevice("student", 1)
		setting = &xSetting.Setting{
			ObjectMeta: metaV1.ObjectMeta{
				Name: "default",
			},
			Spec: xSetting.SettingSpec{
				DefaultGpuPerDevice: 1,
			},
		}
		fakeDb = &database.Faker{}
		fakeDb.Init()
		_ = fakeDb.CreateUser(teacher)
		_ = fakeDb.CreateUser(student)
		initContainer()
	})

	Describe("AuthAndAuthorization test", func() {
		It("open-hydra user list should be expected", func() {
			_, r2 := callApi(http.MethodGet, openHydraUsersURL, createTokenValue(teacher, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))
			result, err := io.ReadAll(r2.Body)
			Expect(err).To(BeNil())
			var users xUserV1.OpenHydraUserList
			err = json.Unmarshal(result, &users)
			Expect(err).To(BeNil())
			Expect(len(users.Items)).To(Equal(2))

			_, r2 = callApi(http.MethodGet, openHydraUsersURL, createTokenValue(student, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusForbidden))
		})

		It("open-hydra user create should be expected", func() {
			body1, err := json.Marshal(newStudent)
			Expect(err).To(BeNil())
			_, r2 := callApi(http.MethodPost, openHydraUsersURL, createTokenValue(teacher, nil), bytes.NewReader(body1))
			Expect(r2.Code).To(Equal(http.StatusCreated))
			result, err := io.ReadAll(r2.Body)
			Expect(err).To(BeNil())
			var user xUserV1.OpenHydraUser
			err = json.Unmarshal(result, &user)
			Expect(err).To(BeNil())
			Expect(user.Name).To(Equal(newStudent.Name))

			body2, err := json.Marshal(newTeacher)
			Expect(err).To(BeNil())
			_, r2 = callApi(http.MethodPost, openHydraUsersURL, createTokenValue(student, nil), bytes.NewReader(body2))
			Expect(r2.Code).To(Equal(http.StatusForbidden))
		})

		It("open-hydra user login should be expected", func() {
			body1, err := json.Marshal(teacher)
			Expect(err).To(BeNil())
			_, r2 := callApi(http.MethodGet, openHydraUsersURL+"/teacher", createTokenValue(teacher, nil), bytes.NewReader(body1))
			Expect(r2.Code).To(Equal(http.StatusOK))

			body2, err := json.Marshal(student)
			Expect(err).To(BeNil())
			_, r2 = callApi(http.MethodGet, openHydraUsersURL+"/student", createTokenValue(student, nil), bytes.NewReader(body2))
			Expect(r2.Code).To(Equal(http.StatusOK))
		})

		It("open-hydra user delete should be expected", func() {
			_, r2 := callApi(http.MethodDelete, openHydraUsersURL+"/teacher", createTokenValue(teacher, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))

			_, r2 = callApi(http.MethodDelete, openHydraUsersURL+"/student", createTokenValue(student, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusForbidden))
		})

		It("open-hydra user update should be expected", func() {
			student.Spec.Description = "test"
			body1, err := json.Marshal(student)
			Expect(err).To(BeNil())
			_, r2 := callApi(http.MethodPut, openHydraUsersURL+"/student", createTokenValue(teacher, nil), bytes.NewReader(body1))
			Expect(r2.Code).To(Equal(http.StatusOK))
			var test *xUserV1.OpenHydraUser
			test, _ = fakeDb.GetUser("student")
			Expect(test.Spec.Description).To(Equal("test"))

			student.Spec.Description = "test2"
			body2, err := json.Marshal(student)
			Expect(err).To(BeNil())
			_, r2 = callApi(http.MethodPut, openHydraUsersURL+"/student", createTokenValue(student, nil), bytes.NewReader(body2))
			Expect(r2.Code).To(Equal(http.StatusForbidden))
		})

		It("open-hydra user get should be expected", func() {
			_, r2 := callApi(http.MethodGet, openHydraUsersURL+"/teacher", createTokenValue(teacher, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))

			_, r2 = callApi(http.MethodGet, openHydraUsersURL+"/student", createTokenValue(student, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))
		})

		It("open-hydra user get should be deny because student do not have right to get other user info", func() {
			_, r2 := callApi(http.MethodGet, openHydraUsersURL+"/teacher", createTokenValue(student, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusForbidden))
		})

		It("open-hydra device list should expected", func() {
			_, r2 := callApi(http.MethodGet, openHydraDevicesURL, createTokenValue(teacher, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))

			_, r2 = callApi(http.MethodGet, openHydraDevicesURL, createTokenValue(student, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusForbidden))
		})

		It("open-hydra device get should be expected", func() {
			_, r2 := callApi(http.MethodGet, openHydraDevicesURL+"/teacher", createTokenValue(teacher, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))

			_, r2 = callApi(http.MethodGet, openHydraDevicesURL+"/student", createTokenValue(student, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))
		})

		It("open-hydra device get should be deny because student do not have right to get other device info", func() {
			_, r2 := callApi(http.MethodGet, openHydraDevicesURL+"/test", createTokenValue(student, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusForbidden))
		})

		It("open-hydra device create should be expected", func() {
			body1, err := json.Marshal(device1)
			Expect(err).To(BeNil())
			_, r2 := callApi(http.MethodPost, openHydraDevicesURL, createTokenValue(teacher, nil), bytes.NewReader(body1))
			Expect(r2.Code).To(Equal(http.StatusOK))
			var target xDeviceV1.Device
			result, err := io.ReadAll(r2.Body)
			Expect(err).To(BeNil())
			err = json.Unmarshal(result, &target)
			Expect(err).To(BeNil())
			Expect(target.Spec.DeviceGpu).To(Equal(uint8(1)))
			Expect(target.Spec.DeviceType).To(Equal("gpu"))
			Expect(target.Spec.DeviceStatus).To(Equal("Creating"))
			Expect(target.Spec.OpenHydraUsername).To(Equal("teacher"))

			target = xDeviceV1.Device{}
			body2, err := json.Marshal(device2)
			Expect(err).To(BeNil())
			_, r2 = callApi(http.MethodPost, openHydraDevicesURL, createTokenValue(student, nil), bytes.NewReader(body2))
			Expect(r2.Code).To(Equal(http.StatusOK))
			result, err = io.ReadAll(r2.Body)
			Expect(err).To(BeNil())
			err = json.Unmarshal(result, &target)
			Expect(err).To(BeNil())
			Expect(target.Spec.DeviceGpu).To(Equal(uint8(0)))
			Expect(target.Spec.DeviceType).To(Equal("cpu"))
			Expect(target.Spec.DeviceStatus).To(Equal("Creating"))
			Expect(target.Spec.OpenHydraUsername).To(Equal("student"))
		})

		It("open-hydra device create should be deny because student do not have right to create device for others", func() {
			body, err := json.Marshal(device1)
			Expect(err).To(BeNil())
			_, r2 := callApi(http.MethodPost, openHydraDevicesURL, createTokenValue(student, nil), bytes.NewReader(body))
			Expect(r2.Code).To(Equal(http.StatusForbidden))
		})

		It("open-hydra device create should be deny because student do not have right to create gpu device", func() {
			body, err := json.Marshal(device3)
			Expect(err).To(BeNil())
			_, r2 := callApi(http.MethodPost, openHydraDevicesURL, createTokenValue(student, nil), bytes.NewReader(body))
			Expect(r2.Code).To(Equal(http.StatusForbidden))
		})

		It("open-hydra device delete should be expected", func() {
			_, r2 := callApi(http.MethodDelete, openHydraDevicesURL+"/teacher", createTokenValue(teacher, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))

			_, r2 = callApi(http.MethodDelete, openHydraDevicesURL+"/student", createTokenValue(student, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))
		})

		It("open-hydra device delete should be deny because student do not have right to delete other's device", func() {
			_, r2 := callApi(http.MethodDelete, openHydraDevicesURL+"/teacher", createTokenValue(student, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusForbidden))
		})

		It("open-hydra user list by student should be ok when disableAuth", func() {
			openHydraConfig.DisableAuth = true
			initContainer()
			_, r2 := callApi(http.MethodGet, openHydraUsersURL, createTokenValue(student, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))
			result, err := io.ReadAll(r2.Body)
			Expect(err).To(BeNil())
			var users xUserV1.OpenHydraUserList
			err = json.Unmarshal(result, &users)
			Expect(err).To(BeNil())
			Expect(len(users.Items)).To(Equal(2))
		})

		It("open-hydra get setting by student should be deny", func() {
			_, r2 := callApi(http.MethodGet, openHydraSettingsURL, createTokenValue(student, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusForbidden))
		})

		It("open-hydra update setting by student should be deny", func() {
			_, r2 := callApi(http.MethodPut, openHydraSettingsURL, createTokenValue(student, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusForbidden))
		})

		It("open-hydra get setting by teacher should be ok", func() {
			_, r2 := callApi(http.MethodGet, openHydraSettingsURL, createTokenValue(teacher, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))
			result, err := io.ReadAll(r2.Body)
			Expect(err).To(BeNil())
			var target xSetting.Setting
			err = json.Unmarshal(result, &target)
			Expect(err).To(BeNil())
			Expect(target.Spec.DefaultGpuPerDevice).To(Equal(uint8(0)))
		})

		It("open-hydra update setting by teacher should be ok", func() {
			body, err := json.Marshal(setting)
			Expect(err).To(BeNil())
			_, r2 := callApi(http.MethodPut, openHydraSettingsURL, createTokenValue(teacher, nil), bytes.NewReader(body))
			Expect(r2.Code).To(Equal(http.StatusOK))
			_, r2 = callApi(http.MethodGet, openHydraSettingsURL, createTokenValue(teacher, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))
			result, err := io.ReadAll(r2.Body)
			Expect(err).To(BeNil())
			var target xSetting.Setting
			err = json.Unmarshal(result, &target)
			Expect(err).To(BeNil())
			Expect(target.Spec.DefaultGpuPerDevice).To(Equal(uint8(1)))
		})

		It("list dataset by teacher should be ok", func() {
			_, r2 := callApi(http.MethodGet, openHydraDatasetsURL, createTokenValue(teacher, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))
		})

		It("list dataset by student should be deny", func() {
			_, r2 := callApi(http.MethodGet, openHydraDatasetsURL, createTokenValue(student, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusForbidden))
		})

		It("create dataset by teacher should be ok", func() {
			uploadResource(openHydraDatasetsURL, openHydraConfig.PublicDatasetBasePath)

			// list it
			_, r2 := callApi(http.MethodGet, openHydraDatasetsURL, createTokenValue(teacher, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))
			var target xDataset.DatasetList
			result, err := io.ReadAll(r2.Body)
			Expect(err).To(BeNil())
			err = json.Unmarshal(result, &target)
			Expect(err).To(BeNil())
			Expect(len(target.Items)).To(Equal(1))

			// get it
			_, r2 = callApi(http.MethodGet, openHydraDatasetsURL+"/unit-test", createTokenValue(teacher, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))
			var target2 xDataset.Dataset
			result, err = io.ReadAll(r2.Body)
			Expect(err).To(BeNil())
			err = json.Unmarshal(result, &target2)
			Expect(err).To(BeNil())
			Expect(target2.Name).To(Equal("unit-test"))
			Expect(target2.Spec.Description).To(Equal("unit-test"))

			util.DeleteDirs("/tmp/test")
			util.DeleteDirs(path.Join(openHydraConfig.PublicDatasetBasePath, "unit-test"))
		})

		It("dataset will reject students", func() {
			uploadResource(openHydraDatasetsURL, openHydraConfig.PublicDatasetBasePath)
			_, r2 := callApi(http.MethodGet, openHydraDatasetsURL, createTokenValue(student, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusForbidden))
			_, r2 = callApi(http.MethodGet, openHydraDatasetsURL+"/unit-test", createTokenValue(student, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusForbidden))
			util.DeleteDirs("/tmp/test")
			util.DeleteDirs(path.Join(openHydraConfig.PublicDatasetBasePath, "unit-test"))
		})

		It("dataset delete by teacher should be ok", func() {
			uploadResource(openHydraDatasetsURL, openHydraConfig.PublicDatasetBasePath)
			_, r2 := callApi(http.MethodDelete, openHydraDatasetsURL+"/unit-test", createTokenValue(teacher, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))
			_, err := os.Stat(path.Join(openHydraConfig.PublicDatasetBasePath, "unit-test"))
			Expect(os.IsNotExist(err)).To(BeTrue())

			// list it
			_, r2 = callApi(http.MethodGet, openHydraDatasetsURL, createTokenValue(teacher, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))
			var target xDataset.DatasetList
			result, err := io.ReadAll(r2.Body)
			Expect(err).To(BeNil())
			err = json.Unmarshal(result, &target)
			Expect(err).To(BeNil())
			Expect(len(target.Items)).To(Equal(0))
		})

		It("create course by teacher should be ok", func() {
			uploadResource(openHydraCoursesURL, openHydraConfig.PublicCourseBasePath)

			// list it
			_, r2 := callApi(http.MethodGet, openHydraCoursesURL, createTokenValue(teacher, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))
			var target xCourseV1.CourseList
			result, err := io.ReadAll(r2.Body)
			Expect(err).To(BeNil())
			err = json.Unmarshal(result, &target)
			Expect(err).To(BeNil())
			Expect(len(target.Items)).To(Equal(1))

			// get it
			_, r2 = callApi(http.MethodGet, openHydraCoursesURL+"/unit-test", createTokenValue(teacher, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))
			var target2 xCourseV1.Course
			result, err = io.ReadAll(r2.Body)
			Expect(err).To(BeNil())
			err = json.Unmarshal(result, &target2)
			Expect(err).To(BeNil())
			Expect(target2.Name).To(Equal("unit-test"))
			Expect(target2.Spec.Description).To(Equal("unit-test"))

			util.DeleteDirs("/tmp/test")
			util.DeleteDirs(path.Join(openHydraConfig.PublicCourseBasePath, "unit-test"))
		})

		It("course will reject students", func() {
			uploadResource(openHydraCoursesURL, openHydraConfig.PublicCourseBasePath)
			_, r2 := callApi(http.MethodGet, openHydraCoursesURL, createTokenValue(student, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusForbidden))
			_, r2 = callApi(http.MethodGet, openHydraCoursesURL+"/unit-test", createTokenValue(student, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusForbidden))
			util.DeleteDirs("/tmp/test")
			util.DeleteDirs(path.Join(openHydraConfig.PublicCourseBasePath, "unit-test"))
		})

		It("course delete by teacher should be ok", func() {
			uploadResource(openHydraCoursesURL, openHydraConfig.PublicCourseBasePath)
			_, r2 := callApi(http.MethodDelete, openHydraCoursesURL+"/unit-test", createTokenValue(teacher, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))
			_, err := os.Stat(path.Join(openHydraConfig.PublicCourseBasePath, "unit-test"))
			Expect(os.IsNotExist(err)).To(BeTrue())

			// list it
			_, r2 = callApi(http.MethodGet, openHydraCoursesURL, createTokenValue(teacher, nil), nil)
			Expect(r2.Code).To(Equal(http.StatusOK))
			var target xCourseV1.CourseList
			result, err := io.ReadAll(r2.Body)
			Expect(err).To(BeNil())
			err = json.Unmarshal(result, &target)
			Expect(err).To(BeNil())
			Expect(len(target.Items)).To(Equal(0))
		})

		AfterEach(func() {
		})
	})
})
