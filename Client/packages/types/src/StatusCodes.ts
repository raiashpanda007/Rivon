export enum HttpStatus {
  STATUS_OK = "STATUS_OK",
  PROCESSED_OK = "PROCESSED_OK",
  CONFLICT = "CONFLICT",
  INTERNAL = "INTERNAL",
  UNAUTHORIZED = "UNAUTHORIZED",
  NOT_FOUND = "NOT_FOUND",
  BAD_REQUEST = "BAD_REQUEST",
  FORBIDDEN = "FORBIDDEN",
  UNPROCESSABLE_ENTITY = "UNPROCESSABLE_ENTITY",
}

export type CODES = 200 | 201 | 409 | 500 | 401 | 404 | 400 | 403 | 422;

export const StatusCodes: Record<CODES, HttpStatus> = {
  200: HttpStatus.STATUS_OK,
  201: HttpStatus.PROCESSED_OK,
  400: HttpStatus.BAD_REQUEST,
  401: HttpStatus.UNAUTHORIZED,
  403: HttpStatus.FORBIDDEN,
  404: HttpStatus.NOT_FOUND,
  409: HttpStatus.CONFLICT,
  422: HttpStatus.UNPROCESSABLE_ENTITY,
  500: HttpStatus.INTERNAL,
};



export interface ApiResponse<T> {
  status: number
  message: string
  heading: string
  data: T
}


