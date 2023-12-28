package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	appv1 "k8s.io/api/apps/v1"
	bachV1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magieAgent/internal/config"
	"github.com/muidea/magieAgent/internal/core/module/k8s/pkg/mariadb"
	"github.com/muidea/magieAgent/pkg/common"
)

func (s *K8s) checkIsMariadb(deploymentPtr *appv1.Deployment) bool {
	return strings.Index(deploymentPtr.ObjectMeta.GetName(), common.Mariadb) != -1
}

func (s *K8s) getMariadbServiceInfo(deploymentPtr *appv1.Deployment, clientSet *kubernetes.Clientset) (ret *common.ServiceInfo, err *cd.Result) {
	ptr := &common.ServiceInfo{
		Name:      deploymentPtr.ObjectMeta.GetName(),
		Namespace: deploymentPtr.ObjectMeta.GetNamespace(),
		Labels:    deploymentPtr.ObjectMeta.Labels,
		Svc: &common.Svc{
			Host: config.GetLocalHost(),
		},
		Replicas: *deploymentPtr.Spec.Replicas,
		Status: &common.Status{
			Available: deploymentPtr.Status.UnavailableReplicas == 0,
		},
	}
	if len(deploymentPtr.Spec.Template.Spec.Containers) > 0 {
		ptr.Image = deploymentPtr.Spec.Template.Spec.Containers[0].Image
		ptr.Spec = &common.Spec{
			CPU:    deploymentPtr.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String(),
			Memory: deploymentPtr.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String(),
		}
	}

	dataPath, dataErr := s.getServiceDataPath(deploymentPtr, clientSet)
	if dataErr != nil {
		err = dataErr
		log.Errorf("getServiceInfoFromDeployment failed, s.getServiceDataPath %v error:%v", deploymentPtr.ObjectMeta.GetName(), dataErr.Error())
		return
	}

	backPath, _ := s.getServiceBackPath(deploymentPtr, clientSet)

	portVal, portErr := s.getHAServicePort(deploymentPtr, clientSet)
	if portErr != nil {
		err = portErr
		log.Errorf("getServiceInfoFromDeployment failed, s.getHAServicePort %v error:%v", deploymentPtr.ObjectMeta.GetName(), portErr.Error())
		return
	}

	ptr.Catalog = common.Mariadb
	ptr.Volumes = &common.Volumes{
		ConfPath: &common.Path{
			Name:  "my.cnf",
			Value: common.DefaultMariadbConfigPath,
			Type:  common.InnerPath,
		},
		DataPath: dataPath,
		BackPath: backPath,
	}
	ptr.Env = s.getEnv(deploymentPtr, clientSet)
	ptr.Svc.Port = portVal
	ret = ptr
	return
}

func (s *K8s) getDefaultMariadbServiceInfo(serviceName string) (ret *common.ServiceInfo) {
	ret = &common.ServiceInfo{
		Name:      serviceName,
		Namespace: s.getNamespace(),
		Catalog:   common.Mariadb,
		Image:     common.DefaultMariadbImage,
		Labels:    common.DefaultLabels,
		Spec:      &common.MariadbDefaultSpec,
		Volumes: &common.Volumes{
			ConfPath: &common.Path{
				Name:  "config",
				Value: common.DefaultMariadbConfigPath,
				Type:  common.InnerPath,
			},
			BackPath: &common.Path{
				Name:  "back-path",
				Value: common.DefaultMariadbBackPath,
				Type:  common.HostPath,
			},
		},
		Env: &common.Env{
			Root:     common.DefaultMariadbRoot,
			Password: common.DefaultMariadbPassword,
		},
		Svc: &common.Svc{
			Host: config.GetLocalHost(),
			Port: common.DefaultMariadbPort,
		},
		Replicas: 1,
	}

	ret.Labels["app"] = serviceName
	ret.Labels["catalog"] = common.Mariadb
	return
}

func (s *K8s) createMariadb(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	// 1、Create pvc
	_, pvcErr := s.clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Create(context.TODO(),
		mariadb.GetPersistentVolumeClaims(serviceInfo),
		metav1.CreateOptions{})
	if pvcErr != nil {
		err = cd.NewError(cd.UnExpected, pvcErr.Error())
		log.Errorf("createMariadb %v pvc failed, s.clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Create error:%s",
			serviceInfo, pvcErr.Error())
		return
	}

	// 2、Create Deployment
	_, deploymentErr := s.clientSet.AppsV1().Deployments(s.getNamespace()).Create(context.TODO(),
		mariadb.GetDeployment(serviceInfo),
		metav1.CreateOptions{})
	if deploymentErr != nil {
		err = cd.NewError(cd.UnExpected, deploymentErr.Error())
		log.Errorf("createMariadb %v deployment failed, s.clientSet.AppsV1().Deployments(s.getNamespace()).Create error:%s",
			serviceInfo, deploymentErr.Error())

		s.clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})
		return
	}

	// 3、Create Service
	_, serviceErr := s.clientSet.CoreV1().Services(s.getNamespace()).Create(context.TODO(),
		mariadb.GetService(serviceInfo),
		metav1.CreateOptions{})
	if serviceErr != nil {
		err = cd.NewError(cd.UnExpected, serviceErr.Error())
		log.Errorf("createMariadb %v service failed, s.clientSet.CoreV1().Services(s.getNamespace()).Create error:%s",
			serviceInfo, serviceErr.Error())

		s.clientSet.AppsV1().Deployments(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})
		s.clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})
		return
	}

	return
}

func (s *K8s) destroyMariadb(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	_ = s.clientSet.CoreV1().Services(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})
	_ = s.clientSet.AppsV1().Deployments(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})
	_ = s.clientSet.CoreV1().PersistentVolumeClaims(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})

	return
}

func (s *K8s) startMariadb(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	deploymentPtr, deploymentErr := s.clientSet.AppsV1().Deployments(s.getNamespace()).Get(context.TODO(), serviceInfo.Name, metav1.GetOptions{})
	if deploymentErr != nil {
		err = cd.NewError(cd.UnExpected, deploymentErr.Error())
		log.Errorf("startMariadb %v failed, get service scale error:%s", serviceInfo, deploymentErr.Error())
		return
	}

	replicas := func() *int32 {
		val := int32(1)
		return &val
	}

	deploymentPtr.Spec.Replicas = replicas()
	deploymentPtr, deploymentErr = s.clientSet.AppsV1().Deployments(s.getNamespace()).Update(
		context.TODO(),
		deploymentPtr,
		metav1.UpdateOptions{},
	)
	if deploymentErr != nil {
		err = cd.NewError(cd.UnExpected, deploymentErr.Error())
		log.Errorf("startMariadb %v failed, set service scale error:%s", serviceInfo, deploymentErr.Error())
		return
	}

	watchPtr, watchErr := s.clientSet.AppsV1().Deployments(s.getNamespace()).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", deploymentPtr.Name),
	})
	if watchErr != nil {
		err = cd.NewError(cd.UnExpected, watchErr.Error())
		log.Errorf("startMariadb %v service failed, s.clientSet.AppsV1().Deployments(s.getNamespace()).Watch error:%s",
			serviceInfo, watchErr.Error())
		return
	}
	defer watchPtr.Stop()

	startTime := time.Now()
	for event := range watchPtr.ResultChan() {
		deploymentPtr, ok := event.Object.(*appv1.Deployment)
		if !ok {
			continue
		}

		if time.Since(startTime) > 30*time.Minute {
			err = cd.NewError(cd.UnExpected, fmt.Sprintf("takes more than 30 minutes"))
			log.Errorf("startMariadb %v service failed, check mariadb status error:%s",
				serviceInfo, err.Error())
			break
		}

		if deploymentPtr.Status.AvailableReplicas != deploymentPtr.Status.Replicas {
			continue
		}

		break
	}
	return
}

func (s *K8s) stopMariadb(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	deploymentPtr, deploymentErr := s.clientSet.AppsV1().Deployments(s.getNamespace()).Get(context.TODO(), serviceInfo.Name, metav1.GetOptions{})
	if deploymentErr != nil {
		err = cd.NewError(cd.UnExpected, deploymentErr.Error())
		log.Errorf("stopMariadb %v failed, get service scale error:%s", serviceInfo, deploymentErr.Error())
		return
	}

	replicas := func() *int32 {
		val := int32(0)
		return &val
	}

	deploymentPtr.Spec.Replicas = replicas()
	deploymentPtr, deploymentErr = s.clientSet.AppsV1().Deployments(s.getNamespace()).Update(
		context.TODO(),
		deploymentPtr,
		metav1.UpdateOptions{},
	)
	if deploymentErr != nil {
		err = cd.NewError(cd.UnExpected, deploymentErr.Error())
		log.Errorf("stopMariadb %v failed, set service scale error:%s", serviceInfo, deploymentErr.Error())
		return
	}

	watchPtr, watchErr := s.clientSet.AppsV1().Deployments(s.getNamespace()).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", deploymentPtr.Name),
	})
	if watchErr != nil {
		err = cd.NewError(cd.UnExpected, watchErr.Error())
		log.Errorf("stopMariadb %v service failed, s.clientSet.AppsV1().Deployments(s.getNamespace()).Watch error:%s",
			serviceInfo, watchErr.Error())
		return
	}
	defer watchPtr.Stop()

	startTime := time.Now()
	for event := range watchPtr.ResultChan() {
		deploymentPtr, ok := event.Object.(*appv1.Deployment)
		if !ok {
			continue
		}

		if time.Since(startTime) > 30*time.Minute {
			err = cd.NewError(cd.UnExpected, fmt.Sprintf("takes more than 30 minutes"))
			log.Errorf("stopMariadb %v service failed, check mariadb status error:%s",
				serviceInfo, err.Error())
			break
		}

		if deploymentPtr.Status.AvailableReplicas != 0 {
			continue
		}

		break
	}
	return
}

func (s *K8s) jobMariadb(serviceInfo *common.ServiceInfo, command []string) (err *cd.Result) {
	// 这里主动清理一次上次遗留下来的同名Job，避免新的job创建失败
	_, preErr := s.clientSet.BatchV1().Jobs(s.getNamespace()).Get(context.TODO(), serviceInfo.Name, metav1.GetOptions{})
	if preErr == nil {
		_ = s.clientSet.BatchV1().Jobs(s.getNamespace()).Delete(context.TODO(), serviceInfo.Name, metav1.DeleteOptions{})
	}

	jobPtr, jobErr := s.clientSet.BatchV1().Jobs(s.getNamespace()).Create(context.TODO(),
		mariadb.GetJob(serviceInfo, command),
		metav1.CreateOptions{})
	if jobErr != nil {
		err = cd.NewError(cd.UnExpected, jobErr.Error())
		log.Errorf("jobMariadb %v service failed, s.clientSet.BatchV1().Jobs(s.getNamespace()).Create error:%s",
			serviceInfo, jobErr.Error())
		return
	}

	watchPtr, watchErr := s.clientSet.BatchV1().Jobs(s.getNamespace()).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", jobPtr.Name),
	})
	if watchErr != nil {
		log.Warnf("jobMariadb %v service failed, s.clientSet.BatchV1().Jobs(s.getNamespace()).Watch error:%s",
			serviceInfo, watchErr.Error())
		return
	}
	defer watchPtr.Stop()

	for event := range watchPtr.ResultChan() {
		job, ok := event.Object.(*bachV1.Job)
		if !ok {
			continue
		}

		// Check if the Job is completed
		if job.Status.Failed > 0 {
			err = cd.NewError(cd.UnExpected, fmt.Sprintf("execution failed"))
			log.Errorf("jobMariadb %v service failed, execution failed", serviceInfo)
			break
		}

		// Check if the Job is completed
		if job.Status.CompletionTime != nil {
			break
		}
	}

	return
}
