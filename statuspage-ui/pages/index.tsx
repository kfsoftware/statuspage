import * as reactSolidIcons from "@heroicons/react/solid";
import {
  CheckCircleIcon,
  ChevronDownIcon,
  SearchIcon,
  SortAscendingIcon,
} from "@heroicons/react/solid";
import type { NextPage } from "next";
import Head from "next/head";
import client from "../apollo/client";
import {
  useGetChecksQuery,
  StatusPageDocument,
  StatusPageQuery,
  StatusPageQueryVariables,
  StatusPagesQuery,
  StatusPagesDocument,
  StatusPagesQueryVariables,
} from "../operations";
import Link from "next/link";

const Home: NextPage<{ query: StatusPagesQuery }> = ({ query }) => {
  return (
    <>
      <Head>
        <title>Status pages</title>
      </Head>
      <div className="max-w-4xl mx-auto mt-6 px-6">
        <div className="md:flex md:items-center md:justify-between">
          <div className="flex-1 min-w-0">
            <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
              Status pages
            </h2>
          </div>
        </div>
        <div className="bg-white shadow overflow-hidden sm:rounded-md mt-8">
          <ul role="list" className="divide-y divide-gray-200">
            {query.statusPages!.map((statusPage) => (
              <li key={statusPage.id}>
                <Link href={`/${statusPage.slug}`}>
                  <a className="block hover:bg-gray-50">
                    <div className="px-4 py-4 sm:px-6">
                      <div className="flex items-center justify-between">
                        <p className="text-sm font-medium text-indigo-600 truncate">
                          {statusPage.title}
                        </p>
                        <div className="ml-2 flex-shrink-0 flex">
                          <p className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
                            {statusPage.checks?.every((i) => i.status === "UP")
                              ? "Up"
                              : "Down"}
                          </p>
                        </div>
                      </div>
                    </div>
                  </a>
                </Link>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </>
  );
};

export async function getStaticProps() {
  const { data } = await client.query<
    StatusPagesQuery,
    StatusPagesQueryVariables
  >({
    query: StatusPagesDocument,
    variables: {},
  });
  return {
    props: {
      query: data,
    },
    revalidate: 1,
  };
}

export default Home;
