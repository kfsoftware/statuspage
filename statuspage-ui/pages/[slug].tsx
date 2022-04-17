import * as reactSolidIcons from "@heroicons/react/solid";
import {
  CheckCircleIcon,
  ChevronDownIcon,
  SearchIcon,
  SortAscendingIcon,
} from "@heroicons/react/solid";
import type { GetStaticPropsContext, NextPage } from "next";
import Head from "next/head";
import { useMemo, useState } from "react";
import client from "../apollo/client";
import {
  StatusPageDocument,
  StatusPageQuery,
  StatusPageQueryVariables,
  StatusPagesQuery,
  StatusPagesDocument,
  StatusPagesQueryVariables,
} from "../operations";

const Home: NextPage<{ statusPage: StatusPage }> = ({ statusPage }) => {
  const [search, setSearch] = useState("");
  const services = useMemo(
    () =>
      search
        ? statusPage.services!.filter((i) =>
            i.name.toLowerCase().includes(search.toLowerCase())
          )
        : statusPage.services!,
    [statusPage, search]
  );
  return (
    <>
      <Head>
        <title>{statusPage.title}</title>
      </Head>
      <div className="max-w-4xl mx-auto mt-6 px-6">
        <div className="md:flex md:items-center md:justify-between">
          <div className="flex-1 min-w-0">
            <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
              {statusPage.title}
            </h2>
          </div>
        </div>
        <div className="rounded-md bg-green-50 p-4 mt-8">
          <div className="flex">
            <div className="flex-shrink-0">
              <CheckCircleIcon
                className="h-5 w-5 text-green-400"
                aria-hidden="true"
              />
            </div>
            <div className="ml-3">
              <p className="text-sm font-medium text-green-800">
                {statusPage.status.message}
              </p>
            </div>
          </div>
        </div>

        <div className="pb-5 border-b border-gray-200 sm:flex sm:items-center sm:justify-between mt-8">
          <h3 className="text-lg leading-6 font-medium text-gray-900">
            Services
          </h3>
          <div className="mt-3 sm:mt-0 sm:ml-4">
            <label htmlFor="mobile-search-service" className="sr-only">
              Search
            </label>
            <label htmlFor="desktop-search-service" className="sr-only">
              Search
            </label>
            <div className="flex rounded-md shadow-sm">
              <div className="relative flex-grow focus-within:z-10">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <SearchIcon
                    className="h-5 w-5 text-gray-400"
                    aria-hidden="true"
                  />
                </div>
                <input
                  type="text"
                  name="mobile-search-service"
                  id="mobile-search-service"
                  className="focus:ring-indigo-500 focus:border-indigo-500 block w-full rounded-r-md rounded-l-md pl-10 sm:hidden border-gray-300"
                  placeholder="Search"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                />
                <input
                  type="text"
                  name="desktop-search-service"
                  id="desktop-search-service"
                  className="hidden focus:ring-indigo-500 focus:border-indigo-500 w-full rounded-r-md rounded-l-md pl-10 sm:block sm:text-sm border-gray-300"
                  placeholder="Search services"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                />
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white shadow overflow-hidden sm:rounded-md">
          <ul role="list" className="divide-y divide-gray-200">
            {services.map((service, idx) => (
              <li key={idx}>
                <a href="#" className="block hover:bg-gray-50">
                  <div className="px-4 py-4 sm:px-6">
                    <div className="flex items-center justify-between">
                      <p className="text-sm font-medium text-indigo-600 truncate">
                        {service.name}
                      </p>
                      <div className="ml-2 flex-shrink-0 flex">
                        <p className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
                          {service.uptime}
                        </p>
                      </div>
                    </div>
                    <div className="mt-2 sm:flex sm:justify-between">
                      <div className="sm:flex">
                        {service.tags.map((tag, idx) => {
                          const Icon = reactSolidIcons[tag.type];
                          return (
                            <p
                              key={idx}
                              className={
                                idx === 0
                                  ? "flex items-center text-sm text-gray-500"
                                  : "mt-2 flex items-center text-sm text-gray-500 sm:mt-0 sm:ml-6"
                              }
                            >
                              <Icon
                                className="flex-shrink-0 mr-1.5 h-5 w-5 text-gray-400"
                                aria-hidden="true"
                              />
                              {tag.value}
                            </p>
                          );
                        })}
                      </div>
                      <div className="mt-2 flex items-center text-sm text-gray-500 sm:mt-0">
                        {service.latestChecks.map((check, idx) => (
                          <span
                            title={check.time}
                            key={idx}
                            className="h-4 w-1 m-0.5 bg-green-400 rounded hover:scale-150"
                          ></span>
                        ))}
                      </div>
                    </div>
                  </div>
                </a>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </>
  );
};
enum StatusCode {
  UP,
  DOWN,
}
interface StatusPage {
  title: string;
  status: {
    message: string;
    emoji: string;
    code: StatusCode;
  };
  services: {
    name: string;
    uptime: string;
    tags: {
      type: ReactType;
      value: string;
    }[];
    latestChecks: {
      time: string;
      status: "UP" | "DOWN";
    }[];
  }[];
}
type ReactType =
  | "AcademicCapIcon"
  | "AdjustmentsIcon"
  | "AnnotationIcon"
  | "ArchiveIcon"
  | "ArrowCircleDownIcon"
  | "ArrowCircleLeftIcon"
  | "ArrowCircleRightIcon"
  | "ArrowCircleUpIcon"
  | "ArrowDownIcon"
  | "ArrowLeftIcon"
  | "ArrowNarrowDownIcon"
  | "ArrowNarrowLeftIcon"
  | "ArrowNarrowRightIcon"
  | "ArrowNarrowUpIcon"
  | "ArrowRightIcon"
  | "ArrowSmDownIcon"
  | "ArrowSmLeftIcon"
  | "ArrowSmRightIcon"
  | "ArrowSmUpIcon"
  | "ArrowUpIcon"
  | "ArrowsExpandIcon"
  | "AtSymbolIcon"
  | "BackspaceIcon"
  | "BadgeCheckIcon"
  | "BanIcon"
  | "BeakerIcon"
  | "BellIcon"
  | "BookOpenIcon"
  | "BookmarkAltIcon"
  | "BookmarkIcon"
  | "BriefcaseIcon"
  | "CakeIcon"
  | "CalculatorIcon"
  | "CalendarIcon"
  | "CameraIcon"
  | "CashIcon"
  | "ChartBarIcon"
  | "ChartPieIcon"
  | "ChartSquareBarIcon"
  | "ChatAlt2Icon"
  | "ChatAltIcon"
  | "ChatIcon"
  | "CheckCircleIcon"
  | "CheckIcon"
  | "ChevronDoubleDownIcon"
  | "ChevronDoubleLeftIcon"
  | "ChevronDoubleRightIcon"
  | "ChevronDoubleUpIcon"
  | "ChevronDownIcon"
  | "ChevronLeftIcon"
  | "ChevronRightIcon"
  | "ChevronUpIcon"
  | "ChipIcon"
  | "ClipboardCheckIcon"
  | "ClipboardCopyIcon"
  | "ClipboardListIcon"
  | "ClipboardIcon"
  | "ClockIcon"
  | "CloudDownloadIcon"
  | "CloudUploadIcon"
  | "CloudIcon"
  | "CodeIcon"
  | "CogIcon"
  | "CollectionIcon"
  | "ColorSwatchIcon"
  | "CreditCardIcon"
  | "CubeTransparentIcon"
  | "CubeIcon"
  | "CurrencyBangladeshiIcon"
  | "CurrencyDollarIcon"
  | "CurrencyEuroIcon"
  | "CurrencyPoundIcon"
  | "CurrencyRupeeIcon"
  | "CurrencyYenIcon"
  | "CursorClickIcon"
  | "DatabaseIcon"
  | "DesktopComputerIcon"
  | "DeviceMobileIcon"
  | "DeviceTabletIcon"
  | "DocumentAddIcon"
  | "DocumentDownloadIcon"
  | "DocumentDuplicateIcon"
  | "DocumentRemoveIcon"
  | "DocumentReportIcon"
  | "DocumentSearchIcon"
  | "DocumentTextIcon"
  | "DocumentIcon"
  | "DotsCircleHorizontalIcon"
  | "DotsHorizontalIcon"
  | "DotsVerticalIcon"
  | "DownloadIcon"
  | "DuplicateIcon"
  | "EmojiHappyIcon"
  | "EmojiSadIcon"
  | "ExclamationCircleIcon"
  | "ExclamationIcon"
  | "ExternalLinkIcon"
  | "EyeOffIcon"
  | "EyeIcon"
  | "FastForwardIcon"
  | "FilmIcon"
  | "FilterIcon"
  | "FingerPrintIcon"
  | "FireIcon"
  | "FlagIcon"
  | "FolderAddIcon"
  | "FolderDownloadIcon"
  | "FolderOpenIcon"
  | "FolderRemoveIcon"
  | "FolderIcon"
  | "GiftIcon"
  | "GlobeAltIcon"
  | "GlobeIcon"
  | "HandIcon"
  | "HashtagIcon"
  | "HeartIcon"
  | "HomeIcon"
  | "IdentificationIcon"
  | "InboxInIcon"
  | "InboxIcon"
  | "InformationCircleIcon"
  | "KeyIcon"
  | "LibraryIcon"
  | "LightBulbIcon"
  | "LightningBoltIcon"
  | "LinkIcon"
  | "LocationMarkerIcon"
  | "LockClosedIcon"
  | "LockOpenIcon"
  | "LoginIcon"
  | "LogoutIcon"
  | "MailOpenIcon"
  | "MailIcon"
  | "MapIcon"
  | "MenuAlt1Icon"
  | "MenuAlt2Icon"
  | "MenuAlt3Icon"
  | "MenuAlt4Icon"
  | "MenuIcon"
  | "MicrophoneIcon"
  | "MinusCircleIcon"
  | "MinusSmIcon"
  | "MinusIcon"
  | "MoonIcon"
  | "MusicNoteIcon"
  | "NewspaperIcon"
  | "OfficeBuildingIcon"
  | "PaperAirplaneIcon"
  | "PaperClipIcon"
  | "PauseIcon"
  | "PencilAltIcon"
  | "PencilIcon"
  | "PhoneIncomingIcon"
  | "PhoneMissedCallIcon"
  | "PhoneOutgoingIcon"
  | "PhoneIcon"
  | "PhotographIcon"
  | "PlayIcon"
  | "PlusCircleIcon"
  | "PlusSmIcon"
  | "PlusIcon"
  | "PresentationChartBarIcon"
  | "PresentationChartLineIcon"
  | "PrinterIcon"
  | "PuzzleIcon"
  | "QrcodeIcon"
  | "QuestionMarkCircleIcon"
  | "ReceiptRefundIcon"
  | "ReceiptTaxIcon"
  | "RefreshIcon"
  | "ReplyIcon"
  | "RewindIcon"
  | "RssIcon"
  | "SaveAsIcon"
  | "SaveIcon"
  | "ScaleIcon"
  | "ScissorsIcon"
  | "SearchCircleIcon"
  | "SearchIcon"
  | "SelectorIcon"
  | "ServerIcon"
  | "ShareIcon"
  | "ShieldCheckIcon"
  | "ShieldExclamationIcon"
  | "ShoppingBagIcon"
  | "ShoppingCartIcon"
  | "SortAscendingIcon"
  | "SortDescendingIcon"
  | "SparklesIcon"
  | "SpeakerphoneIcon"
  | "StarIcon"
  | "StatusOfflineIcon"
  | "StatusOnlineIcon"
  | "StopIcon"
  | "SunIcon"
  | "SupportIcon"
  | "SwitchHorizontalIcon"
  | "SwitchVerticalIcon"
  | "TableIcon"
  | "TagIcon"
  | "TemplateIcon"
  | "TerminalIcon"
  | "ThumbDownIcon"
  | "ThumbUpIcon"
  | "TicketIcon"
  | "TranslateIcon"
  | "TrashIcon"
  | "TrendingDownIcon"
  | "TrendingUpIcon"
  | "TruckIcon"
  | "UploadIcon"
  | "UserAddIcon"
  | "UserCircleIcon"
  | "UserGroupIcon"
  | "UserRemoveIcon"
  | "UserIcon"
  | "UsersIcon"
  | "VariableIcon"
  | "VideoCameraIcon"
  | "ViewBoardsIcon"
  | "ViewGridAddIcon"
  | "ViewGridIcon"
  | "ViewListIcon"
  | "VolumeOffIcon"
  | "VolumeUpIcon"
  | "WifiIcon"
  | "XCircleIcon"
  | "XIcon"
  | "ZoomInIcon"
  | "ZoomOutIcon";

export async function getStaticPaths() {
  const { data } = await client.query<
    StatusPagesQuery,
    StatusPagesQueryVariables
  >({
    query: StatusPagesDocument,
    variables: {},
  });
  const paths = data.statusPages!.map((page) => `/${page.slug}`);
  return { paths, fallback: "blocking" };
}
export async function getStaticProps({
  params,
}: GetStaticPropsContext<{ slug: string }>) {
  const { data } = await client.query<
    StatusPageQuery,
    StatusPageQueryVariables
  >({
    query: StatusPageDocument,
    variables: {
      slug: params?.slug!,
    },
  });
  const statusPageResult = data.statusPage!;
  const status = {
    message: statusPageResult.checks?.every((i) => i.status === "UP")
      ? "All services are up"
      : "Some services are down",
    emoji: "🎉",
    code: StatusCode.UP,
  };
  const statusPage: StatusPage = {
    title: statusPageResult.title,
    services: statusPageResult.checks!.map((check) => {
      return {
        latestChecks: check.latestExecutions!.map((execution) => {
          return {
            time: execution.executionTime!,
            status: execution.status! as "UP" | "DOWN",
          };
        }),
        name: check.name,
        tags: [],
        uptime: `${(check.uptime.uptime24h * 100).toFixed(2)}%`,
      };
    }),
    status: status,
  };
  return {
    props: {
      statusPage,
    },
    revalidate: 10,
  };
}

export default Home;
