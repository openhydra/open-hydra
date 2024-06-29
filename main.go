package main

import (
	"fmt"
	"open-hydra/cmd/open-hydra-server/app/config"
	db "open-hydra/pkg/database"
)

func main() {
	fmt.Println("Hello, World!")
	testConfig := config.DefaultConfig()
	testConfig.AuthDelegateConfig = &config.AuthDelegateConfig{}
	testConfig.AuthDelegateConfig.KeystoneConfig = &config.KeystoneConfig{
		Endpoint: "http://61.241.103.49:30500",
		Username: "admin",
		Password: "a0aVLOT5jUkI7z94tw54dYlX6GZ7BETe",
		DomainId: "default",
	}
	newDB := db.NewMysql(testConfig)
	userList, err := newDB.ListUsers()
	if err != nil {
		fmt.Println(err)
	}

	userAdmin, err := newDB.GetUser("admin")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(userAdmin)

	fmt.Println(userList)

	// err = newDB.CreateUser(&userV1.OpenHydraUser{
	// 	ObjectMeta: v1.ObjectMeta{
	// 		Name: "test",
	// 	},
	// 	Spec: userV1.OpenHydraUserSpec{
	// 		Password: "Thinkbig1",
	// 		Email:    "test@test.com",
	// 		Role:     1,
	// 	},
	// })

	// if err != nil {
	// 	fmt.Println(err)
	// }

	err = newDB.DeleteUser("test")
	if err != nil {
		fmt.Println(err)
	}
}
