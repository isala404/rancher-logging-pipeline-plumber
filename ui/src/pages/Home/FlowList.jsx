import React, { useState, useEffect } from 'react';
import { DataGrid } from '@material-ui/data-grid';
import Container from '@material-ui/core/Container';
import getFlowTests from '../../libs/fetchFlowTests';

const columns = [
  {
    field: 'status', headerName: 'Status', minWidth: 150,
  },
  {
    field: 'name', headerName: 'Name', minWidth: 250,
  },
  {
    field: 'flowType', headerName: 'Flow Type', minWidth: 150,
  },
  {
    field: 'referencePod', headerName: 'Reference Pod', minWidth: 200,
  },
  {
    field: 'referenceFlow', headerName: 'Reference Flow', minWidth: 200,
  },
  {
    field: 'totalTests', headerName: 'Total Tests', minWidth: 150, type: 'number',
  },
  {
    field: 'passedTests', headerName: 'Passed Tests', minWidth: 165, type: 'number',
  },
  {
    field: 'failedTests', headerName: 'Failed Tests', minWidth: 165, type: 'number',
  },
  {
    field: 'createdAt', headerName: 'Created At', minWidth: 175, type: 'dateTime',
  },

];

export default function FlowList() {
  const [data, setData] = useState([]);

  useEffect(async () => {
    setData(await getFlowTests());
  }, []);

  // eslint-disable-next-line no-unreachable
  return (
    <Container style={{ width: '95%', maxWidth: '100%' }}>
      <DataGrid
        rows={data}
        columns={columns}
        checkboxSelection
        autoHeight
        disableExtendRowFullWidth
      />
    </Container>
  );
}
