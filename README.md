# Logging Pipeline Plumber for Rancher


[![Go Report Card](https://goreportcard.com/badge/github.com/MrSupiri/rancher-logging-pipeline-plumber)](https://goreportcard.com/report/github.com/MrSupiri/rancher-logging-pipeline-plumber)

This is a tool that can be used to debug logging pipelines built using [Rancher Logging](https://rancher.com/docs/rancher/v2.5/en/logging/). Once installed users can choose a [Flow or ClusterFlow](https://banzaicloud.com/docs/one-eye/logging-operator/configuration/flow/) along with a pod to simulate and users also can set the log messages that are emitted by the pod.

Then the operator will slice the target flow into `N` permutations where `N` equals the number of Select and Filter statements present in the selected Flow. Then it will schedule an [output](https://banzaicloud.com/docs/one-eye/logging-operator/configuration/output/) for each Flow and if at least one log statement gets passed to the output operator take all the select or filters in that specific flow and mark the as passing and that flow will be deleted to save resources.

When the test hits the timeout (default: 5mins), the operator will clean up all the provisioned resources and users can see which Match or Filter statements are preventing logs from getting to their respective destinations.


## Get started

### Install Logging Plumber 

```sh
git clone https://github.com/MrSupiri/rancher-logging-pipeline-plumber
cd rancher-logging-pipeline-plumber
helm install logging-pipeline-plumber charts/logging-pipeline-plumber
```

### Access the UI

```sh
echo "Visit http://localhost:9090 to get to dashboard"
kubectl port-forward svc/logging-pipeline-plumber 9090:9090
```

### Create Flowtest

- Visit [localhost:9090/create](http://localhost:9090/create)
- Name the test
- Select the target pod and target flow
- Enter log messages that are needed to be tested
   - UI will pre-fill last 10 log messages from the selected pod
- Press create

### Check Results

- From [localhost:9090](http://localhost:9090) select the Flow Test that needs to be inspected 
- Then UI will show filters and match statements which pass at least one log message through to the [output](https://banzaicloud.com/docs/one-eye/logging-operator/configuration/output/)


## Development

### Pre Requisite
- [Go 1.16+](https://golang.org/)
- [Yarn](https://yarnpkg.com/)
- [Docker](https://www.docker.com/)
- [Helm](https://helm.sh/)
- [Yq](https://github.com/mikefarah/yq/)
- A Cluster with [Rancher](https://rancher.com/docs/rancher/v2.5/en/installation/install-rancher-on-k8s/) installed ([k3d](https://k3d.io/)  is prefered)


### Run

To start the operator in develop mode run
```sh
export LOG_OUTPUT_ENDPOINT=http://localhost:8312/
make install
make run
```
this will install CRDs, Roles, and Service account needs to run the operator in your default kubectl context.

When running a FlowTest operator needs to talk to the log-output pod and this pod will be only scheduled while only a test running. So as soon as a new test is started create a port-forwarding to that pod.
```sh
kubectl port-forward svc/logging-plumber-log-aggregator 8312:80
``` 


### Build Locally

```sh
git clone https://github.com/MrSupiri/rancher-logging-pipeline-plumber
cd rancher-logging-pipeline-plumber
make build
```

This will create
  - Docker Image which has an operator with built UI
  - Docker Image of [Pod Simulator](https://github.com/MrSupiri/rancher-logging-pipeline-plumber/tree/main/pod-simulator)
  - Helm chart with updated CRDs and Roles

