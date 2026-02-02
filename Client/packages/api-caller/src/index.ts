import axios, { AxiosRequestConfig } from "axios";
import { config as _config } from 'dotenv'
import {
  type ApiResponse,
} from "@workspace/types/response";
import { store } from "@workspace/store";
import { GetUserMetaDataFromLocolStorage } from "@workspace/store/slices/userSlice";




const BASE_URL = process.env.NEXT_PUBLIC_BASE_API_SERVER_URL;

const INVALID_ERROR: ApiResponse<string> = {
  status: 500,
  heading: "Something went wrong",
  message: "Unexpected error occurred",
  data: "",
};

export enum RequestType {
  POST = "POST",
  GET = "GET",
  PUT = "PUT",
  PATCH = "PATCH",
  DELETE = "DELETE",
}

type QueryParams = Record<string, string | number | boolean | undefined>;

interface ApiCallerParameters<TBody> {
  requestType: RequestType;
  paths?: string[];
  body?: TBody;
  queryParams?: QueryParams;
  retry?: boolean;
  _isRetry?: boolean;
}


export type ApiResult<T> =
  | { ok: true; response: ApiResponse<T> }
  | { ok: false; response: ApiResponse<string> };


let isRefreshing = false;
let refreshSubscribers: ((tokenSuccess: boolean) => void)[] = [];

const subscribeToRefresh = (cb: (tokenSuccess: boolean) => void) => {
  refreshSubscribers.push(cb);
};

const onRefreshed = (tokenSuccess: boolean) => {
  refreshSubscribers.forEach((cb) => cb(tokenSuccess));
  refreshSubscribers = [];
};

async function ApiCaller<TBody, TResp>({
  requestType,
  paths = [],
  body,
  queryParams,
  retry = true,
  _isRetry = false
}: ApiCallerParameters<TBody>): Promise<ApiResult<TResp>> {
  if (!BASE_URL) {
    throw Error("PLEASE PROVIDE BASE API SERVER URL ");
  }

  const url = new URL(BASE_URL);
  const userDetails = store.getState().user.userDetails;


  if (paths.length > 0) {
    url.pathname = `/${paths.map(p => encodeURIComponent(p)).join("/")}`;
  }

  if (queryParams) {
    Object.entries(queryParams).forEach(([key, value]) => {
      if (value !== undefined) {
        url.searchParams.set(key, String(value));
      }
    });
  }

  const config: AxiosRequestConfig = {
    method: requestType,
    url: url.toString(),
    data: body,
    withCredentials: true,
  };

  try {
    const res = await axios<ApiResponse<TResp>>(config);
    return { ok: true, response: res.data };
  } catch (err) {

    if (axios.isAxiosError(err)) {
      const status = err.response?.status;
      if (status == 401 && retry && !_isRetry) {

        if (isRefreshing) {
          console.log("Debug :: Refresh already in progress. Queueing request...", { url: config.url });
          return new Promise((resolve) => {
            subscribeToRefresh((tokenSuccess) => {
              if (tokenSuccess) {
                console.log("Debug :: Processing queued request...", { url: config.url });
                resolve(ApiCaller({
                  requestType,
                  paths,
                  body,
                  queryParams,
                  retry: true,
                  _isRetry: true
                }));
              } else {
                console.log("Debug :: Refresh failed for queued request.", { url: config.url });
                resolve({ ok: false, response: err.response?.data });
              }
            });
          });
        }

        isRefreshing = true;
        console.log("Debug :: 401 error caught. Initiating refresh...", { url: config.url });

        let userId = userDetails?.id;

        if (!userId && typeof window !== 'undefined') {
          const data = GetUserMetaDataFromLocolStorage();
          userId = data?.id;
        }

        if (userId) {
          console.log("Debug :: Found userId for refresh:", userId);
          try {
            const res = await ApiCaller({
              requestType: RequestType.POST,
              paths: ["api", "rivon", "auth", "credentials", "refresh"],
              body: { id: userId },
              retry: false
            });

            if (!res.ok) {
              console.error("Debug :: Refresh failed:", res.response);
              isRefreshing = false;
              onRefreshed(false);
              return { ok: false, response: err.response?.data };
            }

            console.log("Debug :: Refresh successful. Waiting for cookie propagation...");
            // Increase delay to ensure cookie is set
            await new Promise(r => setTimeout(r, 500));

            isRefreshing = false;
            onRefreshed(true);

            console.log("Debug :: Retrying original request...");
            return ApiCaller({
              requestType,
              paths,
              body,
              queryParams,
              retry: true,
              _isRetry: true
            });
          } catch (refreshErr) {
            console.error("Debug :: Exception during refresh:", refreshErr);
            isRefreshing = false;
            onRefreshed(false);
            return { ok: false, response: err.response?.data };
          }
        } else {
          console.warn("Debug :: No userId found for refresh.");
          isRefreshing = false;
          onRefreshed(false);
        }
      }
      return { ok: false, response: err.response?.data };

    } else {
      return { ok: false, response: INVALID_ERROR };
    }

  }
}

export const getOAuthUrl = (provider: string) => {
  return `${BASE_URL}/auth/${provider}`;
}

export default ApiCaller;
