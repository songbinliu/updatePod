# updatePod #

When to reschedule a Pod with our own scheduler, we needs to change the scheduler-name of the Pod.
This piece of code is to test how to update a Pod's scheduler-name.

Usually, Pod will be created by in two ways: by ReplicationControler/ReplicateSet, or created directly.
  * If the Pod is created without a ReplicationController/ReplicateSet, then kill & re-create with the new scheduler.name;
  * If pod is created by ReplicationController, then set the scheduler.name in the ReplicationController.Template;


## Update Pod.scheduler via Client.Pods().Update() API ##
As shown in the function *testUpdatePod()*, it is impossible to update Pod's scheduler name with this API.
```go
# impossible to update Pod's scheduler-name via this API.
kclient.CoreV1().Pods(ns).Update(nPod)
```

```console
failed to update Pod:Pod "myschedule-cpu-80" is invalid: 
spec: Forbidden: pod updates may not change fields other than `containers[*].image` 
or `spec.activeDeadlineSeconds` or `spec.tolerations` (only additions to existing tolerations)
```

## Update Pod.scheduler by kill & create ##
There are four steps:
 * Get the original Pod via client.API;
 * Copy necessary information from the orinial Pod;
 * Modify the new Pod's configuration-- the scheduler-name;
 * Delete the original Pod via client.API;
 * Create a new Pod based on the new configuration via client.API.

As shown in the function *testKillUpdatePod()*, this way works well.  
It should be noted that **If you set the Pod.Spec.NodeName** when call the Create() API:
  * If the NodeName is correct, then the scheduler will bind this Pod to the Node directly;
  * If the NodeName is wrong, the scheduler will schedule it a to a fit Pod.
  
 
 ## Update Pod.scheduler by Update ReplicationController.Template ##
 Some Pod is controlled by ReplicationController (RC), so it is necessary to make assure Pods created by the RC will be scheduled by our scheduler.  Fortunately, we can update the setting of the RC directly via API.
 ```go
 client.CoreV1().ReplicationControllers(nameSpace).Update(newRC)
 ```
 As shown in the function of *testUpdateController()*, we update the scheduler name of by update the RC. Then after one of its
 Pod is deleted, a new Pod will be scheduled by the desinated scheduler.
 
 
## Run it ##

### start up another scheduler for Kubernetes ###
we can build and run the default scheduler of Kubernetes
```console
go get k8s.io/kubernetes
cd $GOPATH/src/k8s.io/kubernetes/plugin/cmd/kube-scheduler
go build

#run it
./kube-scheduler --kubeconfig ./aws.kubeconfig.yaml --scheduler-name xyzscheduler --leader-elect=false --logtostderr --v 3 
```
### test whether the scheduler-name is update ###
build and run this project.
```bash
./updatekube --kubeConfig ./configs/aws.kubeconfig.yaml --nameSpace default --scheduler-name xyzscheduler --alsologtostderr
```
you will see something like this:
```console
I0615 10:27:09.745468    9407 scheduler.go:254] Attempting to schedule pod: default/myschedule-cpu-80
I0615 10:27:09.746095    9407 factory.go:706] Attempting to bind myschedule-cpu-80 to ip-172-23-1-12.us-west-2.compute.internal
I0615 10:27:09.836120    9407 event.go:218] Event(v1.ObjectReference{Kind:"Pod", Namespace:"default", Name:"myschedule-cpu-80", UID:"b7393f1f-51d6-11e7-9ecb-0615046e67da", APIVersion:"v1", ResourceVersion:"10089897", FieldPath:""}): type: 'Normal' reason: 'Scheduled' Successfully assigned myschedule-cpu-80 to ip-172-23-1-12.us-west-2.compute.internal
```
