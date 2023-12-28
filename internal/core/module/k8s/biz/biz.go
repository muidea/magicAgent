package biz

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"strings"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/cache"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magieAgent/internal/config"
	"github.com/muidea/magieAgent/internal/core/base/biz"
	"github.com/muidea/magieAgent/pkg/common"
)

const haSuffix = "-4ha"

func getClusterConfig() (config *rest.Config, err error) {
	if len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0 {
		var addrs []string
		addrs, err = net.LookupHost("kubernetes.default.svc")
		if err != nil {
			return
		}
		os.Setenv("KUBERNETES_SERVICE_HOST", addrs[0])
	}
	if len(os.Getenv("KUBERNETES_SERVICE_PORT")) == 0 {
		os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	}

	return rest.InClusterConfig()
}

type K8s struct {
	biz.Base

	serviceCache cache.KVCache

	clientSet    *kubernetes.Clientset
	clientConfig *rest.Config
}

func New(
	eventHub event.Hub,
	backgroundRoutine task.BackgroundRoutine,
) *K8s {
	clusterConfig, configErr := getClusterConfig()
	if configErr != nil {
		panic(configErr)
	}

	clusterClient, clientErr := kubernetes.NewForConfig(clusterConfig)
	if clientErr != nil {
		panic(clientErr)
	}

	ptr := &K8s{
		Base:         biz.New(common.K8sModule, eventHub, backgroundRoutine),
		serviceCache: cache.NewKVCache(nil),
		clientConfig: clusterConfig,
		clientSet:    clusterClient,
	}

	ptr.SubscribeFunc(common.ExecuteCommand, ptr.ExecuteCommand)
	ptr.SubscribeFunc(common.StartService, ptr.StartService)
	ptr.SubscribeFunc(common.StopService, ptr.StopService)
	ptr.SubscribeFunc(common.ChangeServiceEndpoints, ptr.ChangeServiceEndpoints)
	ptr.SubscribeFunc(common.RestoreServiceEndpoints, ptr.RestoreServiceEndpoints)
	ptr.SubscribeFunc(common.JobService, ptr.JobService)
	ptr.SubscribeFunc(common.ListService, ptr.ListService)
	ptr.SubscribeFunc(common.QueryService, ptr.QueryService)
	ptr.SubscribeFunc(common.CreateService, ptr.CreateService)
	ptr.SubscribeFunc(common.DestroyService, ptr.DestroyService)
	return ptr
}

func (s *K8s) getNamespace() string {
	namespace, found := os.LookupEnv("NAMESPACE")
	if !found {
		namespace = corev1.NamespaceDefault
	}
	return namespace
}

func (s *K8s) Run() {
	s.AsyncTask(func() {
		// 创建一个Watcher来监视Deployment资源变化
		watcher, err := s.clientSet.AppsV1().Deployments(s.getNamespace()).Watch(context.TODO(), metav1.ListOptions{
			LabelSelector: common.GetDefaultLabels(),
		})
		if err != nil {
			log.Criticalf("watch deployment failed, error:%s", err.Error())
			panic(err)
		}

		// 循环监听Watcher的事件
		for ev := range watcher.ResultChan() {
			deployment, ok := ev.Object.(*appv1.Deployment)
			if !ok {
				log.Errorf("Unexpected object type:%v", ev.Object)
				continue
			}

			//if deployment.Status.UnavailableReplicas != 0 || deployment.Status.AvailableReplicas == 0 {
			//	continue
			//}

			// 根据事件类型执行相应操作
			switch ev.Type {
			case watch.Added:
				s.addService(deployment)
			case watch.Modified:
				s.modService(deployment)
			case watch.Deleted:
				s.delService(deployment)
			default:
				log.Warnf("Error occurred, object type:%v", ev.Object)
			}
		}

		// 关闭Watcher
		watcher.Stop()
	})
}

func (s *K8s) addService(deploymentPtr *appv1.Deployment) {
	serviceInfo, serviceErr := s.getServiceInfoFromDeployment(deploymentPtr, s.clientSet)
	if serviceErr != nil {
		log.Errorf("addService failed, s.getServiceInfoFromDeployment %v error:%s", deploymentPtr.ObjectMeta.GetName(), serviceErr.Error())
		return
	}
	// 如果返回空，则表示是不需要处理的服务，直接调过
	if serviceInfo == nil {
		return
	}

	values := event.NewValues()
	values.Set(event.Action, event.Add)
	s.BroadCast(common.NotifyService, values, serviceInfo)

	s.serviceCache.Put(serviceInfo.Name, serviceInfo, cache.MaxAgeValue)
}

func (s *K8s) modService(deploymentPtr *appv1.Deployment) {
	serviceInfo, serviceErr := s.getServiceInfoFromDeployment(deploymentPtr, s.clientSet)
	if serviceErr != nil {
		log.Errorf("addService failed, s.getServiceInfoFromDeployment %v error:%s", deploymentPtr.ObjectMeta.GetName(), serviceErr.Error())
		return
	}
	// 如果返回空，则表示是不需要处理的服务，直接调过
	if serviceInfo == nil {
		return
	}

	values := event.NewValues()
	values.Set(event.Action, event.Mod)
	s.BroadCast(common.NotifyService, values, serviceInfo)

	s.serviceCache.Put(serviceInfo.Name, serviceInfo, cache.MaxAgeValue)
}

func (s *K8s) delService(deploymentPtr *appv1.Deployment) {
	serviceName, serviceCatalog := s.getServiceName(deploymentPtr)
	if serviceCatalog == "" {
		return
	}

	serviceVal := s.serviceCache.Fetch(serviceName)

	values := event.NewValues()
	values.Set(event.Action, event.Del)
	s.BroadCast(common.NotifyService, values, serviceVal.(*common.ServiceInfo))

	s.cleanHAServicePort(deploymentPtr, s.clientSet)
	s.serviceCache.Remove(serviceName)
}

func (s *K8s) getServiceName(deploymentPtr *appv1.Deployment) (name, catalog string) {
	nameVal := deploymentPtr.ObjectMeta.GetName()
	if strings.Index(nameVal, common.Mariadb) != -1 {
		name = nameVal
		catalog = common.Mariadb
		return
	}

	// TODO other databases
	return
}

func (s *K8s) getServiceBackPath(deploymentPtr *appv1.Deployment, _ *kubernetes.Clientset) (ret *common.Path, err *cd.Result) {
	var backVolumes *corev1.Volume
	for _, val := range deploymentPtr.Spec.Template.Spec.Volumes {
		if val.Name == "back-path" {
			backVolumes = &val
			break
		}
	}
	if backVolumes == nil || len(deploymentPtr.Spec.Template.Spec.Volumes) == 0 {
		return
	}

	if backVolumes.HostPath != nil {
		ret = &common.Path{
			Name:  deploymentPtr.Spec.Template.Spec.Volumes[1].Name,
			Value: path.Join(config.GetTmpPath(), deploymentPtr.ObjectMeta.GetName()),
			Type:  common.HostPath,
		}
	}
	return
}

func (s *K8s) getServiceDataPath(deploymentPtr *appv1.Deployment, clientSet *kubernetes.Clientset) (ret *common.Path, err *cd.Result) {
	var dataVolumes *corev1.Volume
	for _, val := range deploymentPtr.Spec.Template.Spec.Volumes {
		if val.Name == deploymentPtr.ObjectMeta.GetName() {
			dataVolumes = &val
			break
		}
	}
	if dataVolumes == nil || dataVolumes.PersistentVolumeClaim == nil {
		return
	}

	pvcInfo, pvcErr := clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Get(
		context.TODO(),
		dataVolumes.PersistentVolumeClaim.ClaimName,
		metav1.GetOptions{})
	if pvcErr != nil {
		err = cd.NewError(cd.UnExpected, pvcErr.Error())
		log.Errorf("getServiceDataPath failed, clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Get %v error:%v",
			dataVolumes.PersistentVolumeClaim.ClaimName,
			pvcErr.Error())
		return
	}

	pvInfo, pvErr := clientSet.CoreV1().PersistentVolumes().Get(context.TODO(), pvcInfo.Spec.VolumeName, metav1.GetOptions{})
	if pvErr != nil {
		err = cd.NewError(cd.UnExpected, pvErr.Error())
		log.Errorf("getServiceDataPath failed, clientSet.CoreV1().PersistentVolumes().Get %v error:%v",
			pvcInfo.Spec.VolumeName,
			pvErr.Error())
		return
	}
	ret = &common.Path{
		Name:  dataVolumes.Name,
		Value: pvInfo.Spec.HostPath.Path,
		Type:  common.LocalPath,
	}
	return
}

func (s *K8s) getServicePort(deploymentPtr *appv1.Deployment, clientSet *kubernetes.Clientset) (ret int32, err *cd.Result) {
	svcInfo, svrErr := clientSet.CoreV1().Services(s.getNamespace()).Get(context.TODO(), deploymentPtr.ObjectMeta.GetName(), metav1.GetOptions{})
	if svrErr != nil {
		err = cd.NewError(cd.UnExpected, svrErr.Error())
		log.Errorf("getServicePort failed, clientSet.CoreV1().Services(s.getNamespace()).Get %v error:%v",
			deploymentPtr.ObjectMeta.GetName(),
			svrErr.Error())
		return
	}

	const defaultVal = "default"
	for _, val := range svcInfo.Spec.Ports {
		if val.Name == defaultVal {
			ret = val.NodePort
			break
		}
	}

	return
}

func (s *K8s) getHAServicePort(deploymentPtr *appv1.Deployment, clientSet *kubernetes.Clientset) (ret int32, err *cd.Result) {
	const defaultVal = "default"
	haSvcInfo, haSvrErr := clientSet.CoreV1().Services(s.getNamespace()).Get(context.TODO(), deploymentPtr.ObjectMeta.GetName()+haSuffix, metav1.GetOptions{})
	if haSvrErr == nil {
		for _, val := range haSvcInfo.Spec.Ports {
			if val.Name == defaultVal {
				ret = val.NodePort
				break
			}
		}

		return
	}

	svcInfo, svrErr := clientSet.CoreV1().Services(s.getNamespace()).Get(context.TODO(), deploymentPtr.ObjectMeta.GetName(), metav1.GetOptions{})
	if svrErr != nil {
		err = cd.NewError(cd.UnExpected, svrErr.Error())
		log.Errorf("getHAServicePort failed, clientSet.CoreV1().Services(s.getNamespace()).Get %v error:%v",
			deploymentPtr.ObjectMeta.GetName(),
			svrErr.Error())
		return
	}

	getDefaultPort := func() (ret []corev1.ServicePort) {
		for _, val := range svcInfo.Spec.Ports {
			if val.Name != defaultVal {
				continue
			}

			ret = append(ret, corev1.ServicePort{
				Name:       val.Name,
				Protocol:   val.Protocol,
				Port:       val.Port,
				TargetPort: val.TargetPort,
			})

			break
		}
		return
	}

	haSvcInfo, haSvrErr = clientSet.CoreV1().Services(s.getNamespace()).Create(context.TODO(),
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploymentPtr.ObjectMeta.GetName() + haSuffix,
				Namespace: deploymentPtr.ObjectMeta.GetNamespace(),
				Labels:    deploymentPtr.ObjectMeta.Labels,
			},
			Spec: corev1.ServiceSpec{
				Ports:    getDefaultPort(),
				Selector: svcInfo.Spec.Selector,
				Type:     corev1.ServiceTypeNodePort,
			},
		},
		metav1.CreateOptions{})
	if haSvrErr != nil {
		log.Errorf("getHAServicePort failed, clientSet.CoreV1().Services(s.getNamespace()).Create %v error:%v",
			deploymentPtr.ObjectMeta.GetName()+haSuffix,
			haSvrErr.Error())
		return
	}

	for _, val := range haSvcInfo.Spec.Ports {
		if val.Name == defaultVal {
			ret = val.NodePort
			break
		}
	}

	return
}

func (s *K8s) getEnv(deploymentPtr *appv1.Deployment, _ *kubernetes.Clientset) (ret *common.Env) {
	const password = "MYSQL_ROOT_PASSWORD"
	ret = &common.Env{
		Root: common.DefaultMariadbRoot,
	}
	if len(deploymentPtr.Spec.Template.Spec.Containers) == 0 {
		return
	}

	envs := deploymentPtr.Spec.Template.Spec.Containers[0].Env
	for _, val := range envs {
		if val.Name == password {
			ret.Password = val.Value
		}
	}

	return
}

func (s *K8s) getServiceInfoFromDeployment(deploymentPtr *appv1.Deployment, clientSet *kubernetes.Clientset) (ret *common.ServiceInfo, err *cd.Result) {
	for {
		if s.checkIsMariadb(deploymentPtr) {
			serviceInfo, serviceErr := s.getMariadbServiceInfo(deploymentPtr, clientSet)
			if serviceErr != nil {
				err = serviceErr
				log.Errorf("getServiceInfoFromDeployment failed, s.getMariadbServiceInfo %v error:%v", deploymentPtr.ObjectMeta.GetName(), serviceErr.Error())
				return

			}

			ret = serviceInfo
			break
		}

		// TODO other databases
		break
	}

	return
}

func (s *K8s) cleanHAServicePort(deploymentPtr *appv1.Deployment, clientSet *kubernetes.Clientset) {
	clientSet.CoreV1().Services(s.getNamespace()).Delete(context.TODO(), deploymentPtr.ObjectMeta.GetName()+haSuffix, metav1.DeleteOptions{})
}

func (s *K8s) Create(serviceName, catalog string) (err *cd.Result) {
	switch catalog {
	case common.Mariadb:
		serviceInfo := s.getDefaultMariadbServiceInfo(serviceName)
		err = s.createService(serviceInfo)
	default:
		panic(fmt.Sprintf("illegal catalog:%v", catalog))
	}

	return
}

func (s *K8s) Destroy(serviceName, catalog string) (err *cd.Result) {
	serviceInfo, serviceErr := s.Query(serviceName, catalog)
	if serviceErr != nil {
		err = serviceErr
		return
	}

	err = s.destroyService(serviceInfo)
	return
}

func (s *K8s) Start(serviceName, catalog string) (err *cd.Result) {
	serviceInfo, serviceErr := s.Query(serviceName, catalog)
	if serviceErr != nil {
		err = serviceErr
		return
	}

	err = s.startService(serviceInfo)
	return
}

func (s *K8s) Stop(serviceName, catalog string) (err *cd.Result) {
	serviceInfo, serviceErr := s.Query(serviceName, catalog)
	if serviceErr != nil {
		err = serviceErr
		return
	}

	err = s.stopService(serviceInfo)
	return
}

func (s *K8s) Query(serviceName, catalog string) (ret *common.ServiceInfo, err *cd.Result) {
	serviceVal := s.serviceCache.Fetch(serviceName)
	if serviceVal == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("%s not exist", serviceName))
		return
	}
	servicePtr := serviceVal.(*common.ServiceInfo)
	if servicePtr.Catalog != catalog {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("%s miss match catalog", serviceName))
		return
	}

	ret = servicePtr
	return
}
