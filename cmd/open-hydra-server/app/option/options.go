package option

import (
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	genericOptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/client-go/util/homedir"
)

var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)
var GroupVersion = schema.GroupVersion{Group: "open-hydra-server.openhydra.io", Version: "v1"}

type OpenHydraServerOption struct {
	ConfigFile     string
	KubeConfigFile string
}

func NewDefaultOpenHydraServerOption() *OpenHydraServerOption {
	return &OpenHydraServerOption{}
}

func NewDefaultApiServerOption() *genericOptions.RecommendedOptions {
	return genericOptions.NewRecommendedOptions("", Codecs.LegacyCodec(GroupVersion))
}

func (option *OpenHydraServerOption) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&option.ConfigFile, "open-hydra-server-config", homedir.HomeDir()+"/.open-hydra-server/config.yaml", "config file location")
	fs.StringVar(&option.KubeConfigFile, "kube-config", homedir.HomeDir()+"/.kube/config", "kube-config file location")
}

type Options struct {
	OpenHydraServerOption *OpenHydraServerOption
	ApiServerOption       *genericOptions.RecommendedOptions
}

func (options *Options) BindFlags(fs *pflag.FlagSet) {
	options.OpenHydraServerOption.BindFlags(fs)
	options.ApiServerOption.AddFlags(fs)
}
