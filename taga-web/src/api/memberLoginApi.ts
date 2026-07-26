import axios from "axios";
import { API_BASE_URL } from "../config/api";

const API = axios.create({
  baseURL: API_BASE_URL,
});

// ✅ LOGIN
export const memberLogin = async (data: {
  email: string;
  password: string;
}) => {
  const res = await API.post("/auth/login", data);
  return res.data;
};

// ✅ FORGOT PASSWORD
export const forgotPassword = async (data: {
  membershipId: string;
  email: string;
  securityQuestion: string;
  securityAnswer: string;
}) => {
  const res = await API.post("/auth/member-forgot-password", data);
  return res.data;
};

// ✅ APPLY MEMBERSHIP
export const applyMembership = async (data: any) => {
  const res = await API.post("/membership/apply", data);
  return res.data;
};
export const getDistricts = async (): Promise<string[]> => {
    console.log("Calling districts API..."); // 👈 ADD

  const response = await API.get('/membership/districts');
  console.log("API URL:", response.config.url); // 👈 ADD
   return response.data.data || [];
};
export const changePassword = async (data: {
  email: string;
  oldPassword: string;
  newPassword: string;
  confirmPassword: string;
}) => {
  const res = await API.post("/member/change-password", data);
  return res.data;
};