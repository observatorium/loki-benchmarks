#!/bin/bash

LOKI_PROJECT_NAME=observatorium-logs-test
LOKI_S3_SECRET_NAME=loki-objectstorage-secret

delete_project_if_exists() {
  PROJECT=$(oc get project | grep $1)
  if [ -n "$PROJECT" ]; then
    echo "Deleting $1 project/namespace"
    oc delete project $1
    while : ; do
      PROJECT=$(oc get project | grep $1)
      if [ -z "$PROJECT" ]; then break; fi
      sleep 1
    done
  fi
}

create_project() {
  oc new-project $1
}

get_observatorium_template() {
  rm -f observatorium-logs-template.yaml
  wget https://raw.githubusercontent.com/rhobs/configuration/main/resources/services/observatorium-logs-template.yaml 
}

create_objectstorage_s3_secret() {
rm -f $LOKI_S3_SECRET_NAME.yaml
cat > $LOKI_S3_SECRET_NAME.yaml << EOF
apiVersion: v1
kind: Secret
stringData :
  endpoint: https://s3.us-east-1.amazonaws.com
  bucketnames: <s3-bucket-name>
  bucket: <s3-bucket-name>
  aws_region: us-east-1
  aws_access_key_id: <aws-access-key>
  aws_secret_access_key: <aws-secret-key>
metadata:
  name: $LOKI_S3_SECRET_NAME
  namespace: $LOKI_PROJECT_NAME
EOF
}

deploy_loki() {
  oc apply -f $LOKI_S3_SECRET_NAME.yaml
  oc process -f observatorium-logs-template.yaml  -p LOKI_INGESTER_REPLICAS="3" -p NAMESPACE=$LOKI_PROJECT_NAME -p LOKI_S3_SECRET=$LOKI_S3_SECRET_NAME | oc apply -f -
}

main () {
  echo "==> Delete project $LOKI_PROJECT_NAME"
  delete_project_if_exists $LOKI_PROJECT_NAME
  echo "==> Create project $LOKI_PROJECT_NAME"
  create_project $LOKI_PROJECT_NAME
  echo "==> Get observatorium template"
  get_observatorium_template
  echo "==> Create object storage secret file"
  create_objectstorage_s3_secret
  echo "==> Deploy loki"
  deploy_loki
}

main
oc get pods
