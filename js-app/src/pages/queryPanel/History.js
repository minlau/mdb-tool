import React, {Component} from "react";
import {Button, MenuItem} from "@blueprintjs/core";
import {Select} from "@blueprintjs/select";
import {highlightText} from "../../utils/select";

const getQueryLabel = (item) => {
    return `${item.queryMode} ${item.groupType} ${item.database != null
        ? `${item.database.groupId}. ${item.database.title}`
        : ''}`;
};

export default class History extends Component {

    constructor(props) {
        super(props);
        this.state = {
            queries: [] // {"timestamp", "queryMode", "groupType", "database", "query"}
        };

        this.addQuery = this.addQuery.bind(this);
    }

    selectItemPredicate(query, item) {
        return `${item.query} ${getQueryLabel(item)}`.toLowerCase().indexOf(query.toLowerCase()) >= 0;
    }

    selectItemRenderer(item, {handleClick, modifiers, query}) {
        if (!modifiers.matchesPredicate) {
            return null;
        }
        return (
            <MenuItem
                active={modifiers.active}
                disabled={modifiers.disabled}
                label={getQueryLabel(item)}
                onClick={handleClick}
                key={`q-${item.timestamp.getTime()}`}
                text={highlightText(item.query, query)}
            />
        );
    }

    addQuery(timestamp, queryMode, groupType, database, query) {
        let queries = this.state.queries;
        if (queries.length > 0) {
            let lastElement = queries[queries.length - 1];
            if (lastElement.queryMode === queryMode && lastElement.groupType === groupType &&
                lastElement.database === database && lastElement.query === query) {
                return;
            }
        }
        queries.push({timestamp, queryMode, groupType, database, query});
        this.setState({queries: queries});
    }

    render() {
        return (
            <Select
                className="query-control-elements-right"
                items={this.state.queries}
                itemPredicate={this.selectItemPredicate}
                itemRenderer={this.selectItemRenderer}
                noResults={<MenuItem disabled text="No results."/>}
                onItemSelect={this.props.onItemSelect}
                resetOnClose
                popoverProps={{minimal: true}}
            >
                <Button
                    icon="history"
                    text={`${this.state.queries.length} items`}
                />
            </Select>
        );
    }

}