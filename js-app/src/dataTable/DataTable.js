import React, {Component} from "react";
import {AgGridReact} from 'ag-grid-react';
import 'ag-grid/dist/styles/ag-grid.css';
import 'ag-grid/dist/styles/ag-theme-balham.css';

class DataTable extends Component {

    constructor(props) {
        super(props);
        this.state = {
            data: props.data,
            columnDefs: [],
            rowData: []
        };
    }

    static getDerivedStateFromProps(nextProps, prevState) {
        if (nextProps.rowData !== prevState.rowData) {
            const columns = [];

            if (nextProps.data === null || nextProps.data.length === 0) {
                return {
                    columnDefs: [],
                    rowData: []
                };
            }

            Object.keys(nextProps.data[0]).forEach((columnName, index) => {
                let element = {headerName: columnName, field: columnName};
                if (columnName === "groupId") {
                    element.sort = 'asc';
                    element.maxWidth = 48;
                    element.type = "numericColumn";
                    element.pinned = "left";
                    element.filter = "agNumberColumnFilter";
                    columns.unshift(element);
                } else {
                    columns.push(element);
                }
            });

            return {
                columnDefs: columns,
                rowData: nextProps.data
            };
        } else return null;
    }

    render() {
        const autoSizeColumns = (params) => {
            let columnIds = [];
            params.columnApi.getAllColumns().forEach(function (column) {
                columnIds.push(column.colId);
            });
            params.columnApi.autoSizeColumns(columnIds);
        };
        return (
            <div
                className="ag-theme-balham flex-container-fill"
                style={{display: 'initial'}}
            >
                <AgGridReact
                    onGridReady={autoSizeColumns}
                    onGridColumnsChanged={autoSizeColumns}
                    enableSorting={true}
                    enableFilter={true}
                    floatingFilter={true}
                    enableColResize={true}
                    defaultColDef={{
                        editable: true
                    }}
                    columnDefs={this.state.columnDefs}
                    rowData={this.state.rowData}
                />
            </div>
        );
    }
}

export default DataTable;