import axios from 'axios';
import snackbarUtils from '../../libs/snackbarUtils';

const flowTestSample = {
  apiVersion: 'loggingpipelineplumber.isala.me/v1alpha1',
  kind: 'FlowTest',
  namespace: 'default',
  metadata: {
    name: 'flowtest-sample',
    labels: {
      'app.kubernetes.io/managed-by': 'logging-pipeline-plumber',
      'app.kubernetes.io/created-by': 'logging-plumber',
      'loggingpipelineplumber.isala.me/flowtest': 'flowtest-sample',
    },
  },
  spec: {
    referencePod: {
      kind: 'Pod',
    },
    referenceFlow: {},
    sentMessages: [],
  },
};

export async function createFlowTest(data) {
  try {
    const flowTest = { ...flowTestSample, ...data };
    flowTest.spec.referencePod.kind = 'Pod';
    const res = await axios.post(`k8s/apis/loggingpipelineplumber.isala.me/v1alpha1/namespaces/${flowTest.namespace}/flowtests`, flowTest);
    if (res.status === 201) {
      snackbarUtils.success('FlowTest Created');
    }
    return true;
  } catch (error) {
    snackbarUtils.error(`[HTTP error]: ${error.message}`);
    snackbarUtils.warning('Failed to create flowtest');
  }
  return false;
}

export const getPods = async (namespace) => {
  const pods = [];
  try {
    const res = await axios.get(`k8s/api/v1/namespaces/${namespace}/pods`);
    res.data.items.forEach((pod) => {
      pods.push(pod.metadata.name);
    });
  } catch (error) {
    snackbarUtils.error(`[HTTP error]: ${error.message}`);
    snackbarUtils.warning('Failed to fetch pods');
  }
  return pods;
};

export const getNamespaces = async () => {
  const namespaces = [];
  try {
    const res = await axios.get('k8s/api/v1/namespaces');
    res.data.items.forEach((namespace) => {
      namespaces.push(namespace.metadata.name);
    });
  } catch (error) {
    snackbarUtils.error(`[HTTP error]: ${error.message}`);
    snackbarUtils.warning('Failed to fetch namespaces');
  }
  return namespaces;
};

export const getFlows = async (namespace, type) => {
  const flows = [];
  try {
    let res;
    if (type === 'Flow') {
      res = await axios.get(`k8s/apis/logging.banzaicloud.io/v1beta1/namespaces/${namespace}/flows`);
    } else if (type === 'ClusterFlow') {
      res = await axios.get(`k8s/apis/logging.banzaicloud.io/v1beta1/namespaces/${namespace}/clusterflows`);
    } else {
      snackbarUtils.warning('Invalid flow type');
      return flows;
    }
    res.data.items.forEach((flow) => {
      flows.push(flow.metadata.name);
    });
  } catch (error) {
    snackbarUtils.error(`[HTTP error]: ${error.message}`);
    snackbarUtils.warning('Failed to fetch flows');
  }
  return flows;
};

export const getLastNlogs = async (pod, namespace, nLines) => {
  try {
    const res = await axios.get(`k8s/api/v1/namespaces/${namespace}/pods/${pod}/log?tailLines=${nLines}`);
    return res.data;
  } catch (error) {
    snackbarUtils.error(`[HTTP error]: ${error.message}`);
    snackbarUtils.warning('Failed to fetch logs');
  }
  return '';
};
