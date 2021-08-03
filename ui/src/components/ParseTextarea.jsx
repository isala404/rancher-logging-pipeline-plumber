import * as React from 'react';
import { TextareaAutosize } from '@material-ui/core';

// eslint-disable-next-line react/prop-types
const ParseTextarea = ({ onChange }) => {
  const [text, setText] = React.useState('');

  const handleChange = (e) => {
    const newValue = e.target.value;
    setText(newValue);
    onChange(newValue.split('\n'));
  };

  return <TextareaAutosize onChange={handleChange} value={text} style={{ height: '180px', width: '750px' }} />;
};

export default ParseTextarea;
