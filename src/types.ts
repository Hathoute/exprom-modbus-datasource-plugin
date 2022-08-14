import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface MyQuery extends DataQuery {
  entity: string
  parameters: {[key: string]: string}
  withStreaming: boolean;
}

export const defaultQuery: Partial<MyQuery> = {
  entity: "Devices",
  parameters: {
    devices: "-1",
    metrics: "-1",
  },
  withStreaming: false,
};

/**
 * These are options configured for each DataSource instance.
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  hostname: string;
  user: string;
  database: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  password: string;
}

export interface MyVariableQuery {
  entity: string
  devices?: string
}
