/* eslint-disable no-param-reassign */
import axios from 'axios';
import snackbarUtils from '../../libs/snackbarUtils';

export const getFlow = async (namespace, kind, name) => {
  try {
    const res = await axios.get(`k8s/apis/logging.banzaicloud.io/v1beta1/namespaces/${namespace}/${kind.toLowerCase()}s/${name}`);
    return res.data;
  } catch (error) {
    snackbarUtils.error(`[HTTP error]: ${error.message}`);
    snackbarUtils.warning(`Failed to fetch ${kind} ${namespace}.${name}`);
  }
  return null;
};

export const getFlowTest = async (namespace, name) => {
  try {
    const res = await axios.get(`k8s/apis/loggingpipelineplumber.isala.me/v1alpha1/namespaces/${namespace}/flowtests/${name}`);
    return res.data;
  } catch (error) {
    snackbarUtils.error(`[HTTP error]: ${error.message}`);
    snackbarUtils.warning(`Failed to fetch flowtest ${namespace}.${name}`);
  }
  return null;
};

export const cleanFlowTest = (flowTest) => {
  delete flowTest?.metadata.generation;
  delete flowTest?.metadata.managedFields;
  return flowTest;
};
