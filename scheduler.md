 # Experiments about Kubernetes scheduling policy #
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


## Results ##
| index | description | result|
|-|-|-|
| 1 | without schedulerName | scheduled by the "default-scheduler" |
| 2 | with schedulerName="default-scheduler" | scheduled by the "default-scheduler" |
| 3 | with schedulerName="xyzscheduler" | scheduled by the "xyzscheduler" |
| 4 | with schedulerName="slow-xyzscheduler" | scheduled, but not by "slow-xyzscheduler", and no event indicating by "default-scheduler" |
| 5 | with schedulerName="none-exist" | scheduled, but no event indicating by "default-scheduler" |
| 6 | with schedulerName="none-exist", and a nodeSelector which cannot be matched | cannot be scheduled, keep pending |

Note1: "default-scheduler" is Kubernetes' default scheduler name;

Note2: the "xyzscheduler" is built from k8s.io/kubernetes/plugin/cmd/kube-scheduler/;

Note3: the "slow-xyzscheduler" is also built from k8s.io/kubernetes/plugin/cmd/kube-scheduler/, with one-line modification: 
sleep 30 seconds in the scheduler.scheduleOne() function before caling shed.schedule(pod);

Note4: we will check the procedure of the pod creatation by "kubectl get events"


## Conclusions ##
Based on the test results, we can get the following conclusion:
```console
 If the cumstomer scheduler is slow, it will lose the oppotunity to schedule the Pod; 
 Even if the Pod is assigned to that customer scheduler.
 ```

## Discussion ##
Based on the source code of k8s.io/kubernetes/plugin/cmd/kube-scheduler/, when the default scheduler gets an unscheduled Pod, the default scheduler will check whether the Pod's schedulerName equals to its own name. If the schedulerName matches, the default scheduler will do the scheduling for the Pod; otherwise, the default scheduler won't schedule the Pod.

But in this experiments, if a Pod has a schedulerName, and the scheduler is slow, then the Pod will be scheduled by someone else. I am afraid that there is some bug, or other constrains in Kubernetes scheduler framework. It is necessary to read the code of the ApiServer to know what it will do in this situation.


    
