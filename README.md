# open-hydra

open-hydra 是一个专注于 `机器学习|深度学习` 教育培训一体机的管理平台， 他可以用来快速创建在线的开发环境。 For English version please see [README-EN.md](docs/README-EN.md),他有以下特性

* 云原生，基于k8s，api 与 k8s api 深度集成 (Aggregation api)
* 提供操作系统的一体机快速部署 iso 镜像
  * iso 镜像预置 nvidia 驱动
    * 支持 nvidia gpu time slicing
* 主要的资源的管理支持 kubectl 命令
* 秒级启动 jupyter lab 和 vscode 环境
  * 终端中户可以自助式的服务
  * 管理员可以管理用户环境可以启动/关闭带有 gpu/cpu 的环境
* 支持上传自定义的 数据集 和 课程
* 支持快速替换自定义的 jupyter lab 和 vscode 的镜像
* 支持关闭授权管理验证，极限简化管理模式

## open-hydra 管理平面组件架构

![open-hydra](images/arch-01.png)

## open-hydra 用户平面的组件架构

![open-hydra](images/arch-02.png)

## 快速开始

### 使用预先打包的 iso 一体机快速安装

我们提供打包的好的带有 ubuntu 操作系统的 iso 镜像方便用户直接快速部署相关组件，详见 [iso 安装指南](docs/iso-installation-guide.md)

### 在已有的 k8s 环境上部署 open-hydra

#### 开始之前

* 目前经过测试的 k8s 版本为 1.23.1 理论上在 1.23.1 + 的版本都可以使用, 如果您没有 k8s 那么可以通过 kubeadm 来快速创建一个，参考 [kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/)
* 如果您没有 gpu 也不会妨碍你搭建环境，只是不能创建带有 gpu 的环境，如果你创建了带有 gpu 设备的环境，那么这个 pod 会进入 pending 状态
* 正确配置 gpu 设备名称，其中 jupyter lab 预装的 cuda 版本对齐了 nvidia 驱动 `525.125.06` 理论上 `535.129.03` 也可以工作

```bash
# 确认您的 gpu 设备的在 k8s 中的名称
$ kubectl describe node
# 假设您的 gpu 设备名称为 nvidia/tesla-v100
# 修改 config map 文件
# 找到 config.yaml: | 下的内容，修改下方的键 defaultGpuDriver 为 nvidia/tesla-v100
$ vi open-hydra-server/deploy/install-open-hydra.yaml
```

* 确认 storage class 设置为默认， storage class 会被 mysql-operator 来使用，所以我们需要一个默认的 sc 如果您没有 sc 可以通过 `rancher.io/local-path` 快速的利用本地某个目录模拟一个

```bash
$ kubectl patch storageclass {you storage class name} -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

* 创建课程和代码的共享目录

```bash
# 创建一个目录用于存放课程和代码
# public-dataset 用来存放公共的数据集
$ mkdir /mnt/public-dataset
# public-course 用来存放你的课程
$ mkdir /mnt/public-course
# jupyter-lab 用来存放用户 jupyter lab 的代码
$ mkdir /mnt/jupyter-lab
# public-vscode 用来存放用户 vscode 的代码
$ mkdir /mnt/public-vscode

# 如果你不使用这些目录，那么请修改 install-open-hydra.yaml 中 config.yaml: | 下的内容，修改下方的键 datasetBasePath，jupyterLabHostBaseDir，vscodeBasePath, courseBasePath 为你自己的目录 
```

#### 在已经部署好的 k8s 集群中安装 open-hydra

```bash
# 创建 mysql 实例
$ kubectl apply -f deploy/mysql-deploy-crds.yaml
$ kubectl apply -f deploy/mysql-operator.yaml
$ kubectl apply -f deploy/mysql-instance.yaml

# 等待片刻后，检查数据库实例已经运行了
$ kubectl get pods -n mysql-operator
# 输出
NAME                                READY   STATUS    RESTARTS   AGE
mycluster-0                         2/2     Running   0          2m
mycluster-router-5d74f97d5b-plpp5   1/1     Running   0          1m
mysql-operator-66bfb7f6df-82zcn     1/1     Running   0          5m

# 部署 open-hydra
$ kubectl apply -f deploy/install-open-hydra.yaml
# 等待片刻后，检查 open-hydra 已经运行了
$ kubectl get pods -n open-hydra
# 输出
$ NAME                                    READY   STATUS    RESTARTS   AGE
open-hydra-server-5c659bf678-n5ldl      1/1     Running   0          60m
# 检查 apiservice 
$ kubectl get apiservice v1.open-hydra-server.openhydra.io
# 输出
NAME                                SERVICE                        AVAILABLE   AGE
v1.open-hydra-server.openhydra.io   open-hydra/open-hydra-server   True        61m
```

### 开始使用

#### 创建管理员用户(可选，如果是 iso 安装或者不需要 ui 可跳过)

* 注意如果您在搭建时将 `disableAuth: true` 那么就无法用 kubectl 来创建 admin 用户和其他相关资源
* 管理员用主要是当你部署 `open-hydra-ui` 时才会需要这个账号，如果您只使用 kubectl 管理则没必要创建这个账号

```bash
# 创建 admin 用户
$ kubectl create -f deploy/user-admin.yaml
# 等待片刻后，检查 admin 用户已经创建了
$ kubectl get openhydrausers
# 输出
NAME    AGE
admin   <unknown>
```

#### 部署 `open-hydra-ui`(可选，如果是 iso 安装或者不需要 ui 可跳过)

* 如果想要使用 html 页面来管理你可以部署 dashboard
* 请注意 `由于 open-hydra-ui 不带后端仅有 html 页面和 js 脚本，所以我们会启动一个反向代理来代理 apiserver 这有一定安全风险，不建议在高安全要求的情况下部署 open-hydra-ui`，您可以自行实现一个后端的页面，请查看 [api 文档](docs/api.md)

```bash
# 部署反向代理
$ kubectl apply -f deploy/reverse-proxy.yaml
# 检查结果
$ kubectl get ep,svc -n open-hydra reverse-proxy
# 输出
NAME                      ENDPOINTS          AGE
endpoints/reverse-proxy   172.25.27.254:80   94m

NAME                    TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
service/reverse-proxy   ClusterIP   10.96.66.183   <none>        80/TCP    94m

# 下载 open-hydra-ui 项目
$ cd open-hydra-ui/deploy
# 修改 nginx 配置里的 {address} 为反向代理的地址
$ proxy=$(sudo kubectl get svc reverse-proxy -o jsonpath='{.spec.clusterIP}' -n open-hydra)
$ sed -i "s/{address}/${proxy}/g" nginx.conf
# 创建 ui 配置
$ kubectl create cm open-hydra-ui-config --from-file nginx.conf -n open-hydra
# 创建 ui 和 服务
$ kubectl create -f deploy.yaml
# 检查结果
$ kubectl get svc,ep -n open-hydra open-hydra-ui
# 输出
NAME                    TYPE       CLUSTER-IP     EXTERNAL-IP   PORT(S)        AGE
service/open-hydra-ui   NodePort   10.111.179.4   <none>        80:30001/TCP   111m

NAME                      ENDPOINTS          AGE
endpoints/open-hydra-ui   172.25.27.255:80   111m

```

#### 创建一个普通用户 user1

```bash
# role 1 = admin
# role 2 = user
$ cat <<EOF > user1.yaml
apiVersion: open-hydra-server.openhydra.io/v1
kind: OpenHydraUser
metadata:
  name: user1
spec:
  password: password
  role: 2
EOF

# 创建
$ kubectl create -f user1.yaml
# 检验结果
$ kubectl get openhydrausers user1
# 输出
NAME    AGE
user1   <unknown>
```

#### 开启 nvidia-gpu 基于时间片的共享(可选)

```bash
# 创建时间切片的配置
# 你可以修改 deploy/time-slicing-gpu.yaml 中的 replicas 字段来调整显卡副本数
$ kubectl apply -f deploy/time-slicing-gpu.yaml
# patch gpu-operator
$ kubectl patch clusterpolicy/cluster-policy     -n gpu-operator --type merge     -p '{"spec": {"devicePlugin": {"config": {"name": "time-slicing-config-all", "default": "any"}}}}'
# 等待片刻后，检查 gpu-operator 已经运行了
$ kubectl get pod -n gpu-operator -w

```

#### 为 user1 创建一个 jupyter lab 环境

```bash
$ cat <<EOF > user1-device.yaml
apiVersion: open-hydra-server.openhydra.io/v1
kind: Device
metadata:
  name: user1
spec:
  openHydraUsername: user1
EOF

# 创建
$ kubectl create -f user1-device.yaml
# 检查结果
$ kubectl get devices user1 -o custom-columns=User:.spec.openHydraUsername,LabUrl:.spec.jupyterLabUrl,Status:.spec.deviceStatus
# 输出
User    LabUrl                       Status
user1   http://172.16.151.70:31001   Running
# 用浏览器打开页面 http://172.16.151.70:31001
```

![open-hydra](images/lab-01.png)

#### 为 user1 关闭设备

* 无论是 `管理员` 还是 `用户` 同以时间只支持 1 个开发环境的运行，所以在启动不同类型的开发环境时

```bash
# 关闭删除设备，但是不用担心用户之前编写的代码会消失
# 删除释放设备需要等待 20 秒左右
$ kubectl delete -f user1-device.yaml
# 检查结果
$ kubectl get devices user1 -o custom-columns=User:.spec.openHydraUsername,LabUrl:.spec.jupyterLabUrl,Status:.spec.deviceStatus
# 输出
User    LabUrl   Status
user1   <none>   Terminating
```

#### 为 user1 创建一个 vscode 环境

```bash
$ cat <<EOF > user1-device-vscode.yaml
apiVersion: open-hydra-server.openhydra.io/v1
kind: Device
metadata:
  name: user1
spec:
  ideType: vscode
  openHydraUsername: user1
EOF

# 创建
$ kubectl create -f user1-device-vscode.yaml
# 检查结果
$ kubectl get devices user1 -o custom-columns=User:.spec.openHydraUsername,LabUrl:.spec.vsCodeUrl,Status:.spec.deviceStatus
# 输出
User    LabUrl                       Status
user1   http://172.16.151.70:30013   Running
```

![open-hydra](images/vscode-01.png)

### 为 user1 创建 gpu 开发环境

```bash
# 删除之前的设备, 等待环境释放
$ kubectl delete -f user1-device-vscode.yaml

# 创建配置文件
$ cat <<EOF > user1-gpu-device.yaml
apiVersion: open-hydra-server.openhydra.io/v1
kind: Device
metadata:
  name: user1
spec:
  openHydraUsername: user1
  deviceGpu: 1
EOF

# 创建
$ kubectl create -f user1-gpu-device.yaml

# 检验结果
kubectl get devices user1 -o custom-columns=User:.spec.openHydraUsername,LabUrl:.spec.jupyterLabUrl,Status:.spec.deviceStatus,GPU:.spec.deviceGpu
# 输出
User    LabUrl                       Status    GPU
user1   http://172.16.151.70:37811   Running   1

# 打开浏览器访问 http://172.16.151.70:37811
```

![open-hydra](images/lab-02.png)