package biz

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magieAgent/pkg/common"
)

func (s *K8s) execInPod(clientSet *kubernetes.Clientset, clientConfig *rest.Config, namespace, podName, containerName, command string) (stdout []byte, stderr []byte, err *cd.Result) {
	log.Infof("execInPod namespace:%s podName:%s containerName:%s command:%v", namespace, podName, containerName, command)
	cmd := []string{
		"sh",
		"-c",
		command,
	}

	req := clientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).SubResource("exec").Param("container", containerName)
	req.VersionedParams(
		&corev1.PodExecOptions{
			Command: cmd,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		},
		scheme.ParameterCodec,
	)

	var stdoutBuff, stderrBuff bytes.Buffer
	execPtr, execErr := remotecommand.NewSPDYExecutor(clientConfig, "POST", req.URL())
	if execErr != nil {
		err = cd.NewError(cd.UnExpected, execErr.Error())
		log.Errorf("execInPod failed, remotecommand.NewSPDYExecutor error:%s", err.Error())
		return
	}

	execErr = execPtr.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdoutBuff,
		Stderr: &stderrBuff,
	})
	if execErr != nil {
		err = cd.NewError(cd.UnExpected, execErr.Error())
		log.Errorf("execInPod failed, execPtr.Stream error:%s", err.Error())
		return
	}

	stdout = stdoutBuff.Bytes()
	stderr = stderrBuff.Bytes()
	return
}

func (s *K8s) ExecuteCommand(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("ExecuteCommand failed, nil param")
		return
	}

	cmdInfoPtr, cmdInfoOK := param.(*common.CmdInfo)
	if !cmdInfoOK || cmdInfoPtr == nil {
		log.Warnf("ExecuteCommand failed, illegal param")
		return
	}

	log.Infof("ExecuteCommand service:%v, serviceInfo:%v", cmdInfoPtr.Service, cmdInfoPtr.ServiceInfo)
	deploymentPtr, deploymentErr := s.clientSet.AppsV1().Deployments(s.getNamespace()).Get(context.TODO(), cmdInfoPtr.Service, metav1.GetOptions{})
	if deploymentErr != nil {
		log.Errorf("ExecuteCommand failed, s.clientSet.AppsV1().Deployments(s.getNamespace()).Get error:%s", deploymentErr.Error())
		if re != nil {
			re.Set(nil, cd.NewError(cd.UnExpected, deploymentErr.Error()))
		}

		return
	}

	const podTemplateHash = "pod-template-hash"
	podTemplateHashVal := deploymentPtr.ObjectMeta.Labels["pod-template-hash"]
	requirementPtr, _ := labels.NewRequirement(podTemplateHash, selection.In, []string{podTemplateHashVal})

	selectorPtr, selectorErr := metav1.LabelSelectorAsSelector(deploymentPtr.Spec.Selector)
	if selectorErr != nil {
		log.Errorf("ExecuteCommand failed, metav1.LabelSelectorAsSelector error:%s", selectorErr.Error())
		if re != nil {
			re.Set(nil, cd.NewError(cd.UnExpected, selectorErr.Error()))
		}

		return
	}
	selectorPtr.Add(*requirementPtr)

	podList, podsErr := s.clientSet.CoreV1().Pods(s.getNamespace()).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selectorPtr.String(),
	})
	if podsErr != nil || len(podList.Items) == 0 {
		if podsErr == nil {
			podsErr = fmt.Errorf("not exist %s pods", cmdInfoPtr.ServiceInfo.Name)
		}

		log.Errorf("ExecuteCommand failed, s.clientSet.CoreV1().Pods(s.getNamespace()).List error:%s", podsErr.Error())
		if re != nil {
			re.Set(nil, cd.NewError(cd.UnExpected, podsErr.Error()))
		}
		return
	}

	podName := podList.Items[0].Name
	containerName := podList.Items[0].Spec.Containers[0].Name
	commandVal := strings.Join(cmdInfoPtr.Command, " ")
	resultData, errorData, resultErr := s.execInPod(s.clientSet, s.clientConfig, s.getNamespace(), podName, containerName, commandVal)
	if re != nil {
		re.Set(resultData, resultErr)
		re.SetVal("stderr", errorData)
	}
}

func (s *K8s) CreateService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("CreateService failed, nil param")
		return
	}

	serviceName, serviceOK := param.(string)
	if !serviceOK {
		log.Warnf("CreateService failed, illegal param")
		return
	}
	catalog := ev.GetData("catalog")
	if catalog == nil {
		log.Warnf("CreateService failed, illegal catalog")
		return
	}

	var err *cd.Result
	switch catalog.(string) {
	case common.Mariadb:
		serviceInfo := s.getDefaultMariadbServiceInfo(serviceName)
		err = s.createMariadb(serviceInfo)
	default:
		panic(fmt.Sprintf("illegal catalog:%v", catalog))
	}

	if re != nil {
		re.Set(nil, err)
	}
}

func (s *K8s) DestroyService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("DestroyService failed, nil param")
		return
	}

	serviceName, serviceOK := param.(string)
	if !serviceOK {
		log.Warnf("DestroyService failed, illegal param")
		return
	}
	catalog := ev.GetData("catalog")
	if catalog == nil {
		log.Warnf("DestroyService failed, illegal catalog")
		return
	}

	var err *cd.Result
	switch catalog.(string) {
	case common.Mariadb:
		serviceInfo, serviceErr := s.Query(serviceName, common.Mariadb)
		if serviceErr == nil {
			err = s.destroyService(serviceInfo)
		} else {
			err = serviceErr
		}
	default:
		panic(fmt.Sprintf("illegal catalog:%v", catalog))
	}

	if re != nil {
		re.Set(nil, err)
	}
}

func (s *K8s) StartService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("StartService failed, nil param")
		return
	}

	cmdInfoPtr, cmdInfoOK := param.(*common.CmdInfo)
	if !cmdInfoOK {
		log.Warnf("StartService failed, illegal param")
		return
	}

	err := s.startService(cmdInfoPtr.ServiceInfo)
	if re != nil {
		re.Set(nil, err)
	}
}

func (s *K8s) StopService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("StopService failed, nil param")
		return
	}

	cmdInfoPtr, cmdInfoOK := param.(*common.CmdInfo)
	if !cmdInfoOK {
		log.Warnf("StopService failed, illegal param")
		return
	}

	err := s.stopService(cmdInfoPtr.ServiceInfo)
	if re != nil {
		re.Set(nil, err)
	}
}

func (s *K8s) ChangeServiceEndpoints(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("ChangeServiceEndpoints failed, nil param")
		return
	}

	serviceInfo, serviceOK := param.(*common.ServiceInfo)
	if !serviceOK {
		log.Warnf("ChangeServiceEndpoints failed, illegal param")
		return
	}
	endpointPtr := ev.GetData("endpoint").(*common.Endpoint)
	err := s.changeServiceEndpoints(serviceInfo, endpointPtr)
	if re != nil {
		re.Set(nil, err)
	}
	return
}

func (s *K8s) RestoreServiceEndpoints(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("RestoreServiceEndpoints failed, nil param")
		return
	}

	serviceInfo, serviceOK := param.(*common.ServiceInfo)
	if !serviceOK {
		log.Warnf("RestoreServiceEndpoints failed, illegal param")
		return
	}
	err := s.restoreServiceEndpoints(serviceInfo)
	if re != nil {
		re.Set(nil, err)
	}
	return
}

func (s *K8s) JobService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("StopService failed, nil param")
		return
	}

	cmdInfoPtr, cmdInfoOK := param.(*common.CmdInfo)
	if !cmdInfoOK {
		log.Warnf("StopService failed, illegal param")
		return
	}

	jobServiceInfo := *cmdInfoPtr.ServiceInfo
	jobServiceInfo.Name = cmdInfoPtr.Service

	err := s.jobService(&jobServiceInfo, cmdInfoPtr.Command)
	if re != nil {
		re.Set(nil, err)
	}
	return
}

func (s *K8s) ListService(ev event.Event, re event.Result) {
	catalog2ServiceList := s.enumService()
	if re != nil {
		re.Set(catalog2ServiceList, nil)
	}
}

func (s *K8s) enumService() common.Catalog2ServiceList {
	catalog2ServiceList := common.Catalog2ServiceList{}

	mariadbList := common.ServiceList{}
	tsDBList := common.ServiceList{}
	mongoList := common.ServiceList{}
	postgreList := common.ServiceList{}
	fileList := common.ServiceList{}
	vxBaseList := common.ServiceList{}
	serviceList := s.serviceCache.GetAll()
	for _, val := range serviceList {
		servicePtr := val.(*common.ServiceInfo)
		switch servicePtr.Catalog {
		case common.Mariadb:
			mariadbList = append(mariadbList, servicePtr.Name)
		case common.TsDB:
			tsDBList = append(tsDBList, servicePtr.Name)
		case common.MongoDB:
			mongoList = append(mongoList, servicePtr.Name)
		case common.PostgreSQL:
			postgreList = append(postgreList, servicePtr.Name)
		case common.FileService:
			fileList = append(fileList, servicePtr.Name)
		case common.VxBase:
			vxBaseList = append(vxBaseList, servicePtr.Name)
		}
	}
	if len(mariadbList) > 0 {
		catalog2ServiceList[common.Mariadb] = mariadbList
	}
	if len(tsDBList) > 0 {
		catalog2ServiceList[common.TsDB] = tsDBList
	}
	if len(mongoList) > 0 {
		catalog2ServiceList[common.MongoDB] = mongoList
	}
	if len(postgreList) > 0 {
		catalog2ServiceList[common.PostgreSQL] = postgreList
	}
	if len(fileList) > 0 {
		catalog2ServiceList[common.FileService] = fileList
	}
	if len(vxBaseList) > 0 {
		catalog2ServiceList[common.VxBase] = vxBaseList
	}

	return catalog2ServiceList
}

func (s *K8s) QueryService(ev event.Event, re event.Result) {
	param := ev.Data()
	if param == nil {
		log.Warnf("QueryService failed, nil param")
		return
	}

	serviceName, serviceOK := param.(string)
	if !serviceOK {
		log.Warnf("QueryService failed, illegal param")
		return
	}

	catalog := ev.GetData("catalog")
	if catalog == nil {
		log.Warnf("QueryService failed, illegal catalog")
		return
	}

	serviceInfoPtr, serviceInfoErr := s.Query(serviceName, catalog.(string))
	if re != nil {
		re.Set(serviceInfoPtr, serviceInfoErr)
	}
}

func (s *K8s) createService(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	switch serviceInfo.Catalog {
	case common.Mariadb:
		err = s.createMariadb(serviceInfo)
	}

	return
}

func (s *K8s) destroyService(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	switch serviceInfo.Catalog {
	case common.Mariadb:
		err = s.destroyMariadb(serviceInfo)
	}

	return
}

func (s *K8s) startService(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	switch serviceInfo.Catalog {
	case common.Mariadb:
		err = s.startMariadb(serviceInfo)
	}

	return
}

func (s *K8s) stopService(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	switch serviceInfo.Catalog {
	case common.Mariadb:
		err = s.stopMariadb(serviceInfo)
	}

	return
}

func (s *K8s) changeServiceEndpoints(serviceInfo *common.ServiceInfo, endpointVal *common.Endpoint) (err *cd.Result) {
	servicePtr, serviceErr := s.clientSet.CoreV1().Services(s.getNamespace()).Get(context.TODO(), serviceInfo.Name, metav1.GetOptions{})
	if serviceErr != nil {
		err = cd.NewError(cd.UnExpected, serviceErr.Error())
		log.Errorf("changeServiceEndpoints %v failed, s.clientSet.CoreV1().Services(s.getNamespace()).Get error:%s", serviceInfo, serviceErr)
		return
	}

	servicePtr.Spec.Selector = nil
	servicePtr, serviceErr = s.clientSet.CoreV1().Services(s.getNamespace()).Update(context.TODO(), servicePtr, metav1.UpdateOptions{})
	if serviceErr != nil {
		err = cd.NewError(cd.UnExpected, serviceErr.Error())
		log.Errorf("changeServiceEndpoints %v failed, s.clientSet.CoreV1().Services(s.getNamespace()).Update error:%s", serviceInfo, serviceErr)
		return
	}

	endpointPtr, endpointErr := s.clientSet.CoreV1().Endpoints(s.getNamespace()).Get(context.TODO(), serviceInfo.Name, metav1.GetOptions{})
	if endpointErr != nil {
		err = cd.NewError(cd.UnExpected, endpointErr.Error())
		log.Errorf("changeServiceEndpoints %v failed, s.clientSet.CoreV1().Endpoints(s.getNamespace()).Get error:%s", serviceInfo, endpointErr)
		return
	}

	endpointPtr.Subsets = []corev1.EndpointSubset{
		{
			Addresses: []corev1.EndpointAddress{
				{IP: endpointVal.Host},
			},
			Ports: []corev1.EndpointPort{
				{
					Name:     "default",
					Port:     endpointVal.Port,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
	endpointPtr, endpointErr = s.clientSet.CoreV1().Endpoints(s.getNamespace()).Update(context.TODO(), endpointPtr, metav1.UpdateOptions{})
	if endpointErr != nil {
		err = cd.NewError(cd.UnExpected, endpointErr.Error())
		log.Errorf("changeServiceEndpoints %v failed, s.clientSet.CoreV1().Endpoints(s.getNamespace()).Update error:%s", serviceInfo, endpointErr)
		return
	}
	return
}

func (s *K8s) restoreServiceEndpoints(serviceInfo *common.ServiceInfo) (err *cd.Result) {
	servicePtr, serviceErr := s.clientSet.CoreV1().Services(s.getNamespace()).Get(context.TODO(), serviceInfo.Name, metav1.GetOptions{})
	if serviceErr != nil {
		err = cd.NewError(cd.UnExpected, serviceErr.Error())
		log.Errorf("restoreServiceEndpoints %v failed, s.clientSet.CoreV1().Services(s.getNamespace()).Get error:%s", serviceInfo, serviceErr)
		return
	}
	servicePtr.Spec.Selector = serviceInfo.Labels
	servicePtr, serviceErr = s.clientSet.CoreV1().Services(s.getNamespace()).Update(context.TODO(), servicePtr, metav1.UpdateOptions{})
	if serviceErr != nil {
		err = cd.NewError(cd.UnExpected, serviceErr.Error())
		log.Errorf("restoreServiceEndpoints %v failed, s.clientSet.CoreV1().Services(s.getNamespace()).Update error:%s", serviceInfo, serviceErr)
		return
	}

	return
}

func (s *K8s) jobService(serviceInfo *common.ServiceInfo, command []string) (err *cd.Result) {
	switch serviceInfo.Catalog {
	case common.Mariadb:
		err = s.jobMariadb(serviceInfo, command)
	}

	return
}
