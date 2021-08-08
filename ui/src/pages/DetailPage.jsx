import * as React from 'react';
import { Container, Grid, Paper } from '@material-ui/core';
import { useParams } from 'react-router-dom';

export default function DetailView() {
  const { name } = useParams();
  return (
    <Container style={{ width: '95%', maxWidth: '100%' }}>
      <h1>{`${name}`}</h1>
      <Paper>
        <Grid container spacing={3}>
          <Grid item xs={12} lg={6}>
            <div>Status: Running</div>
            <div>Reference Flow</div>
            <div style={{ marginLeft: '20px' }}>
              <div>Kind: ClusterFlow</div>
              <div>Name: flow-test</div>
              <div>Namespace: default</div>
            </div>
            <div>Reference Pod</div>
            <div style={{ marginLeft: '20px' }}>
              <div>Name: busybox</div>
              <div>Namespace: default</div>
            </div>
            <div>Testing Logs</div>
            <div style={{ marginLeft: '20px' }}>
              <div>
                [2021-06-10T11:50:06Z] @DEBUG Tam ipsae consuetudo infelix
                adtendi contexo mansuefecisti diutius re. 1373 ::0.403911
              </div>
              <div>
                [2021-06-10T11:50:07Z] @WARNING Ne hi flagitantur alienam
                neglecta. 1374 ::0.474177
              </div>
              <div>
                [2021-06-10T11:50:08Z] @INFO Amo ideoque die se at, caro aer, ad
                cor. 1375 ::0.263548
              </div>
              <div>
                [2021-06-10T11:50:09Z] @INFO Se contexo servis inpiis erogo,
                diligit ita significaret eosdem. 1376 ::0.405282
              </div>
            </div>
          </Grid>
          <Grid item xs={6}>
            <div>Tested Matches</div>
            <div style={{ marginLeft: '20px' }}>
              <pre style={{ color: 'red' }}>
                - select: labels: loggingplumber.isala.me/test: invalid
              </pre>
              <pre>
                - select: labels: loggingplumber.isala.me/test: invalid
              </pre>
              <pre style={{ color: 'red' }}>
                - select: labels: loggingplumber.isala.me/test: invalid
              </pre>
            </div>
            <div>Tested Filters</div>
            <div style={{ marginLeft: '20px' }}>
              <pre style={{ color: 'red' }}>
                - grep:
                regexp:
                - key: first
                pattern: /^5\d\d$/
              </pre>
              <pre>
                - record_modifier:
                records:
                - foo: bar
              </pre>
            </div>
          </Grid>
        </Grid>
      </Paper>
    </Container>
  );
}
