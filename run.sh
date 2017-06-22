#!/bin/bash
set -x

./updatekube -h
#./kubeclient --kubeconfig ./admin.kubeconfig --masterUrl https://enmaster.octurbo.org:8443
#./kubeclient --kubeconfig ./configs/en119.kubeconfig.yaml

options="--kubeConfig ./configs/aws.kubeconfig.yaml "
options=" $options --nameSpace default "
#options=" $options --scheduler-name default-scheduler "
options=" $options --scheduler-name xyzscheduler "
#options=" $options --scheduler-name none-scheduler "
options=" $options --alsologtostderr --v 3"
options=" $options --uuid 6a457c63-551b-11e7-9ecb-0615046e67da "

##testMove
slave1="ip-172-23-1-92.us-west-2.compute.internal"
slave2="ip-172-23-1-12.us-west-2.compute.internal"
options=" $options --podName noscheduler-mem-30-1 "
options=" $options --nodeName $slave1 "



./updatekube  $options
