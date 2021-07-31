import axios from 'axios';
import snackbarUtils from '../../libs/snackbarUtils';

export default async function getFlowTests() {
  const flowTests = [];

  try {
    const res = await axios('k8s/apis/loggingplumber.isala.me/v1alpha1/namespaces/default/flowtests');

    res.data.items.forEach((flowTest, index) => {
      const totalTests = flowTest.status.filterStatus.length + flowTest.status.matchStatus.length;
      // eslint-disable-next-line max-len
      const passedTests = flowTest.status.filterStatus.filter(Boolean).length + flowTest.status.matchStatus.filter(Boolean).length;

      flowTests.push({
        id: index,
        uuid: flowTest.metadata.uid,
        status: flowTest.status.status,
        name: flowTest.metadata.name,
        flowType: flowTest.spec.referenceFlow.kind,
        referencePod: flowTest.spec.referencePod.name,
        referenceFlow: flowTest.spec.referenceFlow.name,
        totalTests,
        passedTests,
        failedTests: totalTests - passedTests,
        createdAt: (new Date(flowTest.metadata.creationTimestamp)).toLocaleString(),
      });
    });
  } catch (error) {
    snackbarUtils.error(`[HTTP error]: ${error.message}`);
    snackbarUtils.warning('Failed to fetch flowtest list');
  }
  return flowTests;
}
