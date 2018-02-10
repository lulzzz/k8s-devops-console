import React from 'react';
import ReactDOM from 'react-dom';
import App from './App';

window.$ = window.jQuery = require("jquery");
(function () {
    let rootEl = document.getElementById("root");
    if (rootEl) {
        ReactDOM.render(<App/>, rootEl);
    }
})();
