#!/bin/bash

set -eou pipefail

BUCKET_NAME=$1

REGION=$(aws configure get region)

delete_bucket() {
    aws s3 rb s3://"$BUCKET_NAME" --region "$REGION" --force  || true
}

main() {
  echo "deleting bucket $BUCKET_NAME (if exists)"
  delete_bucket
}

main
