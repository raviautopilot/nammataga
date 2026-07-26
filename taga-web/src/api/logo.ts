import axios from "axios";
import { API_BASE_URL } from "../config/api";

const API = axios.create({
  baseURL: API_BASE_URL,
});

export const getLogo = async () => {
  const res = await API.get("/logo"); 
  return res.data; // { url: "https://..." }
};