import React from 'react';
import axios from 'axios';
import { SnackbarProvider } from 'notistack';
import {
  BrowserRouter as Router,
  Switch,
  Route,
} from 'react-router-dom';
import { SnackbarUtilsConfigurator } from './libs/snackbarUtils';
import ListView from './pages/ListPage';
import CreateView from './pages/CreatePage';
import DetailView from './pages/DetailPage';

if (process.env.REACT_APP_BASE_URL) {
  axios.defaults.baseURL = process.env.REACT_APP_BASE_URL;
}

function App() {
  return (
    <Router>
      <div className="App">
        <SnackbarProvider maxSnack={5}>
          <SnackbarUtilsConfigurator />
          <Switch>
            <Route exact path="/">
              <ListView />
            </Route>
            <Route path="/create">
              <CreateView />
            </Route>
            <Route path="/flowtest">
              <DetailView />
            </Route>
          </Switch>
        </SnackbarProvider>
      </div>
    </Router>
  );
}

export default App;
