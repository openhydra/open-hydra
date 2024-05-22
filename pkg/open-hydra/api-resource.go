package openhydra

import (
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	OpenHydraUserKind = "OpenHydraUser"
	DeviceKind        = "Device"
	SumUpKind         = "SumUp"
	OpenHydraUserPath = "openhydrausers"
	DevicePath        = "devices"
	SumUpPath         = "sumups"
	DatasetPath       = "datasets"
	DatasetKind       = "Dataset"
	SettingKind       = "Setting"
	SettingPath       = "settings"
	CourseKind        = "Course"
	CoursePath        = "courses"
)

// we should register the api resource here
func ApiResources() []metaV1.APIResource {
	return []metaV1.APIResource{
		{
			Name:         OpenHydraUserPath,
			SingularName: "xuser",
			Namespaced:   false,
			Kind:         OpenHydraUserKind,
			Verbs:        metaV1.Verbs{"get", "list", "watch", "create", "update", "delete", "patch"},
		},
		{
			Name:         DevicePath,
			SingularName: "dev",
			Namespaced:   false,
			Kind:         DeviceKind,
			Verbs:        metaV1.Verbs{"get", "list", "watch", "create", "update", "delete", "patch"},
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
}
