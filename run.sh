#!/bin/bash
set -x

./updatekube -h
#./kubeclient --kubeconfig ./admin.kubeconfig --masterUrl https://enmaster.octurbo.org:8443
#./kubeclient --kubeconfig ./configs/en119.kubeconfig.yaml

options="--kubeConfig ./configs/aws.kubeconfig.yaml "
options=" $options --nameSpace default "
options=" $options --scheduler-name xyzscheduler "
#options=" $options --scheduler-name my-scheduler "
options=" $options --alsologtostderr --v 2"



./updatekube  $options
