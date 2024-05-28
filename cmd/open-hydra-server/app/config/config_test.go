package config

import (
	"open-hydra/pkg/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("open-hydra config test", func() {

	configFile := "/tmp/open-hydra-config.yaml"
	configFile2 := "/tmp/open-hydra-config2.yaml"
	kubeConfigPath := "/tmp/kube-config.yaml"
	testConfig := DefaultConfig()
	testConfig2 := *testConfig
	testConfig2.MySqlConfig.Character = "utf8mb4"
	testConfig2.MySqlConfig.Collation = "utf8mb4_general_ci"
	testConfig2.LeaderElection = nil
	testKubeConfig := `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: test
    server: https://apiserver.cluster.local:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: kubernetes-admin
  name: kubernetes-admin@test-cluster
current-context: kubernetes-admin@test-cluster
kind: Config
preferences: {}
users:
- name: kubernetes-admin
  user:
    client-certificate-data: test
    client-key-data: test`

	BeforeEach(func() {
		err := util.WriteFileWithNosec(kubeConfigPath, []byte(testKubeConfig))
		Expect(err).To(BeNil())
	})

	Describe("config file read write test", func() {
		It("write and read to config file no error", func() {
			err := WriteConfig(configFile, testConfig)
			Expect(err).To(BeNil())
			targetConfig, err := LoadConfig(configFile, kubeConfigPath)
			Expect(err).To(BeNil())
			Expect(targetConfig.DefaultCpuPerDevice).To(Equal(testConfig.DefaultCpuPerDevice))
			Expect(targetConfig.DefaultRamPerDevice).To(Equal(testConfig.DefaultRamPerDevice))
			Expect(targetConfig.DefaultGpuPerDevice).To(Equal(testConfig.DefaultGpuPerDevice))
			Expect(targetConfig.MySqlConfig.Address).To(Equal("mysql.svc.cluster.local"))
			Expect(targetConfig.MySqlConfig.Port).To(Equal(uint16(3306)))
			Expect(targetConfig.MySqlConfig.Username).To(Equal("root"))
			Expect(targetConfig.MySqlConfig.Password).To(Equal("root"))
			Expect(targetConfig.MySqlConfig.DataBaseName).To(Equal("open-hydra"))
			Expect(targetConfig.MySqlConfig.Protocol).To(Equal("tcp"))
			Expect(targetConfig.MySqlConfig.Character).To(Equal("utf8mb4"))
			Expect(targetConfig.MySqlConfig.Collation).To(Equal("utf8mb4_general_ci"))
			Expect(targetConfig.EtcdConfig).To(Equal(DefaultEtcdConfig()))
			Expect(targetConfig.MaximumPortsPerSandbox).To(Equal(uint8(3)))
		})
		It("test default value will not overwrite by empty value ", func() {
			err := WriteConfig(configFile2, &testConfig2)
			Expect(err).To(BeNil())
			targetConfig, err := LoadConfig(configFile2, kubeConfigPath)
			Expect(err).To(BeNil())
			Expect(targetConfig.LeaderElection).To(Equal(testConfig.LeaderElection))
			Expect(targetConfig.AuthDelegateConfig).To(BeNil())
			Expect(targetConfig.MySqlConfig.Character).To(Equal("utf8mb4"))
			Expect(targetConfig.MySqlConfig.Collation).To(Equal("utf8mb4_general_ci"))
		})
		AfterEach(func() {
			_ = util.DeleteFile(configFile)
			_ = util.DeleteFile(configFile2)
		})
	})
})
