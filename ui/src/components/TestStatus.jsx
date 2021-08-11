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
      <Grid item xs={12} md={6}>
        Passing
        <div>
          {tests?.map((match, index) => {
            if (status[index]) {
              return (
                <pre style={{ color: 'green' }}>
                  { YAML.stringify(match) }
                </pre>
              );
            }
            return null;
          })}
        </div>
      </Grid>
      <Grid item xs={12} md={6}>
        Failing
        <div>
          {tests?.map((match, index) => {
            if (!status[index]) {
              return (
                <pre style={{ color: 'red' }}>
                  { YAML.stringify(match) }
                </pre>
              );
            }
            return null;
          })}
        </div>
      </Grid>
    </Grid>
  </>
);

export default TestStatus;
