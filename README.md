# k8s-oversold

> k8s资源超售，通过调整节点资源的超卖比，在可用资源上乘上一个系数来呈现一个超卖之后的节点。目前 Kubernetes 是没有这样的接口，通过 Kubernetes 的扩展机制，Hook 了 Kubelet 向上汇报的过程，实现了该功能

# 一、简介

1. 基于[goadmission](https://github.com/mritd/goadmission) 一个 Kubernetes 动态准入控制的脚手架，进行开发完成
2. 修改[oversold](https://github.com/SecondLifter/oversold) 代码,源项目代码作者未维护,直接无法使用,再此基础上对相关代码做了修改
   1. 修复cpu超卖不生效,计算超卖逻辑存在bug抛出异常
   2. 增加突破node pods的限制功能
   3. cpu mem pods超卖倍数支持浮点数功能



# 二 、背景和相关技术原理
1. [K8S基于MutatingAdmissionWebhook实现资源超卖](https://blog.csdn.net/qq_17305249/article/details/105024493)
2. [腾讯自研业务上云：优化Kubernetes集群负载的技术方案探讨](https://cloud.tencent.com/developer/article/1505214)

# 三、如何使用

克隆本项目到本地，在本地使用docker 对代码进行容器构建。再到k8s集群进行部署

```bash
cd k8s-oversold
docker build -t k8s-oversold:1.1 . ###构建镜像
##docker login 上传镜像到镜像仓库 xxxx请自行替换为镜像仓库地址
docker tag k8s-oversold:1.1   xxxx/k8s-oversold:1.1
docker push  xxxx/k8s-oversold:1.1
kubectl create ns oversold  ###创建命名空间
cd deploy/cfssl
sh create.sh ###生成密钥
cd ../mutatingwebhook/
kubectl apply -f . ###在k8s集群中部署
### 镜像替换
kubectl set image deployment mutating-webhook goadmission=xxxx/k8s-oversold:1.0 -n oversold
```

# 四、如何开启

 ```shell
 kubectl label --overwrite node --all oversold=oversold  
 kubectl label --overwrite node --all overcpu=3 ###cpu超卖倍数3倍，支持浮点数
 kubectl label --overwrite node --all overmem=2 ###内存超卖倍数2倍，支持浮点数
 kubectl label --overwrite node --all overpods=2 ###pods超卖倍数2倍，支持浮点数
 ##使用共有云的K8S服务,如EKS,默认pods节点有限制,同时IP分配也有限制,如需突破限制使用如下命令
##详细请参考:https://docs.aws.amazon.com/zh_cn/eks/latest/userguide/cni-increase-ip-addresses.html
kubectl set env daemonset aws-node -n kube-system ENABLE_PREFIX_DELEGATION=true
kubectl set env ds aws-node -n kube-system WARM_PREFIX_TARGET=1
 ```

# 五、其他
目前仅在AWS EKS v1.21.12上测试是可行的,如需在其他K8S版本使用可自行测试