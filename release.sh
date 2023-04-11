#!/bin/bash

set -xeo pipefail

DIR="$(dirname "${BASH_SOURCE[0]}" &> /dev/null && pwd)"
pushd "$DIR"

command -v gcloud &> /dev/null || (echo "gcloud not found" && exit 1)
command -v ko &> /dev/null || (echo "ko not found" && exit 1)
command -v git &> /dev/null || (echo "git not found" && exit 1)
command -v go &> /dev/null || (echo "go not found" && exit 1)

git pull --rebase
go mod tidy
git diff --exit-code go.mod go.sum

export PROJECT_ID=${GOOGLE_CLOUD_PROJECT:-$DEVSHELL_PROJECT_ID}
PROJECT_ID=${PROJECT_ID:?"Need to set PROJECT_ID"}

function version () {
        local shortsha
        local shortdate
        shortsha=$(git rev-parse --short HEAD) # will output 91d9a52
        shortdate=$(date "+%F")                # will output 2021-01-01
        echo "$shortdate-$shortsha"
}

export REGION="europe-north1"
export SERVICE_NAME="lingonweb"
export RUN_NAME="lingonweb-77cd54dc0c33a59410461ed0aa76f91c"

# LDFLAGS="-X main.Version=$(version) -X main.Date=$(date "+%F") -X main.Commit=$(git rev-parse --short HEAD)"
Version="$(version)"
export Version
Date="$(date "+%F")"
export Date
Commit="$(git rev-parse --short HEAD)"
export Commit

export KO_DOCKER_REPO="$REGION-docker.pkg.dev/$PROJECT_ID/$SERVICE_NAME/$SERVICE_NAME"

gcloud run deploy "$RUN_NAME" --region="$REGION" --image="$(ko build .)"
