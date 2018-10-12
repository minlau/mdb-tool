import React, {Component} from "react";
import {Button, FormGroup, HTMLSelect, MenuItem, Spinner, TextArea} from "@blueprintjs/core";
import {Select} from "@blueprintjs/select";
import axios from "axios";
import DataTable from "../dataTable/DataTable";
import './queryPanel.css';
import QueryErrorDialog from "../queryErrorDialog/QueryErrorDialog";

const highlightText = (text, query) => {
    let lastIndex = 0;
    const words = query
        .split(/\s+/)
        .filter(word => word.length > 0)
        .map(escapeRegExpChars);
    if (words.length === 0) {
        return [text];
    }
    const regexp = new RegExp(words.join("|"), "gi");
    const tokens = [];
    while (true) {
        const match = regexp.exec(text);
        if (!match) {
            break;
        }
        const length = match[0].length;
        const before = text.slice(lastIndex, regexp.lastIndex - length);
        if (before.length > 0) {
            tokens.push(before);
        }
        lastIndex = regexp.lastIndex;
        tokens.push(<strong key={lastIndex}>{match[0]}</strong>);
    }
    const rest = text.slice(lastIndex);
    if (rest.length > 0) {
        tokens.push(rest);
    }
    return tokens;
};

const escapeRegExpChars = (text) => {
    return text.replace(/([.*+?^=!:${}()|\[\]\/\\])/g, "\\$1");
};

const queryModes = [{label: "Multiple", value: "multiple"}, {label: "Single", value: "single"}];

class QueryPanel extends Component {

    constructor(props) {
        super(props);

        this.errorDialog = null;
        this.refHandlers = {
            errorDialog: (ref) => this.errorDialog = ref
        };

        this.initializeData = this.initializeData.bind(this);
        this.handleQueryTextChange = this.handleQueryTextChange.bind(this);
        this.onQueryClick = this.onQueryClick.bind(this);
        this.onErrorClick = this.onErrorClick.bind(this);

        this.onViewModeChange = this.onViewModeChange.bind(this);
        this.onQueryModeChange = this.onQueryModeChange.bind(this);
        this.onGroupTypeChange = this.onGroupTypeChange.bind(this);

        this.onDatabaseSelect = this.onDatabaseSelect.bind(this);
        this.selectItemRenderer = this.selectItemRenderer.bind(this);

        this.state = {
            viewMode: "joined",
            queryMode: "multiple",
            groupTypes: [],
            groupType: "",
            databases: [],
            database: null,
            querying: false,
            query: "select now()",
            data: [],
            errors: []
        };

        this.initializeData();
    }

    static selectItemPredicate(query, item) {
        return `${item.groupId}. ${item.title.toLowerCase()} ${item.groupType}`.indexOf(query.toLowerCase()) >= 0;
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

    selectItemRenderer(item, {handleClick, modifiers, query}) {
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
                text={highlightText(text, query)}
            />
        );
    }

    onQueryClick() {
        this.setState({querying: true, errors: []});

        let singleMode = this.state.queryMode === "single";
        let reqParams = {
            groupId: (singleMode ? this.state.database.groupId : null),
            groupType: this.state.groupType,
            query: this.state.query
        };

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
                            } else if (response.data.data !== null) {
                                data = response.data.data;
                            }
                        } else {
                            response.data.forEach((element) => {
                                if (element.error !== null) {
                                    let err = element.error;
                                    err.groupId = element.groupId;
                                    errors.push(err)
                                } else if (element.data !== null) {
                                    element.data.forEach((e2) => {
                                        e2.groupId = element.groupId;
                                    });
                                    data.push(...element.data);
                                }
                            });
                        }
                        this.setState({data: data, errors: errors, querying: false});
                    } else {
                        console.error("failed to query", response);
                        this.setState({querying: false});
                        alert("failed to query");
                    }
                },
                error => {
                    console.error("failed to query", error);
                    this.setState({querying: false});
                    alert("failed to query");
                })
    }

    onErrorClick() {
        this.errorDialog.setState({isOpen: true});
    }

    handleQueryTextChange(e) {
        this.setState({query: e.target.value})
    }

    onQueryModeChange(e) {
        this.setState({queryMode: e.target.value});

    }

    onViewModeChange(e) {
        this.setState({viewMode: e.target.value});
    }

    onGroupTypeChange(event) {
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

    onDatabaseSelect(value) {
        this.setState({database: value});
    }

    render() {
        const {query, data, querying} = this.state;

        const containsError = this.state.errors !== null && this.state.errors.length > 0;

        return (
            <div style={{height: '100vh'}} className="flex-container c-children-spacing">
                <div style={{
                    minHeight: '100px', maxHeight: '50%', height: '200px',
                    resize: 'vertical', overflow: 'hidden'
                }}>
                    <TextArea
                        style={{fontFamily: 'monospace', height: 'calc(100% - 30px - 4px)', resize: 'none'}}
                        fill={true}
                        placeholder="query"
                        required={true}
                        onChange={this.handleQueryTextChange}
                        value={query}/>

                    <div className="c-query-elements">
                        <FormGroup inline={true} label="Query Mode">
                            <HTMLSelect value={this.state.queryMode}
                                        onChange={this.onQueryModeChange}
                                        options={queryModes}/>
                        </FormGroup>

                        <FormGroup inline={true} label="Group Type">
                            <HTMLSelect value={this.state.groupType}
                                        disabled={this.state.groupTypes === []}
                                        onChange={this.onGroupTypeChange}
                                        options={this.state.groupTypes}/>
                        </FormGroup>

                        <FormGroup inline={true} label="Database">
                            <Select items={this.state.databases}
                                    itemPredicate={QueryPanel.selectItemPredicate}
                                    itemRenderer={this.selectItemRenderer}
                                    noResults={<MenuItem disabled={true} text="No results."/>}
                                    onItemSelect={this.onDatabaseSelect}
                                    popoverProps={{minimal: true}}
                                    disabled={this.state.queryMode === "multiple"}>
                                <Button
                                    icon={"database"}
                                    text={this.state.database !== null ? this.state.database.groupId + ". " + this.state.database.title : "-"}
                                    rightIcon="double-caret-vertical"
                                    disabled={this.state.queryMode === "multiple"}/>
                            </Select>
                        </FormGroup>
                        <Button style={{float: 'right'}}
                                disabled={this.state.querying || (this.state.queryMode === 'single' && this.state.database === null) || this.state.query.length === 0}
                                onClick={this.onQueryClick}
                                icon="play"
                                text="Query"/>
                        {querying && <Spinner className="c-query-panel-spinner"
                                              size={Spinner.SIZE_SMALL}/>}
                        {containsError && <Button style={{float: 'right'}}
                                                  onClick={this.onErrorClick}
                                                  icon="error"
                                                  intent={"danger"}
                                                  text="Errors"/>}
                        {containsError && <QueryErrorDialog errors={this.state.errors}
                                                            key={1}
                                                            ref={this.refHandlers.errorDialog}/>}
                    </div>
                </div>

                <DataTable data={data}/>

            </div>
        );
    }
}

export default QueryPanel;

