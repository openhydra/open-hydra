package database

import (
	"database/sql"
	stdErr "errors"
	"fmt"
	"log/slog"

	"open-hydra/cmd/open-hydra-server/app/config"
	xCourseV1 "open-hydra/pkg/apis/open-hydra-api/course/core/v1"
	xDatasetV1 "open-hydra/pkg/apis/open-hydra-api/dataset/core/v1"
	xUserV1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"
	"open-hydra/pkg/util"

	defaultPlugin "open-hydra/pkg/database/auth-plugin"
	keystoneTrain "open-hydra/pkg/database/auth-plugin/keystone/train"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/sync/singleflight"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewMysql(cfg *config.OpenHydraServerConfig) IDataBase {
	result := &Mysql{
		Config:      cfg,
		singleGroup: new(singleflight.Group),
	}

	if cfg.AuthDelegateConfig != nil {
		if cfg.AuthDelegateConfig.KeystoneConfig != nil {
			slog.Debug("Using keystone auth plugin")
			result.IDataBaseUser = &keystoneTrain.KeystoneAuthPlugin{
				Config: cfg,
			}
		}
	}

	if result.IDataBaseUser == nil {
		slog.Debug("No user auth plugin is set use mysql")
		result.IDataBaseUser = &defaultPlugin.DefaultMysqlAuthPlugin{
			Db: result.getDB,
		}
	}

	return result
}

// Mysql implements IDataBase
type Mysql struct {
	Config        *config.OpenHydraServerConfig
	instance      *sql.DB
	singleGroup   *singleflight.Group
	IDataBaseUser // embed IDataBaseUser
}

// CreateDataset implements IDataBaseDataset creates a new dataset
func (db *Mysql) CreateDataset(dataset *xDatasetV1.Dataset) error {
	ins, err := db.getDB()
	if err != nil {
		return err
	}
	dataset.Spec.LastUpdate = metaV1.Now()
	dataset.CreationTimestamp = metaV1.Now()
	result, err := ins.Exec("INSERT INTO dataset (name, description,create_time, last_update) VALUES (?, ?, ?, ?)", dataset.Name, dataset.Spec.Description, dataset.CreationTimestamp.Time, dataset.Spec.LastUpdate.Time)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to crate dataset %s into database", dataset.Name), err)
		return err
	}
	_, err = result.LastInsertId()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed get create dataset %s result", dataset.Name), err)
	}
	return nil
}

// GetDataset implements IDataBaseDataset gets a dataset by name
func (db *Mysql) GetDataset(name string) (*xDatasetV1.Dataset, error) {
	inst, err := db.getDB()
	if err != nil {
		return nil, err
	}
	var dataset xDatasetV1.Dataset
	util.FillObjectGVK(&dataset)
	row := inst.QueryRow("SELECT name, description, create_time, last_update FROM dataset WHERE name = ?", name)
	err = row.Scan(&dataset.Name, &dataset.Spec.Description, &dataset.CreationTimestamp.Time, &dataset.Spec.LastUpdate.Time)
	if err != nil {
		if stdErr.Is(err, sql.ErrNoRows) {
			dataset.GetResourceVersion()
			return nil, errors.NewNotFound(schema.GroupResource{Group: xUserV1.GroupName, Resource: util.GetObjectKind(&dataset)}, name)
		}
		slog.Error(fmt.Sprintf("Failed to query dataset %s from database", name), err)
		return nil, err
	}
	return &dataset, nil
}

// UpdateDataset implements IDataBaseDataset updates a dataset
func (db *Mysql) UpdateDataset(dataset *xDatasetV1.Dataset) error {
	inst, err := db.getDB()
	if err != nil {
		return err
	}
	dataset.Spec.LastUpdate = metaV1.Now()
	result, err := inst.Exec("UPDATE dataset SET description = ?, last_update = ? WHERE name = ?", dataset.Spec.Description, dataset.Spec.LastUpdate.Time, dataset.Name)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to update dataset %s from database", dataset.Name), err)
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get update dataset %s result", dataset.Name)
	}
	if affected == 0 {
		return errors.NewNotFound(schema.GroupResource{Group: xDatasetV1.GroupName, Resource: util.GetObjectKind(&xDatasetV1.Dataset{})}, dataset.Name)
	}
	return nil
}

// DeleteDataset implements IDataBaseDataset deletes a dataset
func (db *Mysql) DeleteDataset(name string) error {
	inst, err := db.getDB()
	if err != nil {
		return err
	}
	result, err := inst.Exec("DELETE FROM dataset WHERE name = ?", name)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to delete dataset %s from database", name), err)
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed get delete dataset %s result", name), err)
		return err
	}
	if affected == 0 {
		return errors.NewNotFound(schema.GroupResource{Group: xUserV1.GroupName, Resource: util.GetObjectKind(&xDatasetV1.Dataset{})}, name)
	}
	return nil
}

// ListDatasets implements IDataBaseDataset lists all datasets
func (db *Mysql) ListDatasets() (xDatasetV1.DatasetList, error) {
	inst, err := db.getDB()
	if err != nil {
		return xDatasetV1.DatasetList{}, err
	}

	rows, err := inst.Query("SELECT name, description, create_time, last_update FROM dataset ")
	if err != nil {
		return xDatasetV1.DatasetList{}, err
	}
	defer rows.Close()
	var result xDatasetV1.DatasetList
	for rows.Next() {
		var dataset xDatasetV1.Dataset
		util.FillObjectGVK(&dataset)
		err = rows.Scan(&dataset.Name, &dataset.Spec.Description, &dataset.CreationTimestamp.Time, &dataset.Spec.LastUpdate.Time)
		if err != nil {
			return xDatasetV1.DatasetList{}, err
		}
		result.Items = append(result.Items, dataset)
	}

	return result, nil
}

// connectDB connects to mysql database and checks the connection
func (db *Mysql) connectDB() (*sql.DB, error) {
	dbCfg := db.Config.MySqlConfig
	dsnConfig := mysql.Config{
		User:      dbCfg.Username,
		Passwd:    dbCfg.Password,
		Net:       dbCfg.Protocol,
		Addr:      fmt.Sprintf("%s:%d", dbCfg.Address, dbCfg.Port),
		DBName:    dbCfg.DataBaseName,
		ParseTime: true,
	}

	inst, err := sql.Open("mysql", dsnConfig.FormatDSN())
	if err != nil {
		return nil, err
	}
	if err = inst.Ping(); err != nil {
		return nil, err
	}
	return inst, nil
}

// getDB gets the database connect pool instance
func (db *Mysql) getDB() (*sql.DB, error) {
	v, err, _ := db.singleGroup.Do("mysql_db", func() (interface{}, error) {
		var err error
		if db.instance == nil {
			db.instance, err = db.connectDB()
			return db.instance, err
		}
		if err = db.instance.Ping(); err != nil {
			slog.Error("Failed to ping mysql database", err)
			db.instance, err = db.connectDB()
			return db.instance, err
		}
		return db.instance, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*sql.DB), nil
}

// InitDb implements IDataBase init database
func (db *Mysql) InitDb() error {
	// for init we cannot use getDB() because database not been created yet
	inst, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/", db.Config.MySqlConfig.Username, db.Config.MySqlConfig.Password, db.Config.MySqlConfig.Address, db.Config.MySqlConfig.Port))
	if err != nil {
		return err
	}
	if err = inst.Ping(); err != nil {
		return err
	}
	defer inst.Close()

	// create database openhydra
	_, err = inst.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS openhydra CHARACTER SET %s COLLATE %s", db.Config.MySqlConfig.Character, db.Config.MySqlConfig.Collation))
	if err != nil {
		return err
	}

	// use database openhydra
	_, err = inst.Exec("USE openhydra")
	if err != nil {
		return err
	}

	_, err = inst.Exec("CREATE TABLE IF NOT EXISTS user ( id INT AUTO_INCREMENT PRIMARY KEY, username  VARCHAR(255), role INT , ch_name NVARCHAR(255) , description NVARCHAR(255) , email VARCHAR(255) , password VARCHAR(255) , UNIQUE (username) )")
	if err != nil {
		return err
	}
	_, err = inst.Exec("CREATE TABLE IF NOT EXISTS dataset ( id INT AUTO_INCREMENT PRIMARY KEY, name  VARCHAR(255), description NVARCHAR(255) , last_update DATETIME , create_time DATETIME , UNIQUE (name) )")
	if err != nil {
		return err
	}

	_, err = inst.Exec("CREATE TABLE IF NOT EXISTS course ( id INT AUTO_INCREMENT PRIMARY KEY, name  VARCHAR(255), description NVARCHAR(255) , created_by NVARCHAR(255) , last_update DATETIME , create_time DATETIME , file_size BIGINT , level INT , UNIQUE (name) )")
	if err != nil {
		return err
	}
	return nil
}

// CreateCourse implements IDataBaseCourse creates a new course
func (db *Mysql) CreateCourse(course *xCourseV1.Course) error {
	ins, err := db.getDB()
	if err != nil {
		return err
	}
	course.Spec.LastUpdate = metaV1.Now()
	course.CreationTimestamp = metaV1.Now()
	result, err := ins.Exec("INSERT INTO course (name, description, created_by, create_time, last_update, file_size, level) VALUES (?, ?, ?, ?, ?, ?, ?)", course.Name, course.Spec.Description, course.Spec.CreatedBy, course.CreationTimestamp.Time, course.Spec.LastUpdate.Time, course.Spec.Size, course.Spec.Level)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to crate course %s into database", course.Name), err)
		return err
	}
	_, err = result.LastInsertId()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed get create course %s result", course.Name), err)
	}
	return nil
}

// GetCourse implements IDataBaseCourse gets a course by name
func (db *Mysql) GetCourse(name string) (*xCourseV1.Course, error) {
	inst, err := db.getDB()
	if err != nil {
		return nil, err
	}
	var course xCourseV1.Course
	util.FillObjectGVK(&course)
	row := inst.QueryRow("SELECT name, description, created_by, create_time, last_update, file_size, level FROM course WHERE name = ?", name)
	err = row.Scan(&course.Name, &course.Spec.Description, &course.Spec.CreatedBy, &course.CreationTimestamp.Time, &course.Spec.LastUpdate.Time, &course.Spec.Size)
	if err != nil {
		if stdErr.Is(err, sql.ErrNoRows) {
			course.GetResourceVersion()
			return nil, errors.NewNotFound(schema.GroupResource{Group: xUserV1.GroupName, Resource: util.GetObjectKind(&course)}, name)
		}
		slog.Error(fmt.Sprintf("Failed to query course %s from database", name), err)
		return nil, err
	}
	return &course, nil
}

// UpdateCourse implements IDataBaseCourse updates a course
func (db *Mysql) UpdateCourse(course *xCourseV1.Course) error {
	inst, err := db.getDB()
	if err != nil {
		return err
	}
	course.Spec.LastUpdate = metaV1.Now()
	result, err := inst.Exec("UPDATE course SET description = ?, last_update = ? WHERE name = ?", course.Spec.Description, course.Spec.LastUpdate.Time, course.Name)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to update course %s from database", course.Name), err)
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get update course %s result", course.Name)
	}
	if affected == 0 {
		return errors.NewNotFound(schema.GroupResource{Group: xCourseV1.GroupName, Resource: util.GetObjectKind(&xCourseV1.Course{})}, course.Name)
	}
	return nil
}

// DeleteCourse implements IDataBaseCourse deletes a course
func (db *Mysql) DeleteCourse(name string) error {
	inst, err := db.getDB()
	if err != nil {
		return err
	}
	result, err := inst.Exec("DELETE FROM course WHERE name = ?", name)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to delete course %s from database", name), err)
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed get delete course %s result", name), err)
		return err
	}
	if affected == 0 {
		return errors.NewNotFound(schema.GroupResource{Group: xUserV1.GroupName, Resource: util.GetObjectKind(&xCourseV1.Course{})}, name)
	}
	return nil
}

// ListCourses implements IDataBaseCourse lists all courses
func (db *Mysql) ListCourses() (xCourseV1.CourseList, error) {
	inst, err := db.getDB()
	if err != nil {
		return xCourseV1.CourseList{}, err
	}

	rows, err := inst.Query("SELECT name, description, created_by, create_time, last_update, file_size FROM course ")
	if err != nil {
		return xCourseV1.CourseList{}, err
	}
	defer rows.Close()
	var result xCourseV1.CourseList
	for rows.Next() {
		var course xCourseV1.Course
		util.FillObjectGVK(&course)
		err = rows.Scan(&course.Name, &course.Spec.Description, &course.Spec.CreatedBy, &course.CreationTimestamp.Time, &course.Spec.LastUpdate.Time, &course.Spec.Size)
		if err != nil {
			return xCourseV1.CourseList{}, err
		}
		result.Items = append(result.Items, course)
	}

	return result, nil
}
