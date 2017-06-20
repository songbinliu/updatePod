 # Experiment about Kubernetes scheduling policy #
 Before reading the code of Kubernetes ApiServer and kubelete, some tests are done to have a initial understanding
 of Kubernetes scheduling policies.
 
 ## Different situations ##
 
This experiments will test the behaviour of Pod scheduluation in different situations:

* 1. create a Pod without setting the schedulerName;
* 2. create a Pod with schedulerName to "default-scheduler";
* 3. create a Pod with a customer schedulerName (xyzscheduler in the tests);
* 4. create a Pod with a customer schedulerName, but this scheduler is very slow;
    (sleep for about 30 seconds before doing the schedule)
* 5. create a Pod with a non-exist schedulerName;
* 6. create a Pod with a non-exist schedulerName, and the Pod has a nodeSelector which cannot be matched in the cluster;
* 7. create a ReplicationController with a customer schedulerName (xyzscheduler in the tests);;
* 8. create a ReplicationController with non-exist schedulerName;


## Results ##
| index | description | result|
|-|-|-|
| 1 | without schedulerName | scheduled by the "default-scheduler" |
| 2 | with schedulerName="default-scheduler" | scheduled by the "default-scheduler" |
| 3 | with schedulerName="xyzscheduler" | scheduled by the "xyzscheduler" |
| 4 | with schedulerName="slow-xyzscheduler" | scheduled by "slow-xyzscheduler" |
| 5 | with schedulerName="none-exist" | pending |
| 6 | with schedulerName="none-exist", and a nodeSelector which cannot be matched | pending |
| 7 | ReplicationController 3 replicas, schedulerName="xyzscheduler" | all 3 pods are scheduled by "xyzscheduler"|
| 8 | ReplicationController 3 replicas, schedulerName="none-exist" | all 3 pods are pending|


Note1: "default-scheduler" is Kubernetes' default scheduler name;

Note2: the "xyzscheduler" is built from k8s.io/kubernetes/plugin/cmd/kube-scheduler/;

Note3: the "slow-xyzscheduler" is also built from k8s.io/kubernetes/plugin/cmd/kube-scheduler/, with one-line modification: 
sleep 30 seconds in the scheduler.scheduleOne() function before caling shed.schedule(pod);

Note4: we get to know which scheduler did the scheduling by "kubectl get events", and the logs of "xyzscheduler" and "slow-xyzscheduler".


## Conclusions ##
Based on the test results, we can get the following conclusion:
```console
 1. Once scheduler name is set, Pod will wait for the scheduler to assign a node.
 2. If customer scheduler crashed, then all the pods will be pending.
 ```

## Others ##
Based on the source code of k8s.io/kubernetes/plugin/cmd/kube-scheduler/, when the default scheduler gets an unscheduled Pod, the default scheduler will check whether the Pod's schedulerName equals to its own name. If the schedulerName matches, the default scheduler will do the scheduling for the Pod; otherwise, the default scheduler won't schedule the Pod.


    
