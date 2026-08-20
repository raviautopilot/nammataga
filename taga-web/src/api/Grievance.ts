import axios from "axios";
import { API_BASE_URL } from "../config/api";

const BASE = API_BASE_URL;

// Grievance CRUD base
const GRIEVANCE_API = `${BASE}/grievances`;

// Dropdown/meta APIs
const META_API = BASE;

const getAuthHeaders = () => {
  const token = localStorage.getItem('member_token') || localStorage.getItem('admin_token');
  const headers: Record<string, string> = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
};

export const createGrievance = async (data: any) => {
  const response = await axios.post(GRIEVANCE_API, data, {
    headers: getAuthHeaders()
  });
  return response.data;
};

export const getGrievances = async () => {
  const response = await axios.get(GRIEVANCE_API, {
    headers: getAuthHeaders()
  });
  return response.data;
};

export const getGrievanceById = async (id: string) => {
  const response = await axios.get(`${GRIEVANCE_API}/${id}`, {
    headers: getAuthHeaders()
  });
  return response.data;
};

export const updateGrievance = async (id: string, data: any) => {
  const response = await axios.put(`${GRIEVANCE_API}/${id}`, data, {
    headers: getAuthHeaders()
  });
  return response.data;
};

export const deleteGrievance = async (id: string) => {
  const response = await axios.delete(`${GRIEVANCE_API}/${id}`, {
    headers: getAuthHeaders()
  });
  return response.data;
};

// Dropdown metadata
export const getCategories = async () => {
  const res = await axios.get(`${META_API}/categories`, {
    headers: getAuthHeaders()
  });
  return res.data;
};

export const getPriorities = async () => {
  const res = await axios.get(`${META_API}/priorities`, {
    headers: getAuthHeaders()
  });
  return res.data;
};