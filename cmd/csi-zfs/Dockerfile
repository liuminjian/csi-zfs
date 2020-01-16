FROM centos:7.6.1810

ENV CEPH_VERSION "mimic"
ADD ceph.repo /etc/yum.repos.d/
RUN yum install -y epel-release && \
  yum install -y --nogpgcheck ceph-common && \
  yum install -y --nogpgcheck http://download.zfsonlinux.org/epel/zfs-release.el7_6.noarch.rpm && \
  yum install -y --nogpgcheck zfs && \
  yum clean all
ADD csi-zfs-plugin /bin/

ENTRYPOINT ["/bin/csi-zfs-plugin"]
