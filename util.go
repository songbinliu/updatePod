package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func printPods(pods *v1.PodList) {
	fmt.Printf("api version:%s, kind:%s, r.version:%s\n",
		pods.APIVersion,
		pods.Kind,
		pods.ResourceVersion)

	for _, pod := range pods.Items {
		fmt.Printf("%s/%s, phase:%s, node.Name:%s, host:%s\n",
			pod.Namespace,
			pod.Name,
			pod.Status.Phase,
			pod.Spec.NodeName,
			pod.Status.HostIP)
	}
}

func testPod(client *kubernetes.Clientset) {
	pods, err := client.CoreV1().Pods(v1.NamespaceAll).List(metav1.ListOptions{})
	//pods, err := client.Pods("").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	printPods(pods)

	glog.V(2).Info("test finish")
}

func copyPodInfoX(oldPod, newPod *v1.Pod) {
	//typeMeta
	newPod.TypeMeta = oldPod.TypeMeta
	//newPod.Kind = oldPod.Kind
	//newPod.APIVersion = oldPod.APIVersion

	//objectMeta
	newPod.ObjectMeta = oldPod.ObjectMeta
	newPod.SelfLink = ""
	newPod.ResourceVersion = ""
	newPod.Generation = 0
	newPod.CreationTimestamp = metav1.Time{}
	newPod.DeletionTimestamp = nil
	newPod.DeletionGracePeriodSeconds = nil

	//newPod.Name = oldPod.Name
	//newPod.Namespace = oldPod.Namespace
	//newPod.Labels = oldPod.Labels
	//newPod.GenerateName = oldPod.GenerateName
	//newPod.Annotations = oldPod.Annotations
	//newPod.OwnerReferences = oldPod.OwnerReferences
	//newPod.Finalizers = oldPod.Finalizers
	//newPod.ClusterName = oldPod.ClusterName
	//newPod.UID = oldPod.UID

	//podSpec
	spec := oldPod.Spec
	spec.Hostname = ""
	spec.Subdomain = ""
	spec.NodeName = ""
	//spec := v1.PodSpec{
	//	Volumes:                       oldPod.Spec.Volumes,
	//	InitContainers:                oldPod.Spec.Containers,
	//	RestartPolicy:                 oldPod.Spec.RestartPolicy,
	//	TerminationGracePeriodSeconds: oldPod.Spec.TerminationGracePeriodSeconds,
	//	ActiveDeadlineSeconds:         oldPod.Spec.ActiveDeadlineSeconds,
	//	DNSPolicy:                     oldPod.Spec.DNSPolity,
	//	NodeSelector:                  oldPod.Spec.NodeSelector,
	//	ServiceAccountName:            oldPod.Spec.ServiceAccountName,
	//}

	newPod.Spec = spec
	return
}

func copyPodInfo(oldPod, newPod *v1.Pod) {
	//typeMeta
	newPod.Kind = oldPod.Kind
	newPod.APIVersion = oldPod.APIVersion

	////objectMeta
	//newPod.SelfLink = ""
	//newPod.ResourceVersion = ""
	//newPod.Generation = 0
	//newPod.CreationTimestamp = metav1.Time{}
	//newPod.DeletionTimestamp = nil
	//newPod.DeletionGracePeriodSeconds = nil

	newPod.Name = oldPod.Name
	newPod.Namespace = oldPod.Namespace
	newPod.Labels = oldPod.Labels
	newPod.GenerateName = oldPod.GenerateName
	newPod.Annotations = oldPod.Annotations
	newPod.OwnerReferences = oldPod.OwnerReferences
	newPod.Finalizers = oldPod.Finalizers
	newPod.ClusterName = oldPod.ClusterName
	newPod.UID = oldPod.UID

	//podSpec
	spec := oldPod.Spec
	spec.Hostname = ""
	spec.Subdomain = ""
	spec.NodeName = ""
	//spec := v1.PodSpec{
	//	Volumes:                       oldPod.Spec.Volumes,
	//	InitContainers:                oldPod.Spec.Containers,
	//	RestartPolicy:                 oldPod.Spec.RestartPolicy,
	//	TerminationGracePeriodSeconds: oldPod.Spec.TerminationGracePeriodSeconds,
	//	ActiveDeadlineSeconds:         oldPod.Spec.ActiveDeadlineSeconds,
	//	DNSPolicy:                     oldPod.Spec.DNSPolity,
	//	NodeSelector:                  oldPod.Spec.NodeSelector,
	//	ServiceAccountName:            oldPod.Spec.ServiceAccountName,
	//}

	newPod.Spec = spec
	return
}

func getKubeClient() *kubernetes.Clientset {

	if *masterUrl == "" && *kubeConfig == "" {
		fmt.Println("must specify masterUrl or kubeConfig.")
		return nil
	}

	var err error
	var config *restclient.Config

	if *kubeConfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", *kubeConfig)
	} else {
		config, err = clientcmd.BuildConfigFromFlags(*masterUrl, "")
	}

	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

func testUpdatePod(kclient *kubernetes.Clientset, ns, podName, schedulerName string) error {
	client := kclient.CoreV1().Pods(ns)
	//1. get
	option := metav1.GetOptions{}
	pod, err := client.Get(podName, option)
	if err != nil {
		fmt.Printf("failed to get pod:%v\n", err)
		return err
	}

	fmt.Printf("Pod:%v, %v\n", pod.Status.Phase, pod.Spec.SchedulerName)

	//2. update
	newScheduler := schedulerName
	pod.Spec.SchedulerName = newScheduler
	pod, err = client.Update(pod)
	if err != nil {
		fmt.Printf("failed to update Pod:%v\n", err)
		return err
	}

	//3. get it again
	pod, err = client.Get(podName, option)
	if err != nil {
		fmt.Printf("failed to get pod:%v\n", err)
		return err
	}
	fmt.Printf("Pod:%v, %v\n", pod.Status.Phase, pod.Spec.SchedulerName)
	return nil
}

func testKillUpdatePod(kclient *kubernetes.Clientset, nameSpace, podName, schedulerName string) error {
	client := kclient.CoreV1().Pods(nameSpace)
	//1. get
	option := metav1.GetOptions{}
	pod, err := client.Get(podName, option)
	if err != nil {
		fmt.Printf("failed to get pod:%v\n", err)
	} else {
		fmt.Printf("Pod:%v, %v\n", pod.Status.Phase, pod.Spec.SchedulerName)
	}

	//2. kill and create to update
	fmt.Printf("Begin to kill pod:%v\n", podName)
	var duration int64 = 10
	doption := &metav1.DeleteOptions{
		GracePeriodSeconds: &duration,
	}
	err = client.Delete(podName, doption)
	if err != nil {
		fmt.Printf("failed to delete Pod:%v\n", err)
		return err
	}

	time.Sleep(time.Second * 35)
	newScheduler := schedulerName
	npod := &v1.Pod{}
	copyPodInfo(pod, npod)
	npod.Spec.SchedulerName = newScheduler
	/* if we set the NodeName, then the default scheduler will bind it to the Node directly;
	so we don't have to set the scheduler name. */
	//npod.Spec.NodeName = "ip-172-23-1-39.us-west-2.compute.internal"
	npod, err = client.Create(npod)
	if err != nil {
		fmt.Printf("failed to create Pod:%v\n", err)
	}

	//3. get it again
	time.Sleep(time.Second * 10)
	pod, err = client.Get(podName, option)
	if err != nil {
		fmt.Printf("failed to get pod:%v\n", err)
		return err
	}
	fmt.Printf("Pod:%v, %v\n", pod.Status.Phase, pod.Spec.SchedulerName)
	return nil
}

func testUpdateController(client *kubernetes.Clientset, nameSpace, rcName, schedulerName string) error {
	rcClient := client.CoreV1().ReplicationControllers(nameSpace)

	//1. get
	option := metav1.GetOptions{}
	rc, err := rcClient.Get(rcName, option)
	if err != nil {
		glog.Errorf("failed to get ReplicationController:%v\n", err)
		return err
	}
	fmt.Printf("ReplicationController:%v, replicaNum:%v\n",
		rc.Spec.Template.Spec.SchedulerName,
		*rc.Spec.Replicas)

	//2. update
	//*rc.Spec.Replicas = *(rc.Spec.Replicas) + 1
	//p := rc.Spec.Replicas
	//if *p > 3 {
	//	*p -= 1
	//} else {
	//	*p += 1
	//}
	newScheduler := schedulerName
	rc.Spec.Template.Spec.SchedulerName = newScheduler
	rc, err = rcClient.Update(rc)
	if err != nil {
		glog.Warningf("failed to update RC:%v\n", err)
		return err
	}

	//3. get it again
	rc, err = rcClient.Get(rcName, option)
	if err != nil {
		fmt.Printf("failed to get ReplicationController:%v\n", err)
		return err
	}
	fmt.Printf("ReplicationController:%v, replicaNum:%v\n",
		rc.Spec.Template.Spec.SchedulerName,
		*rc.Spec.Replicas)

	return nil
}

func getLabelSelector(rc *v1.ReplicationController) string {
	glog.V(3).Infof("selectors = %d", len(rc.Spec.Selector))
	data := make([]string, len(rc.Spec.Selector))
	i := 0
	for key, value := range rc.Spec.Selector {
		glog.V(2).Infof("key=[%s],val=[%s]", key, value)
		data[i] = key + "=" + value
		i++
	}

	if len(data) == 1 {
		return data[0]
	}
	sort.StringSlice(data).Sort()
	return strings.Join(data, ",")
}

func selectNode(nodes *[]string) string {
	idx := rand.Intn(len(*nodes))
	return (*nodes)[idx]
}

func selectPod(pods *v1.PodList) *v1.Pod {
	idx := rand.Intn(len(pods.Items))
	return &(pods.Items[idx])
}

func genListOption(rc *v1.ReplicationController) *metav1.ListOptions {
	labelSelector := getLabelSelector(rc)
	glog.V(2).Infof("labelSelector:[%v]", labelSelector)
	fieldSelector := "status.phase=" + string(v1.PodRunning)
	loption := metav1.ListOptions{LabelSelector: labelSelector,
		FieldSelector: fieldSelector,
	}

	return &loption
}

func testScaleUpController(client *kubernetes.Clientset, nameSpace, rcName, schedulerName string) error {
	rcClient := client.CoreV1().ReplicationControllers(nameSpace)
	podClient := client.CoreV1().Pods(nameSpace)

	nodes, err := testGetNode(client)
	glog.V(2).Infof("nodes:[%v], [%v]\n", nodes, (*nodes)[0])

	//1. get
	option := metav1.GetOptions{}
	rc, err := rcClient.Get(rcName, option)
	if err != nil {
		glog.Errorf("failed to get ReplicationController:%v\n", err)
		return err
	}
	glog.V(2).Infof("ReplicationController:%v, replicaNum:%v\n",
		rc.Spec.Template.Spec.SchedulerName,
		*rc.Spec.Replicas)

	//2. move: kill one pod, and create another pod.
	lstOption := genListOption(rc)
	pods, err := podClient.List(*lstOption)
	if err != nil {
		glog.Errorf("failed to get Pods for rc:%s\n%v", rcName, err.Error())
		return err
	}

	if len(pods.Items) < 1 {
		glog.Warningf("no living Pods for rc:%s\n", rcName)
		return nil
	}

	//2.1 select a pod, and copy it
	pod := selectPod(pods)
	npod := &v1.Pod{}
	copyPodInfo(pod, npod)
	//npod.Name = pod.Name + "mv"
	npod.Name = pod.Name

	// if NodeName is not set, then ReplicationController will create another pod sooner than me.
	nodeName := selectNode(nodes)
	npod.Spec.NodeName = nodeName
	//pod.Spec.SchedulerName = "xyzscheduler"
	glog.V(3).Infof("nodeName=[%s], [%s]\n", nodeName, npod.Spec.NodeName)

	//2.2 kill the pod
	var duration int64 = 0
	delOption := &metav1.DeleteOptions{
		GracePeriodSeconds: &duration,
	}
	glog.V(2).Infof("begin to kill pod:%v", pod.Name)
	err = podClient.Delete(pod.Name, delOption)
	if err != nil {
		glog.Errorf("failed to delete pod:%v\n%v\n", pod.Name, err)
		return err
	}

	//2.3 create the new Pod
	glog.V(2).Infof("begin to create pod:%v on %v", npod.Name, npod.Spec.NodeName)
	npod, err = podClient.Create(npod)
	if err != nil {
		glog.Errorf("failed to create Pod:%v\n%v\n", npod.Name, err)
		return err
	}

	return nil
}

func testGetNode(client *kubernetes.Clientset) (*[]string, error) {
	rcClient := client.CoreV1().Nodes()

	option := metav1.ListOptions{}
	nodeList, err := rcClient.List(option)
	if err != nil {
		fmt.Printf("failed to get list: %v\n", err.Error())
		return nil, err
	}

	fmt.Printf("There are %v nodes:\n", len(nodeList.Items))

	result := make([]string, len(nodeList.Items))

	for i, node := range nodeList.Items {
		//fmt.Printf("%v\n", node.Name)
		result[i] = node.Name
	}

	return &result, nil
}

func testGetPodbyUUID(client *kubernetes.Clientset, namespace, uid string) {
	podClient := client.CoreV1().Pods(namespace)
	nodeName := "ip-172-23-1-39.us-west-2.compute.internal"
	option := metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
		//FieldSelector: "metadata.uid=" + uid,
	}

	podList, err := podClient.List(option)
	if err != nil {
		glog.Errorf("failed to list Pod:%v", err.Error())
		return
	}

	fmt.Printf("\n%v pods on Node:%v\n", len(podList.Items), nodeName)

	printPods(podList)
}
