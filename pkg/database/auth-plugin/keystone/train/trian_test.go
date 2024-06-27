package train

import (
	"encoding/json"
	"io"
	"net/http"
	"open-hydra/cmd/open-hydra-server/app/config"
	xUserV1 "open-hydra/pkg/apis/open-hydra-api/user/core/v1"
	"open-hydra/pkg/util"
	"time"

	"github.com/emicklei/go-restful"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var typeMetaOpenhydraUser = metaV1.TypeMeta{
	Kind:       "OpenHydraUser",
	APIVersion: "open-hydra-server.openhydra.io/v1",
}

var testUserList = UserContainer{[]User{
	{ID: "test1id", Name: "test1", Password: "test1", Email: "test1@maas.com", Enabled: true},
	{ID: "test2id", Name: "test2", Password: "test2", Enabled: true, OpenhydraUser: &xUserV1.OpenHydraUser{
		TypeMeta: typeMetaOpenhydraUser,
		ObjectMeta: metaV1.ObjectMeta{
			Name: "test2",
		},
		Spec: xUserV1.OpenHydraUserSpec{
			Email:    "test2@openhydra.com",
			Password: "test2",
			Role:     2,
		},
	}},
	{ID: "test3id", Name: "test3", Password: "test3", Email: "test3@maas.com", Enabled: true},
	{ID: "adminid", Name: "admin", Password: "admin", Email: "test3@maas.com", Enabled: true},
},
}

var testRouter = func(ws *restful.WebService) {
	ws.Route(ws.POST("/v3/auth/tokens").To(func(request *restful.Request, response *restful.Response) {
		body, err := io.ReadAll(request.Request.Body)
		if err != nil {
			response.WriteErrorString(http.StatusInternalServerError, err.Error())
			return
		}
		var auth AuthRequest
		err = json.Unmarshal(body, &auth)
		if err != nil {
			response.WriteErrorString(http.StatusBadRequest, "failed to parse request body")
			return
		}

		if auth.Auth.Identity.Password.User.Name == "admin" && auth.Auth.Identity.Password.User.Password == "admin" {
			response.AddHeader("X-Subject-Token", "test-token")
			response.WriteAsJson(&TokenResponse{
				Token: Token{
					Methods:   []string{"password"},
					User:      User{Name: "admin", ID: "admin"},
					AuditIds:  []string{"test"},
					ExpiresAt: "2021-08-10T14:00:00Z",
					IssuedAt:  "2021-08-10T13:00:00Z",
					Domain:    Domain{Name: "default"},
					Roles:     []Role{{ID: "test", Name: "test"}},
					Catalog:   []Catalog{{Endpoints: []Endpoint{{ID: "test", Interface: "public", RegionID: "test", URL: "http://test", Region: "test"}}, ID: "test", Type: "test", Name: "test"}},
				},
			})
			return
		}
		response.WriteHeader(http.StatusUnauthorized)
	}))
	ws.Route(ws.GET("/v3/users").To(func(request *restful.Request, response *restful.Response) {
		token := request.HeaderParameter("X-Auth-Token")
		if token != "test-token" {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		usersJson, err := json.Marshal(testUserList)
		if err != nil {
			response.WriteErrorString(http.StatusInternalServerError, err.Error())
			return
		}
		response.Write(usersJson)
	}))
	ws.Route(ws.GET("/v3/users/{user_id}").To(func(request *restful.Request, response *restful.Response) {
		token := request.HeaderParameter("X-Auth-Token")
		if token != "test-token" {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		userID := request.PathParameter("user_id")
		if userID == "" {
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		if userID == "test1id" {
			userJson, err := json.Marshal(&struct {
				User *User `json:"user"`
			}{User: &testUserList.Users[0]})
			if err != nil {
				response.WriteErrorString(http.StatusInternalServerError, err.Error())
				return
			}
			response.Write(userJson)
			return
		}
		if userID == "test2id" {
			userJson, err := json.Marshal(&struct {
				User *User `json:"user"`
			}{User: &testUserList.Users[1]})
			if err != nil {
				response.WriteErrorString(http.StatusInternalServerError, err.Error())
				return
			}
			response.Write(userJson)
			return
		}
		if userID == "test3id" {
			userJson, err := json.Marshal(&struct {
				User *User `json:"user"`
			}{User: &testUserList.Users[2]})
			if err != nil {
				response.WriteErrorString(http.StatusInternalServerError, err.Error())
				return
			}
			response.Write(userJson)
			return
		}
		if userID == "adminid" {
			userJson, err := json.Marshal(&struct {
				User *User `json:"user"`
			}{User: &User{Name: "admin", ID: "admin", Password: "admin"}})
			if err != nil {
				response.WriteErrorString(http.StatusInternalServerError, err.Error())
				return
			}
			response.Write(userJson)
			return
		}
		response.WriteHeader(http.StatusNotFound)
	}))
	ws.Route(ws.POST("/v3/users").To(func(request *restful.Request, response *restful.Response) {
		token := request.HeaderParameter("X-Auth-Token")
		if token != "test-token" {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		body, err := io.ReadAll(request.Request.Body)
		if err != nil {
			response.WriteErrorString(http.StatusInternalServerError, err.Error())
			return
		}
		var userContainer struct {
			User *User `json:"user"`
		}
		err = json.Unmarshal(body, &userContainer)
		if err != nil {
			response.WriteErrorString(http.StatusBadRequest, "failed to parse request body")
			return
		}

		if userContainer.User.Name == "test1" || userContainer.User.Name == "test2" || userContainer.User.Name == "test3" {
			response.WriteErrorString(http.StatusConflict, "user already exists")
			return
		}

		response.WriteHeader(http.StatusCreated)
	}))
	ws.Route(ws.DELETE("/v3/users/{user_id}").To(func(request *restful.Request, response *restful.Response) {
		token := request.HeaderParameter("X-Auth-Token")
		if token != "test-token" {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		userID := request.PathParameter("user_id")
		if userID == "" {
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		response.WriteHeader(http.StatusOK)
	}))
}

var _ = Describe("open-hydra-server util test", func() {
	var serverConfig *config.OpenHydraServerConfig
	var keystone *KeystoneAuthPlugin
	BeforeEach(func() {
		serverConfig = &config.OpenHydraServerConfig{
			AuthDelegateConfig: &config.AuthDelegateConfig{
				KeystoneConfig: &config.KeystoneConfig{
					Endpoint:  "http://localhost:20081",
					Username:  "admin",
					Password:  "admin",
					DomainId:  "default",
					ProjectId: "default",
				},
			},
		}
		keystone = &KeystoneAuthPlugin{
			Config: serverConfig,
		}
	})
	Describe("RequestToken test", func() {
		It("should be expected", func() {
			stopChan := make(chan struct{}, 1)
			go util.StartMockServer(20081, testRouter, stopChan)
			time.Sleep(2 * time.Second)
			defer close(stopChan)
			token, tokenResp, err := keystone.RequestToken(serverConfig.AuthDelegateConfig.KeystoneConfig.Username, serverConfig.AuthDelegateConfig.KeystoneConfig.Password, true)
			Expect(err).To(BeNil())
			Expect(token).To(Equal("test-token"))
			Expect(tokenResp.Token.User.Name).To(Equal("admin"))

		})
	})

	Describe("getDomain test", func() {
		It("should be the default value", func() {
			Expect(keystone.getDomainId()).To(Equal("default"))
		})

		It("should be equal to test", func() {
			serverConfig.AuthDelegateConfig.KeystoneConfig.DomainId = "test"
			Expect(keystone.getDomainId()).To(Equal("test"))
		})
	})

	Describe("getToken test", func() {
		It("should be expected", func() {
			stopChan := make(chan struct{}, 1)
			go util.StartMockServer(20081, testRouter, stopChan)
			time.Sleep(2 * time.Second)
			defer close(stopChan)
			token, err := keystone.getToken()
			Expect(err).To(BeNil())
			Expect(token).To(Equal("test-token"))
			token, err = keystone.getToken()
			Expect(err).To(BeNil())
			Expect(token).To(Equal("test-token"))
		})
	})

	Describe("buildPath test", func() {
		It("should be expected", func() {
			Expect(keystone.buildPath("/v3/auth/tokens")).To(Equal("http://localhost:20081/v3/auth/tokens"))
			Expect(keystone.buildPath("v3/users")).To(Equal("http://localhost:20081/v3/users"))
		})
	})

	Describe("commentRequestAutoRenewToken test", func() {
		It("should be expected", func() {
			serverConfig.AuthDelegateConfig.KeystoneConfig.Endpoint = "http://localhost:20082"
			keystone.token = "wrong token"
			stopChan := make(chan struct{}, 1)
			go util.StartMockServer(20082, testRouter, stopChan)
			time.Sleep(2 * time.Second)
			defer close(stopChan)
			_, _, httpCode, err := keystone.commentRequestAutoRenewToken("/v3/users", http.MethodGet, nil)
			Expect(err).To(BeNil())
			Expect(httpCode).To(Equal(http.StatusOK))

		})
	})

	Describe("GetRawKeystoneUserList test", func() {
		It("should be expected", func() {
			serverConfig.AuthDelegateConfig.KeystoneConfig.Endpoint = "http://localhost:20088"
			stopChan := make(chan struct{}, 1)
			go util.StartMockServer(20088, testRouter, stopChan)
			time.Sleep(2 * time.Second)
			defer close(stopChan)
			users, err := keystone.GetRawKeystoneUserList()
			Expect(err).To(BeNil())
			Expect(len(users.Users)).To(Equal(len(testUserList.Users)))
		})
	})

	Describe("GetUserIdFromName test", func() {
		It("should be expected", func() {
			serverConfig.AuthDelegateConfig.KeystoneConfig.Endpoint = "http://localhost:20089"
			stopChan := make(chan struct{}, 1)
			go util.StartMockServer(20089, testRouter, stopChan)
			time.Sleep(2 * time.Second)
			defer close(stopChan)
			id, err := keystone.GetUserIdFromName("test1")
			Expect(err).To(BeNil())
			Expect(id).To(Equal("test1id"))
		})
	})

	Describe("ListUsers test", func() {
		It("should be expected", func() {
			serverConfig.AuthDelegateConfig.KeystoneConfig.Endpoint = "http://localhost:20083"
			stopChan := make(chan struct{}, 1)
			go util.StartMockServer(20083, testRouter, stopChan)
			time.Sleep(2 * time.Second)
			defer close(stopChan)
			users, err := keystone.ListUsers()
			Expect(err).To(BeNil())
			Expect(len(users.Items)).To(Equal(len(testUserList.Users)))
			Expect(users.Items[0].Name).To(Equal(testUserList.Users[0].Name))
			Expect(users.Items[0].Spec.ChineseName).To(Equal(testUserList.Users[0].Name))
			Expect(users.Items[0].Spec.Password).To(Equal("*********"))
			Expect(users.Items[0].Spec.Role).To(Equal(1))
			Expect(users.Items[0].TypeMeta).To(Equal(typeMetaOpenhydraUser))
			Expect(users.Items[0].Spec.Email).To(Equal(testUserList.Users[0].Email))
			Expect(string(users.Items[0].UID)).To(Equal(testUserList.Users[0].ID))
			Expect(users.Items[1].Name).To(Equal(testUserList.Users[1].Name))
			Expect(users.Items[1].Spec.ChineseName).To(Equal(""))
			Expect(users.Items[1].Spec.Password).To(Equal(testUserList.Users[1].Password))
			Expect(users.Items[1].Spec.Role).To(Equal(2))
			Expect(users.Items[1].TypeMeta).To(Equal(typeMetaOpenhydraUser))
			Expect(users.Items[1].Spec.Email).To(Equal(testUserList.Users[1].OpenhydraUser.Spec.Email))
			Expect(string(users.Items[1].UID)).To(Equal(testUserList.Users[1].ID))
			Expect(users.Items[2].Name).To(Equal(testUserList.Users[2].Name))
			Expect(users.Items[2].Spec.ChineseName).To(Equal(testUserList.Users[2].Name))
			Expect(users.Items[2].Spec.Password).To(Equal("*********"))
			Expect(users.Items[2].Spec.Role).To(Equal(1))
			Expect(users.Items[2].TypeMeta).To(Equal(typeMetaOpenhydraUser))
			Expect(users.Items[2].Spec.Email).To(Equal(testUserList.Users[2].Email))
			Expect(string(users.Items[2].UID)).To(Equal(testUserList.Users[2].ID))
		})
	})

	Describe("GetUser test", func() {
		It("should be expected", func() {
			serverConfig.AuthDelegateConfig.KeystoneConfig.Endpoint = "http://localhost:20084"
			stopChan := make(chan struct{}, 1)
			go util.StartMockServer(20084, testRouter, stopChan)
			time.Sleep(2 * time.Second)
			defer close(stopChan)
			user, err := keystone.GetUser("test1")
			Expect(err).To(BeNil())
			Expect(string(user.UID)).To(Equal(testUserList.Users[0].ID))
			Expect(user.Name).To(Equal(testUserList.Users[0].Name))
			Expect(user.Spec.ChineseName).To(Equal(testUserList.Users[0].Name))
			Expect(user.Spec.Password).To(Equal("*********"))
			Expect(user.Spec.Role).To(Equal(1))
			Expect(user.TypeMeta).To(Equal(typeMetaOpenhydraUser))
			Expect(user.Spec.Email).To(Equal(testUserList.Users[0].Email))
			Expect(string(user.UID)).To(Equal(testUserList.Users[0].ID))
		})
	})

	Describe("CreateUser test", func() {
		It("should be expected", func() {
			serverConfig.AuthDelegateConfig.KeystoneConfig.Endpoint = "http://localhost:20085"
			stopChan := make(chan struct{}, 1)
			go util.StartMockServer(20085, testRouter, stopChan)
			time.Sleep(2 * time.Second)
			defer close(stopChan)
			userToCreate := &xUserV1.OpenHydraUser{
				TypeMeta: typeMetaOpenhydraUser,
				ObjectMeta: metaV1.ObjectMeta{
					Name: "test4",
				},
				Spec: xUserV1.OpenHydraUserSpec{
					Email:    "",
					Password: "test4",
					Role:     1,
				},
			}
			err := keystone.CreateUser(userToCreate)
			Expect(err).To(BeNil())
			// test conflict user will be rejected
			userToCreate.Name = "test1"
			err = keystone.CreateUser(userToCreate)
			Expect(err).NotTo(BeNil())
		})
	})

	Describe("DeleteUser test", func() {
		It("should be expected", func() {
			serverConfig.AuthDelegateConfig.KeystoneConfig.Endpoint = "http://localhost:20086"
			stopChan := make(chan struct{}, 1)
			go util.StartMockServer(20086, testRouter, stopChan)
			time.Sleep(2 * time.Second)
			defer close(stopChan)
			err := keystone.DeleteUser("admin")
			Expect(err).NotTo(BeNil())
			err = keystone.DeleteUser("test1")
			Expect(err).To(BeNil())
		})
	})

	Describe("LoginUser test", func() {
		It("should be expected", func() {
			serverConfig.AuthDelegateConfig.KeystoneConfig.Endpoint = "http://localhost:20087"
			stopChan := make(chan struct{}, 1)
			go util.StartMockServer(20087, testRouter, stopChan)
			time.Sleep(2 * time.Second)
			defer close(stopChan)
			user, err := keystone.LoginUser("admin", "admin")
			Expect(err).To(BeNil())
			Expect(user.Name).To(Equal("admin"))
			Expect(user.Spec.Password).To(Equal("*********"))
		})
	})
})
