# openhydra 最小对接 keystone(alpha 功能)

本文档描述了如何将 openhydra 的验证和 keystone 对接，我们目前只支持将 openhydra 用户信息存放到 keystone 的 user 结构体中

## 开始之前

* openhydra 已经正常运行中
* keystone 服务已经部署好了，并正常工作
* 注意在和 keystone 对接后，会以 keystone 中而 user id 做为用户在 openhydra

## 修改 openhydra 配置文件

登陆到服务器上并运行以下命令

```bash
# 备份 openhydra 配置文件
$ kubectl get cm -n open-hydra open-hydra-config -o yaml > open-hydra-config.yaml
# 修改 openhydra 配置文件
$ kubectl edit cm -n open-hydra open-hydra-config
# 在 resourceNamespace 下加入以下配置后保存退出
# 替换 keystone-user 为你的 keystone 的账号和密码，我们这里建议您使用 admin 账号
authDelegateConfig:
  keystoneConfig:
    endpoint: http://[keystone ip]:[port]
    username: keystone-user
    password: keystone-user-password
    domainId: default

# 重启 openhydra server 
$ kubectl scale deployment -n open-hydra open-hydra-server --replicas=0
$ kubectl scale deployment -n open-hydra open-hydra-server --replicas=1
```

## 检验结果

在服务器上运行以下命令，注意当您集成 keystone 后，如果您尝试删除 admin 和 service 账号的操作是会被拒绝的

```bash
# 列出所有用户
$ kubectl get openhydrausers
# 输出
NAME                               AGE
595af254b5fe491fb7fa2c0a42cb299b   <unknown>
19ff8aee587f4dbeabff2fb88003c6d7   <unknown>
0c3eee2a36da4c2d8ba8b3a073eaeec3   <unknown>
e521cf70761d432ba09d093257ac7e35   <unknown>
4bce15e7abc240c3860bca2c0aaa271b   <unknown>
d75eaa9345984278ad6e6106f48c0d30   <unknown>

# 创建用户
# 由于 keystone 使用 id 作为唯一的主键会和 k8s 理念冲突，所以则中做法，将 id 变为登陆名，创建的时候一切照旧，但是返回看到的时候是 keystone 中的 id 作为登陆 id
$ cat <<EOF > user1.yaml
apiVersion: open-hydra-server.openhydra.io/v1
kind: OpenHydraUser
metadata:
  name: user1
spec:
  password: Openhydra123
  role: 2
EOF

# 验证结果
$ kubectl get openhydrausers -o custom-columns=Name:metadata.name,User:.spec.chineseName
# 输出结果
Name                               User
.......
18ca18f419e54de7b040617b5c065c7f   user1


# 尝试登陆
# 打开 openhydra 访问页面，输入用户名 18ca18f419e54de7b040617b5c065c7f 和密码 Openhydra123

# 删除用户
$ kubectl delete openhydrauser 18ca18f419e54de7b040617b5c065c7f

```