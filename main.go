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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//global variables
var (
	masterUrl     *string
	kubeConfig    *string
	nameSpace     *string
	podName       *string
	rcName        *string
	schedulerName *string
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
	pods, err := client.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	printPods(pods)
}

func setFlags() {
	masterUrl = flag.String("masterUrl", "", "master url")
	kubeConfig = flag.String("kubeConfig", "", "absolute path to the kubeconfig file")
	nameSpace = flag.String("nameSpace", "default", "kubernetes object namespace")
	podName = flag.String("podName", "myschedule-cpu-80", "the podName to be handled")
	rcName = flag.String("rcName", "cpu-group", "the ReplicationController name")
	schedulerName = flag.String("scheduler-name", "default-scheduler", "the name of the scheduler")

	flag.Parse()

	fmt.Printf("kubeConfig=%s, masterUrl=%s\n", *kubeConfig, *masterUrl)
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

	testPod(clientset)
	return clientset
}

func testUpdatePod(kclient *kubernetes.Clientset, ns, podName, schedulerName string) error {
	//ns := "default"
	//podName := "myschedule-cpu-80"

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

func copyPodInfo(oldPod, newPod *v1.Pod) {
	//typeMeta
	newPod.Kind = oldPod.Kind
	newPod.APIVersion = oldPod.APIVersion
	//objectMeta
	newPod.Name = oldPod.Name
	newPod.Namespace = oldPod.Namespace
	newPod.Labels = oldPod.Labels
	newPod.Annotations = oldPod.Annotations
	newPod.OwnerReferences = oldPod.OwnerReferences
	newPod.Finalizers = oldPod.Finalizers
	newPod.ClusterName = oldPod.ClusterName

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
		fmt.Printf("failed to get ReplicationController:%v\n", err)
		return err
	}
	fmt.Printf("ReplicationController:%v\n", rc.Spec.Template.Spec.SchedulerName)

	//2. update
	newScheduler := schedulerName
	rc.Spec.Template.Spec.SchedulerName = newScheduler
	rc, err = rcClient.Update(rc)
	if err != nil {
		fmt.Printf("failed to update RC:%v\n", err)
		return err
	}

	//3. get it again
	rc, err = rcClient.Get(rcName, option)
	if err != nil {
		fmt.Printf("failed to get ReplicationController:%v\n", err)
		return err
	}
	fmt.Printf("ReplicationController:%v\n", rc.Spec.Template.Spec.SchedulerName)

	return nil
}

func main() {

	setFlags()
	fmt.Printf("kubeConfig=%s, masterUrl=%s\n", *kubeConfig, *masterUrl)

	kubeclient := getKubeClient()
	if kubeclient == nil {
		fmt.Println("failed to get kubeclient")
		return
	}

	testPod(kubeclient)

	//will fail:failed to update Pod:Pod "myschedule-cpu-80" is invalid: spec: 
	// Forbidden: pod updates may not change fields other than `containers[*].image` or `spec.activeDeadlineSeconds` or `spec.tolerations` 
	//(only additions to existing tolerations)
	testUpdatePod(kubeclient, *nameSpace, *podName, *schedulerName)

	//Kill & reCreate it
	testKillUpdatePod(kubeclient, *nameSpace, *podName, *schedulerName)

	//Update ReplicationController, kill & wait for RC to reCreate it.
	testUpdateController(kubeclient, *nameSpace, *rcName, *schedulerName)
}
