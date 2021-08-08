import React, { useState, useEffect } from 'react';
import { DataGrid } from '@material-ui/data-grid';
import { Button } from '@material-ui/core';
import { useHistory } from 'react-router-dom';
import fetchFlowTests from '../utils/flowtests/fetchFlowTests';
import deleteFlowTest from '../utils/flowtests/deleteFlowTest';

const columns = [
  {
    field: 'status', headerName: 'Status', minWidth: 150,
  },
  {
    field: 'name', headerName: 'Name', minWidth: 230,
  },
  {
    field: 'flowType', headerName: 'Flow Type', minWidth: 150,
  },
  {
    field: 'namespace', headerName: 'Namespace', minWidth: 160,
  },
  {
    field: 'referencePod', headerName: 'Reference Pod', minWidth: 175,
  },
  {
    field: 'referenceFlow', headerName: 'Reference Flow', minWidth: 180,
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
  const history = useHistory();

  useEffect(async () => {
    setLoading(false);
    setData(await fetchFlowTests());
  }, []);

  const getSelectedFlows = (e) => {
    const selectedIDs = new Set(e);
    const selectedRowData = data.filter((row) => selectedIDs.has(row.id));
    setSeletedFlows(selectedRowData);
  };

  return (
    <div>
      <Button style={{ marginBottom: '10px' }} variant="contained" color="secondary" onClick={() => { deleteFlowTest(selectedFlows); }}>
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
        onRowClick={(row) => history.push(`flowtest/${row.row.name}`)}
      />
    </div>
  );
}
