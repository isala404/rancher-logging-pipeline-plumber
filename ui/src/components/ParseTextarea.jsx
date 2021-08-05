/* eslint-disable react/prop-types */
import * as React from 'react';
import { TextareaAutosize } from '@material-ui/core';
import { getLastNlogs } from '../utils/flowtests/createFlowTest';

const ParseTextarea = ({
  onChange, pod, namespace, nLines,
}) => {
  const [text, setText] = React.useState('');

  React.useEffect(async () => {
    if (!pod || !namespace || !nLines) {
      return;
    }
    const logs = await getLastNlogs(pod, namespace, nLines);
    setText(logs);
  }, [pod, namespace]);

  const handleChange = (e) => {
    const newValue = e.target.value;
    setText(newValue);
    onChange(newValue.split('\n'));
  };

  return <TextareaAutosize onChange={handleChange} value={text} style={{ height: '180px', width: '750px' }} />;
};

export default ParseTextarea;
