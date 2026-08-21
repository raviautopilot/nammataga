import axios from "axios";
import { API_BASE_URL } from "../config/api";

const API = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

// Add token interceptor
API.interceptors.request.use(
  (config) => {
    try {
      const token = localStorage.getItem("member_token");
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    } catch {
      // localStorage may throw in private browsing on some mobile browsers
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor for token expiry
API.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      const message = error.response?.data?.error;
      if (message === "Token has expired. Please login again" || message === "Invalid token") {
        try {
          // Only wipe and redirect if token still exists (prevents double-redirect loops)
          const token = localStorage.getItem("member_token");
          if (token) {
            localStorage.removeItem("member_token");
            localStorage.removeItem("member_token_expiry");
            localStorage.removeItem("user");
            window.location.href = "/member-login";
          }
        } catch {
          // localStorage may throw in private browsing on some mobile browsers
        }
      }
    }
    return Promise.reject(error);
  }
);

export interface Member {
  id?: string;
  tagaId: string;
  username: string;
  name: string;
  initial: string;
  gender: string;
  fatherName: string;
  motherName: string;
  educationalQualification: string;
  designation: string;
  workingDistrict: string;
  nativeDistrict: string;
  recruitmentBatch: string;
  seniorityNumber: string;
  residentialAddress: string;
  permanentAddress: string;
  dateOfBirth: string;
  mobileNumber: string;
  emailId: string;
  tbfNumber: string;
  cpsGpfNumber: string;
  firstLogin: boolean;
}

// Member Login with JWT
export const memberLoginAPI = async (credentials: { email: string; password: string }) => {
  const response = await API.post("/member/login", credentials);
  
  if (response.data.token) {
    localStorage.setItem("member_token", response.data.token);
    localStorage.setItem("member_role", response.data.role);
    localStorage.setItem("user", JSON.stringify(response.data.user));
    
    // Use backend-provided expires_in (already configured server-side, default 1 week)
    const expiresIn = response.data.expires_in;
    if (expiresIn && expiresIn > 0) {
      const expiresAt = Date.now() + (expiresIn * 1000);
      localStorage.setItem("member_token_expiry", expiresAt.toString());
    }
  }
  
  return response.data;
};

// Member Logout
export const memberLogoutAPI = () => {
  try {
    localStorage.removeItem("member_token");
    localStorage.removeItem("member_token_expiry");
    localStorage.removeItem("member_role");
    localStorage.removeItem("user");
  } catch {
    // localStorage may throw in private browsing on some mobile browsers
  }
};

// Get Member Profile (using JWT token)
export const getMemberProfile = async (): Promise<Member> => {
  const response = await API.get("/member/profile");
  
  const data = response.data.user;
  
  return {
    id: data.id,
    username: data.username || data.emailId,
    tagaId: data.taga_id || data.tagaId,
    name: data.name,
    initial: data.initial,
    gender: data.gender,
    fatherName: data.father_name || data.fatherName,
    motherName: data.mother_name || data.motherName,
    educationalQualification: data.educational_qualification || data.educationalQualification,
    designation: data.designation,
    workingDistrict: data.working_district || data.workingDistrict,
    nativeDistrict: data.native_district || data.nativeDistrict,
    recruitmentBatch: data.recruitment_batch || data.recruitmentBatch,
    seniorityNumber: data.seniority_number || data.seniorityNumber,
    residentialAddress: data.residential_address || data.residentialAddress,
    permanentAddress: data.permanent_address || data.permanentAddress,
    dateOfBirth: data.date_of_birth || data.dateOfBirth,
    mobileNumber: data.mobile_number || data.mobileNumber,
    emailId: data.emailId,
    tbfNumber: data.tbf_number || data.tbfNumber,
    cpsGpfNumber: data.cps_gpf_number || data.cpsGpfNumber,
    firstLogin: data.first_login || data.firstLogin,
  };
};

// Check if member is logged in and token not expired
export const isMemberLoggedIn = (): boolean => {
  try {
    const token = localStorage.getItem("member_token");
    const expiry = localStorage.getItem("member_token_expiry");
    
    if (!token) return false;
    if (!expiry) return true;
    
    return Date.now() < parseInt(expiry);
  } catch {
    // localStorage may throw in private browsing on some mobile browsers
    return false;
  }
};

// Get remaining token time in minutes
export const getTokenRemainingMinutes = (): number => {
  try {
    const expiry = localStorage.getItem("member_token_expiry");
    if (!expiry) return 0;
    const remaining = parseInt(expiry) - Date.now();
    return Math.max(0, Math.floor(remaining / 60000));
  } catch {
    return 0;
  }
};

export const createEditRequest = async (data: any) => {
  const response = await API.post("/member/edit-request", data);
  return response.data;
};