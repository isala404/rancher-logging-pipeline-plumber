import React, { useEffect, useState } from 'react';
import {
  Container, Grid, Paper,
} from '@material-ui/core';
import { useParams } from 'react-router-dom';
import { getFlow, getFlowTest } from '../utils/flowtests/flowDetails';
import TestStatus from '../components/TestStatus';

export default function DetailView() {
  const { namespace, name } = useParams();
  const [flowTest, setFlowTest] = useState(undefined);
  const [flow, setFlow] = useState(undefined);

  useEffect(async () => {
    if (flowTest?.spec?.referenceFlow) {
      setFlow(await getFlow(
        flowTest.spec.referenceFlow.namespace,
        flowTest.spec.referenceFlow.kind,
        flowTest.spec.referenceFlow.name,
      ));
    }
  }, [flowTest]);

  useEffect(async () => {
    setFlowTest(await getFlowTest(namespace, name));
  }, []);

  return (
    <Container style={{ width: '95%', maxWidth: '100%' }}>
      <h1>{name}</h1>
      <h3>Details</h3>
      <Paper>
        <Grid container spacing={3}>
          <Grid item xs={12} xl={6}>
            <div style={{ margin: '10px' }}>
              Status:
              {` ${flowTest?.status?.status}`}
            </div>
            <div style={{ margin: '10px' }}>Reference Flow</div>
            <div style={{ marginLeft: '30px' }}>
              <div>{`Kind: ${flowTest?.spec.referenceFlow.kind}`}</div>
              <div>{`Namespace: ${flowTest?.spec.referenceFlow.namespace}`}</div>
              <div>{`Name: ${flowTest?.spec.referenceFlow.name}`}</div>
            </div>
            <div style={{ margin: '10px' }}>Reference Pod</div>
            <div style={{ marginLeft: '30px' }}>
              <div>{`Namespace: ${flowTest?.spec.referencePod.namespace}`}</div>
              <div>{`Name: ${flowTest?.spec.referencePod.name}`}</div>
            </div>
            <div style={{ margin: '10px' }}>Testing Logs</div>
            <div style={{ marginLeft: '30px' }}>
              {flowTest?.spec?.sentMessages?.map((message) => (
                <pre>
                  { message }
                </pre>
              ))}
            </div>
          </Grid>
          <Grid item xs={12} xl={6}>
            <TestStatus type="Matches" tests={flow?.spec?.match} status={flowTest?.status?.matchStatus} />
            <TestStatus type="Filters" tests={flow?.spec?.filters} status={flowTest?.status?.filterStatus} />
          </Grid>
        </Grid>
      </Paper>
    </Container>
  );
}
