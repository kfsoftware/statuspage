import fetch from "cross-fetch";
import {
  ApolloClient,
  ApolloProvider,
  createHttpLink,
  InMemoryCache,
} from "@apollo/client";
import { RetryLink } from "@apollo/client/link/retry";
import { setContext } from "@apollo/client/link/context";
import { onError } from "apollo-link-error";
import React, { useMemo } from "react";
import { useState } from "react";
interface AuthorizedApolloProviderProps {
  url: string;
  children: JSX.Element;
}
export const AuthorizedApolloProvider = ({
  children,
  url,
}: AuthorizedApolloProviderProps) => {
  const [errorMsg, setErrorMsg] = useState("");
  const httpLink = createHttpLink({
    uri: url,
    credentials: "same-origin",
    fetch,
  });
  const authLink = useMemo(
    () =>
      setContext(async () => {
        const headers: any = {};
        return {
          headers: headers,
        };
      }),
    []
  );
  const retryLink = useMemo(
    () =>
      new RetryLink({
        delay: {
          initial: 300,
          max: Infinity,
          jitter: true,
        },
        attempts: {
          max: 0,
          retryIf: (error, _operation) => !!error,
        },
      }),
    []
  );

  const errorLink = onError(({ graphQLErrors, networkError, operation }) => {
    if (graphQLErrors) {
      graphQLErrors.map(({ message, path }) =>
        console.error(
          `[GraphQL error]: Message: ${message}, Operation: ${operation.operationName}, Path: ${path}`
        )
      );
    }
    if (networkError) console.error(`[Network error]: ${networkError}`);
  });
  const apolloClient = useMemo(
    () =>
      new ApolloClient({
        link: authLink
          .concat(errorLink as any)
          .concat(retryLink)
          .concat(httpLink),
        cache: new InMemoryCache(),
        connectToDevTools: true,
        defaultOptions: {
          watchQuery: {
            fetchPolicy: "no-cache",
            errorPolicy: "ignore",
          },
          query: {
            fetchPolicy: "no-cache",
            errorPolicy: "all",
          },
        },
      }),
    [authLink, errorLink, httpLink, retryLink]
  );

  return <ApolloProvider client={apolloClient}>{children}</ApolloProvider>;
};
