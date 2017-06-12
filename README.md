# updatePod #

When to reschedule a Pod with our own scheduler, we needs to change the scheduler-name of the Pod.
This piece of code is to test how to update a Pod's scheduler-name:
(1) If pod is created by ReplicationController, then set the scheduler.name in the ReplicationController.Template;
(2) If the Pod is created without a ReplicationController/ReplicateSet, then kill & re-create with the new scheduler.name;

## Update Pod.scheduler via Client.Pods().Update() API ##
As shown in the function *testUpdatePod()*, it is impossible to update Pod's scheduler name with this API.
```go
# impossible to update Pod's scheduler-name via this API.
kclient.CoreV1().Pods(ns).Update(nPod)
```

```console
failed to update Pod:Pod "myschedule-cpu-80" is invalid: spec: Forbidden: pod updates may not change fields other than `containers[*].image` or `spec.activeDeadlineSeconds` or `spec.tolerations` (only additions to existing tolerations)
```

## Update Pod.scheduler by kill & create ##
