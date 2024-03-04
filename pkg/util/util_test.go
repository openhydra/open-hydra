package util

import (
	"fmt"
	"open-hydra/cmd/open-hydra-server/app/option"
	xDeviceV1 "open-hydra/pkg/apis/open-hydra-api/device/core/v1"
	xUserV1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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

})
