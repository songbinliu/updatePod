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
First copy the original Pod's necessary information, and modify the scheduler-name; 
second, delete the original Pod; 
third, create the new Pod.

It works well in this way.
