import React from 'react';
import { useSnackbar } from 'notistack';

const InnerSnackbarUtilsConfigurator = (props) => {
  // eslint-disable-next-line react/destructuring-assignment
  props.setUseSnackbarRef(useSnackbar());
  return null;
};

let useSnackbarRef;
const setUseSnackbarRef = (useSnackbarRefProp) => {
  useSnackbarRef = useSnackbarRefProp;
};

export const SnackbarUtilsConfigurator = () => (
  <InnerSnackbarUtilsConfigurator setUseSnackbarRef={setUseSnackbarRef} />
);

export default {
  success(msg) {
    this.toast(msg, 'success');
  },
  warning(msg) {
    this.toast(msg, 'warning');
  },
  info(msg) {
    this.toast(msg, 'info');
  },
  error(msg) {
    this.toast(msg, 'error');
  },
  toast(msg, variant = 'default') {
    useSnackbarRef.enqueueSnackbar(msg, { variant });
  },
};
