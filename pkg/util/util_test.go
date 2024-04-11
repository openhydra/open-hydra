package util

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"open-hydra/cmd/open-hydra-server/app/option"
	xDeviceV1 "open-hydra/pkg/apis/open-hydra-api/device/core/v1"
	xUserV1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"
	"time"

	"github.com/emicklei/go-restful"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type TestStruct struct {
	Name       string
	TotalCount int
}

var testRouter = func(ws *restful.WebService) {
	ws.Route(ws.GET("/route1-test").To(func(request *restful.Request, response *restful.Response) {
		response.AddHeader("test", "test")
		response.WriteAsJson(&TestStruct{Name: "test", TotalCount: 2})
	}))
}

var _ = Describe("open-hydra-server util test", func() {
	BeforeEach(func() {
	})
	Describe("FillKinkAndApiVersion test", func() {
		It("should be expected", func() {
			device := xDeviceV1.Device{}
			FillKindAndApiVersion(&device.TypeMeta, "OpenHydraUser")
			Expect(device.Kind).To(Equal("OpenHydraUser"))
			Expect(device.APIVersion).To(Equal(fmt.Sprintf("%s/%s", option.GroupVersion.Group, option.GroupVersion.Version)))
		})

		AfterEach(func() {
		})
	})

	Describe("FillObjectGVK test", func() {
		It("should be expected", func() {
			xUser := &xUserV1.OpenHydraUser{}
			FillObjectGVK(xUser)
			Expect(xUser.APIVersion, fmt.Sprintf("%s/%s", option.GroupVersion.Group, option.GroupVersion.Version))
			Expect(xUser.Kind, "OpenHydraUser")
		})
		AfterEach(func() {
		})
	})

	Describe("FillObjectGVK test", func() {
		It("should be expected", func() {
			xUser := &xUserV1.OpenHydraUser{}
			FillObjectGVK(xUser)
			Expect(xUser.APIVersion, fmt.Sprintf("%s/%s", option.GroupVersion.Group, option.GroupVersion.Version))
			Expect(xUser.Kind, "OpenHydraUser")
		})
		AfterEach(func() {
		})
	})

	Describe("ZipDir test", func() {
		It("should be expected", func() {
			err := CreateDirIfNotExists("/tmp/test")
			Expect(err).To(BeNil())
			err = CreateDirIfNotExists("/tmp/test/test1")
			Expect(err).To(BeNil())
			err = WriteFileWithNosec("/tmp/test/test1/test.txt", []byte("test"))
			Expect(err).To(BeNil())
			err = ZipDir("/tmp/test", "/tmp/test.zip")
			Expect(err).To(BeNil())
		})
		AfterEach(func() {
			DeleteDirs("/tmp/test")
			DeleteFile("/tmp/test.zip")
		})
	})

	Describe("CommonRequest test", func() {
		It("should be expected", func() {
			stopChan := make(chan struct{}, 1)
			go StartMockServer(20080, testRouter, stopChan)
			time.Sleep(2 * time.Second)
			defer func() {
				stopChan <- struct{}{}
			}()
			rawResult, header, httpCode, err := CommonRequest("http://localhost:20080/route1-test", http.MethodGet, "", json.RawMessage{}, map[string]string{}, true, true, 3*time.Second)
			Expect(err).To(BeNil())
			Expect(httpCode).To(Equal(http.StatusOK))
			var result TestStruct
			err = json.Unmarshal(rawResult, &result)
			Expect(err).To(BeNil())
			Expect(result.Name).To(Equal("test"))
			Expect(result.TotalCount).To(Equal(2))
			headerValue := header.Get("test")
			Expect(headerValue).To(Equal("test"))
		})
		AfterEach(func() {
		})
	})

	Describe("StartMockServer test", func() {
		It("should be expected", func() {
			stopChan := make(chan struct{}, 1)
			go StartMockServer(28080, testRouter, stopChan)
			time.Sleep(2 * time.Second)
			defer close(stopChan)
			// code check tcp port
			conn, err := net.DialTimeout("tcp", "localhost:28080", 3*time.Second)
			Expect(err).To(BeNil())
			defer conn.Close()
		})
		AfterEach(func() {
		})
	})

	Describe("GetStringValueOrDefault test", func() {
		It("should be expected", func() {
			value := GetStringValueOrDefault("test", "test", "default")
			Expect(value).To(Equal("test"))
			value = GetStringValueOrDefault("", "", "default")
			Expect(value).To(Equal("default"))
		})
	})

})
