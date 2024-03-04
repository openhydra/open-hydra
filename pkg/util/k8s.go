package util

import (
	"fmt"

	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"open-hydra/cmd/open-hydra-server/app/option"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func FillKindAndApiVersion(mType *metaV1.TypeMeta, kind string) {
	mType.Kind = kind
	mType.APIVersion = fmt.Sprintf("%s/%s", option.GroupVersion.Group, option.GroupVersion.Version)
}

func FillObjectGVK(obj schema.ObjectKind) {
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   option.GroupVersion.Group,
		Kind:    GetObjectKind(obj),
		Version: option.GroupVersion.Version,
	})
}

func GetObjectKind(obj schema.ObjectKind) string {
	v, err := conversion.EnforcePtr(obj)
	if err != nil {
		panic(err)
	}
	return v.Type().Name()
}
