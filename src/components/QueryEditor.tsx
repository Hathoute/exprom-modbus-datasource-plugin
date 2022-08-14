import {defaults} from 'lodash';

import React, {useEffect, useState} from 'react';
import {ActionMeta, LegacyForms, Select} from '@grafana/ui';
import {MetricFindValue, QueryEditorProps, SelectableValue} from '@grafana/data';
import {DataSource} from '../datasource';
import {defaultQuery, MyDataSourceOptions, MyQuery} from '../types';

const { Switch } = LegacyForms;

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor(props: Props) {

    const onChange = (query: MyQuery) => {
        props.onChange(query);
        props.onRunQuery();
    }

    const createSelect = (label: string, options: SelectableValue[], value: any,
                          onChange: (e: SelectableValue, actionMeta: ActionMeta) => void) => (
        <>
            <span className="gf-form-label width-10">{label}</span>
            <Select
                options={options}
                value={value}
                onChange={onChange}
                allowCustomValue={false}
                closeMenuOnSelect={true}
                isClearable={false}
                isMulti={true}
            />
        </>
    )

    const toSelectableValue = (entity: MetricFindValue): SelectableValue<string> => {
        return {
            label: entity.text,
            value: entity.value as string
        }
    };

    const setParam = (p: string, v: any) => {
        setQuery({
            ...query,
            parameters: {
                ...query.parameters,
                [p]: v
            }
        });
    }

    const onDevicesChange = (e: SelectableValue<string | number>) => {
        let devices = "" + e.value;
        if(e instanceof Array) {
            devices = e.map(x => x.value).join(",")
        }
        setParam("devices", devices);
    }

    const onMetricsChange = (e: SelectableValue<string | number>) => {
        let metrics = "" + e.value;
        if(e instanceof Array) {
            metrics = e.map(x => x.value).join(",")
        }
        setParam("metrics", metrics);
    }

    const [query, setQuery] = useState<MyQuery>(defaults(props.query, defaultQuery));

    const [allDevices, setAllDevices] = useState<MetricFindValue[]>([]);
    const [allMetrics, setAllMetrics] = useState<MetricFindValue[]>([]);

    const selectedDevices = query.parameters["devices"];
    const selectedDevicesArray = selectedDevices.split(",").map(x => parseInt(x, 10) || x);
    const selectedMetrics = query.parameters["metrics"];
    const selectedMetricsArray = selectedMetrics.split(",").map(x => parseInt(x, 10) || x);

    useEffect(() => {
        props.datasource.metricFindQuery({entity: "Devices"}).then(r => setAllDevices(r));
    }, [props.datasource]);

    useEffect(() => {
        props.datasource.metricFindQuery({entity: "Metrics", devices: selectedDevices ?? "-1"})
            .then(r => setAllMetrics(r));
    }, [props.datasource, selectedDevices])

    useEffect(() => {
        if(query.entity !== "MetricsData") {
            setQuery({...query, entity: "MetricsData"});
            return;   // onChange will get executed next time (dependency on query)
        }

        if(query.parameters["filter"] !== "metrics") {
            setParam("filter", "metrics");  // We are displaying metrics, so enforce metrics as filter.
            return;
        }

        onChange(query)

        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [query]);

    return (
        <div>
            <div className="gf-form">
                <span className="gf-form-label width-10">FROM</span>
                {createSelect("Devices",
                    allDevices.map(toSelectableValue),
                    allDevices.filter(x => selectedDevicesArray.includes(x.value!)),
                    onDevicesChange
                )}
            </div>

            <div className="gf-form">
                <span className="gf-form-label width-10">SELECT</span>
                {createSelect("Metrics",
                    allMetrics.map(toSelectableValue),
                    allMetrics.filter(x => selectedMetricsArray.includes(x.value!)),
                    onMetricsChange
                )}
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
