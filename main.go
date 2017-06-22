/*
When to reschedule a Pod with our own scheduler, we needs to change the scheduler-name of the Pod.

This piece of code is to test how to update a Pod's scheduler-name:
(1) If pod is created by ReplicationController, then set the scheduler.name in the ReplicationController.Template;
(2) If the Pod is created without a ReplicationController/ReplicateSet, then kill & re-create with the new scheduler.name;
*/

// Note: only works with kubernetes 1.6+.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
)

//global variables
var (
	masterUrl     *string
	kubeConfig    *string
	nameSpace     *string
	podName       *string
	rcName        *string
	schedulerName *string
	uuid          *string
	nodeName      *string
)

func setFlags() {
	masterUrl = flag.String("masterUrl", "", "master url")
	kubeConfig = flag.String("kubeConfig", "", "absolute path to the kubeconfig file")
	nameSpace = flag.String("nameSpace", "default", "kubernetes object namespace")
	podName = flag.String("podName", "myschedule-cpu-80", "the podName to be handled")
	rcName = flag.String("rcName", "cpu-group", "the ReplicationController name")
	schedulerName = flag.String("scheduler-name", api.DefaultSchedulerName, "the name of the scheduler")
	uuid = flag.String("uuid", "", "the UUID of object")
	nodeName = flag.String("nodeName", "", "Destination of move")

	flag.Parse()
	flag.Set("logtostderr", "true")
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func testMovePod(client *kubernetes.Clientset, namespace, podName, nodeName string) {
	if namespace == "" || podName == "" || nodeName == "" {
		glog.Errorf("should not be emtpy: ns=[%v], podName=[%v], nodeName=[%v]",
			namespace, podName, nodeName)
		return
	}

	podClient := client.CoreV1().Pods(namespace)
	if podClient == nil {
		glog.Errorf("cannot get Pod client for nameSpace:%v", namespace)
		return
	}

	//1. get original pod
	getOption := metav1.GetOptions{}
	pod, err := podClient.Get(podName, getOption)
	if err != nil {
		err = fmt.Errorf("move-failed: get original pod:%v/%v\n%v",
			namespace, podName, err.Error())
		glog.Error(err.Error())
		return
	}

	if pod.Spec.NodeName == nodeName {
		err = fmt.Errorf("move-abort: pod %v/%v is already on node: %v",
			namespace, podName, nodeName)
		glog.Error(err.Error())
		return
	}

	glog.V(2).Infof("move-pod: begin to move %v/%v from %v to %v",
		namespace, pod.Name, pod.Spec.NodeName, nodeName)

	//2. copy and kill original pod
	npod := &v1.Pod{}
	copyPodInfo(pod, npod)
	npod.Spec.NodeName = nodeName

	var grace int64 = 0
	delOption := &metav1.DeleteOptions{GracePeriodSeconds: &grace}
	err = podClient.Delete(pod.Name, delOption)
	if err != nil {
		err = fmt.Errorf("move-failed: failed to delete original pod: %v/%v\n%v",
			namespace, pod.Name, err.Error())
		glog.Error(err.Error())
		return
	}

	//3. create (and bind) the new Pod
	_, err = podClient.Create(npod)
	if err != nil {
		err = fmt.Errorf("move-failed: failed to create new pod: %v/%v\n%v",
			namespace, npod.Name, err.Error())
		glog.Error(err.Error())
		return
	}

	//4. check the new Pod
	time.Sleep(time.Second * 1)
	if err = checkPodLive(client, namespace, npod.Name); err != nil {
		glog.Errorf("move-failed: check failed:%v\n", err.Error())
		return
	}

	glog.V(2).Infof("move-success: %v/%v from %v to %v",
		namespace, pod.Name, pod.Spec.NodeName, nodeName)

	return
}

func checkPodLive(client *kubernetes.Clientset, namespace, name string) error {
	podClient := client.CoreV1().Pods(namespace)
	xpod, err := podClient.Get(name, metav1.GetOptions{})
	if err != nil {
		err = fmt.Errorf("fail to get Pod: %v/%v\n%v\n",
			namespace, name, err.Error())
		glog.Errorf(err.Error())
		return err
	}

	goodStatus := map[v1.PodPhase]bool{
		v1.PodRunning: true,
		v1.PodPending: true,
	}

	ok := goodStatus[xpod.Status.Phase]
	if ok {
		return nil
	}

	err = fmt.Errorf("pod.status=%v\n", xpod.Status.Phase)
	return err
}

func main() {
	setFlags()
	fmt.Printf("kubeConfig=%v, masterUrl=%v\n", *kubeConfig, *masterUrl)
	glog.V(1).Info("begin tests")
	defer glog.V(1).Info("end of tests")

	kubeclient := getKubeClient()
	if kubeclient == nil {
		fmt.Println("failed to get kubeclient")
		return
	}

	testPod(kubeclient)

	//will fail:failed to update Pod:Pod "myschedule-cpu-80" is invalid: spec:
	// Forbidden: pod updates may not change fields other than `containers[*].image` or `spec.activeDeadlineSeconds` or `spec.tolerations`
	//(only additions to existing tolerations)
	//testUpdatePod(kubeclient, *nameSpace, *podName, *schedulerName)

	////Kill & reCreate it
	testKillUpdatePod(kubeclient, *nameSpace, *podName, *schedulerName)

	////Update ReplicationController, kill & wait for RC to reCreate it.
	//testUpdateController(kubeclient, *nameSpace, *rcName, *schedulerName)
	//time.Sleep(time.Second * 5)

	//testScaleUpController(kubeclient, *nameSpace, *rcName, *schedulerName)

	//testGetPodbyUUID(kubeclient, *nameSpace, *uuid)

	//testMovePod(kubeclient, *nameSpace, *podName, *nodeName)
}
