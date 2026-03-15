import type { RequestOptions } from "../interface";
import axios from "axios"

export const defaultRequestOptions: RequestOptions = {
  showErrorMessage: true,
  allowRepeatRequest: false,
  returnRawResponse: false,
  successStatus: [200, 201, 204],
};

const baseURL =
  import.meta.env.MODE === "development"
    ? ""
    : import.meta.env.VITE_API_BASE_URL;

export const axiosInstance = axios.create({
  baseURL: baseURL,
  timeout: 15000,
  headers: {
    "Content-Type": "application/json",
  },
  withCredentials: false,
});
