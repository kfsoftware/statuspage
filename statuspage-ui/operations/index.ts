import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
const defaultOptions = {} as const;
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: string;
  String: string;
  Boolean: boolean;
  Int: number;
  Float: number;
  Time: any;
};

export type Check = {
  errorMsg: Scalars['String'];
  executions?: Maybe<Array<CheckExecution>>;
  frecuency: Scalars['String'];
  id: Scalars['ID'];
  latestCheck?: Maybe<Scalars['Time']>;
  latestExecutions?: Maybe<Array<CheckExecution>>;
  message: Scalars['String'];
  name: Scalars['String'];
  namespace: Scalars['String'];
  status: Scalars['String'];
  uptime: CheckUptime;
};


export type CheckExecutionsArgs = {
  from?: InputMaybe<Scalars['Time']>;
  until?: InputMaybe<Scalars['Time']>;
};


export type CheckLatestExecutionsArgs = {
  limit: Scalars['Int'];
};

export type CheckExecution = {
  __typename?: 'CheckExecution';
  errorMsg: Scalars['String'];
  executionTime: Scalars['Time'];
  id: Scalars['ID'];
  message: Scalars['String'];
  status: Scalars['String'];
};

export type CheckUptime = {
  __typename?: 'CheckUptime';
  uptime7d: Scalars['Float'];
  uptime24h: Scalars['Float'];
  uptime30d: Scalars['Float'];
};

export type CreateHttpCheckInput = {
  frecuency: Scalars['String'];
  name: Scalars['String'];
  namespace: Scalars['String'];
  statusCode: Scalars['Int'];
  url: Scalars['String'];
};

export type CreateIcmpCheckInput = {
  address: Scalars['String'];
  frecuency: Scalars['String'];
  name: Scalars['String'];
  namespace: Scalars['String'];
};

export type CreateStatusPageInput = {
  checkSlugs: Array<Scalars['String']>;
  name: Scalars['String'];
  namespace: Scalars['String'];
  title: Scalars['String'];
};

export type CreateTcpCheckInput = {
  address: Scalars['String'];
  frecuency: Scalars['String'];
  name: Scalars['String'];
  namespace: Scalars['String'];
};

export type CreateTlsCheckInput = {
  address: Scalars['String'];
  frecuency: Scalars['String'];
  name: Scalars['String'];
  namespace: Scalars['String'];
  rootCAs?: InputMaybe<Scalars['String']>;
};

export type DeleteResponse = {
  __typename?: 'DeleteResponse';
  id: Scalars['ID'];
};

export type HttpCheck = Check & {
  __typename?: 'HttpCheck';
  errorMsg: Scalars['String'];
  executions?: Maybe<Array<CheckExecution>>;
  frecuency: Scalars['String'];
  id: Scalars['ID'];
  latestCheck?: Maybe<Scalars['Time']>;
  latestExecutions?: Maybe<Array<CheckExecution>>;
  message: Scalars['String'];
  name: Scalars['String'];
  namespace: Scalars['String'];
  status: Scalars['String'];
  uptime: CheckUptime;
  url: Scalars['String'];
};


export type HttpCheckExecutionsArgs = {
  from?: InputMaybe<Scalars['Time']>;
  until?: InputMaybe<Scalars['Time']>;
};


export type HttpCheckLatestExecutionsArgs = {
  limit: Scalars['Int'];
};

export type IcmpCheck = Check & {
  __typename?: 'IcmpCheck';
  address: Scalars['String'];
  errorMsg: Scalars['String'];
  executions?: Maybe<Array<CheckExecution>>;
  frecuency: Scalars['String'];
  id: Scalars['ID'];
  latestCheck?: Maybe<Scalars['Time']>;
  latestExecutions?: Maybe<Array<CheckExecution>>;
  message: Scalars['String'];
  name: Scalars['String'];
  namespace: Scalars['String'];
  status: Scalars['String'];
  uptime: CheckUptime;
};


export type IcmpCheckExecutionsArgs = {
  from?: InputMaybe<Scalars['Time']>;
  until?: InputMaybe<Scalars['Time']>;
};


export type IcmpCheckLatestExecutionsArgs = {
  limit: Scalars['Int'];
};

export type Mutation = {
  __typename?: 'Mutation';
  createHttpCheck: Check;
  createIcmpCheck: Check;
  createStatusPage: StatusPage;
  createTcpCheck: Check;
  createTlsCheck: Check;
  deleteCheck: DeleteResponse;
  deleteStatusPage: DeleteResponse;
  poll?: Maybe<PollResult>;
};


export type MutationCreateHttpCheckArgs = {
  input: CreateHttpCheckInput;
};


export type MutationCreateIcmpCheckArgs = {
  input: CreateIcmpCheckInput;
};


export type MutationCreateStatusPageArgs = {
  input: CreateStatusPageInput;
};


export type MutationCreateTcpCheckArgs = {
  input: CreateTcpCheckInput;
};


export type MutationCreateTlsCheckArgs = {
  input: CreateTlsCheckInput;
};


export type MutationDeleteCheckArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
};


export type MutationDeleteStatusPageArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
};

export type Namespace = {
  __typename?: 'Namespace';
  id: Scalars['ID'];
  name: Scalars['String'];
};

export type PollResult = {
  __typename?: 'PollResult';
  took: Scalars['Int'];
};

export type Query = {
  __typename?: 'Query';
  check?: Maybe<Check>;
  checks?: Maybe<Array<Check>>;
  execution?: Maybe<CheckExecution>;
  executions?: Maybe<Array<CheckExecution>>;
  latestExecutions?: Maybe<Array<CheckExecution>>;
  namespaces?: Maybe<Array<Namespace>>;
  statusPage?: Maybe<StatusPage>;
  statusPages?: Maybe<Array<StatusPage>>;
};


export type QueryCheckArgs = {
  checkId: Scalars['ID'];
};


export type QueryChecksArgs = {
  namespace?: InputMaybe<Scalars['String']>;
};


export type QueryExecutionArgs = {
  execId: Scalars['ID'];
};


export type QueryExecutionsArgs = {
  checkId: Scalars['ID'];
  from?: InputMaybe<Scalars['Time']>;
  until?: InputMaybe<Scalars['Time']>;
};


export type QueryLatestExecutionsArgs = {
  checkId: Scalars['ID'];
  limit?: Scalars['Int'];
};


export type QueryStatusPageArgs = {
  slug: Scalars['String'];
};


export type QueryStatusPagesArgs = {
  namespace?: InputMaybe<Scalars['String']>;
};

export type StatusPage = {
  __typename?: 'StatusPage';
  checks?: Maybe<Array<Check>>;
  id: Scalars['ID'];
  name: Scalars['String'];
  namespace: Scalars['String'];
  slug: Scalars['String'];
  title: Scalars['String'];
};

export type TcpCheck = Check & {
  __typename?: 'TcpCheck';
  address: Scalars['String'];
  errorMsg: Scalars['String'];
  executions?: Maybe<Array<CheckExecution>>;
  frecuency: Scalars['String'];
  id: Scalars['ID'];
  latestCheck?: Maybe<Scalars['Time']>;
  latestExecutions?: Maybe<Array<CheckExecution>>;
  message: Scalars['String'];
  name: Scalars['String'];
  namespace: Scalars['String'];
  status: Scalars['String'];
  uptime: CheckUptime;
};


export type TcpCheckExecutionsArgs = {
  from?: InputMaybe<Scalars['Time']>;
  until?: InputMaybe<Scalars['Time']>;
};


export type TcpCheckLatestExecutionsArgs = {
  limit: Scalars['Int'];
};

export type TlsCheck = Check & {
  __typename?: 'TlsCheck';
  address: Scalars['String'];
  errorMsg: Scalars['String'];
  executions?: Maybe<Array<CheckExecution>>;
  frecuency: Scalars['String'];
  id: Scalars['ID'];
  latestCheck?: Maybe<Scalars['Time']>;
  latestExecutions?: Maybe<Array<CheckExecution>>;
  message: Scalars['String'];
  name: Scalars['String'];
  namespace: Scalars['String'];
  status: Scalars['String'];
  uptime: CheckUptime;
};


export type TlsCheckExecutionsArgs = {
  from?: InputMaybe<Scalars['Time']>;
  until?: InputMaybe<Scalars['Time']>;
};


export type TlsCheckLatestExecutionsArgs = {
  limit: Scalars['Int'];
};

export type GetChecksQueryVariables = Exact<{ [key: string]: never; }>;


export type GetChecksQuery = { __typename?: 'Query', checks?: Array<{ __typename?: 'HttpCheck', id: string, name: string, namespace: string, uptime: { __typename?: 'CheckUptime', uptime24h: number, uptime7d: number, uptime30d: number } } | { __typename?: 'IcmpCheck', id: string, name: string, namespace: string, uptime: { __typename?: 'CheckUptime', uptime24h: number, uptime7d: number, uptime30d: number } } | { __typename?: 'TcpCheck', id: string, name: string, namespace: string, uptime: { __typename?: 'CheckUptime', uptime24h: number, uptime7d: number, uptime30d: number } } | { __typename?: 'TlsCheck', id: string, name: string, namespace: string, uptime: { __typename?: 'CheckUptime', uptime24h: number, uptime7d: number, uptime30d: number } }> | null };

export type StatusPageQueryVariables = Exact<{
  slug: Scalars['String'];
}>;


export type StatusPageQuery = { __typename?: 'Query', statusPage?: { __typename?: 'StatusPage', id: string, name: string, namespace: string, title: string, checks?: Array<{ __typename?: 'HttpCheck', id: string, name: string, namespace: string, status: string, uptime: { __typename?: 'CheckUptime', uptime24h: number, uptime7d: number, uptime30d: number }, latestExecutions?: Array<{ __typename?: 'CheckExecution', executionTime: any, status: string }> | null } | { __typename?: 'IcmpCheck', id: string, name: string, namespace: string, status: string, uptime: { __typename?: 'CheckUptime', uptime24h: number, uptime7d: number, uptime30d: number }, latestExecutions?: Array<{ __typename?: 'CheckExecution', executionTime: any, status: string }> | null } | { __typename?: 'TcpCheck', id: string, name: string, namespace: string, status: string, uptime: { __typename?: 'CheckUptime', uptime24h: number, uptime7d: number, uptime30d: number }, latestExecutions?: Array<{ __typename?: 'CheckExecution', executionTime: any, status: string }> | null } | { __typename?: 'TlsCheck', id: string, name: string, namespace: string, status: string, uptime: { __typename?: 'CheckUptime', uptime24h: number, uptime7d: number, uptime30d: number }, latestExecutions?: Array<{ __typename?: 'CheckExecution', executionTime: any, status: string }> | null }> | null } | null };

export type StatusPagesQueryVariables = Exact<{ [key: string]: never; }>;


export type StatusPagesQuery = { __typename?: 'Query', statusPages?: Array<{ __typename?: 'StatusPage', id: string, name: string, namespace: string, title: string, slug: string, checks?: Array<{ __typename?: 'HttpCheck', status: string, name: string, namespace: string, uptime: { __typename?: 'CheckUptime', uptime24h: number, uptime7d: number, uptime30d: number } } | { __typename?: 'IcmpCheck', status: string, name: string, namespace: string, uptime: { __typename?: 'CheckUptime', uptime24h: number, uptime7d: number, uptime30d: number } } | { __typename?: 'TcpCheck', status: string, name: string, namespace: string, uptime: { __typename?: 'CheckUptime', uptime24h: number, uptime7d: number, uptime30d: number } } | { __typename?: 'TlsCheck', status: string, name: string, namespace: string, uptime: { __typename?: 'CheckUptime', uptime24h: number, uptime7d: number, uptime30d: number } }> | null }> | null };


export const GetChecksDocument = gql`
    query GetChecks {
  checks {
    id
    name
    namespace
    uptime {
      uptime24h
      uptime7d
      uptime30d
    }
  }
}
    `;

/**
 * __useGetChecksQuery__
 *
 * To run a query within a React component, call `useGetChecksQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetChecksQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetChecksQuery({
 *   variables: {
 *   },
 * });
 */
export function useGetChecksQuery(baseOptions?: Apollo.QueryHookOptions<GetChecksQuery, GetChecksQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetChecksQuery, GetChecksQueryVariables>(GetChecksDocument, options);
      }
export function useGetChecksLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetChecksQuery, GetChecksQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetChecksQuery, GetChecksQueryVariables>(GetChecksDocument, options);
        }
export type GetChecksQueryHookResult = ReturnType<typeof useGetChecksQuery>;
export type GetChecksLazyQueryHookResult = ReturnType<typeof useGetChecksLazyQuery>;
export type GetChecksQueryResult = Apollo.QueryResult<GetChecksQuery, GetChecksQueryVariables>;
export const StatusPageDocument = gql`
    query statusPage($slug: String!) {
  statusPage(slug: $slug) {
    id
    name
    namespace
    title
    checks {
      id
      name
      namespace
      status
      uptime {
        uptime24h
        uptime7d
        uptime30d
      }
      latestExecutions(limit: 50) {
        executionTime
        status
      }
    }
  }
}
    `;

/**
 * __useStatusPageQuery__
 *
 * To run a query within a React component, call `useStatusPageQuery` and pass it any options that fit your needs.
 * When your component renders, `useStatusPageQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useStatusPageQuery({
 *   variables: {
 *      slug: // value for 'slug'
 *   },
 * });
 */
export function useStatusPageQuery(baseOptions: Apollo.QueryHookOptions<StatusPageQuery, StatusPageQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<StatusPageQuery, StatusPageQueryVariables>(StatusPageDocument, options);
      }
export function useStatusPageLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<StatusPageQuery, StatusPageQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<StatusPageQuery, StatusPageQueryVariables>(StatusPageDocument, options);
        }
export type StatusPageQueryHookResult = ReturnType<typeof useStatusPageQuery>;
export type StatusPageLazyQueryHookResult = ReturnType<typeof useStatusPageLazyQuery>;
export type StatusPageQueryResult = Apollo.QueryResult<StatusPageQuery, StatusPageQueryVariables>;
export const StatusPagesDocument = gql`
    query statusPages {
  statusPages {
    id
    name
    namespace
    title
    slug
    checks {
      status
      name
      namespace
      uptime {
        uptime24h
        uptime7d
        uptime30d
      }
    }
  }
}
    `;

/**
 * __useStatusPagesQuery__
 *
 * To run a query within a React component, call `useStatusPagesQuery` and pass it any options that fit your needs.
 * When your component renders, `useStatusPagesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useStatusPagesQuery({
 *   variables: {
 *   },
 * });
 */
export function useStatusPagesQuery(baseOptions?: Apollo.QueryHookOptions<StatusPagesQuery, StatusPagesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<StatusPagesQuery, StatusPagesQueryVariables>(StatusPagesDocument, options);
      }
export function useStatusPagesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<StatusPagesQuery, StatusPagesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<StatusPagesQuery, StatusPagesQueryVariables>(StatusPagesDocument, options);
        }
export type StatusPagesQueryHookResult = ReturnType<typeof useStatusPagesQuery>;
export type StatusPagesLazyQueryHookResult = ReturnType<typeof useStatusPagesLazyQuery>;
export type StatusPagesQueryResult = Apollo.QueryResult<StatusPagesQuery, StatusPagesQueryVariables>;