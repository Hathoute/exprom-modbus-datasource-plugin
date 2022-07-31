import {defaults} from 'lodash';

import React, {useEffect, useState} from 'react';
import {LegacyForms, Select} from '@grafana/ui';
import {QueryEditorProps, SelectableValue} from '@grafana/data';
import {DataSource} from '../datasource';
import {defaultQuery, MyDataSourceOptions, MyQuery} from '../types';

const { Switch } = LegacyForms;

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor(props: Props) {

  const onChange = (query: MyQuery) => {
      props.onChange(query);
      props.onRunQuery();
  }

  const createInput = (label: string, name: string, value: any,
                 onChange: (e: React.ChangeEvent<HTMLInputElement>) => void) => (
      <>
        <span className="gf-form-label width-10">{label}</span>
        <input
            name={name}
            className="gf-form-input"
            onChange={onChange}
            onBlur={props.onRunQuery}
            value={value}
        />
      </>
  )

  const createSelect = (label: string, options: SelectableValue[], value: any,
                  onChange: (e: SelectableValue) => void) => (
    <>
      <span className="gf-form-label width-10">{label}</span>
      <Select
          options={options}
          value={value}
          onChange={onChange}
          allowCustomValue={false}
          closeMenuOnSelect={true}
          isClearable={false}
          isMulti={false}
      />
    </>
  )

  const toSelectableValue = (entity: string): SelectableValue<string> => {
    return {
      label: entity,
      value: entity
    }
  };

  const [query, setQuery] = useState<MyQuery>(defaults(props.query, defaultQuery));

  useEffect(() => {
      if(query.entity !== "MetricsData") {
          setQuery({...query, entity: "MetricsData"})
          return;   // onChange will get executed next time (dependency on query)
      }
      onChange(query)

      // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query]);

  const setParam = (p: string, v: any) => {
    setQuery({
      ...query,
      parameters: {
        ...query.parameters,
        [p]: v
      }
    });
  }

  return (
      <div>
        <div className="gf-form">
          <span className="gf-form-label width-10">SELECT</span>
          <span className="gf-form-label width-10">MetricsData</span>
        </div>

        <div className="gf-form">
          {createSelect("WHERE",
              ["devices", "metrics"].map(toSelectableValue),
              query.parameters["filter"],
              e => setParam("filter", e.value)
          )}
          {query.parameters.filter && createInput("IN",
              "filter",
              query.parameters[query.parameters.filter],
              e => setParam(query.parameters.filter, e.target.value)
          )
          }
        </div>

        <div className="gf-form">
          <Switch checked={query.withStreaming}
                  label="Enable streaming (v8+)"
                  onChange={e => setQuery({
                      ...query,
                      withStreaming: e.currentTarget.checked
                  })} />
        </div>
      </div>
    );
}
