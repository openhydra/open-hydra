package database

import (
	xCourseV1 "open-hydra/pkg/apis/open-hydra-api/course/core/v1"
	xDatasetV1 "open-hydra/pkg/apis/open-hydra-api/dataset/core/v1"
	xUserV1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"
)

type IDataBase interface {
	IDataBaseDataset
	IDataBaseUser
	IDataBaseCourse
	InitDb() error
}

type IDataBaseUser interface {
	// Create a new user
	CreateUser(user *xUserV1.OpenHydraUser) error
	// Get a user by name
	GetUser(name string) (*xUserV1.OpenHydraUser, error)
	// Update a user
	UpdateUser(user *xUserV1.OpenHydraUser) error
	// Delete a user
	DeleteUser(name string) error
	// List all users
	ListUsers() (xUserV1.OpenHydraUserList, error)
	// Login a user
	LoginUser(name, password string) (*xUserV1.OpenHydraUser, error)
}

type IDataBaseDataset interface {
	// Create a new dataset
	CreateDataset(dataset *xDatasetV1.Dataset) error
	// Get a dataset by name
	GetDataset(name string) (*xDatasetV1.Dataset, error)
	// Update a dataset
	UpdateDataset(dataset *xDatasetV1.Dataset) error
	// Delete a dataset
	DeleteDataset(name string) error
	// List all datasets
	ListDatasets() (xDatasetV1.DatasetList, error)
}

type IDataBaseCourse interface {
	// Create a new course
	CreateCourse(course *xCourseV1.Course) error
	// Get a course by name
	GetCourse(name string) (*xCourseV1.Course, error)
	// Update a course
	UpdateCourse(course *xCourseV1.Course) error
	// Delete a course
	DeleteCourse(name string) error
	// List all courses
	ListCourses() (xCourseV1.CourseList, error)
}
