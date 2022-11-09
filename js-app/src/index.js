import React from 'react';
import ReactDOM from 'react-dom';
import QueryPanel from "./pages/queryPanel/QueryPanel";
import './index.css';
import "@blueprintjs/core/lib/css/blueprint.css";
import "@blueprintjs/select/lib/css/blueprint-select.css";
import "@blueprintjs/popover2/lib/css/blueprint-popover2.css";
import "normalize.css/normalize.css";

ReactDOM.render(<QueryPanel/>, document.getElementById('root'));
