
# open-hydra

[![codecov](https://codecov.io/github/openhydra/open-hydra/graph/badge.svg?token=YIC9CCFA3D)](https://codecov.io/github/openhydra/open-hydra)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/695338f25de94dc69d5b222c49770f2a)](https://app.codacy.com/gh/openhydra/open-hydra/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)

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

## 源码编译

```bash
# 使用 make 镜像编译
# 输出在 cmd/open-hydra-server 下
$ make go-build
```

## 快速开始

* 我们要在 docker 里快速运行我们的 open-hydra 项目
* 请预先准备 [kind](https://kind.sigs.k8s.io/docs/user/quick-start)
* 请确认 docker 已经正常工作
* cpu >= 4
* 内存 >= 8G
* 磁盘 >= 60G(如果你想在 jupyter lab 和 vscode 以为运行 keras 和 paddle 则建议 100G)
* （可选）gpu 支持--理论上 kind 目前没有支持 gpu 的版本，但是我们可以通过修改 kind 的 docker 镜像来支持 gpu，如果有兴趣可以自行研究 [kind 支持 gpu(linux only)](https://jacobtomlinson.dev/posts/2022/quick-hack-adding-gpu-support-to-kind/)

```bash
# 使用 kind 创建集群
$ kind create cluster
# 输出
Creating cluster "kind" ...
 ✓ Ensuring node image (kindest/node:v1.29.2) 🖼 
 ✓ Preparing nodes 📦  
 ✓ Writing configuration 📜 
 ✓ Starting control-plane 🕹️ 
 ✓ Installing CNI 🔌 
 ✓ Installing StorageClass 💾 
Set kubectl context to "kind-kind"
You can now use your cluster with:

kubectl cluster-info --context kind-kind

Have a nice day! 👋

# 检查容器
$ docker ps | grep -i kindest
74f2c42b481e   kindest/node:v1.29.2   "/usr/local/bin/entr…"   2 minutes ago   Up 2 minutes   127.0.0.1:42199->6443/tcp   kind-control-plane

# 进入容器
$ docker exec -it 74f2c42b481e /bin/bash

# 安装 git 工具
root@kind-control-plane:/# cd && apt update && apt install -y git

# 下载 open-hydra 项目
root@kind-control-plane:# git clone https://github.com/openhydra/open-hydra.git

# 部署 local-path
root@kind-control-plane:# mkdir /opt/local-path-provisioner
root@kind-control-plane:# kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/v0.0.26/deploy/local-path-storage.yaml
# 检查结果
root@kind-control-plane:# kubectl get pods -n local-path-storage
# 输出
NAME                                      READY   STATUS    RESTARTS   AGE
local-path-provisioner-58b65c8d5d-mntd2   1/1     Running   0          3m4s

# 部署 mysql-operator
root@kind-control-plane:# cd open-hydra
root@kind-control-plane:# kubectl apply -f deploy/mysql-operator-crds.yaml
root@kind-control-plane:# kubectl apply -f deploy/mysql-operator.yaml
# 等待几分钟，直到 mysql-operator 运行
root@kind-control-plane:# kubectl get pods -n mysql-operator
NAME                              READY   STATUS    RESTARTS   AGE
mysql-operator-754799c79b-r4gv8   1/1     Running   0          99s
# 部署 mysql 实例
root@kind-control-plane:# kubectl apply -f deploy/mysql-instance.yaml
# 等待几分钟，直到 mysql 实例运行，取决于您的网配置可能需要3到10分钟左右
# 目前看到有特定几率 mycluster-0 会长时间卡在 init 2/3 状态大家请耐心等待下
root@kind-control-plane:# kubectl get pods -n mysql-operator
# 输出，一个实例会有一个 router 出现
NAME                                READY   STATUS    RESTARTS   AGE
mycluster-0                         2/2     Running   0          4m6s
mycluster-router-5c6646bfd5-r5q5q   1/1     Running   0          43s

# 部署 open-hydra
root@kind-control-plane:# mkdir /mnt/public-dataset
root@kind-control-plane:# mkdir /mnt/public-course
root@kind-control-plane:# mkdir /mnt/workspace
root@kind-control-plane:# kubectl create ns open-hydra
# 替换显示 ip 为你的容器 ip
root@kind-control-plane:# ip=$(ip a show dev eth0 | grep -w inet | awk '{print $2}' | cut -d "/" -f 1)
root@kind-control-plane:# sed -i "s/localhost/$ip/g" deploy/install-open-hydra.yaml
# 降低 lab 的消耗的资源
# 降低为使用 1 cpu
root@kind-control-plane:# sed  -i "s/2000/1000/g" deploy/install-open-hydra.yaml
# 降低为内存 4g
root@kind-control-plane:# sed  -i "s/8192/4096/g" deploy/install-open-hydra.yaml
# 创建 open-hydra deployment
root@kind-control-plane:# kubectl apply -f deploy/install-open-hydra.yaml
# 检查结果
root@kind-control-plane:# kubectl get pods -n open-hydra
# 输出
NAME                                 READY   STATUS    RESTARTS   AGE
open-hydra-server-5fcdff6645-94h46   1/1     Running   0          109s

# 创建一个 admin 账号
root@kind-control-plane:# kubectl create -f deploy/user-admin.yaml
# 检查结果
root@kind-control-plane:# kubectl get openhydrausers -o yaml
# 输出
apiVersion: v1
items:
- apiVersion: open-hydra-server.openhydra.io/v1
  kind: OpenHydraUser
  metadata:
    creationTimestamp: null
    name: admin
  spec:
    chineseName: admin
    description: admin
    password: openhydra
    role: 1
  status: {}
kind: List
metadata:
  resourceVersion: ""

# 手动下载 lab 镜像，由于装有 cuda 的镜像很大，我们手动下载这个镜像
root@kind-control-plane:# ctr -n k8s.io i pull registry.cn-shanghai.aliyuncs.com/openhydra/jupyter:Python-3.8.18-dual-lan
# 等待片刻后，检查镜像是否下载成功
registry.cn-shanghai.aliyuncs.com/openhydra/jupyter:Python-3.8.18-dual-lan:       resolved       |++++++++++++++++++++++++++++++++++++++| 
manifest-sha256:5c4fa3b3103bdbc1feacdd0ed0880be4b3ddd8913e46d3b7ade3e7b0f1d5ebd1: done           |++++++++++++++++++++++++++++++++++++++| 
config-sha256:999c96811ac8bac0a4d41c67bb628dc01b4e529794133a791b953f11fc7f4039:   done           |++++++++++++++++++++++++++++++++++++++| 
layer-sha256:82c434eb639ddb964f5089c4489d84ab87f6e6773766a5db3e90ba4576aa1fcd:    done           |++++++++++++++++++++++++++++++++++++++| 
layer-sha256:827606935cb54e3918e80f62abe94946b2b42b7dba0da6d6451c4a040fa8d873:    done           |++++++++++++++++++++++++++++++++++++++| 
layer-sha256:4f4fb700ef54461cfa02571ae0db9a0dc1e0cdb5577484a6d75e68dc38e8acc1:    done           |++++++++++++++++++++++++++++++++++++++| 
layer-sha256:3dd181f9be599de628e1bc6d868d517125e07f968824bcf7b7ed8d28ad1026b1:    done           |++++++++++++++++++++++++++++++++++++++| 
elapsed: 638.3s                                                                   total:  60.4 M (96.8 KiB/s) 

# 部署 ui 服务
root@kind-control-plane:# kubectl apply -f deploy/reverse-proxy.yaml
# 验证结果
root@kind-control-plane:# kubectl get deploy,svc,ep -n open-hydra reverse-proxy
# 输出
NAME                            READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/reverse-proxy   1/1     1            1           95s

NAME                    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
service/reverse-proxy   ClusterIP   10.96.146.137   <none>        80/TCP    95s

NAME                      ENDPOINTS        AGE
endpoints/reverse-proxy   10.244.0.12:80   95s

# 下载 open-hydra-ui 项目
root@kind-control-plane:# cd && git clone https://github.com/openhydra/open-hydra-ui.git
root@kind-control-plane:# cd open-hydra-ui
root@kind-control-plane:# proxy=$(kubectl get svc reverse-proxy -o jsonpath='{.spec.clusterIP}' -n open-hydra)
root@kind-control-plane:# sed -i "s/{address}/${proxy}/g" deploy/nginx.conf
root@kind-control-plane:# kubectl create cm open-hydra-ui-config --from-file deploy/nginx.conf -n open-hydra
root@kind-control-plane:# kubectl apply -f deploy/deploy.yaml

# 大功告成，查看 ip 地址，退出容器
root@kind-control-plane:# echo $ip
# 输出
172.18.0.2
# 退出
root@kind-control-plane:# exit

# 访问 dashboard 
# 打开浏览器访问 http://172.18.0.2:30001
# 使用 admin/openhydra 登录
```

## 安装部署

### 使用预先打包的 iso 一体机快速安装

我们提供打包的好的带有 ubuntu 操作系统的 iso 镜像方便用户直接快速部署相关组件，详见 [iso 安装指南](https://github.com/openhydra/core-api/blob/main/docs/installation/iso-installation-guide.md)

### 在已有的 k8s 环境上部署 open-hydra

#### 开始之前

* 目前经过测试的 k8s 版本为 1.23.1 理论上在 1.23.1 + 的版本都可以使用, 如果您没有 k8s 那么可以通过 kubeadm 来快速创建一个，参考 [kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/)
* 如果您没有 gpu 也不会妨碍你搭建环境，只是不能创建带有 gpu 的环境，如果你创建了带有 gpu 设备的环境，那么这个 pod 会进入 pending 状态
* 正确配置 gpu 设备名称，其中 jupyter lab 预装的 cuda 版本对齐了 nvidia 驱动 `525.125.06` 理论上 `535.129.03` 也可以工作, 我们更推荐使用 [gpu-operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/gpu-operator-rdma.html#configuring-gpudirect-rdma-using-nvidia-peermem) 来管理 gpu 设备

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
# 用来存放用户编辑的内容的总目录
# 比如你为 user1 启动了 jupyter lab 环境，那么就会有这么一个目录 /mnt/workspace/jupyter-lab/test
$ mkdir /mnt/workspace

# 如果你不使用这些目录，那么请修改 install-open-hydra.yaml 中 config.yaml: | 下的内容，修改下方的键 datasetBasePath，courseBasePath 为你自己的目录 
```

#### 在已经部署好的 k8s 集群中安装 open-hydra

```bash
# 创建 mysql 实例
$ kubectl apply -f deploy/mysql-operator-crds.yaml
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
$ kubectl create ns open-hydra
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
$ proxy=$(kubectl get svc reverse-proxy -o jsonpath='{.spec.clusterIP}' -n open-hydra)
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
  sandboxName: jupyter-lab # 默认我们已经在 config map 里配置了 jupyter-lab
EOF

# 创建
$ kubectl create -f user1-device.yaml
# 检查结果
$ kubectl get devices user1 -o custom-columns=User:.spec.openHydraUsername,LabUrl:.spec.sandboxURLs,Status:.spec.deviceStatus
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
$ kubectl get devices user1 -o custom-columns=User:.spec.openHydraUsername,LabUrl:.spec.sandboxURLs,Status:.spec.deviceStatus
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
  sandboxName: vscode
  openHydraUsername: user1
EOF

# 创建
$ kubectl create -f user1-device-vscode.yaml
# 检查结果
$ kubectl get devices user1 -o custom-columns=User:.spec.openHydraUsername,LabUrl:.spec.sandboxURLs,Status:.spec.deviceStatus
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
  sandboxName: jupyter-lab
  # 如果多节点情况下你可以通过 affinity 实现亲和
  #affinity:
  #  nodeAffinity:
  #    requiredDuringSchedulingIgnoredDuringExecution:
  #      nodeSelectorTerms:
  #      - matchExpressions:
  #        - key: "kubernetes.io/hostname"
  #          operator: In
  #          values:
  #          - "your-node"
EOF

# 创建
$ kubectl create -f user1-gpu-device.yaml

# 检验结果
kubectl get devices user1 -o custom-columns=User:.spec.openHydraUsername,LabUrl:.spec.sandboxURLs,Status:.spec.deviceStatus,GPU:.spec.deviceGpu
# 输出
User    LabUrl                       Status    GPU
user1   http://172.16.151.70:37811   Running   1

# 打开浏览器访问 http://172.16.151.70:37811
```

![open-hydra](images/lab-02.png)

## 在 openhydra 里添加自定义的镜像

* openhydra 支持发布用户自动以的镜像并暴露出对应的端口,此处我们将指引大家如何快速发布自己的镜像到 openhydra 中
* 检查现有配置

```bash
# 检查现有配置
$ kubectl get settings test -o json
# 你会看到 sandboxed 下方有很多比如 PaddlePaddle， jupyter-lab， keras， vscode
{
    apiVersion: v1
data:
  plugins: |
    {
        "sandboxes": {
            "xedu": {
                "cpuImageName": "registry.cn-shanghai.aliyuncs.com/openhydra/jupyter:Python-3.8.18-dual-lan",
                "gpuImageName": "registry.cn-shanghai.aliyuncs.com/openhydra/jupyter:Python-3.8.18-dual-lan",
                "icon_name": "jupyter-lab.png",
                "command": [],
                "args": [],
                "description": "XEdu的全名为OpenXLabEdu，是基于OpenXLab的教育版，为上海人工智能实验室浦育团队开发的一套完整的AI开发和学习工具。XEdu核心工具为计算机视觉库MMEdu、神经网络库BaseNN和传统机器学习库BaseML，以及数据集处理工具BaseDT和通用模型推理工具XEduHub等。",
                "developmentInfo": [
                    "python-version: 3.8",
                    "jupyter-lab-version: 4.0.9",
                    "torch-version: 1.8.1+cu111",
                    "cuda-version: 11.1",
                    "lan: chinese | english"
                ],
                "ports": [
                  8888
                ],
                "volume_mounts": [
                    {
                        "name": "jupyter-lab",
                        "mount_path": "/root/notebook"
                    },
                    {
                        "name": "public-dataset",
                        "mount_path": "/root/notebook/dataset-public",
                        "read_only": true
                    },
                    {
                        "name": "public-course",
                        "mount_path": "/root/notebook/course-public",
                        "read_only": true
                    },
                    {
                        "name": "shm",
                        "mount_path": "/dev/shm",
                        "read_only": false
                    }
                ],
                "volumes": [
                    {
                      "host_path": {
                        "name": "jupyter-lab",
                        "path": "{workspace}/jupyter-lab/{username}",
                        "type": "DirectoryOrCreate"
                      }
                    },
                    {
                      "host_path": {
                        "name": "public-dataset",
                        "path": "{dataset-public}",
                        "type": "DirectoryOrCreate"
                      }
                    },
                    {
                      "host_path": {
                        "name": "public-course",
                        "path": "{course-public}",
                        "type": "DirectoryOrCreate"
                      }
                    },
                    {
                      "empty_dir": {
                        "name": "shm",
                        "medium": "Memory",
                        "size_limit": 2048
                      }
                    }
                ]
            },
            "keras": {
                "cpuImageName": "registry.cn-shanghai.aliyuncs.com/openhydra/tensorflow:2.15.0-jupyter-cpu",
                "gpuImageName": "registry.cn-shanghai.aliyuncs.com/openhydra/tensorflow:latest-gpu-with-jlab",
                "icon_name": "keras.png",
                "command": [],
                "args": [],
                "description": "Keras是在谷歌TensorFlow的基础上发展起来的深度学习开发框架。可以把Keras看成是TensorFlow的入门简化版本，开发门槛更低，拥有大量的用户。",
                "developmentInfo": [
                    "keras-version: 2.6.1",
                    "tensorflow-version: 2.6.0",
                    "jupyter-lab-version: 4.0.9",
                    "lan: chinese | english"
                ],
                "ports": [
                  8888
                ],
                "volume_mounts": [
                    {
                        "name": "keras",
                        "mount_path": "/home/workspace"
                    },
                    {
                        "name": "public-dataset",
                        "mount_path": "/root/notebook/dataset-public",
                        "read_only": true
                    },
                    {
                        "name": "public-course",
                        "mount_path": "/root/notebook/course-public",
                        "read_only": true
                    }
                ],
                "volumes": [
                    {
                      "host_path": {
                        "name": "keras",
                        "path": "{workspace}/keras/{username}",
                        "type": "DirectoryOrCreate"
                      }
                    },
                    {
                      "host_path": {
                        "name": "public-dataset",
                        "path": "{dataset-public}",
                        "type": "DirectoryOrCreate"
                      }
                    },
                    {
                      "host_path": {
                        "name": "public-course",
                        "path": "{course-public}",
                        "type": "DirectoryOrCreate"
                      }
                    }
                ]
            },
            "PaddlePaddle": {
                "cpuImageName": "registry.cn-shanghai.aliyuncs.com/openhydra/paddle:2.6.0-jlab",
                "gpuImageName": "registry.cn-shanghai.aliyuncs.com/openhydra/paddle:2.6.0-gpu-cuda12.0-cudnn8.9-trt8.6-jlab",
                "icon_name": "PaddlePaddle.png",
                "command": [],
                "args": [],
                "description":"PaddlePaddle是百度开发的深度学习平台。PaddlePaddle集核心框架、基础模型库、端到端开发套件、丰富的工具组件、用户社区于一体，是中国首个自主研发、功能丰富、开源开放的产业级深度学习平台。",
                "developmentInfo": [
                    "PaddlePaddle-version: 2.1.0",
                    "jupyter-lab-version: 4.0.9",
                    "cuda-version: 12.0",
                    "lan: chinese | english"
                ],
                "ports": [
                  8888
                ],
                "volume_mounts": [
                    {
                        "name": "paddle",
                        "mount_path": "/home/workspace"
                    },
                    {
                        "name": "public-dataset",
                        "mount_path": "/root/notebook/dataset-public",
                        "source_path": "{dataset-public}",
                        "read_only": true
                    },
                    {
                        "name": "public-course",
                        "mount_path": "/root/notebook/course-public",
                        "source_path": "{course-public}",
                        "read_only": true
                    }
                ],
                "volumes": [
                    {
                      "host_path": {
                        "name": "paddle",
                        "path": "{workspace}/paddle/{username}",
                        "type": "DirectoryOrCreate"
                      }
                    },
                    {
                      "host_path": {
                        "name": "public-dataset",
                        "path": "{dataset-public}",
                        "type": "DirectoryOrCreate"
                      }
                    },
                    {
                      "host_path": {
                        "name": "public-course",
                        "path": "{course-public}",
                        "type": "DirectoryOrCreate"
                      }
                    }
                ]
            }
        }
    }
kind: ConfigMap
metadata:
  name: openhydra-plugin
  namespace: open-hydra
}
```

* 插入自定义的镜像

```bash
# 假设你的镜像名字叫 my-app
# 假设你的镜像 cpu 和 gpu 使用同一个镜像 docker.io/my-reg/my-app:latest
# 假设你的镜像暴露 port 3000 web服务
# 假设你的镜像同时暴露 jupyter lab 服务 8888 端口
# 假设你的镜像希望在运行时为每个用户启动一个私有的目录
# 假设你的镜像也希望挂在公共的数据集目录
# 假设你的镜像也希望挂在公共的课程目录
# 修改 config map
$  kubectl edit cm -n open-hydra openhydra-plugin
# 在最后一个元素，比如上面输出最后一个是 vscode , 那么我们在结尾处加入以下 json
                "my-app": {
                    "cpuImageName": "docker.io/my-reg/my-app:latest",
                    "description": "my-app 是在线的开发环境，帮助您快速学习大模型的开发使用",
                    "developmentInfo": [
                        "version: 0.01",
                        "python-version": "3.18"
                    ],
                    "gpuImageName": "docker.io/my-reg/my-app:latest",
                    "ports": [
                        3000, // 对应您的 3000 宽口
                        8888 // 对应您的 8888 端口
                    ],
                    "volume_mounts": [
                        {
                            "mount_path": "/home/my-app", // 这里对应了你的镜像里的 working dir 目录
                            "name": "my-app", // 名字可以自己定义，注意符合 dns 规范，建议 xx-xx 开头不要用数字
                            "read_only": false // 保持默认，除非你不想让用户写入
                        },
                        {
                            "mount_path": "/root/notebook/dataset-public", // 这里对应了你的镜像里的数据集目录，假设你制作镜像的时候 mkdir -p /root/notebook/dataset-public 这里就保持默认
                            "name": "public-dataset", // 保持默认即可 
                            "read_only": true // 我们不希望用户能直接修改挂在公共数据集里的内容我们这里设置为 true
                        },
                        {
                            "mount_path": "/root/notebook/course-public", # 同理如上
                            "name": "public-course", # 保持默认
                            "read_only": true # 同理如上
                        }
                    ],
                    "volumes": [
                        {
                        "host_path": {
                            "name": "my-app",
                            "path": "{workspace}/my-app/{username}",
                            "type": "DirectoryOrCreate"
                        }
                        },
                        {
                        "host_path": {
                            "name": "public-dataset",
                            "path": "{dataset-public}",
                            "type": "DirectoryOrCreate"
                        }
                        },
                        {
                        "host_path": {
                            "name": "public-course",
                            "path": "{course-public}",
                            "type": "DirectoryOrCreate"
                        }
                        }
                    ]
                }


# 创建一个 yaml
$ cat <<EOF > user1-device-my-app.yaml
apiVersion: open-hydra-server.openhydra.io/v1
kind: Device
metadata:
  name: user1
spec:
  sandboxName: my-app # 注意对应 my-app
  openHydraUsername: user1
EOF

# 创建他
$ kubectl create -f user1-device-my-app.yaml

# 获取对外的服务
$ kubectl get svc -n open-hydra -l openhydra-user=user1
# 输出
NAME                      TYPE       CLUSTER-IP       EXTERNAL-IP   PORT(S)                       AGE
openhydra-service-user1   NodePort   10.106.240.195   <none>        3000:32570/TCP,8888:30112     33s

```

## 解决问题

常见问题的解决的方法见文档 [错误解决](docs/trouble-shooting.md)
