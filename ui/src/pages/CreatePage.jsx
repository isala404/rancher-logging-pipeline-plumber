/* eslint-disable no-unused-vars */
/* eslint-disable react/jsx-props-no-spreading */
import * as React from 'react';
import Container from '@material-ui/core/Container';
import { useForm, Controller } from 'react-hook-form';
import {
  Checkbox, TextField, Select, FormLabel, Paper, Grid, Button, TextareaAutosize,
} from '@material-ui/core';
import ParseTextarea from '../components/ParseTextarea';
import { createFlowTest } from '../utils/flowtests/createFlowTest';

export default function CreateView() {
  const { control, handleSubmit } = useForm();

  const onSubmit = (data) => {
    createFlowTest(data);
  };

  return (
    <Container style={{ width: '95%', maxWidth: '100%' }}>
      <h1>Create FlowTest</h1>
      <form onSubmit={handleSubmit(onSubmit)}>
        <Paper style={{ padding: 16 }}>
          <Grid container alignItems="flex-start" spacing={2}>
            <Grid item xs={12}>
              <Controller
                render={({ field }) => <TextField {...field} label="Test Name" variant="outlined" />}
                name="metadata.name"
                control={control}
                defaultValue=""
              />
            </Grid>
            <Grid item xs={12}>
              <Controller
                render={({ field }) => <TextField {...field} label="Pod Name" variant="outlined" />}
                name="spec.referencePod.name"
                control={control}
                defaultValue=""
              />
              <Controller
                render={({ field }) => <TextField {...field} label="Namespace" variant="outlined" />}
                name="spec.referencePod.namespace"
                control={control}
                defaultValue=""
              />
            </Grid>
            <Grid item xs={12}>
              <Controller
                render={({ field }) => <TextField {...field} label="Flow Type" variant="outlined" />}
                name="spec.referenceFlow.kind"
                control={control}
                defaultValue=""
              />
              <Controller
                render={({ field }) => <TextField {...field} label="Flow Name" variant="outlined" />}
                name="spec.referenceFlow.name"
                control={control}
                defaultValue=""
              />
              <Controller
                render={({ field }) => <TextField {...field} label="Namespace" variant="outlined" />}
                name="spec.referenceFlow.namespace"
                control={control}
                defaultValue=""
              />
            </Grid>
            <Grid item xs={12}>
              <Controller
                // without ref=null, it logs an error
                render={({ field }) => <ParseTextarea {...field} ref={null} />}
                name="spec.sentMessages"
                control={control}
              />
            </Grid>
            <Grid item xs={12}>
              <Button type="submit" variant="contained" color="primary">
                Create
              </Button>
              <Button type="reset" variant="contained" color="secondary">
                Rest
              </Button>
            </Grid>
          </Grid>
        </Paper>
      </form>
    </Container>
  );
}
