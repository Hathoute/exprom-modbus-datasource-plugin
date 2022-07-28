import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

const { SecretFormField, FormField } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> {}

interface State {}

interface CfgFormFieldProps {
  label: string,
  field: string,
  value: any
}

export class ConfigEditor extends PureComponent<Props, State> {

  onFieldChange = (event: ChangeEvent<HTMLInputElement>, field: string, isSecret?: boolean) => {
    const { onOptionsChange, options } = this.props;

    const parentFieldName = isSecret ? "secureJsonData" : "jsonData";
    const data = {
      ...options[parentFieldName],
      [field]: event.target.value,
    };
    onOptionsChange({ ...options, [parentFieldName]: data });
  };

  onResetPassword = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        password: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        password: '',
      },
    });
  };

  CfgFormField = (props: CfgFormFieldProps) => (
      <div className="gf-form">
        <FormField
            label={props.label}
            labelWidth={6}
            inputWidth={20}
            onChange={e => this.onFieldChange(e, props.field)}
            value={props.value || ''}
            placeholder="json field returned to frontend"
        />
      </div>
  )

  render() {
    const { options } = this.props;
    const { jsonData, secureJsonFields } = options;
    const secureJsonData = (options.secureJsonData || {}) as MySecureJsonData;

    return (
      <div className="gf-form-group">
        <this.CfgFormField label="Hostname" field="hostname" value={jsonData.hostname}/>
        <this.CfgFormField label="User" field="user" value={jsonData.user}/>
        <div className="gf-form-inline">
          <div className="gf-form">
            <SecretFormField
              isConfigured={(secureJsonFields && secureJsonFields.password) as boolean}
              value={secureJsonData.password || ''}
              label="Password"
              placeholder="MYSQL password"
              labelWidth={6}
              inputWidth={20}
              onReset={this.onResetPassword}
              onChange={e => this.onFieldChange(e, "password", true)}
            />
          </div>
        </div>
        <this.CfgFormField label="Database" field="database" value={jsonData.database}/>
      </div>
    );
  }
}
