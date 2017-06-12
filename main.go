/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
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
	//ns := "default"
	//podName := "cpu-group-mc6md"

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
	//ns := "default"
	//rcName := "cpu-group"

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

	testUpdatePod(kubeclient, *nameSpace, *podName, *schedulerName)

	testKillUpdatePod(kubeclient, *nameSpace, *podName, *schedulerName)

	testUpdateController(kubeclient, *nameSpace, *rcName, *schedulerName)
}
