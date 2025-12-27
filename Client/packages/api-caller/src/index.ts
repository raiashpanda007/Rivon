import axios, { AxiosRequestConfig } from "axios";
import dotenv from "dotenv";

import {
  StatusCodes,
  type CODES,
  type ApiResponse,
} from "@workspace/types/response";

dotenv.config();

const BASE_URL = process.env.BASE_API_SERVER_URL ?? "http://localhost:8000";

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
  retry: boolean
}

export type ApiResult<T> =
  | { ok: true; response: ApiResponse<T> }
  | { ok: false; response: ApiResponse<string> };

async function ApiCaller<TBody, TResp>({
  requestType,
  paths = [],
  body,
  queryParams,
  retry = false
}: ApiCallerParameters<TBody>): Promise<ApiResult<TResp>> {
  const url = new URL(BASE_URL);


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
      if (retry) {
        const status = err.response?.status;
        if (status == 401) {
          const res = await ApiCaller({
            requestType,
            paths: ["auth", "credentials", "refresh"],
            body,
            queryParams,
            retry: false
          });
          if (!res.ok) {
            return res;
          }
          return ApiCaller({
            requestType,
            paths,
            body,
            queryParams,
            retry: true
          });

        } else {
          return { ok: false, response: err.response?.data };
        }
      } else {
        return { ok: false, response: err.response?.data };
      }

    } else {
      return { ok: false, response: INVALID_ERROR };
    }

  }
}
export default ApiCaller;
