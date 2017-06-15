 # notes of the pattern Kubernetes scheduling patterns #
 Before reading the code of Kubernetes ApiServer and kubelete, some experiments are done to have a initial understanding
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

Note1: "default-scheduler" is Kubernetes' default scheduler name;
Note2: the "xyzscheduler" is built from k8s.io/Kubernetes/plugin/cmd/kube-scheduler/;
Note3: we will check the procedure of the pod creatation by "kubectl get events"

## Results ##
| index | description | result|
|-|-|-|
| 1 | without schedulerName | scheduled by the "default-scheduler" |
| 2 | with schedulerName="default-scheduler" | scheduled by the "default-scheduler" |
| 3 | with schedulerName="xyzscheduler" | scheduled by the "xyzscheduler" |
| 4 | with schedulerName="slow-xyzscheduler" | scheduled, but not by "slow-xyzscheduler", and no event indicating by "default-scheduler" |
| 5 | with schedulerName="none-exist" | scheduled, but no event indicating by "default-scheduler" |
| 6 | with schedulerName="none-exist", and a nodeSelector which cannot be matched | cannot be scheduled, keep pending |


## Conclusions ##


    
