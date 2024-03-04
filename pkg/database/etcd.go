package database

import (
	"open-hydra/cmd/open-hydra-server/app/config"
	xCourseV1 "open-hydra/pkg/apis/open-hydra-api/course/core/v1"
	xDatasetV1 "open-hydra/pkg/apis/open-hydra-api/dataset/core/v1"
	xUserV1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"
)

type Etcd struct {
	Config *config.OpenHydraServerConfig
}

// implements IDataBaseUser creates a new user
func (db *Etcd) CreateUser(user *xUserV1.OpenHydraUser) error {
	return nil
}

// implements IDataBaseUser gets a user by name
func (db *Etcd) GetUser(name string) (*xUserV1.OpenHydraUser, error) {
	return nil, nil
}

// implements IDataBaseUser updates a user
func (db *Etcd) UpdateUser(user *xUserV1.OpenHydraUser) error {
	return nil
}

// implements IDataBaseUser deletes a user
func (db *Etcd) DeleteUser(name string) error {
	return nil
}

// implements IDataBaseUser lists all users
func (db *Etcd) ListUsers() (xUserV1.OpenHydraUserList, error) {
	return xUserV1.OpenHydraUserList{}, nil
}

// implements IDataBaseDataset creates a new dataset
func (db *Etcd) CreateDataset(dataset *xDatasetV1.Dataset) error {
	return nil
}

// implements IDataBaseDataset gets a dataset by name
func (db *Etcd) GetDataset(name string) (*xDatasetV1.Dataset, error) {
	return nil, nil
}

// implements IDataBaseDataset updates a dataset
func (db *Etcd) UpdateDataset(dataset *xDatasetV1.Dataset) error {
	return nil
}

// implements IDataBaseDataset deletes a dataset
func (db *Etcd) DeleteDataset(name string) error {
	return nil
}

// implements IDataBaseDataset lists all datasets
func (db *Etcd) ListDatasets() (xDatasetV1.DatasetList, error) {
	return xDatasetV1.DatasetList{}, nil
}

func (db *Etcd) LoginUser(name, password string) (*xUserV1.OpenHydraUser, error) {
	return nil, nil
}

func (db *Etcd) InitDb() error {
	return nil
}

// implements IDataBaseCourse creates a new course
func (db *Etcd) CreateCourse(course *xCourseV1.Course) error {
	return nil
}

// implements IDataBaseCourse gets a course by name
func (db *Etcd) GetCourse(name string) (*xCourseV1.Course, error) {
	return nil, nil
}

// implements IDataBaseCourse updates a course
func (db *Etcd) UpdateCourse(course *xCourseV1.Course) error {
	return nil
}

// implements IDataBaseCourse deletes a course
func (db *Etcd) DeleteCourse(name string) error {
	return nil
}

// implements IDataBaseCourse lists all courses
func (db *Etcd) ListCourses() (xCourseV1.CourseList, error) {
	return xCourseV1.CourseList{}, nil
}
