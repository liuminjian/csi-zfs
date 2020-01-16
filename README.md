# csi-zfs [![Build Status](https://travis-ci.org/liuminjian/csi-zfs.svg?branch=master)](https://travis-ci.org/liuminjian/csi-zfs)
这是个k8s的csi插件，后台用ceph提供块存储，挂载的时候使用zfs，使pod容器可以方便使用zfs的压缩和快照功能，
支持ceph和zfs的动态扩容，目前快照功能还没和VolumeSnapshot整合。
主要参考了csi-digitalocean的代码，https://github.com/digitalocean/csi-digitalocean。

### 准备环境

controller节点和worker节点都要安装ceph-common和zfs，并modprobe rbd和zfs，并将ceph的配置文件放到各个节点/etc/ceph目录下。
创建configmap 和 secret。

```yaml
kubectl create configmap ceph-config --from-file=/etc/ceph/ceph.conf

kubectl create secret generic ceph-key --from-file=/etc/ceph/ceph.client.admin.keyring
```

如果使用Kubernetes 1.14 或者 1.15，需要kubelet启动的时候增加参数--feature-gates=ExpandCSIVolumes=true,
ExpandInUsePersistentVolumes=true才能使用扩容特性

### 使用方式
```yaml
// 编译代码和创建docker镜像
make publish 
// 在各个工作节点下载镜像
// 进入deploy目录，用helm安装
helm install csi-zfs ./csi-zfs
// 用example目录的例子测试
kubectl apply -f pvc.yaml
kubectl apply -f pod.yaml
```