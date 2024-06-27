# openhydra interfacing with keystone (alpha feature)

This document describes how to interface openhydra with keystone for authentication. Currently, we only support storing openhydra user information in the keystone user structure.

## Before you begin

* openhydra is running normally
* The keystone service has been deployed and is working properly
* Note that after interfacing with keystone, the username in keystone will be used as the account in openhydra

## Modify the openhydra configuration file

Log in to the server and run the following command

```bash
# Backup the openhydra configuration file
$ kubectl get cm -n open-hydra open-hydra-config -o yaml > open-hydra-config.yaml
# Modify the openhydra configuration file
$ kubectl edit cm -n open-hydra open-hydra-config
# Add the following configuration under the key "resourceNamespace" and save and exit
# Replace keystone-user with your keystone account and password. We recommend that you use the admin account here
authDelegateConfig:
  keystoneConfig:
    endpoint: http://[keystone ip]:[port]
    username: keystone-user
    password: keystone-user-password
    domainId: default

# Restart the openhydra server
$ kubectl scale deployment -n open-hydra open-hydra-server --replicas=0
$ kubectl scale deployment -n open-hydra open-hydra-server --replicas=1
```

## Verify the results

Run the following command on the server. Note that if you try to delete the admin and service accounts after integrating with keystone, the operation will be rejected

```bash
# List all users
$ kubectl get openhydrausers
# Output
NAME                               AGE
595af254b5fe491fb7fa2c0a42cb299b   <unknown>
19ff8aee587f4dbeabff2fb88003c6d7   <unknown>
0c3eee2a36da4c2d8ba8b3a073eaeec3   <unknown>
e521cf70761d432ba09d093257ac7e35   <unknown>
4bce15e7abc240c3860bca2c0aaa271b   <unknown>
d75eaa9345984278ad6e6106f48c0d30   <unknown>

# Creating a user
# Since keystone uses the id as the unique primary key, which conflicts with the k8s concept, we will change the id to the login name. When creating a user,
# everything is the same as usual, but when you see the return, the id in keystone will be used as the account
$ cat <<EOF > user1.yaml
apiVersion: open-hydra-server.openhydra.io/v1
kind: OpenHydraUser
metadata:
  name: user1
spec:
  password: Openhydra123
  role: 2
EOF


# Verify the results
$ kubectl get openhydrausers -o custom-columns=Name:metadata.name,User:.spec.chineseName
# Output
Name                               User
.......
18ca18f419e54de7b040617b5c065c7f   user1


# Attempt to log in
# Open the openhydra access page and enter the username 18ca18f419e54de7b040617b5c065c7f and password Openhydra123

# Delete user
# note openhydra will refused to delete user admin for security reasons when use keystone as the authentication plugin
$ kubectl delete openhydrauser 18ca18f419e54de7b040617b5c065c7f
```