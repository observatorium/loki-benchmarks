#!/bin/bash

set -eou pipefail

NAMESPACE=$1
BUCKET_NAME=$2

REGION=""
ENDPOINT=""
ACCESS_KEY_ID=""
SECRET_ACCESS_KEY=""

set_credentials_from_aws() {
  AWS_CONFIG_FILE_NAME=$HOME/.aws/config
  AWS_CREDENTIALS_FILE_NAME=$HOME/.aws/credentials

  REGION="$(grep -m 1 region < $AWS_CONFIG_FILE_NAME | awk '{print $3}')"
  ENDPOINT="https://s3.${REGION}.amazonaws.com"
  ACCESS_KEY_ID="$(grep -m 1 aws_access_key_id < $AWS_CREDENTIALS_FILE_NAME | awk '{print $3}')"
  SECRET_ACCESS_KEY="$(grep -m 1 aws_secret_access_key < $AWS_CREDENTIALS_FILE_NAME | awk '{print $3}')"
}

create_secret() {
  kubectl -n $NAMESPACE delete secret test ||:
  kubectl -n $NAMESPACE create secret generic test \
    --from-literal=endpoint=$(echo -n "$ENDPOINT") \
    --from-literal=aws_region=$(echo -n "$REGION") \
    --from-literal=bucket=$(echo -n "$BUCKET_NAME") \
    --from-literal=bucketnames=$(echo -n "$BUCKET_NAME") \
    --from-literal=aws_access_key_id=$(echo -n "$ACCESS_KEY_ID") \
    --from-literal=aws_secret_access_key=$(echo -n "$SECRET_ACCESS_KEY")
}

main() {
  set_credentials_from_aws
  create_secret
}

main
