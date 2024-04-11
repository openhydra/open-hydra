package auth_plugin

import (
	"database/sql"
	stdErr "errors"
	"fmt"
	"log/slog"
	xUserV1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"
	"open-hydra/pkg/util"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type DefaultMysqlAuthPlugin struct {
	Db func() (*sql.DB, error)
}

func (db *DefaultMysqlAuthPlugin) CreateUser(user *xUserV1.OpenHydraUser) error {
	inst, err := db.Db()
	if err != nil {
		return err
	}
	defer inst.Close()
	_, err = inst.Exec("INSERT INTO user (username, email, password, ch_name, description, role) VALUES (?, ?, ?, ?, ?, ?)", user.Name, user.Spec.Email, user.Spec.Password, user.Spec.ChineseName, user.Spec.Description, user.Spec.Role)
	if err != nil {
		return err
	}

	return nil
}

// GetUser implements IDataBaseUser gets a user by name
func (db *DefaultMysqlAuthPlugin) GetUser(name string) (*xUserV1.OpenHydraUser, error) {
	inst, err := db.Db()
	if err != nil {
		return nil, err
	}
	var user xUserV1.OpenHydraUser
	util.FillObjectGVK(&user)
	row := inst.QueryRow("SELECT username, email, password, ch_name, description, role FROM user WHERE username = ?", name)
	err = row.Scan(&user.Name, &user.Spec.Email, &user.Spec.Password, &user.Spec.ChineseName, &user.Spec.Description, &user.Spec.Role)
	if err != nil {
		if stdErr.Is(err, sql.ErrNoRows) {
			user.GetResourceVersion()
			return nil, errors.NewNotFound(schema.GroupResource{Group: xUserV1.GroupName, Resource: util.GetObjectKind(&user)}, name)
		}
		slog.Error(fmt.Sprintf("Failed to query user %s from database", name), err)
		return nil, err
	}
	return &user, nil
}

// UpdateUser implements IDataBaseUser updates a user
func (db *DefaultMysqlAuthPlugin) UpdateUser(user *xUserV1.OpenHydraUser) error {
	inst, err := db.Db()
	if err != nil {
		return err
	}
	_, err = inst.Exec("UPDATE user SET email = ?, password = ?, ch_name = ?, description = ?, role = ? WHERE username = ?", user.Spec.Email, user.Spec.Password, user.Spec.ChineseName, user.Spec.Description, user.Spec.Role, user.Name)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to update user %s from database", user.Name), err)
		return err
	}
	return nil
}

// DeleteUser implements IDataBaseUser deletes a user
func (db *DefaultMysqlAuthPlugin) DeleteUser(name string) error {
	inst, err := db.Db()
	if err != nil {
		return err
	}
	result, err := inst.Exec("DELETE FROM user WHERE username = ?", name)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to delete user %s from database", name), err)
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed get delete user %s result", name), err)
		return err
	}
	if affected == 0 {
		return errors.NewNotFound(schema.GroupResource{Group: xUserV1.GroupName, Resource: "OpenHydraUser"}, name)
	}
	return nil
}

// ListUsers implements IDataBaseUser lists all users
func (db *DefaultMysqlAuthPlugin) ListUsers() (xUserV1.OpenHydraUserList, error) {
	inst, err := db.Db()
	if err != nil {
		return xUserV1.OpenHydraUserList{}, err
	}
	rows, err := inst.Query("SELECT username, email, password, ch_name, description, role FROM user")
	if err != nil {
		return xUserV1.OpenHydraUserList{}, err
	}
	defer rows.Close()
	result := xUserV1.OpenHydraUserList{}
	for rows.Next() {
		var user xUserV1.OpenHydraUser
		util.FillObjectGVK(&user)
		err = rows.Scan(&user.Name, &user.Spec.Email, &user.Spec.Password, &user.Spec.ChineseName, &user.Spec.Description, &user.Spec.Role)
		if err != nil {
			return xUserV1.OpenHydraUserList{}, err
		}
		result.Items = append(result.Items, user)
	}

	return result, nil
}

func (db *DefaultMysqlAuthPlugin) LoginUser(name, password string) (*xUserV1.OpenHydraUser, error) {
	inst, err := db.Db()
	if err != nil {
		return nil, err
	}
	defer inst.Close()
	rows, err := inst.Query("SELECT username, email, password, ch_name, description, role FROM user WHERE username = ? AND password = ?", name, password)
	if err != nil {
		return nil, err
	}
	var user xUserV1.OpenHydraUser
	util.FillObjectGVK(&user)
	for rows.Next() {
		err = rows.Scan(&user.Name, &user.Spec.Email, &user.Spec.Password, &user.Spec.ChineseName, &user.Spec.Description, &user.Spec.Role)
		if err != nil {
			return nil, err
		}
	}

	if user.Name == "" {
		return nil, fmt.Errorf("user %s not found", name)
	}

	return &user, nil
}
