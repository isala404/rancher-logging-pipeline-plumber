import React, { useState, useEffect, Fragment } from 'react';
import { DataGrid } from '@material-ui/data-grid';
import { Button } from '@material-ui/core';
import getFlowTests from '../../libs/fetchFlowTests';
import deleteFlowTest from '../../libs/deleteFlowTest';

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
  const [loading, setLoading] = useState(true);
  const [selectedFlows, setSeletedFlows] = useState([]);

  useEffect(async () => {
    setLoading(false);
    setData(await getFlowTests());
  }, []);

  const getSelectedFlows = (e) => {
    const selectedIDs = new Set(e);
    const selectedRowData = data.filter((row) => selectedIDs.has(row.id));
    setSeletedFlows(selectedRowData);
  };

  return (
    <>
      <Button variant="contained" color="primary" onClick={() => { deleteFlowTest(selectedFlows); }}>
        Delete
      </Button>
      <DataGrid
        rows={data}
        loading={loading}
        columns={columns}
        checkboxSelection
        autoHeight
        disableExtendRowFullWidth
        onSelectionModelChange={getSelectedFlows}
      />
    </>
  );
}
