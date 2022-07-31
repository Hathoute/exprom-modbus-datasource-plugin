import React, {useState, useEffect} from 'react';

import {Select} from '@grafana/ui';
import {SelectableValue} from "@grafana/data";
import {MyVariableQuery} from "../types";

interface VariableQueryProps {
    query: MyVariableQuery;
    onChange: (query: MyVariableQuery, definition: string) => void;
}

export default function VariableQueryEditor ({ onChange, query }: VariableQueryProps) {
    const [state, setState] = useState(query);
    const entities = ["Devices", "Metrics"];

    const saveQuery = () => {
        onChange(state, `${state.entity} (${state.devices ?? ''})`);
    };

    useEffect(() => {
        saveQuery();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [state])

    const toSelectableValue = (entity: string): SelectableValue<string> => {
        return {
            label: entity,
            value: entity
        }
    };

    return (
        <>
            <div className="gf-form">
                <span className="gf-form-label width-10">Select Entity</span>
                <Select
                    options={entities.map(toSelectableValue)}
                    value={state.entity}
                    onChange={e => setState({...state, entity: e.value ?? entities[0]})}
                    allowCustomValue={false}
                    closeMenuOnSelect={true}
                    isClearable={false}
                    isMulti={false}
                />
                { state.entity === "Metrics" && (
                    <>
                        <span className="gf-form-label width-10">From Device(s)</span>
                        <input
                        name="rawQuery"
                        className="gf-form-input"
                        onBlur={saveQuery}
                        onChange={e => setState({...state, devices: e.target.value})}
                        value={state.devices ?? ""}
                        />
                    </>
                )}
            </div>
        </>
    );
};
