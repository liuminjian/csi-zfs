module github.com/liuminjian/csi-zfs

go 1.12

require (
	github.com/container-storage-interface/spec v1.2.0
	github.com/go-delve/delve v1.3.2 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2
	github.com/kubernetes-csi/csi-test v2.2.0+incompatible
	github.com/sirupsen/logrus v1.4.2
	golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3 // indirect
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys v0.0.0-20190826190057-c7b8b68b1456
	golang.org/x/tools v0.0.0-20190524140312-2c0ae7006135 // indirect
	google.golang.org/grpc v1.26.0
	honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc // indirect
	k8s.io/apimachinery v0.17.0
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v0.0.0-00010101000000-000000000000
	k8s.io/utils v0.0.0-20200108110541-e2fb8e668047
)

replace k8s.io/kubernetes => github.com/kubernetes/kubernetes v1.14.0
