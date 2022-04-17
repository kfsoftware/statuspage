import "../styles/globals.css";
import type { AppProps } from "next/app";
import { AuthorizedApolloProvider } from "../apollo/apollo";

function MyApp({ Component, pageProps }: AppProps) {
  return (
    <AuthorizedApolloProvider url={"http://localhost:8888/graphql"}>
      <Component {...pageProps} />
    </AuthorizedApolloProvider>
  );
}

export default MyApp;
