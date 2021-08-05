/* eslint-disable no-unused-vars */
/* eslint-disable react/jsx-props-no-spreading */
import * as React from 'react';
import Container from '@material-ui/core/Container';
import { useForm, Controller, useWatch } from 'react-hook-form';
import {
  InputLabel, TextField, Select, MenuItem, Paper, Grid, Button, FormControl,
} from '@material-ui/core';
import ParseTextarea from '../components/ParseTextarea';
import ControlledAutocomplete from '../components/ControlledAutocomplete';
import {
  createFlowTest, getNamespaces, getPods, getFlows,
} from '../utils/flowtests/createFlowTest';

export default function CreateView() {
  const { control, handleSubmit } = useForm();
  const referencePodNS = useWatch({ control, name: 'spec.referencePod.namespace' });
  const referenceFlowKind = useWatch({ control, name: 'spec.referenceFlow.kind', defaultValue: '' });
  const referenceFlowNS = useWatch({ control, name: 'spec.referenceFlow.namespace' });

  const onSubmit = (data) => {
    // createFlowTest(data);
    console.log(data);
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
              <ControlledAutocomplete
                control={control}
                label="Namespace"
                variant="outlined"
                name="spec.referencePod.namespace"
                fetchfunc={async () => getNamespaces()}
                style={{ display: 'inline-block', width: '252px' }}
              />
              <ControlledAutocomplete
                control={control}
                label="Pod Name"
                variant="outlined"
                name="spec.referencePod.name"
                fetchfunc={async () => getPods(referencePodNS)}
                style={{ display: 'inline-block', width: '252px' }}
                disabled={!referencePodNS}
              />
            </Grid>
            <Grid item xs={12}>
              <Controller
                render={({ field }) => (
                  <FormControl variant="outlined">
                    <InputLabel id="spec-referenceFlow-kind-label">FlowType</InputLabel>
                    <Select
                      labelId="spec-referenceFlow-kind-label"
                      id="spec-referenceFlow-kind"
                      label="FlowType"
                      {...field}
                      style={{ display: 'inline-block', width: '252px' }}
                    >
                      <MenuItem value="Flow">Flow</MenuItem>
                      <MenuItem value="ClusterFlow">ClusterFlow</MenuItem>
                    </Select>
                  </FormControl>
                )}
                name="spec.referenceFlow.kind"
                control={control}
              />
              <ControlledAutocomplete
                control={control}
                label="Namespace"
                variant="outlined"
                name="spec.referenceFlow.namespace"
                fetchfunc={async () => getNamespaces()}
                style={{ display: 'inline-block', width: '252px' }}
                disabled={!referenceFlowKind}
              />
              <ControlledAutocomplete
                control={control}
                label="Flow Name"
                variant="outlined"
                name="spec.referenceFlow.name"
                fetchfunc={async () => getFlows(referenceFlowNS, referenceFlowKind)}
                style={{ display: 'inline-block', width: '252px' }}
                disabled={!referenceFlowNS}
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
