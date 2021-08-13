/* eslint-disable react/prop-types */
import * as React from 'react';
import { Grid } from '@material-ui/core';
import YAML from 'json-to-pretty-yaml';

const TestStatus = ({
  type, tests, status,
}) => (
  <>
    <div style={{ margin: '10px' }}>
      {`Tested ${type}`}
    </div>
    <Grid container style={{ marginLeft: '30px' }}>
      <Grid item xs={12}>
        <div>
          {tests?.map((match, index) => (
            <div key={YAML.stringify(match)} style={{ display: 'table' }}>
              {
                status[index]
                  ? <div className="badge-wrapper"><span className="badge badge-pass">Pass</span></div>
                  : <div className="badge-wrapper"><span className="badge badge-fail">Fail</span></div>
              }
              <pre style={{ display: 'inline-block', color: status[index] ? 'green' : 'red' }}>
                { YAML.stringify(match) }
              </pre>
            </div>
          ))}
        </div>
      </Grid>
    </Grid>
  </>
);

export default TestStatus;
