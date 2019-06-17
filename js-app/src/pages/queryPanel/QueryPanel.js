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

        this.state = {
            queryMode: queryModes[0].value,
            groupTypes: [],
            groupType: "",
            databases: [],
            database: null,
            executingQuery: false,
            query: "select now()",
            data: [],
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
                        let data = [];
                        let errors = [];

                        if (singleMode) {
                            if (response.data.error !== null) {
                                let err = response.data.error;
                                err.groupId = reqParams.groupId;
                                errors.push(err);
                            }
                            if (response.data.data !== null) {
                                data = response.data.data;
                            }
                        } else {
                            response.data.forEach((element) => {
                                if (element.error !== null) {
                                    let err = element.error;
                                    err.groupId = element.groupId;
                                    errors.push(err)
                                }
                                if (element.data !== null) {
                                    element.data.forEach((e2) => {
                                        e2.groupId = element.groupId;
                                    });
                                    data.push(...element.data);
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
        const containsError = this.state.errors !== null && this.state.errors.length > 0;
        const queryExecutionDisabled = this.state.executingQuery
            || (this.state.queryMode === 'single' && this.state.database === null)
            || this.state.query.length === 0;

        return (
            <div style={{height: '100vh'}} className="flex-container c-children-spacing">
                <div className="query-editor-container">
                    <CodeMirror
                        className="query-editor"
                        value={this.state.query}
                        options={{
                            mode: 'text/x-sql',
                            theme: 'eclipse',
                            lineNumbers: true,
                            scrollbarStyle: 'native'
                        }}
                        onBeforeChange={this.handleQueryTextChange}
                        ref={this.refHandlers.codeMirror}
                    />

                    <div className="query-control-elements">
                        <FormGroup inline label="Query Mode">
                            <HTMLSelect
                                value={this.state.queryMode}
                                onChange={this.handleQueryModeChange}
                                options={queryModes}
                            />
                        </FormGroup>

                        <FormGroup inline label="Group Type">
                            <HTMLSelect
                                value={this.state.groupType}
                                disabled={this.state.groupTypes === []}
                                onChange={this.handleGroupTypeChange}
                                options={this.state.groupTypes}
                            />
                        </FormGroup>

                        <FormGroup inline label="Database">
                            <Select
                                items={this.state.databases}
                                itemPredicate={QueryPanel.selectItemPredicate}
                                itemRenderer={this.renderDatabase}
                                noResults={<MenuItem disabled text="No results."/>}
                                onItemSelect={this.handleDatabaseSelect}
                                popoverProps={{minimal: true}}
                                disabled={this.state.queryMode === "multiple"}
                            >
                                <Button
                                    icon="database"
                                    text={this.state.database !== null
                                        ? this.state.database.groupId + ". " + this.state.database.title
                                        : '-'}
                                    rightIcon="double-caret-vertical"
                                    disabled={this.state.queryMode === "multiple"}
                                />
                            </Select>
                        </FormGroup>

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

                        {this.state.executingQuery && <Spinner
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
                            errors={this.state.errors}
                            key="QueryErrorDialog"
                            ref={this.refHandlers.errorDialog}
                        />}
                    </div>
                </div>

                <DataTable data={this.state.data}/>
            </div>
        );
    }
}

