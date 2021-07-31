import * as React from 'react';
import Container from '@material-ui/core/Container';
import FlowList from './FlowList';

export default function DataTable() {
  return (
    <Container style={{ width: '95%', maxWidth: '100%' }}>
      <FlowList />
    </Container>
  );
}
