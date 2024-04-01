# 错误解决

本文档描述了如何解决一些常见的错误的方法

## usb 安装系统后正常启动但是在 openhydra 全部完成前服务器挂了

* 这种情况多半无法自动恢复，我们需要一下步骤来完全重置后重新安装

```bash
# 登陆到服务器上
# 切换到 root 账号
# 开始重置
$ systemctl stop kubelet
$ systemctl disable kubelet
# 停止并禁用 openhydra 安装服务,因为如果 /etc/kubernetes 不存在的话 systemd 服务会在服务器重启后自动运行
# 会将事情复杂化，所以我们需要停止和禁用服务以简化事情
$ systemctl stop maas
$ systemctl disable maas
# 如果 kubeadm reset -f 这个命令各卡住超过 2 分钟， 则直接重启服务器即可
$ kubeadm reset -f
# 清理 kubeclipper 数据库 
$ kcctl clean -Af
# 检查是否有僵尸容器，如果之前重启过服务器那么大概率不会有僵尸容器，直接跳过这一步
$ ctr -n k8s.io container list
# 注意 如果上面的命令返回有容器则需要重启服务器
$ reboot
# 重启后
$ kubeadm reset -f
$ rm -rf /etc/kubernetes

# 开始重新安装
$ systemctl start maas

# 查看日志
$ journalctl -u maas -f
```

## 如何更新 open-hydra server 的镜像

* 登陆 open-hydra server 的服务器运行以下命令

```bash
# 手动下载镜像
$ ctr -n k8s.io i pull docker.io/99cloud/open-hydra-server:latest
# 重启 open-hydra server
$ kubectl scale deployment open-hydra-server --replicas=0 -n open-hydra
# 等待 3 秒后
$ kubectl scale deployment open-hydra-server --replicas=1 -n open-hydra
```
