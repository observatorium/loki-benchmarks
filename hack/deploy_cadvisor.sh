#!/bin/bash

CADVISOR_FOLDER=/tmp/cadvisor
CADVISOR_PROJECT_NAME=cadvisor
CADVISOR_BRANCH=v0.37.5
check_oc_works() {
  OC_WORKS=$(oc whoami > /dev/null 2>&1 ; echo $?)
  if ! [ "$OC_WORKS" == "0" ]; then
    echo "==> ERROR: oc command not working,  make sure you are logged into OCP cluster and that oc command works"
    exit
  else
    echo "OK, OCP cluster working"
  fi
}

check_kustomize_works() {
  OC_WORKS=$(kustomize version > /dev/null 2>&1 ; echo $?)
  if ! [ "$OC_WORKS" == "0" ]; then
    echo "==> ERROR: kustomize command not working, install kustomize and make sure it works"
    exit
  else
    echo "OK, kustomize installed and working"
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

get_cadvisor() {
  rm -r -f $CADVISOR_FOLDER
  git clone https://github.com/google/cadvisor.git --branch $CADVISOR_BRANCH $CADVISOR_FOLDER
}

deploy_cadvisor() {
  oc adm policy add-scc-to-user privileged -z cadvisor
  oc adm policy add-cluster-role-to-user cluster-reader -z cadvisor
  kustomize build /tmp/cadvisor/deploy/kubernetes/base | kubectl apply -f -
}

main () {
  echo "==> Check that oc works"
  check_oc_works
  echo "==> Check that kustomize works"
  check_kustomize_works
  echo "==> Get cadvisor"
  get_cadvisor
  echo "==> Delete project $CADVISOR_PROJECT_NAME"
  delete_project_if_exists $CADVISOR_PROJECT_NAME
  echo "==> Create project $CADVISOR_PROJECT_NAME"
  create_project $CADVISOR_PROJECT_NAME
  echo "==> Deploy cadvisor"
  deploy_cadvisor
}

main
