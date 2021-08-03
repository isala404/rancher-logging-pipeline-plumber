import * as React from 'react';
import Container from '@material-ui/core/Container';
import Button from '@material-ui/core/Button';
import Grid from '@material-ui/core/Grid';
import Typography from '@material-ui/core/Typography';
import FlowList from '../components/FlowList';

export default function ListView() {
  return (
    <Container style={{ width: '95%', maxWidth: '100%', marginTop: '50px' }}>

      <Grid
        justifyContent="space-between"
        container
      >
        <Grid item>
          <Typography variant="h4" gutterBottom>
            FlowTests
          </Typography>
        </Grid>

        <Grid item>
          <Button variant="contained" color="primary" href="/create">Create FlowTest</Button>
        </Grid>
      </Grid>
      <FlowList />
    </Container>
  );
}
