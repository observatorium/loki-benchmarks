#!/bin/bash

LOKI_PROJECT_NAME=observatorium-logs-test
LOKI_S3_SECRET_NAME=loki-objectstorage-secret
LOKI_S3_SECRET_FILE=/tmp/$LOKI_S3_SECRET_NAME.yaml
LOKI_TEMPLATE_FILE=/tmp/observatorium-logs-template.yaml
AWS_CREDENTIALS_FILE_NAME=~/.aws/credentials

LOKI_INGESTER_REPLICAS=${LOKI_INGESTER_REPLICAS:-"3"}

get_aws_credentials() {
  # get AWS credentials from local user account file
  USER_AWS_ACCESS_KEY_ID=$(grep -m 1 aws_access_key_id < $AWS_CREDENTIALS_FILE_NAME | awk '{print $3}')
  USER_AWS_SECRET_ACCESS_KEY=$(grep -m 1 aws_secret_access_key < $AWS_CREDENTIALS_FILE_NAME | awk '{print $3}')

  # Get AWS credentials from environment variables
  AWS_S3_END_POINT="${AWS_S3_END_POINT:-https://s3.us-east-1.amazonaws.com}"
  AWS_S3_REGION="${AWS_S3_REGION:-us-east-1}"
  AWS_S3_LOKI_BUCKET_NAME="${AWS_S3_LOKI_BUCKET_NAME:-loki-bucket}"
  AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-$USER_AWS_ACCESS_KEY_ID}"
  AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-$USER_AWS_SECRET_ACCESS_KEY}"

  if [ -z "$AWS_ACCESS_KEY_ID" ]; then
    echo "==> ERROR: Provide AWS access_key using AWS_ACCESS_KEY_ID env.variable or using $AWS_CREDENTIALS_FILE_NAME file"
    exit
  fi

  if [ -z "$AWS_SECRET_ACCESS_KEY" ]; then
    echo "==> ERROR: Provide AWS secret_key using AWS_SECRET_ACCESS_KEY env.variable or using $AWS_CREDENTIALS_FILE_NAME file"
    exit
  fi

  echo "
Using AWS S3 configuration for loki bucket::
-==-=-=-=-=-=-
AWS S3 end-point: $AWS_S3_END_POINT
AWS S3 region: $AWS_S3_REGION
bucket name: $AWS_S3_LOKI_BUCKET_NAME

(Note: Use environment variables as needed to change defaults)
"
}

check_oc_works() {
  OC_WORKS=$(oc whoami > /dev/null 2>&1 ; echo $?)
  if ! [ "$OC_WORKS" == "0" ]; then
    echo "==> ERROR: oc command not working,  make sure you are logged into OCP cluster and that oc command works"
    exit
  else
    echo "OK, OCP cluster working"
  fi
}

delete_project_if_exists() {
  PROJECT=$(oc get project | grep "$1")
  if [ -n "$PROJECT" ]; then
    echo "Deleting $1 project/namespace"
    oc delete project "$1"
    while : ; do
      PROJECT=$(oc get project | grep "$1")
      if [ -z "$PROJECT" ]; then break; fi
      sleep 1
    done
  fi
}

create_project() {
  oc new-project "$1" > /dev/null 2>&1
  echo "Done"
}

get_observatorium_template() {
  rm -f $LOKI_TEMPLATE_FILE
  wget -nv -O $LOKI_TEMPLATE_FILE https://raw.githubusercontent.com/rhobs/configuration/main/resources/services/observatorium-logs-template.yaml
}

create_s3_secret() {
  rm -f $LOKI_S3_SECRET_FILE
  cat > $LOKI_S3_SECRET_FILE << EOF
apiVersion: v1
kind: Secret
data:
  endpoint: $(echo "$AWS_S3_END_POINT" | base64)
  aws_region: $(echo "$AWS_S3_REGION" | base64)
  bucketnames: $(echo "$AWS_S3_LOKI_BUCKET_NAME" | base64)
  bucket: $(echo "$AWS_S3_LOKI_BUCKET_NAME" | base64)
  aws_access_key_id: $(echo "$AWS_ACCESS_KEY_ID" | base64)
  aws_secret_access_key: $(echo "$AWS_SECRET_ACCESS_KEY" | base64)
metadata:
  name: $LOKI_S3_SECRET_NAME
  namespace: $LOKI_PROJECT_NAME
EOF
    echo "Done"
}

deploy_loki() {
  echo "Deploying Loki s3 secret from = $LOKI_S3_SECRET_NAME.yaml"
  oc apply -f $LOKI_S3_SECRET_FILE
  echo "Deploying Loki manifests with ingester replicas = $LOKI_INGESTER_REPLICAS into project $LOKI_PROJECT_NAME"
  oc process -f $LOKI_TEMPLATE_FILE  -p LOKI_INGESTER_REPLICAS="$LOKI_INGESTER_REPLICAS" -p NAMESPACE=$LOKI_PROJECT_NAME -p LOKI_S3_SECRET=$LOKI_S3_SECRET_NAME | oc apply -f -
  echo "Done"
}

main () {
  echo "==> Check that oc works"
  check_oc_works
  echo "==> AWS credentials for Loki Bucket"
  get_aws_credentials
  echo "==> Create object storage secret file"
  create_s3_secret
  echo "==> Get observatorium template"
  get_observatorium_template
  echo "==> Delete project $LOKI_PROJECT_NAME"
  delete_project_if_exists $LOKI_PROJECT_NAME
  echo "==> Create project $LOKI_PROJECT_NAME"
  create_project $LOKI_PROJECT_NAME
  echo "==> Deploy loki"
  deploy_loki
}

main
