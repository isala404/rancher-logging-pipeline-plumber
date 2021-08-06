import * as yup from 'yup';

const schema = yup.object().shape({
  metadata: yup.object().shape({
    name: yup.string().required('FlowTest Name is required'),
  }),
  spec: yup.object().shape({
    referencePod: yup.object().shape({
      namespace: yup.string().required('Pod namespace is required'),
      name: yup.string().required('Pod name is required'),
    }),
    referenceFlow: yup.object().shape({
      kind: yup.string().oneOf(['Flow', 'ClusterFlow']),
      namespace: yup.string().required('Flow namespace is required'),
      name: yup.string().required('Flow name is required'),
    }),
    sentMessages: yup.array().of(yup.string()).required('FlowTest must have at least one message'),
  }),
});

export default schema;
