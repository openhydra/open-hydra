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
$ kubectl apply -f deploy/mysql-operator-crds.yaml
$ kubectl apply -f deploy/mysql-operator.yaml

# ensure mysql operator it's ready
# then create mysql cluster
$ kubectl apply -f deploy/mysql-instance.yaml

# check see if mysql cluster is ready
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
# input following setting
podAllocatableLimit: -1 # not pod limit is set will use pod limit of k8s node by default
defaultCpuPerDevice: 2000 # cpu per pod default 2
defaultRamPerDevice: 8192 # memory per pod default 8Gi
defaultGpuPerDevice: 0 # gpu per pod default 0, keep it 0 unless you have tons of gpus
datasetBasePath: /mnt/public-dataset # where dataset keep on server dir
datasetVolumeType: hostpath # so far we only support hostpath
jupyterLabHostBaseDir: /mnt/jupyter-lab # where user custom code of jupyter-lab on server dir
imageRepo: "docker.io/99cloud/jupyter:Python-3.8.18"
vscodeImageRepo: "docker.io/99cloud/vscode:1.85.1"
defaultGpuDriver: nvidia.com/gpu
serverIP: "172.16.151.70"
patchResourceNotRelease: true
disableAuth: true
mysqlConfig:
  address: mycluster-instances.mysql-operator.svc
  port: 3306
  username: root
  password: openhydra
  databaseName: openhydra
  protocol: tcp
leaderElection:
  leaderElect: false
  leaseDuration: 30s
  renewDeadline: 15s
  retryPeriod: 5s
  resourceLock: endpointsleases
  resourceName: open-hydra-api-leader-lock
  resourceNamespace: open-hydra


# debug it
# checkout out .vscode/launch.json if you use vscode

```

## db create table(option)

* It's ok not to create table manually, table will be created automatically when you start the server by `mysql.go`

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
        "chineseName": "first user,
        "description": "user1",
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
$ vi user2.yaml
apiVersion: open-hydra-server.openhydra.io/v1
kind: OpenHydraUser
metadata:
  name: user2
spec:
  chineseName: 2nd users
  description: user2
  email: user2@gmail.com
  password: password
  role: 2

# create user
$ kubectl apply -f user2.yaml
# list user
$ kubectl get openhydrausers

# create device
$ vi device1.yaml
apiVersion: open-hydra-server.openhydra.io/v1
kind: Device
metadata:
  name: user2
spec:
  studentName: user2

# create device for user2
$ kubectl create -f device1.yaml

# list device
$ kubectl get devices
```

## reverse proxy

deploy a reverse proxy to access open-hydra-server api directly but not secure, you should use it in your local environment

```bash
# deploy reverse proxy
$ kubectl create -f deploy/reverse-proxy.yaml

# check it
$ kubectl get svc -n open-hydra
NAME            TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
reverse-proxy   ClusterIp  10.98.192.222   <none>        80:80/TCP      1m

# check reverse proxy working
$ curl http://reverse-proxy.open-hydra.svc/api/apis

# check it over cluster ip
$ curl http://10.98.192.222/api/apis
```

## code commit

* Please see [code of conduct](../code-of-conduct.md) before you commit your code

## modify api

* After you modify you api property you should always run `make update-openapi` to update you openapi spec
