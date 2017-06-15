 # notes of the pattern Kubernetes scheduling patterns #
 Before reading the code of Kubernetes ApiServer and kubelete, some experiments are done to have a initial understanding
 of Kubernetes scheduling policies.
 
 ## different situations ##
 
The behaviour of Pod scheduluation in different situations including:

* create a Pod without setting the schedulerName;
* create a Pod with schedulerName to "default-scheduler";
* create a Pod with a customer schedulerName (xyzscheduler in the tests);
* create a Pod with a non-exist schedulerName;
* create a Pod with a customer schedulerName, but this scheduler is very slow;
    (sleep for about 30 seconds before doing the schedule)
    
