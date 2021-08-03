import axios from 'axios';
import snackbarUtils from '../../libs/snackbarUtils';

const flowTestSample = {
  apiVersion: 'loggingplumber.isala.me/v1alpha1',
  kind: 'FlowTest',
  namespace: 'default',
  metadata: {
    name: 'flowtest-sample',
    labels: {
      'app.kubernetes.io/managed-by': 'logging-pipeline-plumber',
      'app.kubernetes.io/created-by': 'logging-plumber',
      'loggingplumber.isala.me/flowtest': 'flowtest-sample',
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
    const res = await axios.post(`k8s/apis/loggingplumber.isala.me/v1alpha1/namespaces/${flowTest.namespace}/flowtests`, flowTest);
    if (res.status === 201) {
      snackbarUtils.success('FlowTest Created');
    }
  } catch (error) {
    snackbarUtils.error(`[HTTP error]: ${error.message}`);
    snackbarUtils.warning('Failed to create flowtest');
  }
}

export const getPods = async (namespace) => {
  const pods = [];
  try {
    const res = await axios.get(`k8s/api/v1/namespaces/${namespace}/pods`);
    res.data.items.forEach((pod) => {
      pods.push({
        name: pod.metadata.name,
        namespace: pod.metadata.namespace,
      });
    });
  } catch (error) {
    snackbarUtils.error(`[HTTP error]: ${error.message}`);
    snackbarUtils.warning('Failed to fetch pods');
  }
  return pods;
};
