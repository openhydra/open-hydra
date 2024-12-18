package k8s

import (
	"fmt"
	"open-hydra/pkg/open-hydra/apis"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("K8s", func() {
	var cpuMemorySet CpuMemorySet
	var gpuSet apis.GpuSet
	var deployParameter *DeploymentParameters
	var volumeMounts []apis.VolumeMount
	var volumes []apis.Volume
	var affinity *coreV1.Affinity
	BeforeEach(func() {
		cpuMemorySet = CpuMemorySet{
			CpuRequest:    "100m",
			CpuLimit:      "200m",
			MemoryRequest: "100Mi",
			MemoryLimit:   "200Mi",
		}

		gpuSet = apis.GpuSet{
			Gpu:           1,
			GpuDriverName: "nvidia.com/gpu",
		}

		volumeMounts = []apis.VolumeMount{
			{
				Name:       "test",
				MountPath:  "/test/mount",
				SourcePath: "/test/source",
			},
		}

		volumes = []apis.Volume{
			{
				HostPath: &apis.HostPath{
					Name: "test",
					Path: "/test/path",
					Type: "Directory",
				},
			},
		}

		affinity = &coreV1.Affinity{
			NodeAffinity: &coreV1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &coreV1.NodeSelector{
					NodeSelectorTerms: []coreV1.NodeSelectorTerm{
						{
							MatchExpressions: []coreV1.NodeSelectorRequirement{
								{
									Key:      "testKey",
									Operator: "In",
									Values:   []string{"testValue"},
								},
							},
						},
					},
				},
			},
		}

		deployParameter = &DeploymentParameters{
			Username:     "testUsername",
			Namespace:    "testNamespace",
			Image:        "testImage",
			SandboxName:  "testSandbox",
			CpuMemorySet: cpuMemorySet,
			GpuSet:       gpuSet,
			VolumeMounts: volumeMounts,
			Command:      []string{"testCommand"},
			Args:         []string{"--testArgs argValue1"},
			Ports:        map[string]int{"testPort": 8080},
			Volumes:      volumes,
			Affinity:     affinity,
		}

	})

	Describe("createResource", func() {
		It("should be expected", func() {
			reqRequest, limRequest := createResource(cpuMemorySet, gpuSet)
			Expect(reqRequest.Cpu().String()).To(Equal("100m"))
			Expect(limRequest.Cpu().String()).To(Equal("200m"))
			Expect(reqRequest.Memory().String()).To(Equal("100Mi"))
			Expect(limRequest.Memory().String()).To(Equal("200Mi"))
			Expect(reqRequest["nvidia.com/gpu"]).To(Equal(resource.MustParse(strconv.Itoa(int(gpuSet.Gpu)))))
			Expect(limRequest["nvidia.com/gpu"]).To(Equal(resource.MustParse(strconv.Itoa(int(gpuSet.Gpu)))))
		})
		It("should be expected without gpu set", func() {
			gpuSet = apis.GpuSet{}
			reqRequest, limRequest := createResource(cpuMemorySet, gpuSet)
			Expect(reqRequest.Cpu().String()).To(Equal("100m"))
			Expect(limRequest.Cpu().String()).To(Equal("200m"))
			Expect(reqRequest.Memory().String()).To(Equal("100Mi"))
			Expect(limRequest.Memory().String()).To(Equal("200Mi"))
			Expect(reqRequest["nvidia.com/gpu"]).To(Equal(resource.Quantity{}))
			Expect(limRequest["nvidia.com/gpu"]).To(Equal(resource.Quantity{}))
		})
	})

	Describe("createVolume", func() {
		It("should be expected", func() {
			volume := createVolume(volumes)
			Expect(volume[0].Name).To(Equal("test"))
			Expect(volume[0].HostPath.Path).To(Equal("/test/path"))
			Expect(volume[0].EmptyDir).To(BeNil())
		})
		It("should be expected with empty dir", func() {
			volumes = []apis.Volume{
				{
					EmptyDir: &apis.EmptyDir{
						Name:   "test",
						Medium: "Memory",
					},
				},
			}
			volume := createVolume(volumes)
			Expect(volume[0].Name).To(Equal("test"))
			Expect(volume[0].HostPath).To(BeNil())
			Expect(string(volume[0].EmptyDir.Medium)).To(Equal("Memory"))
		})
	})

	Describe("createContainers", func() {
		It("should be expected", func() {
			baseName := fmt.Sprintf(OpenHydraDeployNameTemplate, deployParameter.Username)
			reqRequest, limRequest := createResource(cpuMemorySet, gpuSet)
			containers := createContainers(baseName, deployParameter.Image, "", deployParameter.VolumeMounts, reqRequest, limRequest, deployParameter.Command, deployParameter.Args, deployParameter.Ports)
			Expect(containers[0].Name).To(Equal(fmt.Sprintf("%s-%s", baseName, "container")))
			Expect(containers[0].Image).To(Equal(deployParameter.Image))
			Expect(containers[0].VolumeMounts[0].Name).To(Equal("test"))
			Expect(containers[0].VolumeMounts[0].MountPath).To(Equal("/test/mount"))
			Expect(containers[0].VolumeMounts[0].ReadOnly).To(Equal(false))
			Expect(containers[0].Resources.Requests).To(Equal(reqRequest))
			Expect(containers[0].Resources.Limits).To(Equal(limRequest))
			Expect(containers[0].Command).To(Equal(deployParameter.Command))
			Expect(containers[0].Args).To(Equal(deployParameter.Args))
			Expect(containers[0].Ports[0].ContainerPort).To(Equal(int32(8080)))
		})
		It("should be expected without args and command", func() {
			deployParameter.Command = nil
			deployParameter.Args = nil
			baseName := fmt.Sprintf(OpenHydraDeployNameTemplate, deployParameter.Username)
			reqRequest, limRequest := createResource(cpuMemorySet, gpuSet)
			containers := createContainers(baseName, deployParameter.Image, "", deployParameter.VolumeMounts, reqRequest, limRequest, deployParameter.Command, deployParameter.Args, deployParameter.Ports)
			Expect(containers[0].Name).To(Equal(fmt.Sprintf("%s-%s", baseName, "container")))
			Expect(containers[0].Image).To(Equal(deployParameter.Image))
			Expect(containers[0].VolumeMounts[0].Name).To(Equal("test"))
			Expect(containers[0].VolumeMounts[0].MountPath).To(Equal("/test/mount"))
			Expect(containers[0].VolumeMounts[0].ReadOnly).To(Equal(false))
			Expect(containers[0].Resources.Requests).To(Equal(reqRequest))
			Expect(containers[0].Resources.Limits).To(Equal(limRequest))
			Expect(containers[0].Command).To(BeNil())
			Expect(containers[0].Args).To(BeNil())
			Expect(containers[0].Ports[0].ContainerPort).To(Equal(int32(8080)))
		})
	})

	Describe("createDeployment", func() {
		It("should be expected", func() {
			deployment := createDeployment(deployParameter)
			baseName := fmt.Sprintf(OpenHydraDeployNameTemplate, deployParameter.Username)
			reqRequest, limRequest := createResource(cpuMemorySet, gpuSet)
			Expect(deployment.Name).To(Equal(baseName))
			Expect(deployment.Namespace).To(Equal(deployParameter.Namespace))
			Expect(deployment.Labels[OpenHydraUserLabelKey]).To(Equal(deployParameter.Username))
			Expect(deployment.Labels[OpenHydraWorkloadLabelKey]).To(Equal(OpenHydraWorkloadLabelValue))
			Expect(deployment.Labels[OpenHydraIDELabelKey]).To(Equal(OpenHydraIDELabelUnset))
			Expect(deployment.Labels[OpenHydraSandboxKey]).To(Equal(deployParameter.SandboxName))
			Expect(*deployment.Spec.Replicas).To(Equal(int32(1)))
			Expect(deployment.Spec.Template.Spec.Containers[0].Name).To(Equal(fmt.Sprintf("%s-%s", fmt.Sprintf(OpenHydraDeployNameTemplate, deployParameter.Username), "container")))
			Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(Equal(deployParameter.Image))
			Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name).To(Equal("test"))
			Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath).To(Equal("/test/mount"))
			Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].ReadOnly).To(Equal(false))
			Expect(deployment.Spec.Template.Spec.Containers[0].Resources.Requests).To(Equal(reqRequest))
			Expect(deployment.Spec.Template.Spec.Containers[0].Resources.Limits).To(Equal(limRequest))
			Expect(deployment.Spec.Template.Spec.Containers[0].Command).To(Equal(deployParameter.Command))
			Expect(deployment.Spec.Template.Spec.Containers[0].Args).To(Equal(deployParameter.Args))
			Expect(deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).To(Equal(int32(8080)))
			Expect(deployment.Spec.Template.Spec.Volumes[0].HostPath.Path).To(Equal("/test/path"))
			Expect(deployment.Spec.Template.Spec.Affinity).To(Equal(affinity))
		})
		It("should be expected label proper set", func() {
			deployParameter.CustomLabels = map[string]string{
				"testLabel": "testValue",
			}
			deployment := createDeployment(deployParameter)
			Expect(deployment.Labels["testLabel"]).To(Equal("testValue"))
		})
		It("should be expected custom label key should be ignored", func() {
			deployParameter.CustomLabels = map[string]string{
				OpenHydraUserLabelKey: "testValue",
			}
			deployment := createDeployment(deployParameter)
			Expect(deployment.Labels[OpenHydraUserLabelKey]).To(Equal(deployParameter.Username))
		})

	})
})
