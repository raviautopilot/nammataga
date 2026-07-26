import axios from "axios";
import { API_BASE_URL } from "../config/api";

const API_BASE = API_BASE_URL;

export interface SubscriptionPlan {
  id: string;
  name: string;
  description: string;
  amount?: number;
  frequency: string;
  status: string;
  lastPaidDate?: string;
  nextDueDate?: string;
  allowCustomAmount?: boolean;
  oneTime?: boolean;
  needBased?: boolean;
}

export interface CreateOrderResponse {
  orderId: string;
  key: string;
  amount: number;
  currency: string;
}

// Get all subscription plans
export const getSubscriptions = async (): Promise<SubscriptionPlan[]> => {
  const res = await axios.get(`${API_BASE}/subscriptions`);
  return res.data.data;
};

// Create Razorpay order for subscription
export const createSubscriptionOrder = async (
  subscriptionId: string,
  amount: number,
  email: string,
  notes?: Record<string, string>
): Promise<CreateOrderResponse> => {
  const token = localStorage.getItem("member_token");
  const res = await axios.post(
    `${API_BASE}/subscriptions/create-order`,
    { subscriptionId, amount, email, notes },
    { headers: { Authorization: `Bearer ${token}` } }
  );
  return res.data;
};

// Verify subscription payment
export const verifySubscriptionPayment = async (
  subscriptionId: string,
  orderId: string,
  paymentId: string,
  signature: string,
  email: string,
  amount: number
): Promise<any> => {
  const token = localStorage.getItem("member_token");
  const res = await axios.post(
    `${API_BASE}/subscriptions/verify-payment`,
    {
      subscriptionId,
      razorpay_order_id: orderId,
      razorpay_payment_id: paymentId,
      razorpay_signature: signature,
      email,
      amount,
    },
    { headers: { Authorization: `Bearer ${token}` } }
  );
  return res.data;
};

// Get member's subscription status
export const getMemberSubscriptionStatus = async (email: string): Promise<any> => {
  const token = localStorage.getItem("member_token");
  const res = await axios.get(`${API_BASE}/subscriptions/status?email=${email}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  return res.data;
};