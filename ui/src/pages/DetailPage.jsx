import React, { useEffect, useState } from 'react';
import {
  Container, Grid, Paper, Tooltip,
} from '@material-ui/core';
import { useParams } from 'react-router-dom';
import YAML from 'json-to-pretty-yaml';
import { getFlow, getFlowTest } from '../utils/flowtests/flowDetails';

export default function DetailView() {
  const { namespace, name } = useParams();
  const [flowTest, setFlowTest] = useState(undefined);
  const [flow, setFlow] = useState(undefined);

  const matchPassingText = 'Logs pass though to the output when the match statement is present';
  const matchFailingText = 'This match statement blocks logs from passing to output';
  const filterPassingText = 'Logs pass though to the Output when the filter statement is present';
  const filterFailingText = 'This filter statement blocks logs from passing to output';

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
          <Grid item xs={12} lg={6}>
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
          <Grid item xs={6}>
            <div style={{ margin: '10px' }}>Tested Matches</div>
            <div style={{ marginLeft: '30px' }}>
              {flow?.spec?.match?.map((match, index) => (
                <Tooltip
                  placement="bottom-start"
                  title={flowTest.status?.matchStatus[index] ? matchPassingText : matchFailingText}
                >
                  <pre style={{ color: flowTest.status?.matchStatus[index] ? 'green' : 'red' }}>
                    { YAML.stringify(match) }
                  </pre>
                </Tooltip>
              ))}
            </div>
            <div style={{ margin: '10px' }}>Tested Filters</div>
            <div style={{ marginLeft: '30px' }}>
              {flow?.spec?.filters?.map((filter, index) => (
                <Tooltip
                  placement="bottom-start"
                  title={
                    flowTest.status?.filterStatus[index] ? filterPassingText : filterFailingText
                  }
                >
                  <pre style={{ color: flowTest.status?.filterStatus[index] ? 'green' : 'red' }}>
                    { YAML.stringify(filter) }
                  </pre>
                </Tooltip>
              ))}
            </div>
          </Grid>
        </Grid>
      </Paper>
    </Container>
  );
}
