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
# 如果 kubeadm reset -f 这个命令各卡住超过 2 分钟， 则直接重启服务器即可
$ kubeadm reset -f
# 检查是否有僵尸容器
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