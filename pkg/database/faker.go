package database

import (
	"fmt"
	xCourseV1 "open-hydra/pkg/apis/open-hydra-api/course/core/v1"
	xDatasetV1 "open-hydra/pkg/apis/open-hydra-api/dataset/core/v1"
	xUserV1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"
	"open-hydra/pkg/util"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// faker is for test purpose only
type Faker struct {
	fakeUsers    map[string]*xUserV1.OpenHydraUser
	fakeDatasets map[string]*xDatasetV1.Dataset
	fakeCourses  map[string]*xCourseV1.Course
}

func (f *Faker) Init() {
	f.fakeUsers = make(map[string]*xUserV1.OpenHydraUser)
	f.fakeDatasets = make(map[string]*xDatasetV1.Dataset)
	f.fakeCourses = make(map[string]*xCourseV1.Course)
}

// implements IDataBaseUser creates a new user
func (db *Faker) CreateUser(user *xUserV1.OpenHydraUser) error {
	if _, found := db.fakeUsers[user.Name]; found {
		return fmt.Errorf("user %s already exists", user.Name)
	}
	db.fakeUsers[user.Name] = user
	return nil
}

// implements IDataBaseUser gets a user by name
func (db *Faker) GetUser(name string) (*xUserV1.OpenHydraUser, error) {
	if user, found := db.fakeUsers[name]; found {
		return user, nil
	}
	return nil, fmt.Errorf("user %s not found", name)
}

// implements IDataBaseUser updates a user
func (db *Faker) UpdateUser(user *xUserV1.OpenHydraUser) error {
	if _, found := db.fakeUsers[user.Name]; !found {
		return fmt.Errorf("user %s not found", user.Name)
	}
	db.fakeUsers[user.Name] = user
	return nil
}

// implements IDataBaseUser deletes a user
func (db *Faker) DeleteUser(name string) error {
	delete(db.fakeUsers, name)
	return nil
}

// implements IDataBaseUser lists all users
func (db *Faker) ListUsers() (xUserV1.OpenHydraUserList, error) {
	result := xUserV1.OpenHydraUserList{}
	result.Kind = "List"
	result.APIVersion = "v1"
	for _, user := range db.fakeUsers {
		result.Items = append(result.Items, *user)
	}
	return result, nil
}

// implements IDataBaseDataset creates a new dataset
func (db *Faker) CreateDataset(dataset *xDatasetV1.Dataset) error {
	if _, found := db.fakeDatasets[dataset.Name]; found {
		return fmt.Errorf("dataset %s already exists", dataset.Name)
	}
	db.fakeDatasets[dataset.Name] = dataset
	return nil
}

// implements IDataBaseDataset gets a dataset by name
func (db *Faker) GetDataset(name string) (*xDatasetV1.Dataset, error) {
	if dataset, found := db.fakeDatasets[name]; found {
		return dataset, nil
	}
	var dataset xDatasetV1.Dataset
	util.FillObjectGVK(&dataset)
	return nil, errors.NewNotFound(schema.GroupResource{Group: xUserV1.GroupName, Resource: util.GetObjectKind(&dataset)}, name)
}

// implements IDataBaseDataset updates a dataset
func (db *Faker) UpdateDataset(dataset *xDatasetV1.Dataset) error {
	if _, found := db.fakeDatasets[dataset.Name]; !found {
		return fmt.Errorf("dataset %s not found", dataset.Name)
	}
	db.fakeDatasets[dataset.Name] = dataset
	return nil
}

// implements IDataBaseDataset deletes a dataset
func (db *Faker) DeleteDataset(name string) error {
	delete(db.fakeDatasets, name)
	return nil
}

// implements IDataBaseDataset lists all datasets
func (db *Faker) ListDatasets() (xDatasetV1.DatasetList, error) {
	result := xDatasetV1.DatasetList{}
	result.Kind = "List"
	result.APIVersion = "v1"
	for _, dataset := range db.fakeDatasets {
		result.Items = append(result.Items, *dataset)
	}
	return result, nil
}

func (db *Faker) LoginUser(name, password string) (*xUserV1.OpenHydraUser, error) {
	if user, found := db.fakeUsers[name]; found {
		if user.Spec.Password == password {
			return user, nil
		}
		return nil, fmt.Errorf("wrong password")
	}
	return nil, fmt.Errorf("user %s not found", name)
}

func (db *Faker) InitDb() error {
	return nil
}

// implements IDataBaseCourse creates a new course
func (db *Faker) CreateCourse(course *xCourseV1.Course) error {
	if _, found := db.fakeCourses[course.Name]; found {
		return fmt.Errorf("course %s already exists", course.Name)
	}
	db.fakeCourses[course.Name] = course
	return nil
}

// implements IDataBaseCourse gets a course by name
func (db *Faker) GetCourse(name string) (*xCourseV1.Course, error) {
	if course, found := db.fakeCourses[name]; found {
		return course, nil
	}
	var course xDatasetV1.Dataset
	util.FillObjectGVK(&course)
	return nil, errors.NewNotFound(schema.GroupResource{Group: xUserV1.GroupName, Resource: util.GetObjectKind(&course)}, name)
}

// implements IDataBaseCourse updates a course
func (db *Faker) UpdateCourse(course *xCourseV1.Course) error {
	if _, found := db.fakeCourses[course.Name]; !found {
		return fmt.Errorf("course %s not found", course.Name)
	}
	db.fakeCourses[course.Name] = course
	return nil
}

// implements IDataBaseCourse deletes a course
func (db *Faker) DeleteCourse(name string) error {
	delete(db.fakeCourses, name)
	return nil
}

// implements IDataBaseCourse lists all courses
func (db *Faker) ListCourses() (xCourseV1.CourseList, error) {
	result := xCourseV1.CourseList{}
	result.Kind = "List"
	result.APIVersion = "v1"
	for _, course := range db.fakeCourses {
		result.Items = append(result.Items, *course)
	}
	return result, nil
}
