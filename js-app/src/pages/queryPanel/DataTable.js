import React, {PureComponent} from "react";
import {AgGridReact} from 'ag-grid-react';
import 'ag-grid-community/dist/styles/ag-grid.css';
import 'ag-grid-community/dist/styles/ag-theme-balham.css';

export default class DataTable extends PureComponent {

    constructor(props) {
        super(props);
        this.state = {
            columnDefs: []
        };
    }

    componentWillReceiveProps(nextProps) {
        if (nextProps.data !== this.props.data) {
            const columns = [];

            if (nextProps.data !== null && nextProps.data.columns.length !== 0) {
                nextProps.data.columns.forEach((columnName, index) => {
                    let element = {
                        headerName: columnName,
                        valueGetter: (row) => row.data[index]
                    };
                    if (columnName === "groupId") {
                        element.sort = 'asc';
                        element.maxWidth = 48;
                        element.type = "numericColumn";
                        element.pinned = "left";
                        element.filter = "agNumberColumnFilter";
                    }
                    columns.push(element);
                });
            }
            this.setState({
                columnDefs: columns
            });
        }
    }

    autoSizeColumns(params) {
        let columnIds = [];
        params.columnApi.getAllColumns().forEach(function (column) {
            columnIds.push(column.colId);
        });
        params.columnApi.autoSizeColumns(columnIds);
    };

    render() {
        return (
            <div
                className="ag-theme-balham flex-container-fill"
                style={{display: 'initial'}}
            >
                <AgGridReact
                    onGridReady={this.autoSizeColumns}
                    onGridColumnsChanged={this.autoSizeColumns}
                    floatingFilter
                    defaultColDef={{
                        editable: true,
                        sortable: true,
                        filter: true,
                        resizable: true
                    }}
                    columnDefs={this.state.columnDefs}
                    rowData={this.props.data.rows}
                />
            </div>
        );
    }
}