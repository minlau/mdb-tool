import React, {PureComponent} from "react";
import {AgGridReact} from 'ag-grid-react';
import 'ag-grid-community/dist/styles/ag-grid.css';
import 'ag-grid-community/dist/styles/ag-theme-balham.css';

export default class DataTable extends PureComponent {

    autoSizeColumns(params) {
        let columnIds = [];
        params.columnApi.getColumns().forEach(function (column) {
            columnIds.push(column.colId);
        });
        params.columnApi.autoSizeColumns(columnIds);
    };

    render() {
        const columnDefs = [];
        this.props.data.columns.forEach(column => {
            let element = {
                headerName: column.name,
                field: column.fieldName
            };
            if (column.name === "groupName") {
                element.sort = 'asc';
                element.maxWidth = 96;
                element.pinned = "left";
                element.filter = "agTextColumnFilter";
            }
            columnDefs.push(element);
        });

        return (
            <div
                className="ag-theme-balham flex-container-fill"
                style={{display: 'initial'}}
            >
                <AgGridReact
                    onGridColumnsChanged={this.autoSizeColumns}
                    pagination={true}
                    paginationPageSize={10000}
                    defaultColDef={{
                        editable: true,
                        sortable: true,
                        filter: true,
                        resizable: true,
                        floatingFilter: true
                    }}
                    columnDefs={columnDefs}
                    rowData={this.props.data.rows}
                />
            </div>
        );
    }
}