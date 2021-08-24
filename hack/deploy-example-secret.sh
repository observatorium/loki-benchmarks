#!/bin/bash

set -eou pipefail

NAMESPACE=$1
BUCKET_NAME=$2

kubectl -n $NAMESPACE delete secret test ||:
kubectl -n $NAMESPACE create secret generic test \
    --from-literal=aws_region=$(echo -n $(aws configure get region)) \
    --from-literal=bucket=$(echo -n "$BUCKET_NAME") \
    --from-literal=aws_access_key_id=$(echo -n $(aws configure get aws_access_key_id)) \
    --from-literal=aws_secret_access_key=$(echo -n $(aws configure get aws_secret_access_key))
