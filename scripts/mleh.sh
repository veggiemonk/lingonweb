#!/usr/bin/env bash
# Copyright (c) 2023 Volvo Car Corporation
# SPDX-License-Identifier: Apache-2.0


## HELM in reverse is MLEH

echo
echo '███    ███ ██      ███████ ██   ██ '
echo '████  ████ ██      ██      ██   ██ '
echo '██ ████ ██ ██      █████   ███████ '
echo '██  ██  ██ ██      ██      ██   ██ '
echo '██      ██ ███████ ███████ ██   ██ '
echo '                                   '


set -e
set -x
set -u
set -o pipefail

command -v helm > /dev/null
command -v kustomize > /dev/null
command -v go > /dev/null
command -v git > /dev/null
command -v wget > /dev/null

ROOT_DIR=$(git rev-parse --show-toplevel)
VALUES_DIR="$ROOT_DIR"/scripts
TEMPD="$VALUES_DIR"/out
KYGO="$TEMPD"/kygo

DEBUG=0
pushd "$ROOT_DIR"



# build a version of kygo with all possible CRDs
function tool() {
  go mod tidy
  go build -o kygo ./cmd/kygo && mv kygo "$TEMPD"
  [ $DEBUG ] && rm -rf "$TEMPD"/lingonweb
}

function install_repo() {
  helm repo add external-secrets https://charts.external-secrets.io
  helm repo add external-dns https://kubernetes-sigs.github.io/external-dns/
  helm repo add eks https://aws.github.io/eks-charts
  helm repo add aws-ebs-csi-driver https://kubernetes-sigs.github.io/aws-ebs-csi-driver
  helm repo add aws-efs-csi-driver https://kubernetes-sigs.github.io/aws-efs-csi-driver
  helm repo add kube-state-metrics https://kubernetes.github.io/kube-state-metrics
  helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
  helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server
  helm repo add autoscaler https://kubernetes.github.io/autoscaler
  helm repo add nats https://nats-io.github.io/k8s/helm/charts/
  helm repo add benthos https://benthosdev.github.io/benthos-helm-chart/
  helm repo add vm https://victoriametrics.github.io/helm-charts/
  helm repo add grafana https://grafana.github.io/helm-charts
  helm repo add vector https://helm.vector.dev


  helm repo update
}

function manifests() {
  rm -rf "$TEMPD"/manifests
  mkdir -p $TEMPD/manifests && pushd $TEMPD/manifests > /dev/null

  #
  # EXTERNAL SECRET
  #
  helm template external-secrets external-secrets/external-secrets \
    --namespace=external-secrets --create-namespace --set installCRDs=true | \
    $KYGO -out "externalsecrets" -app external-secrets -pkg externalsecrets

  #
  # EXTERNAL DNS
  #
  # docs: https://github.com/kubernetes-sigs/external-dns/blob/master/charts/external-dns/README.md
  helm template external-dns external-dns/external-dns | \
    $KYGO -out "externaldns" -app external-dns -pkg externaldns

  #
  # AWS LB
  #
  # using IAM Roles for service account
#  helm template aws-load-balancer-controller eks/aws-load-balancer-controller \
#    --set clusterName=my-cluster \
#    -n kube-system \
#    --set serviceAccount.create=false \
#    --set serviceAccount.name=aws-load-balancer-controller | \
#    $KYGO -out "awslb" -app awslb -pkg awslb

  #
  # CSI - AWS EFS & EBS
  #
  helm template aws-efs-csi-driver  aws-efs-csi-driver/aws-efs-csi-driver \
    --namespace=kube-system \
    --set image.repository=REPLACE_ME_602401143452.dkr.ecr.region-code.amazonaws.com/eks/aws-efs-csi-driver \
    --set controller.serviceAccount.create=false \
    --set controller.serviceAccount.name=efs-csi-controller-sa | \
    $KYGO -out "csi/efs" -app efs -pkg efs

  kustomize build "https://github.com/kubernetes-sigs/aws-ebs-csi-driver//deploy/kubernetes/overlays/stable/ecr-public/?ref=v1.19.0" | \
    $KYGO -out "csi/ebs" -app ebs -pkg ebs

  #
  # MONITORING
  #
  helm template metrics-server metrics-server/metrics-server --namespace=monitoring --values="$VALUES_DIR"/metrics-server.values.yaml | \
    $KYGO -out "monitoring/metrics-server" -app metrics-server -pkg metricsserver

  helm template promcrd prometheus-community/prometheus-operator-crds | \
    $KYGO -out "monitoring/promcrd" -app prometheus -pkg promcrd -group=false -clean-name=false

  helm template kube-promtheus-stack prometheus-community/kube-prometheus-stack --namespace=monitoring | \
    $KYGO -out "monitoring/promstack" -app kubeprometheusstack -pkg promstack

  helm template vm vm/victoria-metrics-single --namespace=monitoring --values "$VALUES_DIR"/victoriametrics-single.values.yaml | \
    $KYGO -out "monitoring/victoriametrics" -app victoria-metrics -pkg victoriametrics

  wget https://raw.githubusercontent.com/VictoriaMetrics/helm-charts/master/charts/victoria-metrics-k8s-stack/crds/crd.yaml -O - | \
    $KYGO -out "monitoring/vmcrd" -app victoriametrics -pkg vmcrd -group=false -clean-name=false

  helm template vmk8s vm/victoria-metrics-k8s-stack --namespace=monitoring --values "$VALUES_DIR"/vmk8s.values.yaml | \
    $KYGO -out "monitoring/vmk8s" -app vmk8s -pkg vmk8s

  helm template grafana grafana/grafana --namespace=monitoring | \
    $KYGO -out "monitoring/grafana" -app grafana -pkg grafana

  helm template vector vector/vector --namespace=monitoring | \
    $KYGO -out "monitoring/vector" -app vector -pkg vector

  #
  # NATS
  #
  helm template nats nats/nats --namespace=nats --values "$VALUES_DIR"/nats.values.yaml | \
    $KYGO -out "nats" -app nats -pkg nats

  helm template surveyor nats/surveyor --namespace=surveyor --values "$VALUES_DIR"/surveyor.values.yaml | \
    $KYGO -out "nats/surveyor" -app surveyor -pkg surveyor

  helm template benthos benthos/benthos --namespace=benthos --values "$VALUES_DIR"/benthos.values.yaml | \
    $KYGO -out "nats/benthos" -app benthos -pkg benthos

  wget https://github.com/nats-io/nack/releases/latest/download/crds.yml -O - | \
    $KYGO -out "nats/jetstream" -app jetstream -pkg jetstream -group=false -clean-name=false

  #
  # Karpenter
  #
  helm template karpenter oci://public.ecr.aws/karpenter/karpenter --namespace=karpenter \
    --create-namespace \
    --version "v0.27.5" \
    --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"=REPLACE_ME_KARPENTER_IAM_ROLE_ARN \
    --set settings.aws.clusterName=REPLACE_ME_CLUSTER_NAME \
    --set settings.aws.defaultInstanceProfile=KarpenterNodeInstanceProfile-REPLACE_ME_CLUSTER_NAME \
    --set settings.aws.interruptionQueueName=REPLACE_ME_CLUSTER_NAME \
    --set controller.resources.requests.cpu=1 \
    --set controller.resources.requests.memory=1Gi \
    --set controller.resources.limits.cpu=1 \
    --set controller.resources.limits.memory=1Gi | \
     $KYGO -out "karpenter" -app karpenter -pkg karpenter

  # CRDs
  {
    wget https://raw.githubusercontent.com/aws/karpenter/main/pkg/apis/crds/karpenter.sh_provisioners.yaml -O -
    wget https://raw.githubusercontent.com/aws/karpenter/main/pkg/apis/crds/karpenter.sh_machines.yaml -O -
    wget https://raw.githubusercontent.com/aws/karpenter/main/pkg/apis/crds/karpenter.k8s.aws_awsnodetemplates.yaml -O -
  } >> karpenter/mani.yaml

  $KYGO -in karpenter/mani.yaml -out "karpenter/crd" -pkg karpentercrd -app karpenter -group=false -clean-name=false

  helm template autoscaler autoscaler/cluster-autoscaler  --set 'autoDiscovery.clusterName'="REPLACE_ME"  | \
    $KYGO -out "autoscaler" -pkg autoscaler -app autoscaler

  popd

}



function step() {
  set +x
  local name="$1"
  echo
  echo '   #' "$name"
  echo '   ======================'
  echo
  set -x
}

function main {

  mkdir -p "$TEMPD"

  step "build kygo"
  [ ! -f "$KYGO" ] && tool
  "$KYGO" -version

  step "install/update repo"
  install_repo || true

  step "generate manifests"
  manifests


}

main
