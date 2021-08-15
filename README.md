# Logging Pipeline Plumber for Rancher


[![Go Report Card](https://goreportcard.com/badge/github.com/MrSupiri/rancher-logging-pipeline-plumber)](https://goreportcard.com/report/github.com/MrSupiri/rancher-logging-pipeline-plumber)

This a tool which can be used to debug logging piplines built using [Rancher Logging](https://rancher.com/docs/rancher/v2.5/en/logging/).


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
- Select the target pod and tartget flow
- Enter log messages that's needed to be tested
   - UI will pre fill last 10 log messages from the selected pod
- Press create

### Check Results

- From [localhost:9090](http://localhost:9090) select the Flow Test that's need to be inspcted 
- Then UI will show filters and match statements which pass atleast one log message though to the [output](https://banzaicloud.com/docs/one-eye/logging-operator/configuration/output/)


## Development

### Pre Requisite
- [Go 1.16+](https://golang.org/)
- [Yarn](https://yarnpkg.com/)
- [Docker](https://www.docker.com/)
- [Helm](https://helm.sh/)
- A Cluster with [Rancher](https://rancher.com/docs/rancher/v2.5/en/installation/install-rancher-on-k8s/) installed ([k3d](https://k3d.io/)  is prefered)


### Run

To start the operator in develop mode run
```sh
export LOG_OUTPUT_ENDPOINT=http://localhost:8312/
make install
make run
```
this will install CRDs, Roles and Service account nneeds to run the operator in your default kubectl context.

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
  - Docker Image which has operator with built UI
  - Docker Image of [Pod Simulator](https://github.com/MrSupiri/rancher-logging-pipeline-plumber/tree/main/pod-simulator)
  - Helm chart with updated CRDs and Roles

