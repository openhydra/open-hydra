# 使用预先打包的 iso 一体机快速安装指南

本文档描述了如何使用预先打包好的镜像来安装操作已经 open-hydra 相关的组件

## 下载镜像

* 链接: https://pan.baidu.com/s/1GsHhDEnRMBBUrQXpusmtSA?pwd=uc2e 提取码: uc2e

## 最小硬件要求

* CPU core >= 8
* 250 GB 硬盘 >= 1
* memory >= 32 GB
* (可选)nvidia cuda gpu >= 1(driver: 525.125.06)

## usb 包含的组件介绍

* ubuntu-server-20.04.6-live-server-amd64.iso
  * k8s aio cluster deploy with [kubeclipper](https://github.com/kubeclipper/kubeclipper)
  * open-hydra
    * server
    * 反向代理 api-server(建议在安全要求搞的情况下删除)
    * dashboard(不建议在安全要求搞的情况下使用)
  * gpu-operator
    * nvidia-driver
    * nvidia-container-toolkit
    * nvidia-dcgm-exporter
    * nvidia-device-plugin
    * nvidia-cuda-validator

## 安装流程介绍

1. 插入 usb 到主机
2. 重启主机
3. 进入 boot 菜单(根据主板不同，进入方式不同，比如有些主板是按 F12，有些主板是按 F8 或者 F11)
4. 选择 usb 启动（这里可能会看到 usb 和 usb-partition 1 这是我们用 usb-partition 1）
5. 开始安装流程
6. 安装完成后，重启主机
7. 移除 usb
8. 进入系统
9. 等待完成

## 安装过程详细说明

### 步骤5--开始安装流程

* 进入 usb boot 界面后会有一端时间进行 iso 数据验证的过程大概需要持续3-5分钟不等，如下图

![installation](../images/installation-1.png)

* 语言选择--在弹出的语言学则界面我们这里选择 "English" 并按下回车键，如下图

![installation](../images/installation-2.png)

* 语言选择--在接下去弹出的对话框中直接按回车确认，如下图

![installation](../images/installation-3.png)

* 网络配置--在弹出的网络配置界，我们按 "键盘上" 键选择默认选中的第一张网卡，如果下图

![installation](../images/installation-4.png)

* 网络配置--按下回车键后，在弹出的网络配置界面，然后选择 "Edit IPv4" 后按下回车，如下图

![installation](../images/installation-6.png)

* 网络配置--在弹出的网络配置界面，我们选择 "Manual" 后按下回车，如下图

![installation](../images/installation-7.png)
![installation](../images/installation-8.png)

* 网络配置--在弹出的地址输入框中输入地址，我们这里为了后期网络配置方便我们设成 172.16.151.70，然后把光标指向 "Save" 按下回车，如下图

![installation](../images/installation-9.png)

* 网络配置--等待 10 秒后会出现 "Done" 选项在下方，我们按 "方向键下" 选中 "Done" 然后按回车，如下图

![installation](../images/installation-10.png)

* proxy 配置--跳过直接按回车，如下图

![installation](../images/installation-11.png)

* mirror 配置--跳过直接按回车，如下图

![installation](../images/installation-12.png)

* 升级页面--因为不通外部 intelnet 所有一段时间后会自动跳过
* 磁盘配置--我们把光标往上操作对准 "Set up this disk as an LVM group" 然后按下 "空格" 键，如下图

![installation](../images/installation-13.png)

* 磁盘配置--取消选择 "Set up this disk as an LVM group"，如下图

![installation](../images/installation-14.png)

* 磁盘配置--把光标对准下方的 "Done" 按下回车，如下图

![installation](../images/installation-15.png)

* 磁盘配置--在弹出的界面再次点击 "Done"，大家不用纠结涂上的磁盘大小，因为截图的机器不一定和你最终的磁盘大小一致，如下图

![installation](../images/installation-16.png)

* 磁盘配置--在弹出的界面，移动光标选择 "Continue" 然后按下回车确定，如下图

![installation](../images/installation-17.png)

* 用户配置--在弹出的界面，为了方便起见我们全部输入 "test" 然后操作光标对准 "Done" 然后按下回车，如下图

![installation](../images/installation-18.png)

* ssh配置--在弹出的是否安装 ssh 光标默认会对准 "Install OpenSSH server"，如下图

![installation](../images/installation-19.png)

* ssh配置--我们直接 "空格" 键，移动光标选择下方的 "Done" 然后按下回车确定，如下图

![installation](../images/installation-20.png)

* 等待安装完成，如下图

![installation](../images/installation-21.png)

* 等待出现 "Reboot Now" 选项时，移动光标选中 "Reboot now"

* 移除 usb 正常安装正常进入系统

![installation](../images/installation-22.jpeg)

## 查看安装输出日志

```bash
$ journalctl -u maas -f

# 输出
# 如果安装完成那么我们应该可以看到 maas.service: Succeeded 的输出
-- Logs begin at Sun 2023-12-10 10:58:29 UTC. --
...
Dec 10 11:29:14 test systemd[1]: maas.service: Succeeded.
Dec 10 11:29:14 test systemd[1]: Finished maas service.

# 检验 pod 的状态
$ kubectl get pod -A

# 输出，忽略 AGE 列，因为图是预先截取的
NAMESPACE            NAME                                                          READY   STATUS             RESTARTS       AGE
default              gpu-pod                                                       0/1     Completed          0              5d4h
gpu-operator         gpu-feature-discovery-w5nkp                                   1/1     Running            0              5d5h
gpu-operator         gpu-operator-67f8b59c9b-94qc8                                 1/1     Running            1 (3d7h ago)   5d5h
gpu-operator         nfd-node-feature-discovery-gc-5644575d55-dlfvj                1/1     Running            0              5d5h
gpu-operator         nfd-node-feature-discovery-master-5bd568cf5c-dj4qj            1/1     Running            0              5d5h
gpu-operator         nfd-node-feature-discovery-worker-vz7jb                       1/1     Running            3 (3d7h ago)   5d5h
gpu-operator         nvidia-container-toolkit-daemonset-q97hp                      1/1     Running            0              5d5h
gpu-operator         nvidia-cuda-validator-xk4dm                                   0/1     Completed          0              5d5h
gpu-operator         nvidia-dcgm-exporter-b2kl6                                    1/1     Running            0              5d5h
gpu-operator         nvidia-device-plugin-daemonset-mxtz7                          1/1     Running            0              5d5h
gpu-operator         nvidia-driver-daemonset-5.15.0-91-generic-ubuntu22.04-4g5sj   1/1     Running            0              5d5h
gpu-operator         nvidia-operator-validator-5kgs8                               1/1     Running            0              5d5h
kube-system          calico-kube-controllers-6c557bcb96-dx6kl                      1/1     Running            0              5d5h
kube-system          calico-node-skkbx                                             1/1     Running            0              5d5h
kube-system          coredns-bd6b6df9f-bgf6x                                       1/1     Running            0              5d5h
kube-system          coredns-bd6b6df9f-qmxh5                                       1/1     Running            0              5d5h
kube-system          dashboard-cluster-metrics-scraper-7d84b7bd69-fz4fl            1/1     Running            0              5d4h
kube-system          dashboard-user1-metrics-scraper-b45b9c7b5-crtzw               1/1     Running            0              5d4h
kube-system          etcd-test                                                     1/1     Running            0              5d5h
kube-system          kc-csi-controller-96579d646-vjwnw                             9/9     Running            8 (3d7h ago)   5d5h
kube-system          kc-csi-node-plugin-sflfb                                      6/6     Running            0              5d5h
kube-system          kc-csi-reloader-f44f86d76-dhj8c                               1/1     Running            0              5d5h
kube-system          kc-kubectl-7f854bbd9b-6lzn7                                   1/1     Running            0              5d5h
kube-system          kube-apiserver-test                                           1/1     Running            0              3d7h
kube-system          kube-controller-manager-test                                  1/1     Running            1 (3d7h ago)   5d5h
kube-system          kube-proxy-n5f7g                                              1/1     Running            0              5d5h
kube-system          kube-scheduler-test                                           1/1     Running            1 (3d7h ago)   5d5h
kube-system          kubeflow-prometheus-adapter-8f5cc9479-vwq9l                   1/1     Running            0              5d4h
kube-system          kubernetes-dashboard-cluster-8c65bdb47-j95pv                  1/1     Running            1 (5d4h ago)   5d4h
kube-system          kubernetes-dashboard-user1-97865494c-tdbfp                    1/1     Running            0              5d4h
kubeflow             minio-7f4cc98d67-q8tsk                                        1/1     Running            0              5d4h
kubeflow             training-operator-5545844cb8-t4mjj                            1/1     Running            0              5d4h
kubeflow             volcano-controllers-5fdc7dd79d-s5bkh                          1/1     Running            0              5d4h
kubeflow             volcano-scheduler-8695f4fc84-m4b7c                            1/1     Running            0              5d4h
kubeflow             workflow-controller-558df5fd-qtfxg                            1/1     Running            3 (3d7h ago)   5d4h
local-path-storage   local-path-provisioner-58b65c8d5d-784lq                       1/1     Running            0              3d1h
monitoring           grafana-7c754b9cc5-7whb2                                      1/1     Running            0              5d4h
monitoring           node-exporter-cwdnz                                           2/2     Running            0              5d4h
monitoring           prometheus-k8s-0                                              2/2     Running            1 (5d4h ago)   5d4h
monitoring           prometheus-operator-7688cdc9bd-x7thd                          1/1     Running            0              5d4h
mysql-operator       mycluster-0                                                   2/2     Running            0              3d1h
mysql-operator       mycluster-router-5d74f97d5b-plpp5                             1/1     Running            0              3d1h
mysql-operator       mysql-operator-66bfb7f6df-82zcn                               1/1     Running            0              3d1h
open-hydra           open-hydra-server-5c659bf678-n5ldl                            1/1     Running            0              3d1h
open-hydra           open-hydra-ui-79c444fc4c-w59ws                                1/1     Running            0              3d1h
open-hydra           reverse-proxy-6b4d55d79-77s65                                 1/1     Running            0              3d1h
```

## 访问 dashboard

* 打开浏览器输入地址 `http://[ip]:30001`
  * login with admin/openhydra

## 屏幕唤醒黑屏的问题

* 测试有些带有 gpu 的服务器当你让服务器自动安装时，如果屏幕进入待机状态后唤醒后会出现花瓶现象，这是由于安装过程中同时安装了 gpu 驱动可能会导致显示有问题，不用担心可以重启一次即可，或者你可以直接 ssh 进行管理那么就不需要重启了

## 禁用自动升级

由于 gpu operator 是基于 linux 内核构建的，我们使用离线驱动程序镜像 `-generic` 来离线安装 gpu 驱动程序，这就是为什么每次 linux 内核升级时 gpu operator 将无法工作，因此我们需要禁用自动升级，除非您打算升级。

```bash
# 禁止自动升级
$ apt-mark hold linux-headers-generic
$ apt-mark hold linux-image-generic
```
