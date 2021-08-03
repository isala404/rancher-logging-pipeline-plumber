import * as React from 'react';
import Container from '@material-ui/core/Container';
import { useParams } from 'react-router-dom';

export default function DetailView() {
  const { uuid } = useParams();
  return (
    <Container style={{ width: '95%', maxWidth: '100%' }}>
      <h1>
        {`Detail View for flowtest ${uuid}`}
      </h1>
    </Container>
  );
}
