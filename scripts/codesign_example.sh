#! /usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1

set -euo pipefail

# first makes some assertions about the environment and set some shared
# variables before starting the script.
if ! command -v jq > /dev/null 2>&1; then
  echo "This script requires jq to work properly."
  exit 1
fi

if ! command -v sha256sum > /dev/null 2>&1; then
  if !command -v gsha256sum > /dev/null 2>&1; then
    echo "This script requires sha256sum (linux) or gsha256sum (osx) to work properly."
    exit 1
  else
    SHASUM_PROG=gsha256sum
  fi
else
  SHASUM_PROG=sha256sum
fi

PRODUCT_NAME="${PRODUCT_NAME:-""}"
if [ -z "$PRODUCT_NAME" ]; then
  echo "Missing required product name: ${PRODUCT_NAME}"
  exit 1
fi

TARGET_ZIP="${TARGET_ZIP:-""}"
if [ -z "$TARGET_ZIP" ]; then
  echo "Missing required target path"
  exit 1
fi

# Artifactory configuration
ARTIFACTORY_ENDPOINT="${ARTIFACTORY_ENDPOINT:-"https://artifactory.hashicorp.engineering/artifactory"}"
ARTIFACTORY_INPUT_REPO="${ARTIFACTORY_INPUT_REPO:-"hc-signing-input"}"
ARTIFACTORY_OUTPUT_REPO="${ARTIFACTORY_OUTPUT_REPO:-"hc-signing-output"}"

ARTIFACTORY_TOKEN="${ARTIFACTORY_TOKEN:-""}"
ARTIFACTORY_USER="${ARTIFACTORY_USER:-""}"

if [[ -z "$ARTIFACTORY_TOKEN" || -z "$ARTIFACTORY_USER" ]]; then
  echo "Missing required Artifactory credentials"
  exit 1
fi

# Create the sign/notarize ID "SN_ID"
if command -v uuidgen > /dev/null 2>&1; then
  uuid="$(uuidgen)"
elif [ -f /proc/sys/kernel/random/uuid ]; then
  uuid="$(cat /proc/sys/kernel/random/uuid)"
else
  echo "This script needs some way to generate a uuid."
  exit 1
fi
SN_ID="$uuid"

# CircleCI configuration
CIRCLE_ENDPOINT="${CIRCLE_ENDPOINT:-"https://circleci.com/api/v2"}"
CIRCLE_PROJECT="${CIRCLE_PROJECT:-"project/github/hashicorp/circle-codesign"}"

CIRCLE_TOKEN="${CIRCLE_TOKEN:-""}"
if [ -z "$CIRCLE_TOKEN" ]; then
  echo "Missing required CircleCI credentials"
  exit 1
fi

# Next, upload an unsigned zip file to the Artifactory at
# https://artifactory.hashicorp.engineering/artifactory/hc-signing-input/{PRODUCT}/{ID}.zip
echo "Uploading unsigned zip to ${ARTIFACTORY_ENDPOINT}/${ARTIFACTORY_INPUT_REPO}/${PRODUCT_NAME}/${SN_ID}.zip"

curl --show-error --silent --fail \
  --user "${ARTIFACTORY_USER}:${ARTIFACTORY_TOKEN}" \
  --request PUT \
  "${ARTIFACTORY_ENDPOINT}/${ARTIFACTORY_INPUT_REPO}/${PRODUCT_NAME}/${SN_ID}.zip" \
  --upload-file "$TARGET_ZIP" > /dev/null

# Next, start the CircleCI Pipeline, then wait for a Workflow
# to start.
echo "Executing CircleCI job"

res="$(curl --show-error --silent --fail --user "${CIRCLE_TOKEN}:" \
  --request POST \
  --header 'Content-Type: application/json' \
  --header 'Accept: application/json' \
  --data "{ \"branch\": \"main\" ,\"parameters\": { \"PRODUCT\": \"${PRODUCT_NAME}\", \"PKG_NAME\": \"${SN_ID}.zip\" } }" \
  "${CIRCLE_ENDPOINT}/${CIRCLE_PROJECT}/pipeline")"
pipeline_id="$(echo "$res" | jq -r '.id')"
echo "CircleCI Pipeline $pipeline_id started"

echo -n "Retrieving CircleCI Workflow ID"
# 24 * 5 seconds = 2 minutes
counter=12
workflow_id=""
# wait until a Workflow ID is found
until [ "$workflow_id" != "" ]; do
  echo -n "."
  workflow_id=$(curl --silent --fail --user "${CIRCLE_TOKEN}:" \
    --request GET \
    --header 'Accept: application/json' \
    "${CIRCLE_ENDPOINT}/pipeline/${pipeline_id}/workflow" \
    | jq -r '.items[].id'
  )
 if [ "$counter" -eq "0" ]; then
    echo "Tried too many times, but Pipeline ${pipeline_id} still has no Workflows"
    exit 1
  fi
  counter=$((counter - 1))
  sleep 5
done
echo ""

echo "CircleCI Workflow $workflow_id started"

# Next, wait for the Workflow to reach a terminal state, then fails if it isn't
# "success"
echo -n "Waiting for CircleCI Workflow ID: ${workflow_id}"
# 360 * 5 seconds = 30 minutes
counter=360
finished="not_run"
# wait for one of the terminal states: ["success", "failed", "error", "canceled"]
until [[ "$finished" == "success" || "$finished" == "failed" || "$finished" == "error" || "$finished" == "canceled" ]]; do
  echo -n "."
  finished=$(curl --silent --fail --user "${CIRCLE_TOKEN}:" \
    --header 'Accept: application/json' \
    "${CIRCLE_ENDPOINT}/workflow/${workflow_id}" \
    | jq -r '.status'
  )
  if [ "$counter" -eq "0" ]; then
    echo "Tried too many times, but workflow is still in state ${finished}"
    exit 1
  fi
  counter=$((counter - 1))
  sleep 5
done
echo ""

if [ "$finished" != "success" ]; then
  echo "Workflow ID ${workflow_id} ${finished}"
  exit 1
fi

# Next, download the signed zip from Artifactory at
# https://artifactory.hashicorp.engineering/artifactory/hc-signing-output/{PRODUCT}/{ID}.zip
echo "Retrieving signed zip from ${ARTIFACTORY_ENDPOINT}/${ARTIFACTORY_OUTPUT_REPO}/${PRODUCT_NAME}/${SN_ID}.zip"

curl --show-error --silent --fail --user "${ARTIFACTORY_USER}:${ARTIFACTORY_TOKEN}" \
  --request GET \
  "${ARTIFACTORY_ENDPOINT}/${ARTIFACTORY_OUTPUT_REPO}/${PRODUCT_NAME}/${SN_ID}.zip" \
  --output "signed_${SN_ID}.zip"

signed_checksum=$(
  curl --silent --show-error --fail --user "${ARTIFACTORY_USER}:${ARTIFACTORY_TOKEN}" \
    --head \
    "${ARTIFACTORY_ENDPOINT}/${ARTIFACTORY_OUTPUT_REPO}/${PRODUCT_NAME}/${SN_ID}.zip" \
    | grep -i "x-checksum-sha256" | awk 'gsub("[\r\n]", "", $2) {print $2;}'
)

echo "${signed_checksum}  signed_${SN_ID}.zip" | $SHASUM_PROG -c
if [ $? -ne 0 ]; then
  exit 1
fi

mv "signed_${SN_ID}.zip" "$TARGET_ZIP"
