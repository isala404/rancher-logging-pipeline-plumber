import React from 'react';
import axios from 'axios';
import HomePage from './pages/Home/HomePage';

if (process.env.REACT_APP_BASE_URL) {
  axios.defaults.baseURL = process.env.REACT_APP_BASE_URL;
}

function App() {
  return (
    <div className="App">
      <HomePage />
    </div>
  );
}

export default App;
