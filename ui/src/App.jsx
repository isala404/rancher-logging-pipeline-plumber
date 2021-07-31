import React from 'react';
import axios from 'axios';
import { SnackbarProvider } from 'notistack';
import { SnackbarUtilsConfigurator } from './libs/snackbarUtils';
import HomePage from './pages/Home/HomePage';

if (process.env.REACT_APP_BASE_URL) {
  axios.defaults.baseURL = process.env.REACT_APP_BASE_URL;
}

function App() {
  return (
    <div className="App">
      <SnackbarProvider maxSnack={5}>
        <SnackbarUtilsConfigurator />
        <HomePage />
      </SnackbarProvider>
    </div>
  );
}

export default App;
