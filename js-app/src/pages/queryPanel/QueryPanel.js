import React, {Component} from "react";
import {Button, FormGroup, HTMLSelect, MenuItem, Spinner, Switch} from "@blueprintjs/core";
import {Select} from "@blueprintjs/select";
import axios from "axios";

import {Controlled as CodeMirror} from 'react-codemirror2'
import 'codemirror/lib/codemirror.css';
import 'codemirror/theme/eclipse.css';
import 'codemirror/mode/sql/sql';
import 'codemirror/addon/hint/show-hint';
import 'codemirror/addon/hint/sql-hint';
import 'codemirror/addon/hint/show-hint.css';

import DataTable from "./DataTable";
import History from "./History";
import ErrorDialog from "./ErrorDialog";
import {highlightText} from "../../utils/select";
import './QueryPanel.css';

export default class QueryPanel extends Component {

    constructor(props) {
        super(props);

        this.refHandlers = {
            errorDialog: (ref) => this.errorDialog = ref,
            codeMirror: (ref) => this.codeMirror = ref,
            queryHistory: (ref) => this.queryHistory = ref
        };

        this.initDatabases = this.initDatabases.bind(this);
        this.initTablesMetadata = this.initTablesMetadata.bind(this);

        this.handleQueryTextChange = this.handleQueryTextChange.bind(this);
        this.handleExecuteClick = this.handleExecuteClick.bind(this);
        this.handleErrorClick = this.handleErrorClick.bind(this);

        this.handleGroupModeChange = this.handleGroupModeChange.bind(this);
        this.handleGroupTypeChange = this.handleGroupTypeChange.bind(this);

        this.handleDatabaseSelect = this.handleDatabaseSelect.bind(this);
        this.renderDatabase = this.renderDatabase.bind(this);

        this.handleQuerySelect = this.handleQuerySelect.bind(this);

        this.state = {
            groupMode: true,
            groupType: "",
            database: null,
            executingQuery: false,
            query: "select now()",
            data: {columns: [], rows: []},
            errors: [],

            groupTypes: [],
            databases: [],

            sqlTypeMode: null,
            tablesMetadata: null,
        };

        this.initDatabases();
    }

    static selectItemPredicate(query, item) {
        return `${item.groupId}. ${item.title} ${item.groupType}`.toLowerCase().indexOf(query.toLowerCase()) >= 0;
    }

    static getSqlTypeMode(sqlType) {
        switch (sqlType) {
            case 'postgresql':
                return 'text/x-pgsql';
            case 'mysql':
                return 'text/x-mysql';
            default:
                return 'text/x-sql';
        }
    }

    static getSqlType(groupMode, groupType, database, databases) {
        if (groupMode === false) {
            if (database !== null) {
                return database.type;
            }
            return null;
        } else {
            const types = new Set();
            databases.forEach(function (item) {
                if (groupType === item.groupType) {
                    types.add(item.type);
                }
            });
            if (types.size === 1) {
                return types.values().next().value
            }
            return null;
        }
    }

    static getTablesMetadata(groupMode, groupType, database, databases) {
        //single database mode
        if (groupMode === false) {
            if (database === null) {
                return null;
            }
            return database.tablesMetadata;
        }

        //multiple database mode
        function mergeUnique(arr1, arr2) {
            return arr1.concat(arr2.filter(function (item) {
                return arr1.indexOf(item) === -1;
            }));
        }

        let joinedTablesMetadata = {};
        databases.forEach(function (item) {
            if (item.groupType !== groupType) {
                return;
            }
            if (item.tablesMetadata === undefined) {
                return;
            }

            for (const [key, value] of Object.entries(item.tablesMetadata)) {
                if (joinedTablesMetadata[key] === undefined) {
                    joinedTablesMetadata[key] = value;
                } else {
                    joinedTablesMetadata[key] = mergeUnique(joinedTablesMetadata[key], value);
                }
            }
        });
        return joinedTablesMetadata;
    }

    setState(state, callback) {
        if (state.groupMode !== undefined
            || state.groupType !== undefined
            || state.database !== undefined
            || state.databases !== undefined) {
            let oldState = {
                groupMode: this.state.groupMode,
                groupType: this.state.groupType,
                database: this.state.database,
                databases: this.state.databases,
                tablesMetadata: this.state.tablesMetadata,
            }
            let mergedState = {...oldState, ...state}
            state.sqlTypeMode = QueryPanel.getSqlTypeMode(QueryPanel.getSqlType(
                mergedState.groupMode,
                mergedState.groupType,
                mergedState.database,
                mergedState.databases
            ));
            state.tablesMetadata = QueryPanel.getTablesMetadata(
                mergedState.groupMode,
                mergedState.groupType,
                mergedState.database,
                mergedState.databases
            );
        }
        super.setState(state, callback);
    }

    initTablesMetadata() {
        const {groupMode, groupType, database, databases} = this.state;
        if (groupMode === false) {
            if (database == null) {
                return null;
            }
            if (database.tablesMetadata !== undefined) {
                return;
            }

            let tablesMetadata = this.getDatabaseTablesMetadata(database.groupId, database.groupType);
            tablesMetadata.then(data => {
                database.tablesMetadata = data;
                this.setState({
                    database: database
                });
            });
        } else {
            let requests = [];
            databases.forEach((item) => {
                if (groupType !== item.groupType) {
                    return;
                }
                if (item.tablesMetadata !== undefined) {
                    return;
                }

                let tablesMetadataPromise = this.getDatabaseTablesMetadata(item.groupId, item.groupType);
                requests.push(tablesMetadataPromise.then(data => {
                    item.tablesMetadata = data;
                }));
            })

            if (requests.length === 0) {
                return;
            }

            Promise.all(requests).then((values) => {
                this.setState({
                    databases: databases
                });
            });
        }
    }

    getDatabaseTablesMetadata(groupId, groupType) {
        return axios.get('/tables-metadata', {params: {groupId: groupId, groupType: groupType}})
            .then(response => {
                if (response.status === 200) {
                    if (response.data !== null) {
                        return response.data;
                    }
                } else {
                    console.error("request error", response);
                    return null;
                }
            });
    }

    initDatabases() {
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

                        const groupTypes = new Set();
                        data.forEach(function (item) {
                            groupTypes.add(item.groupType);
                        });

                        this.setState({
                            databases: data,
                            groupTypes: Array.from(groupTypes),
                            groupType: data.length > 0 ? data[0].groupType : "",
                            database: data.length > 0 ? data[0] : null
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
                disabled={modifiers.disabled}
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
        let singleMode = this.state.groupMode === false;
        let reqParams = {
            groupId: (singleMode ? this.state.database.groupId : null),
            groupType: this.state.groupType,
            query: (selection.length === 0 ? this.state.query : selection)
        };

        this.queryHistory.addQuery(new Date(), this.state.groupMode, this.state.groupType, this.state.database,
            reqParams.query);

        axios.get('/query', {params: reqParams})
            .then(response => {
                    if (response.status !== 200) {
                        console.error("failed to execute query", response);
                        this.setState({executingQuery: false});
                        alert("failed to execute query");
                        return;
                    }

                    let data = {columns: [], rows: []};
                    let errors = [];
                    let dbColumns = [];

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

                            if (element.data === null
                                || element.data.columns === null
                                || element.data.rows === null) {
                                return;
                            }

                            dbColumns.push({groupId: element.groupId, columns: element.data.columns});

                            //add groupId column and set rows groupId field value
                            element.data.columns.unshift({name: "groupId", fieldName: "groupId"});
                            element.data.rows.forEach(row => {
                                row.groupId = element.groupId;
                            });

                            data.rows.push(...element.data.rows);
                        });
                    }

                    //collect all unique columns
                    let allColumns = [];
                    dbColumns.forEach(e => {
                        e.columns.forEach(e2 => {
                            if (allColumns.findIndex(v => v.fieldName === e2.fieldName) === -1) {
                                allColumns.push(e2);
                            }
                        })
                    });
                    data.columns = allColumns;

                    //find missing columns and add errors
                    dbColumns.forEach(e => {
                        let missingColumns = [];
                        allColumns.forEach(e2 => {
                            if (e.columns.findIndex(v => v.fieldName === e2.fieldName) === -1) {
                                missingColumns.push(e2);
                            }
                        })
                        if (missingColumns.length > 0) {
                            let err = {
                                groupId: e.groupId,
                                message: "missing columns",
                                err: missingColumns
                            };
                            errors.push(err);
                        }
                    });

                    this.setState({data: data, errors: errors, executingQuery: false});
                },
                error => {
                    console.error("failed to execute query", error);
                    this.setState({executingQuery: false});
                    alert("failed to execute query");
                });
    }

    handleErrorClick() {
        this.errorDialog.setState({isOpen: true});
    }

    handleQueryTextChange(editor, data, value) {
        this.setState({query: value})
    }

    handleGroupModeChange(e) {
        this.setState({groupMode: !this.state.groupMode});
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
            database: selectedDatabase,
        });
    }

    handleDatabaseSelect(value) {
        this.setState({
            groupType: value.groupType,
            database: value
        });
    }

    handleQuerySelect(value) {
        this.setState({
            groupMode: value.groupMode,
            groupType: value.groupType,
            database: value.database,
            query: value.query
        });
    }

    render() {
        const {
            query, groupMode, groupType, groupTypes, databases, database,
            executingQuery, errors, data, sqlTypeMode, tablesMetadata
        } = this.state;
        const containsError = errors !== null && errors.length > 0;
        const queryExecutionDisabled = executingQuery
            || (groupMode === false && database === null)
            || query.length === 0;

        return (
            <div style={{height: '100vh'}} className="flex-container c-children-spacing">
                <div className="query-editor-container">
                    <QueryEditor
                        value={query}
                        onBeforeChange={this.handleQueryTextChange}
                        forwardRef={this.refHandlers.codeMirror}
                        sqlTypeMode={sqlTypeMode}
                        tablesMetadata={tablesMetadata}
                        onAutocomplete={this.initTablesMetadata}
                    />

                    <div className="query-control-elements">
                        <GroupModeSelect
                            checked={groupMode}
                            onChange={this.handleGroupModeChange}
                        />

                        {groupMode && <GroupTypeSelect
                            value={groupType}
                            onChange={this.handleGroupTypeChange}
                            options={groupTypes}
                        />}

                        {!groupMode && <DatabaseSelect
                            items={databases}
                            itemRenderer={this.renderDatabase}
                            onItemSelect={this.handleDatabaseSelect}
                            database={database}
                        />}

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
    const {forwardRef, sqlTypeMode, tablesMetadata, onAutocomplete, ...newProps} = props;
    return <CodeMirror
        {...newProps}
        className="query-editor"
        options={{
            mode: sqlTypeMode,
            theme: 'eclipse',
            lineNumbers: true,
            scrollbarStyle: 'native',
            extraKeys: {
                "Ctrl-Space": ((cm, eventObj) => {
                    cm.execCommand('autocomplete');
                    onAutocomplete();
                })
            },
            closeOnUnfocus: false,
            hintOptions: {
                tables: tablesMetadata
            }
        }}
        ref={forwardRef}
    />;
});

const GroupModeSelect = React.memo((props) => {
    return <FormGroup inline label="Group mode">
        <Switch {...props}/>
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
                    ? props.database.groupId + ". " + props.database.title + " │" + props.database.groupType
                    : '-'}
                rightIcon="double-caret-vertical"
                disabled={props.disabled}
            />
        </Select>
    </FormGroup>;
});
