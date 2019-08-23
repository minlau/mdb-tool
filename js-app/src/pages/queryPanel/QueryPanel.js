import React, {Component} from "react";
import {Button, FormGroup, HTMLSelect, MenuItem, Spinner} from "@blueprintjs/core";
import {Select} from "@blueprintjs/select";
import axios from "axios";

import {Controlled as CodeMirror} from 'react-codemirror2'
import 'codemirror/lib/codemirror.css';
import 'codemirror/theme/eclipse.css';
import 'codemirror/mode/sql/sql';

import DataTable from "./DataTable";
import History from "./History";
import ErrorDialog from "./ErrorDialog";
import {highlightText} from "../../utils/select";
import './QueryPanel.css';

const queryModes = [{label: "Multiple", value: "multiple"}, {label: "Single", value: "single"}];

export default class QueryPanel extends Component {

    constructor(props) {
        super(props);

        this.refHandlers = {
            errorDialog: (ref) => this.errorDialog = ref,
            codeMirror: (ref) => this.codeMirror = ref,
            queryHistory: (ref) => this.queryHistory = ref
        };

        this.initializeData = this.initializeData.bind(this);
        this.handleQueryTextChange = this.handleQueryTextChange.bind(this);
        this.handleExecuteClick = this.handleExecuteClick.bind(this);
        this.handleErrorClick = this.handleErrorClick.bind(this);

        this.handleQueryModeChange = this.handleQueryModeChange.bind(this);
        this.handleGroupTypeChange = this.handleGroupTypeChange.bind(this);

        this.handleDatabaseSelect = this.handleDatabaseSelect.bind(this);
        this.renderDatabase = this.renderDatabase.bind(this);

        this.handleQuerySelect = this.handleQuerySelect.bind(this);

        this.state = {
            queryMode: queryModes[0].value,
            groupTypes: [],
            groupType: "",
            databases: [],
            database: null,
            executingQuery: false,
            query: "select now()",
            data: {columns: [], rows: []},
            errors: []
        };

        this.initializeData();
    }

    static selectItemPredicate(query, item) {
        return `${item.groupId}. ${item.title} ${item.groupType}`.toLowerCase().indexOf(query.toLowerCase()) >= 0;
    }

    initializeData() {
        axios.get('/databases')
            .then(response => {
                if (response.status === 200) {
                    if (response.data !== null) {
                        let data = response.data;
                        data.sort(function (a, b) {
                            let o1 = a.groupId;
                            let o2 = b.groupId;

                            if (o1 < o2) return -1;
                            if (o1 > o2) return 1;

                            let p1 = a.title.toLowerCase();
                            let p2 = b.title.toLowerCase();

                            if (p1 < p2) return -1;
                            if (p1 > p2) return 1;

                            let r1 = a.groupType.toLowerCase();
                            let r2 = b.groupType.toLowerCase();

                            if (r1 < r2) return -1;
                            if (r1 > r2) return 1;

                            return 0;
                        });

                        let typeMatches = false;
                        let groupType = this.state.groupType;
                        const groupTypes = new Set();
                        data.forEach(function (item) {
                            groupTypes.add(item.groupType);
                            if (groupType === item.groupType) {
                                typeMatches = true;
                            }
                        });

                        if (!typeMatches) {
                            if (groupTypes.size > 0) {
                                groupType = groupTypes.values().next().value;
                            } else {
                                groupType = "";
                            }
                        }

                        this.setState({
                            databases: data,
                            groupTypes: Array.from(groupTypes),
                            groupType: groupType
                        });
                    }
                } else {
                    console.error("request error: do something", response);
                }
            })
    }

    renderDatabase(item, {handleClick, modifiers, query}) {
        if (!modifiers.matchesPredicate) {
            return null;
        }
        const text = `${item.groupId}. ${item.title}`;
        return (
            <MenuItem
                active={modifiers.active}
                disabled={modifiers.disabled || this.state.groupType !== item.groupType}
                label={item.groupType}
                onClick={handleClick}
                key={`${item.groupId}.${item.groupType}`}
                text={highlightText(text, query)}
            />
        );
    }

    handleExecuteClick() {
        this.setState({executingQuery: true, errors: []});

        let selection = this.codeMirror.editor.getSelection().trim();
        let singleMode = this.state.queryMode === "single";
        let reqParams = {
            groupId: (singleMode ? this.state.database.groupId : null),
            groupType: this.state.groupType,
            query: (selection.length === 0 ? this.state.query : selection)
        };

        this.queryHistory.addQuery(new Date(), this.state.queryMode, this.state.groupType, this.state.database,
            reqParams.query);

        axios.get('/query', {params: reqParams})
            .then(response => {
                    if (response.status === 200) {
                        let data = {columns: [], rows: []};
                        let errors = [];

                        if (singleMode) {
                            if (response.data.error !== null) {
                                let err = response.data.error;
                                err.groupId = reqParams.groupId;
                                errors.push(err);
                            }
                            if (response.data.data !== null
                                && response.data.data.rows !== null
                                && response.data.data.columns !== null) {
                                data = response.data.data;
                            }
                        } else {
                            response.data.forEach((element) => {
                                if (element.error !== null) {
                                    let err = element.error;
                                    err.groupId = element.groupId;
                                    errors.push(err);
                                }
                                if (element.data !== null
                                    && element.data.columns !== null
                                    && element.data.rows !== null) {
                                    element.data.columns.unshift("groupId");
                                    if (data.columns.length === 0) {
                                        data.columns = element.data.columns;
                                    } else {
                                        let missingColumns = element.data.columns.filter(
                                            item => !data.columns.includes(item)
                                        );
                                        if (missingColumns.length > 0) {
                                            let err = {
                                                groupId: element.groupId,
                                                message: "contains new columns",
                                                err: missingColumns
                                            };
                                            errors.push(err);
                                        }
                                        data.columns = data.columns.concat(missingColumns);
                                    }

                                    for (let i = 0; i < element.data.rows.length; i++) {
                                        element.data.rows[i].groupId = element.groupId;
                                    }
                                    data.rows.push(...element.data.rows);
                                }
                            });
                        }
                        this.setState({data: data, errors: errors, executingQuery: false});
                    } else {
                        console.error("failed to execute query", response);
                        this.setState({executingQuery: false});
                        alert("failed to execute query");
                    }
                },
                error => {
                    console.error("failed to execute query", error);
                    this.setState({executingQuery: false});
                    alert("failed to execute query");
                })
    }

    handleErrorClick() {
        this.errorDialog.setState({isOpen: true});
    }

    handleQueryTextChange(editor, data, value) {
        this.setState({query: value})
    }

    handleQueryModeChange(e) {
        this.setState({queryMode: e.target.value});
    }

    handleGroupTypeChange(event) {
        let groupType = event.currentTarget.value;
        let selectedDatabase = this.state.database;
        if (selectedDatabase != null) {
            for (const item of this.state.databases) {
                if (item.groupId === selectedDatabase.groupId && item.groupType === groupType) {
                    selectedDatabase = item;
                    break;
                }
            }
            if (groupType !== selectedDatabase.groupType) {
                selectedDatabase = null;
            }
        }

        this.setState({
            groupType: event.currentTarget.value,
            database: selectedDatabase
        });
    }

    handleDatabaseSelect(value) {
        this.setState({database: value});
    }

    handleQuerySelect(value) {
        this.setState({
            queryMode: value.queryMode,
            groupType: value.groupType,
            database: value.database,
            query: value.query
        });
    }

    render() {
        const {
            query, queryMode, groupType, groupTypes, databases, database,
            executingQuery, errors, data
        } = this.state;
        const containsError = errors !== null && errors.length > 0;
        const queryExecutionDisabled = executingQuery
            || (queryMode === 'single' && database === null)
            || query.length === 0;

        return (
            <div style={{height: '100vh'}} className="flex-container c-children-spacing">
                <div className="query-editor-container">
                    <QueryEditor
                        value={query}
                        onBeforeChange={this.handleQueryTextChange}
                        forwardRef={this.refHandlers.codeMirror}
                    />

                    <div className="query-control-elements">
                        <QueryModeSelect
                            value={queryMode}
                            onChange={this.handleQueryModeChange}
                        />

                        <GroupTypeSelect
                            value={groupType}
                            onChange={this.handleGroupTypeChange}
                            options={groupTypes}
                        />

                        <DatabaseSelect
                            items={databases}
                            itemRenderer={this.renderDatabase}
                            onItemSelect={this.handleDatabaseSelect}
                            disabled={queryMode === "multiple"}
                            database={database}
                        />

                        <Button
                            className="query-control-elements-right"
                            disabled={queryExecutionDisabled}
                            onClick={this.handleExecuteClick}
                            icon="play"
                            text="Execute"
                        />

                        <History
                            className="query-control-elements-right"
                            onItemSelect={this.handleQuerySelect}
                            ref={this.refHandlers.queryHistory}
                        />

                        {executingQuery && <Spinner
                            className={['query-control-elements-right', 'query-executing-spinner']}
                            size={Spinner.SIZE_SMALL}
                        />}
                        {containsError && <Button
                            className="query-control-elements-right"
                            onClick={this.handleErrorClick}
                            icon="error"
                            intent="danger"
                            text="Errors"
                        />}
                        {containsError && <ErrorDialog
                            errors={errors}
                            key="QueryErrorDialog"
                            ref={this.refHandlers.errorDialog}
                        />}
                    </div>
                </div>

                <DataTable data={data}/>
            </div>
        );
    }
}

const QueryEditor = React.memo((props) => {
    return <CodeMirror
        {...props}
        className="query-editor"
        options={{
            mode: 'text/x-sql',
            theme: 'eclipse',
            lineNumbers: true,
            scrollbarStyle: 'native'
        }}
        ref={props.forwardRef}
    />;
});

const QueryModeSelect = React.memo((props) => {
    return <FormGroup inline label="Query Mode">
        <HTMLSelect
            {...props}
            options={queryModes}
        />
    </FormGroup>;
});

const GroupTypeSelect = React.memo((props) => {
    return <FormGroup inline label="Group Type">
        <HTMLSelect
            {...props}
            disabled={props.options === []}
        />
    </FormGroup>;
});

const DatabaseSelect = React.memo((props) => {
    return <FormGroup inline label="Database">
        <Select
            {...props}
            itemPredicate={QueryPanel.selectItemPredicate}
            noResults={<MenuItem disabled text="No results."/>}
            popoverProps={{minimal: true}}
        >
            <Button
                icon="database"
                text={props.database !== null
                    ? props.database.groupId + ". " + props.database.title
                    : '-'}
                rightIcon="double-caret-vertical"
                disabled={props.disabled}
            />
        </Select>
    </FormGroup>;
});
