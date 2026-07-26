import axios from "axios";
import { API_BASE_URL } from "../config/api";

const BASE = API_BASE_URL;

// Grievance CRUD base
const GRIEVANCE_API = `${BASE}/grievances`;

// Dropdown/meta APIs
const META_API = BASE;

export const createGrievance = async (data: any) => {
  const response = await axios.post(GRIEVANCE_API, data);
  return response.data;
};

export const getGrievances = async () => {
  const response = await axios.get(GRIEVANCE_API);
  return response.data;
};

export const getGrievanceById = async (id: string) => {
  const response = await axios.get(`${GRIEVANCE_API}/${id}`);
  return response.data;
};

export const updateGrievance = async (id: string, data: any) => {
  const response = await axios.put(`${GRIEVANCE_API}/${id}`, data);
  return response.data;
};

export const deleteGrievance = async (id: string) => {
  const response = await axios.delete(`${GRIEVANCE_API}/${id}`);
  return response.data;
};

// ✅ FIXED BELOW
export const getCategories = async () => {
  const res = await axios.get(`${META_API}/categories`);
  return res.data;
};

export const getPriorities = async () => {
  const res = await axios.get(`${META_API}/priorities`);
  return res.data;
};