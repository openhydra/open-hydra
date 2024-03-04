# open-hydra server is a api server for handling open-hydra requests

## how to debug it locally without actually deploy it to k8s

* k8s cluster 1.23+
  * mysql-operator
* golang 1.21.4+

```bash
# create service account
$ kubectl create sa admin
$ kubectl create rolebinding my-service-account-cluster-admin --clusterrole=cluster-admin --serviceaccount=default:admin
$ kubectl create rolebinding -n kube-system my-service-account-cluster-admin --clusterrole=cluster-admin --serviceaccount=default:admin
$ kubectl create clusterrolebinding  my-service-account-cluster-admin --clusterrole=cluster-admin --serviceaccount=default:admin

# create a pod with service account to fake the pod environment locally
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  serviceAccountName: admin
  containers:
  - name: my-container
    image: centos
    command: ["init"]

# copy ca.crt namespace token in /var/run/secrets/kubernetes.io/serviceaccount to local
$ kubectl exec -it my-pod -- bash

# create mysql cluster with operator
$ kubectl apply -f asserts/mysql-deploy-crds.yaml
$ kubectl apply -f asserts/mysql-deploy-operator.yaml

# ensure mysql operator it's ready
# then create mysql cluster
$ kubectl apply -f asserts/mysql-aio-cluster.yaml

# 检查是否都启动成功了
$ kubectl get pods -n mysql-operator 
NAME                                READY   STATUS    RESTARTS   AGE
mycluster-0                         2/2     Running   0          18h
mycluster-router-5d74f97d5b-wpqj2   1/1     Running   0          17h

# expose mysql service
$ kubectl expose pod mycluster-0 -n mysql-operator --type=NodePort

# create a dir .open-hydra-server for open-hydra-server config
$ mkdir .open-hydra-server
$ cd .open-hydra-server
$ vi config.yaml
# 输入以下内容
podAllocatableLimit: -1
defaultCpuPerDevice: 2
defaultRamPerDevice: 8192
defaultGpuPerDevice: 0
datasetBasePath: /open-hydra/public-dataset
datasetVolumeType: hostpath
datasetStudentMountPath: /root/public-dataset
mysqlConfig:
  address: 10.0.0.1 # 修改为你的mysql地址
  port: 3306 # 修改为你的mysql端口
  username: root # 修改为你的mysql用户名
  password: root # 修改为你的mysql密码
  databaseName: open-hydra
  protocol: tcp
leaderElection:
  leaderElect: false
  leaseDuration: 30s
  renewDeadline: 15s
  retryPeriod: 5s
  resourceLock: endpointsleases
  resourceName: open-hydra-api-leader-lock
  resourceNamespace: default


# debug it
# checkout out .vscode/launch.json if you use vscode

```

## db create table

目前由于项目比较简单，我们暂时不使用orm，直接使用sql语句操作数据库

```sql
# 创建 user 表
CREATE TABLE user ( id INT AUTO_INCREMENT PRIMARY KEY, username  VARCHAR(255), role INT , ch_name NVARCHAR(255) , description NVARCHAR(255) , email VARCHAR(255) , password VARCHAR(255) , UNIQUE (username) );

# 创建 Dataset 表
CREATE TABLE dataset ( id INT AUTO_INCREMENT PRIMARY KEY, name  VARCHAR(255), description NVARCHAR(255) , last_update DATETIME , create_time DATETIME , UNIQUE (name) );

# 创建 Course 表
CREATE TABLE course ( id INT AUTO_INCREMENT PRIMARY KEY, name  VARCHAR(255), description NVARCHAR(255) , created_by NVARCHAR(255) , last_update DATETIME , create_time DATETIME , UNIQUE (name) );
```

## debug route with curl

* if disableAuth=false, you need to use basic auth to access apiserver by adding `--header 'open-hydra-auth: Bearer xxxxxxx'` you can get your access token(xxxxxxx) by running command `echo -n 'admin:admin' | base64 -w 0`

```bash
# create a user without basic auth
$ curl -k --location -XPOST 'https://localhost:10443/apis/open-hydra-server.openhydra.io/v1/openhydrausers' \
--header 'Content-Type: application/json' --cert pki/apiserver-kubelet-client.crt --key pki/apiserver-kubelet-client.key \
--data-raw '{
    "metadata": {
        "name": "user1"
    },
    "spec": {
        "chineseName": "第一个学员",
        "description": "student1",
        "password": "password",
        "email": "student1@gmail.com",
        "role": 1
    },
    "status": {}
}'


# list users without basic auth
$ curl -k --location 'https://localhost:10443/apis/open-hydra-server.openhydra.io/v1/openhydrausers' --cert pki/apiserver-kubelet-client.crt --key pki/apiserver-kubelet-client.key

# create device for user1
$ curl -k --location -XPOST 'https://localhost:10443/apis/open-hydra-server.openhydra.io/v1/devices' --cert pki/apiserver-kubelet-client.crt --key pki/apiserver-kubelet-client.key \
--header 'Content-Type: application/json' \
--data '{
    "metadata": {
        "name": "user3"
    },
    "spec": {
        "openHydraUsername": "user3"
    }
}'

# delete device for user1
$ curl -k --location -XDELETE 'https://localhost:10443/apis/open-hydra-server.openhydra.io/v1/devices/user1' --cert pki/apiserver-kubelet-client.crt --key pki/apiserver-kubelet-client.key

# login
$ curl -k --location -XPOST 'https://localhost:10443/apis/open-hydra-server.openhydra.io/v1/openhydrausers/login/user1' \
--header 'Content-Type: application/json' --cert pki/apiserver-kubelet-client.crt --key pki/apiserver-kubelet-client.key \
--data-raw '{
    "metadata": {
        "name": "user1"
    },
    "spec": {
        "password": "password"
    }
}'

# update gpu settting
$ curl -k --location -XPUT 'https://localhost:10443/apis/open-hydra-server.openhydra.io/v1/settings/default' --cert pki/apiserver-kubelet-client.crt --key pki/apiserver-kubelet-client.key \                                                                                                                                                                                 4:01:49 PM
--header 'Content-Type: application/json' \
--data '{
    "metadata": {
        "name": "default"
    },
    "spec": {
        "default_gpu_per_device": 0 
    }
}'

# get gpu setting
curl -k --location 'https://localhost:10443/apis/open-hydra-server.openhydra.io/v1/settings/default' --cert pki/apiserver-kubelet-client.crt --key pki/apiserver-kubelet-client.key

# output
{"metadata":{"name":"default"},"spec":{"default_gpu_per_device":0},"status":{}}
```

## try manage everything with kubectl

```bash
# when 'disableAuth' set to 'true' you can use kubectl to manage everything
$ vi user1.yaml
apiVersion: open-hydra-server.openhydra.io/v1
kind: OpenHydraUser
metadata:
  name: student2
spec:
  chineseName: 第二个学员
  description: student2
  email: student2@gmail.com
  password: password
  role: 2

# create user
$ kubectl apply -f user1.yaml
# list user
$ kubectl get openhydrausers

# create device
$ vi device1.yaml
apiVersion: open-hydra-server.openhydra.io/v1
kind: Device
metadata:
  name: student2-device
spec:
  studentName: student2

# create device for sutdent2
$ kubectl create -f device1.yaml

# list device
$ kubectl get devices
```

## reverse proxy

有时为了快速演示或者一体机相对安全情况下，我们可以不需要后台，直接使用反向代理暴露 apiserver 给到前端

```bash
# 部署反向代理
$ kubectl create -f deploy/reverse-proxy.yaml

# 检查
$ kubectl get svc -n open-hydra
NAME            TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
reverse-proxy   ClusterIp  10.98.192.222   <none>        80:80/TCP      1m

# 集群 pod 访问
$ curl http://reverse-proxy.open-hydra.svc/api/apis

# 集群外访问
$ curl http://10.98.192.222/api/apis
```

## 提交代码

大家尽力编写单元测试，保证代码质量， 目前人手不足可以不编写，但是你的写的方法应该是尽可能被单元测试的

* ginkgo 框架已经在代码里有了，可以直接使用
* make test-all && make fmt 应被执行在你的提交代码之前
* 请使用 gerrit 代码提交彼此 review

## how to add a new api

* add a new api in pkg/apis/open-hydra-api/{new-api-name}/v1
* create a deepcopy and register-gen make command
* call it to generate the code
* add pkg/apis/open-hydra-api/{new-api-name}/v1 to `update-openapi` in Makefile
* make update-openapi
