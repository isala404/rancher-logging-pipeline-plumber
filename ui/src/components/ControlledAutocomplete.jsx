/* eslint-disable react/jsx-props-no-spreading */
import React from 'react';
import TextField from '@material-ui/core/TextField';
import Autocomplete from '@material-ui/lab/Autocomplete';
import { Controller } from 'react-hook-form';

const ControlledAutocomplete = (props) => {
  const [open, setOpen] = React.useState(false);
  const [loading, setLoading] = React.useState(false);
  const [options, setOptions] = React.useState([]);

  const {
    // eslint-disable-next-line react/prop-types
    control, label, variant, name, style, fetchfunc, disabled, required,
  } = props;

  React.useEffect(async () => {
    if (!open) {
      return;
    }
    setLoading(true);
    const data = await fetchfunc();
    setOptions(data);
    setLoading(false);
  }, [open]);

  return (
    <Controller
      render={({ field }) => (
        <Autocomplete
          {...field}
          onOpen={() => {
            setOpen(true);
          }}
          onClose={() => {
            setOpen(false);
          }}
          disabled={disabled}
          options={options}
          loading={loading}
          style={style}
          renderInput={(params) => (
            <TextField
              {...params}
              style={style}
              label={label}
              variant={variant}
              required={required}
            />
          )}
          onChange={(_, data) => {
            field.onChange(data);
          }}
        />
      )}
      name={name}
      control={control}
    />
  );
};

export default ControlledAutocomplete;
