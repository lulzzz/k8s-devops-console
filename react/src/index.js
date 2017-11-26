import React from 'react';
import ReactDOM from 'react-dom';
import App from './App';

window.$ = window.jQuery = require("jquery");
ReactDOM.render(<App/>, document.getElementById("root"));

