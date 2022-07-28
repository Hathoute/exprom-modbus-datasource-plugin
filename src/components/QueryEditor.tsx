import {defaults} from 'lodash';

import React, {ChangeEvent, PureComponent, SyntheticEvent} from 'react';
import {LegacyForms} from '@grafana/ui';
import {QueryEditorProps, SelectableValue} from '@grafana/data';
import {DataSource} from '../datasource';
import {defaultQuery, MyDataSourceOptions, MyQuery} from '../types';

const { Switch } = LegacyForms;

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  onEntityChange = (value: SelectableValue<string>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, entity: value.value! });
  };

  onParameterValueChange = (event: ChangeEvent<HTMLInputElement>, key: string) => {
    const { onChange, query } = this.props;
    query.entity = "MetricsData"
    onChange({
      ...query,
      parameters: {
        ...query.parameters,
        [key]: event.target.value
      }
    });
  }

  onWithStreamingChange = (event: SyntheticEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, withStreaming: event.currentTarget.checked });
    // executes the query
    onRunQuery();
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { parameters, withStreaming } = query;

    return (
      <div className="gf-form">
        <span className="gf-form-label width-10">Select data from metric(s)</span>
        <input
            name="metrics"
            className="gf-form-input"
            onChange={e => this.onParameterValueChange(e, "metrics")}
            value={parameters["metrics"] ?? ""}
        />
        <Switch checked={withStreaming || false} label="Enable streaming (v8+)" onChange={this.onWithStreamingChange} />
      </div>
    );
  }
}
