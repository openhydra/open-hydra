package train

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"open-hydra/cmd/open-hydra-server/app/config"
	xUserV1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"
	"open-hydra/pkg/util"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type KeystoneAuthPlugin struct {
	Config *config.OpenHydraServerConfig
	token  string
}

// implement IDataBaseUser

func (k *KeystoneAuthPlugin) CreateUser(user *xUserV1.OpenHydraUser) error {

	_, err := k.GetUser(user.ObjectMeta.Name)
	if err != nil {
		if !errors.IsNotFound(err) {
			slog.Error("Failed to check user existence", err)
			return err
		}
	}

	// for keystone we keep name in chinese name
	user.Spec.ChineseName = user.ObjectMeta.Name
	util.FillKindAndApiVersion(&user.TypeMeta, "OpenHydraUser")

	userPost := &User{
		Name:          user.ObjectMeta.Name,
		Email:         user.Spec.Email,
		Password:      user.Spec.Password,
		Enabled:       true,
		OpenhydraUser: user,
		Options: Options{
			IgnorePasswordExpiry:             true,
			IgnoreChangePasswordUponFirstUse: true,
			IgnoreLockoutFailureAttempts:     true,
		},
	}

	postBody, err := json.Marshal(&struct {
		User *User `json:"user"`
	}{User: userPost})
	if err != nil {
		slog.Error("Failed to marshal user", err)
		return err
	}

	_, _, _, err = k.commentRequestAutoRenewToken("/v3/users", http.MethodPost, postBody)
	if err != nil {
		slog.Error("Failed to create user", err)
		return err
	}

	return nil
}

// Get a user by name
func (k *KeystoneAuthPlugin) GetUser(name string) (*xUserV1.OpenHydraUser, error) {
	body, _, code, err := k.commentRequestAutoRenewToken(fmt.Sprintf("/v3/users/%s", name), http.MethodGet, nil)
	if code == http.StatusNotFound {
		return nil, errors.NewNotFound(xUserV1.Resource("user"), name)
	}
	if err != nil {
		slog.Error("Failed to get user", err)
		return nil, err
	}

	var userContainer struct {
		User *User `json:"user"`
	}
	err = json.Unmarshal(body, &userContainer)
	if err != nil {
		slog.Error("Failed to unmarshal user", err)
		return nil, err
	}

	var result *xUserV1.OpenHydraUser

	if userContainer.User.OpenhydraUser == nil {
		meta := &metaV1.TypeMeta{}
		util.FillKindAndApiVersion(meta, "OpenHydraUser")
		result = &xUserV1.OpenHydraUser{
			TypeMeta:   *meta,
			ObjectMeta: metaV1.ObjectMeta{Name: userContainer.User.ID, UID: types.UID(userContainer.User.ID)}, // note keep keystone user id in object meta UID is important
			Spec: xUserV1.OpenHydraUserSpec{
				Email:       userContainer.User.Email,
				ChineseName: userContainer.User.Name, // put keystone account as display name
				Description: "keystone user",
				Password:    "*********",
				Role:        1, // all user consider as a admin in openhydra which OpenhydraUser is set to nil in keystone
			},
		}
	} else {
		userContainer.User.OpenhydraUser.ObjectMeta.UID = types.UID(userContainer.User.ID)
		result = userContainer.User.OpenhydraUser
	}

	return result, nil
}

// Update a user
func (k *KeystoneAuthPlugin) UpdateUser(user *xUserV1.OpenHydraUser) error {
	// todo: update user
	// so for openhydra do not support update user, we just delete and recreate it
	// we leave it blank here
	return nil
}

// Delete a user
func (k *KeystoneAuthPlugin) DeleteUser(name string) error {
	// we cannot delete admin user in keystone which will cause a big mess here
	if name == "admin" || name == "service" {
		return fmt.Errorf("build in user can not be deleted")
	}
	_, _, _, err := k.commentRequestAutoRenewToken(fmt.Sprintf("/v3/users/%s", name), http.MethodDelete, nil)
	if err != nil {
		slog.Error("Failed to delete user", err)
		return err
	}
	return nil
}

// List all users
func (k *KeystoneAuthPlugin) ListUsers() (xUserV1.OpenHydraUserList, error) {
	body, _, _, err := k.commentRequestAutoRenewToken("/v3/users", http.MethodGet, nil)
	if err != nil {
		slog.Error("Failed to list users", err)
		return xUserV1.OpenHydraUserList{}, err
	}

	var userCollection UserContainer
	err = json.Unmarshal(body, &userCollection)
	if err != nil {
		slog.Error("Failed to unmarshal users", err)
		return xUserV1.OpenHydraUserList{}, err
	}

	var userList xUserV1.OpenHydraUserList

	for index, user := range userCollection.Users {
		if !user.Enabled {
			// skip disabled user
			continue
		}
		if user.OpenhydraUser == nil {
			meta := &metaV1.TypeMeta{}
			util.FillKindAndApiVersion(meta, "OpenHydraUser")
			// we use id as account in openhydra to fit openhydra user management behavior
			userList.Items = append(userList.Items, xUserV1.OpenHydraUser{
				TypeMeta:   *meta,
				ObjectMeta: metaV1.ObjectMeta{Name: user.ID, UID: types.UID(user.ID)}, // note keep keystone user id in object meta UID is important
				Spec: xUserV1.OpenHydraUserSpec{
					Email:       user.Email,
					Password:    "*********",
					ChineseName: user.Name, // put keystone account as display name
					Description: "keystone user",
					Role:        1, // all user consider as a admin in openhydra which OpenhydraUser is set to nil in keystone
				},
			})
		} else {
			userCollection.Users[index].OpenhydraUser.ObjectMeta.UID = types.UID(user.ID)
			userCollection.Users[index].OpenhydraUser.Name = user.ID
			userList.Items = append(userList.Items, *userCollection.Users[index].OpenhydraUser)
		}
	}

	return userList, nil
}

// Login a user
func (k *KeystoneAuthPlugin) LoginUser(name, password string) (*xUserV1.OpenHydraUser, error) {
	// for keystone the name is keystone id
	// so we have to get the user id first
	user, err := k.GetUser(name)
	if err != nil {
		slog.Error("Failed to get user", err)
		return nil, err
	}

	_, _, err = k.RequestToken(user.Spec.ChineseName, password, false)
	if err != nil {
		slog.Error("Failed to login user", err)
		return nil, err
	}

	return user, nil
}

func (k *KeystoneAuthPlugin) RequestToken(name, password string, enableScoped bool) (string, *TokenResponse, error) {
	authReq := &AuthRequest{
		Auth: AuthDetails{
			Identity: IdentityDetails{
				Methods: []string{"password"},
				Password: PasswordDetails{
					User: User{
						Name:     name,
						Password: password,
						Domain: &Domain{
							Id: k.getDomainId(),
						},
					},
				},
			},
		},
	}

	if enableScoped {
		authReq.Auth.Scope = &ScopeDetails{
			Domain: DomainDetails{
				Id: k.getDomainId(),
			},
		}
	}

	postBody, err := json.Marshal(authReq)
	if err != nil {
		slog.Error("Failed to marshal auth request", err)
		return "", nil, err
	}

	resp, header, _, err := util.CommonRequest(k.buildPath("/v3/auth/tokens"), http.MethodPost, "", postBody, nil, false, false, 3*time.Second)
	if err != nil {
		slog.Error("Failed to request token", err)
		return "", nil, err
	}

	key := util.GetStringValueOrDefault("Token in request", k.Config.AuthDelegateConfig.KeystoneConfig.TokenKeyInResponse, "X-Subject-Token")

	token := header.Get(key)
	if token == "" {
		slog.Error("No token in response header")
		return "", nil, fmt.Errorf("no token in response header")
	}

	tokenResp := &TokenResponse{}
	err = json.Unmarshal(resp, tokenResp)
	if err != nil {
		slog.Error("Failed to unmarshal token response", err)
		return "", nil, err
	}

	return token, tokenResp, nil
}

func (k *KeystoneAuthPlugin) getToken() (string, error) {
	if k.token == "" {
		token, _, err := k.RequestToken(k.Config.AuthDelegateConfig.KeystoneConfig.Username, k.Config.AuthDelegateConfig.KeystoneConfig.Password, true)
		if err != nil {
			return "", err
		}
		k.token = token
	}
	return k.token, nil
}

func (k *KeystoneAuthPlugin) getDomainId() string {
	if k.Config.AuthDelegateConfig.KeystoneConfig.DomainId == "" {
		return "default"
	}
	return k.Config.AuthDelegateConfig.KeystoneConfig.DomainId
}

func (k *KeystoneAuthPlugin) buildPath(path string) string {
	baseURL, _ := url.Parse(k.Config.AuthDelegateConfig.KeystoneConfig.Endpoint)
	endpoint, _ := url.Parse(path)
	return baseURL.ResolveReference(endpoint).String()
}

func (k *KeystoneAuthPlugin) commentRequestAutoRenewToken(path, method string, body json.RawMessage) ([]byte, http.Header, int, error) {
	tokenHeaderKey := util.GetStringValueOrDefault("Token header key", k.Config.AuthDelegateConfig.KeystoneConfig.TokenKeyInRequest, "X-Auth-Token")
	token, err := k.getToken()
	if err != nil {
		slog.Error("Failed to get token", err)
		return nil, nil, -1, err
	}
	reqURL := k.buildPath(path)
	result, header, code, err := util.CommonRequest(reqURL, method, "", body, map[string]string{tokenHeaderKey: token}, false, false, 3*time.Second)

	if code == http.StatusUnauthorized {
		slog.Warn("Token may expired, attempt to renew the token and retry for one shot")
		// token may expired, request a new one
		newToken, _, err := k.RequestToken(k.Config.AuthDelegateConfig.KeystoneConfig.Username, k.Config.AuthDelegateConfig.KeystoneConfig.Password, true)
		if err != nil {
			slog.Error("Failed to renew token", err)
			return nil, nil, -1, err
		}
		k.token = newToken
		return util.CommonRequest(reqURL, method, "", body, map[string]string{tokenHeaderKey: newToken}, false, false, 3*time.Second)
	}

	if err != nil {
		slog.Error(fmt.Sprintf("Failed to request %s", reqURL), err)
		return nil, nil, -1, err
	}

	if code != http.StatusOK && code != http.StatusCreated && code != http.StatusNoContent {
		slog.Error(fmt.Sprintf("Failed to request %s, response code: %d", reqURL, code))
		return nil, nil, code, fmt.Errorf("failed to request %s, response code: %d", reqURL, code)
	}

	// if everything is fine, return the result
	return result, header, code, nil
}
