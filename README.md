# vm-crd-example
使用[code-generator](https://github.com/kubernetes/code-generator)生成代码，此repository实现了一个简单的控制器，用于监视使用[CRD](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/) 定义的“VM”资源。

Golang版本号需要大于等于go1.19.4，k8s集群版本需要大于等于v1.20.15，此示例将向您展示如何执行基本操作，例如：

* [x] How to create a custom resource of type `VM` using CRD API.
* [x] How to operate instances of type `VM`.
* [x] How to implement a controller for handling an instance of type `VM` to move the current state towards the desired state.
* [x] How to use Finalizer on instances of type `VM`.
* [x] How to implement LeaseLock for multiple controllers.
* [ ] How to expose metrics of the controller.
## 构建
在工作目录中克隆repo，然后运行以下命令：
```
$ git clone https://github.com/zeratullich/vm-crd-create.git
$ cd vm-crd-create
$ go build -o vm-crd-create .
```
## 运行
> Note：
> 可查看使用参数`go run main.go --help`，可启动多个程序，远程docker机器地址为`192.168.0.100`，可以把`DOCKER_HOST=tcp://192.168.0.100:2375`写入`/etc/profile`文件中。
允许docker远程连接：
```
$ vim /usr/lib/systemd/system/docker.service
##修改如下
ExecStart=/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock -H tcp://0.0.0.0:2375
$ systemctl daemon-reload
$ systemctl restart docker
```
创建CRD资源：
```
$ kubectl apply -f artifacts/example/vm-crd.yaml
```
运行以下命令进行调试：
```
$ export DOCKER_HOST=tcp://192.168.0.100:2375 
$ go run main.go --kubeconfig=$HOME/.kube/config \
    -v=3 --logtostderr \
    --vm-driver=docker
```
可使用[Leader Election](https://github.com/kubernetes/client-go/tree/master/tools/leaderelection)模式启动：
```
# first terminal 
$ export DOCKER_HOST=tcp://192.168.0.100:2375 
$ POD_NAME=test1 go run cmd/main.go --kubeconfig=$HOME/.kube/config -v=3 --logtostderr --lease-lock-namespace=default --leader-elect

# second terminal 
$ export DOCKER_HOST=tcp://192.168.0.100:2375 
$ POD_NAME=test2 go run cmd/main.go --kubeconfig=$HOME/.kube/config -v=3 --logtostderr --lease-lock-namespace=default --leader-elect

# third terminal
$ export DOCKER_HOST=tcp://192.168.0.100:2375 
$ POD_NAME=test3 go run cmd/main.go --kubeconfig=$HOME/.kube/config -v=3 --logtostderr --lease-lock-namespace=default --leader-elect
```
创建CR资源:
```
$ kubectl apply -f artifacts/example/example-vm.yaml
```
查看资源详情：
```
$ kubectl get vms  
$ kubectl describe vms
```
> 项目来源：[Controller101](https://github.com/kairen/controller101)，修改了部分代码与目录结构，感谢原作者，可以观看原作者关于此项目的详细[博客](https://ithelp.ithome.com.tw/users/20120251/ironman/2407)。