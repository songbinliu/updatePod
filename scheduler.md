 # notes of the pattern Kubernetes scheduling patterns #
 Before reading the code of Kubernetes ApiServer and kubelete, some experiments are done to have a initial understanding
 of Kubernetes scheduling policies.
 
 ## Different situations ##
 
The behaviour of Pod scheduluation in different situations including:

* 1. create a Pod without setting the schedulerName;
* 2. create a Pod with schedulerName to "default-scheduler";
* 3. create a Pod with a customer schedulerName (xyzscheduler in the tests);
* 4. create a Pod with a non-exist schedulerName;
* 5. create a Pod with a customer schedulerName, but this scheduler is very slow;
    (sleep for about 30 seconds before doing the schedule)
    
