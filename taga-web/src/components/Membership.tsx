import React, { useEffect, useState } from 'react';
import API_BASE_URL from '../config/api';
import axios from 'axios';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from './ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from './ui/tabs';
import { Avatar, AvatarFallback } from './ui/avatar';
import { Badge } from './ui/badge';
import { format } from 'date-fns';
import { getMemberProfile } from "../api/member";
import { Button } from "./ui/button";
import { Edit, User, Bell, Megaphone, AlertCircle, AlertTriangle } from "lucide-react";
import { toast } from 'sonner';
import { getSubscriptions, getMemberSubscriptionStatus } from "../api/subscription";
import { createEditRequest } from "../api/member";
import { createSubscriptionOrder, verifySubscriptionPayment } from "../api/subscription";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./ui/select";
import { loadRazorpayScript } from '../utils/razorpay';

interface MembershipProps {
  isLoggedIn: boolean;
  isPaidMember: boolean;
}

interface Member {
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
}

interface EditRequest {
  mobileNumber: string;
  mailId: string;
  designation: string;
  workingDistrict: string;
  residentialAddress: string;
  permanentAddress: string;
  remarks: string;
}

interface Subscription {
  id: string;
  name: string;
  description: string;
  amount?: number;
  frequency: string;
  status: string;
  lastPaidDate?: string;
  nextDueDate?: string;
  allowCustomAmount?: boolean;
  oneTime?: boolean;    // new
  needBased?: boolean;  // new
}

interface MemberNotification {
  id: string;
  title: string;
  message: string;
  priority: string;
  is_read: boolean;
  created_at: string;
}

export function Membership({ isLoggedIn, isPaidMember }: MembershipProps) {
  const [memberProfile, setMemberProfile] = useState<Member | null>(null);
  const [subscriptions, setSubscriptions] = useState<Subscription[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [paymentDialogOpen, setPaymentDialogOpen] = useState(false);
  const [selectedSubscription, setSelectedSubscription] = useState<Subscription | null>(null);
  const [customAmount, setCustomAmount] = useState("");
  const [isProcessingPayment, setIsProcessingPayment] = useState(false);
  const bannerImage = `${API_BASE_URL}/images/banner-image.jpg`;
  const API_BASE = API_BASE_URL;
  const [editDialogOpen, setEditDialogOpen] = useState(false);

  // Notifications state
  const [notifications, setNotifications] = useState<MemberNotification[]>([]);
  const [notificationsLoading, setNotificationsLoading] = useState(true);

  // Track paid subscriptions by ID
  const [paidSubscriptions, setPaidSubscriptions] = useState<Set<string>>(new Set());

  const [editRequest, setEditRequest] = useState({
    mobileNumber: "",
    mailId: "",
    designation: "",
    workingDistrict: "",
    residentialAddress: "",
    permanentAddress: "",
    remarks: "",
  });
  const [districtsList, setDistrictsList] = useState<string[]>([]);
  const [districtsLoading, setDistrictsLoading] = useState(false);

  const getStoredUser = JSON.parse(localStorage.getItem("user") || "{}");

  const formatDate = (dateStr: string) => {
    if (!dateStr) return "-";

    let date: Date;
    if (dateStr.includes("-")) {
      const parts = dateStr.split("-");
      if (parts.length === 3) {
        if (parts[0].length === 4) {
          // YYYY-MM-DD
          const [year, month, day] = parts.map(Number);
          date = new Date(year, month - 1, day);
        } else {
          // DD-MM-YYYY
          const [day, month, year] = parts.map(Number);
          date = new Date(year, month - 1, day);
        }
      } else {
        return "-";
      }
    } else if (dateStr.includes("/")) {
      const parts = dateStr.split("/");
      if (parts.length === 3) {
        if (parts[2].length === 4) {
          // DD/MM/YYYY
          const [day, month, year] = parts.map(Number);
          date = new Date(year, month - 1, day);
        } else {
          // YYYY/MM/DD
          const [year, month, day] = parts.map(Number);
          date = new Date(year, month - 1, day);
        }
      } else {
        return "-";
      }
    } else {
      date = new Date(dateStr);
    }

    return isNaN(date.getTime())
      ? "-"
      : format(date, "dd MMM yyyy");
  };

  // For dates in YYYY-MM-DD format (subscription plans)
  const formatISODate = (dateStr: string) => {
    if (!dateStr) return "-";
    const date = new Date(dateStr);
    return isNaN(date.getTime()) ? "-" : format(date, "dd MMM yyyy");
  };

  const isInGracePeriod = () => {
    const now = new Date();
    const year = now.getFullYear();
    const graceStart = new Date(year, 3, 1); // April 1
    const graceEnd = new Date(year, 4, 31, 23, 59, 59); // May 31
    return now >= graceStart && now <= graceEnd;
  };

  // Fetch notifications
  const fetchNotifications = async () => {
    try {
      const token = localStorage.getItem('member_token');
      const user = JSON.parse(localStorage.getItem("user") || "{}");

      if (!user.id && !user.emailId) {
        setNotificationsLoading(false);
        return;
      }

      // Get member ID from stored user
      const memberId = user.id;

      if (!memberId) {
        setNotificationsLoading(false);
        return;
      }

      const response = await axios.get(`${API_BASE}/member/notifications?member_id=${memberId}`, {
        headers: { Authorization: `Bearer ${token}` }
      });

      // Sort by date (newest first) and limit to recent 10
      const sorted = (response.data || []).sort((a: MemberNotification, b: MemberNotification) =>
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      );
      setNotifications(sorted.slice(0, 10));
    } catch (error) {
      console.error('Failed to fetch notifications:', error);
    } finally {
      setNotificationsLoading(false);
    }
  };

  // Mark notification as read
  const markAsRead = async (notificationId: string) => {
    try {
      const token = localStorage.getItem('member_token');
      await axios.put(`${API_BASE}/member/notifications/${notificationId}/read`, {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      // Update local state
      setNotifications(prev =>
        prev.map(n => n.id === notificationId ? { ...n, is_read: true } : n)
      );
    } catch (error) {
      console.error('Failed to mark as read:', error);
    }
  };

  // Get priority styles
  const getPriorityStyles = (priority: string) => {
    switch (priority?.toLowerCase()) {
      case 'urgent':
        return { bg: 'bg-red-50 dark:bg-red-950/30', border: 'border-red-500', icon: AlertCircle, color: 'text-red-600' };
      case 'high':
        return { bg: 'bg-orange-50 dark:bg-orange-950/30', border: 'border-orange-500', icon: AlertTriangle, color: 'text-orange-600' };
      default:
        return { bg: 'bg-blue-50 dark:bg-blue-950/30', border: 'border-blue-500', icon: Bell, color: 'text-blue-600' };
    }
  };

  // Fetch subscriptions
  useEffect(() => {
    const fetchSubscriptions = async () => {
      try {
        const data = await getSubscriptions();
        setSubscriptions(data);
      } catch {
        console.error("Subscriptions error");
      }
    };

    fetchSubscriptions();
  }, []);

  // Fetch paid subscriptions for this member
  useEffect(() => {
    const fetchPaidSubscriptions = async () => {
      if (!isLoggedIn) return;

      try {
        const user = JSON.parse(localStorage.getItem("user") || "{}");
        if (user.emailId) {
          const token = localStorage.getItem("member_token");
          const response = await axios.get(`${API_BASE}/subscriptions/member-paid?email=${user.emailId}`, {
            headers: { Authorization: `Bearer ${token}` }
          });
          setPaidSubscriptions(new Set(response.data));
        }
      } catch (error) {
        console.error("Failed to fetch paid subscriptions:", error);
      }
    };

    fetchPaidSubscriptions();
  }, [isLoggedIn, API_BASE]);

  // Fetch notifications when component mounts
  useEffect(() => {
    if (isLoggedIn) {
      fetchNotifications();
    }
  }, [isLoggedIn]);

  useEffect(() => {
    if (memberProfile) {
      setEditRequest({
        mobileNumber: memberProfile.mobileNumber || "",
        mailId: memberProfile.emailId || "",
        designation: memberProfile.designation || "",
        workingDistrict: memberProfile.workingDistrict || "",
        residentialAddress: memberProfile.residentialAddress || "",
        permanentAddress: memberProfile.permanentAddress || "",
        remarks: "",
      });
    }
  }, [memberProfile]);

  const [isSubmittingEdit, setIsSubmittingEdit] = useState(false);

  const handleEditRequest = async () => {
    try {
      setIsSubmittingEdit(true);
      const storedUser = JSON.parse(localStorage.getItem("user") || "{}");

      await createEditRequest({
        ...editRequest,
        email: storedUser.emailId,
      });

      toast.success("Request submitted successfully");
      setEditDialogOpen(false);

    } catch (error) {
      console.error(error);
      toast.error("Failed to submit request");
    } finally {
      setIsSubmittingEdit(false);
    }
  };

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        const storedUser = JSON.parse(localStorage.getItem("user") || "{}");

        if (!storedUser?.emailId) {
          setError("User not found. Please login again.");
          setLoading(false);
          return;
        }

        const data = await getMemberProfile();
        setMemberProfile(data);
      } catch (err) {
        console.error(err);
        setError("Failed to load profile");
      } finally {
        setLoading(false);
      }
    };

    if (isLoggedIn) {
      fetchProfile();
    }
  }, [isLoggedIn]);

  // Fetch districts for the edit request dropdown
  useEffect(() => {
    if (!editDialogOpen) return;
    const fetchDistricts = async () => {
      setDistrictsLoading(true);
      try {
        const res = await fetch(`${API_BASE}/office-bearers/districts`);
        const data = await res.json();
        setDistrictsList(data || []);
      } catch {
        console.error("Failed to load districts");
      } finally {
        setDistrictsLoading(false);
      }
    };
    fetchDistricts();
  }, [editDialogOpen]);

  const handlePaymentSubmit = async () => {
    try {
      setIsProcessingPayment(true);

      const loaded = await loadRazorpayScript();
      if (!loaded || !(window as any).Razorpay) {
        toast.error("Payment system failed to load. Please try again.");
        setIsProcessingPayment(false);
        return;
      }

      const user = JSON.parse(localStorage.getItem("user") || "{}");
      const amount = selectedSubscription?.allowCustomAmount
        ? Number(customAmount)
        : selectedSubscription?.amount || 0;

      if (!selectedSubscription) {
        toast.error("No subscription selected");
        return;
      }

      if (!user.emailId) {
        toast.error("User not found. Please login again.");
        return;
      }

      // Prepare notes for Razorpay (for admin email notification)
      const notes = {
        subscription_id: selectedSubscription.id,
        subscription_name: selectedSubscription.name,
        subscription_type: selectedSubscription.name,
        member_name: user.name || memberProfile?.name || "",
        member_taga_id: user.tagaId || "",
        member_email: user.emailId,
      };

      // Create Razorpay order with notes
      const order = await createSubscriptionOrder(
        selectedSubscription.id,
        amount * 100,
        user.emailId,
        notes
      );

      const options = {
        key: order.key,
        amount: order.amount,
        currency: order.currency,
        name: "TAGA Membership",
        description: selectedSubscription.name,
        order_id: order.orderId,
        //image: `${import.meta.env.VITE_API_BASE_URL}/images/logo/logo1.jpg`,
        prefill: {
          name: user.name || memberProfile?.name || "",
          email: user.emailId || "",
          contact: user.mobileNumber || "",
        },
        method: {
          upi: true,
          card: true,
          netbanking: true,
          wallet: true,
          paylater: true,
          emi: true,
          paylater_options: {
            paylater_providers: ['paytm', 'amazon', 'flexmoney', 'lazypay', 'zestmoney', 'olapay', 'mobikwik', 'slice', 'icic', 'hdfc', 'kissht']
          },
          upi_methods: [
            'google_pay',
            'phonepe',
            'amazon_pay',
            'paytm',
            'whatsapp_pay',
            'bhim',
            'freecharge',
            'mobikwik',
            'airtel_money',
            'axis_pay',
            'icici_pay',
            'hdfc_pay',
            'sbi_pay'
          ],
          display_priority: {
            upi: 1,
            card: 2,
            netbanking: 3,
            wallet: 4,
            paylater: 5,
            emi: 6
          }
        },
        theme: {
          color: '#297248',
        },
        modal: {
          ondismiss: function () {
            toast.info("Payment cancelled");
          }
        },
        handler: async (response: any) => {
          try {
            await verifySubscriptionPayment(
              selectedSubscription.id,
              response.razorpay_order_id,
              response.razorpay_payment_id,
              response.razorpay_signature,
              user.emailId,
              amount * 100
            );

            if (selectedSubscription.id === 'annual-subscription') {
              toast.success("Payment Successful! Your subscription was activated successfully! 🎉");
            } else {
              toast.success(`Payment Successful! Thank you for your contribution to ${selectedSubscription.name}. ✅`);
            }
            setPaymentDialogOpen(false);

            // Update paid subscriptions list ONLY if it's NOT need‑based
            // (need‑based payments should not be marked as "paid" in the UI)
            if (!selectedSubscription.needBased) {
              setPaidSubscriptions(prev => new Set([...prev, selectedSubscription.id]));
            }

            // 🔥 Only mark as paid if Annual Subscription was paid
            const updatedUser = {
              ...user,
              isPaid: selectedSubscription.id === 'annual-subscription',
              subscription_active: selectedSubscription.id === 'annual-subscription'
            };
            localStorage.setItem("user", JSON.stringify(updatedUser));

          } catch (err: any) {
            toast.error(err.message || "Payment verification failed");
          }
        },
      };

      const rzp = new (window as any).Razorpay(options);
      rzp.open();
    } catch (err: any) {
      toast.error(err.message || "Payment Failed");
    } finally {
      setIsProcessingPayment(false);
    }
  };

  if (!isLoggedIn) {
    return <p className="text-center mt-10" data-testid="testid-membership-login-required-message">Please login</p>;
  }

  if (loading) {
    return <p className="text-center mt-10" data-testid="testid-membership-loading-message">Loading profile...</p>;
  }

  if (error) {
    return <p className="text-center text-red-500 mt-10" data-testid="testid-membership-error-message">{error}</p>;
  }

  if (!memberProfile) {
    return null;
  }

  const unreadCount = notifications.filter(n => !n.is_read).length;

  return (
    <div className="space-y-6" data-testid="testid-membership-page">

      {/* Header */}
      <div className="relative overflow-hidden rounded-2xl shadow-2xl">
        <div className="absolute inset-0">
          <img
            src={bannerImage}
            className="w-full h-full object-cover"
            alt="Member Banner"
          />
          <div className="absolute inset-0 bg-green-900/80" />
        </div>

        <div className="relative p-10 text-white">
          <Badge className="mb-4 bg-green-600">
            <User className="w-3 h-3 mr-1" />
            Member: {memberProfile?.name || 'Loading...'}
          </Badge>

          <h1 className="text-4xl font-bold mb-2">
            My Profile
          </h1>

          <p>
            View and manage your TAGA profile and subscriptions
          </p>
        </div>
      </div>

      <Tabs defaultValue="profile" className="space-y-6" data-testid="testid-membership-tabs-form">

        {/* TAB BUTTONS */}
        <TabsList className="grid w-full grid-cols-3" data-testid="testid-membership-tabs-list">
          <TabsTrigger value="profile" data-testid="testid-member-profile-button">Member Profile</TabsTrigger>
          <TabsTrigger value="subscriptions" data-testid="testid-member-subscriptions-button">Subscriptions</TabsTrigger>
          <TabsTrigger value="announcements" data-testid="testid-member-announcements-button">
            <Megaphone className="w-4 h-4 mr-2" />
            Announcements
            {unreadCount > 0 && (
              <Badge className="ml-2 bg-red-500 text-white">
                {unreadCount}
              </Badge>
            )}
          </TabsTrigger>
        </TabsList>

        {/* PROFILE TAB */}
        <TabsContent value="profile" className="space-y-6">

          {/* PROFILE HEADER CARD */}
          <Card>
            <CardContent className="pt-6 flex items-center space-x-6">
              <Avatar className="w-20 h-20">
                <AvatarFallback>
                  {memberProfile.initial || memberProfile.name?.charAt(0)}
                </AvatarFallback>
              </Avatar>

              <div>
                <div className="flex items-center gap-3">
                  <h2 className="text-2xl font-bold">
                    {memberProfile.name}
                  </h2>

                  <Badge className="bg-green-600 text-white border border-green-300">
                    {isPaidMember ? "Subscriber" : "Member"}
                  </Badge>
                </div>
              </div>

              <Button
                onClick={() => setEditDialogOpen(true)}
                variant="outline"
                className="ml-auto flex items-center gap-2"
                data-testid="testid-request-profile-edit-button"
              >
                <Edit className="w-4 h-4" />
                Request Profile Edit
              </Button>
            </CardContent>
          </Card>

          {/* PERSONAL */}
          <Card>
            <CardHeader>
              <CardTitle>Personal Information</CardTitle>
            </CardHeader>
            <CardContent className="grid md:grid-cols-2 gap-4">
              <Field label="Name" value={memberProfile.name} />
              <Field label="Initial" value={memberProfile.initial} />
              <Field label="Gender" value={memberProfile.gender} />
              <Field label="Date of Birth" value={formatDate(memberProfile.dateOfBirth)} />
              <Field label="Father Name" value={memberProfile.fatherName} />
              <Field label="Mother Name" value={memberProfile.motherName} />
            </CardContent>
          </Card>

          {/* PROFESSIONAL */}
          <Card>
            <CardHeader>
              <CardTitle>Professional Details</CardTitle>
            </CardHeader>
            <CardContent className="grid md:grid-cols-2 gap-4">
              <Field label="Qualification" value={memberProfile.educationalQualification} />
              <Field label="Designation" value={memberProfile.designation} />
              <Field label="Working District" value={memberProfile.workingDistrict} />
              <Field label="Native District" value={memberProfile.nativeDistrict} />
              <Field label="Batch" value={memberProfile.recruitmentBatch} />
              <Field label="Seniority No" value={memberProfile.seniorityNumber} />
              <Field label="TBF Number" value={memberProfile.tbfNumber} />
              <Field label="CPS/GPF" value={memberProfile.cpsGpfNumber} />
            </CardContent>
          </Card>

          {/* CONTACT */}
          <Card>
            <CardHeader>
              <CardTitle>Contact</CardTitle>
            </CardHeader>
            <CardContent className="grid md:grid-cols-2 gap-4">
              <Field label="Mobile" value={memberProfile.mobileNumber} />
              <Field label="Email" value={memberProfile.emailId} />
              <Field label="Residential Address" value={memberProfile.residentialAddress} />
              <Field label="Permanent Address" value={memberProfile.permanentAddress} />
            </CardContent>
          </Card>

        </TabsContent>

        {/* SUBSCRIPTIONS TAB */}
        <TabsContent value="subscriptions" className="space-y-6">

          <Card className="p-6">
            <CardHeader className="p-0 mb-6">
              <CardTitle className="text-lg font-semibold">Subscription & Payments</CardTitle>
              <p className="text-sm text-muted-foreground">
                Manage your TAGA membership and optional contributions
              </p>
            </CardHeader>

            <div className="grid gap-4">
              {subscriptions.map((sub) => {
                const isPaid = paidSubscriptions.has(sub.id);
                const isAnnual = sub.id === "annual-subscription";
                const inGrace = isInGracePeriod();

                // Determine button/status area
                let actionComponent;
                // One‑time subscriptions that are already paid -> show "Paid" badge, no button
                if (sub.oneTime && isPaid) {
                  actionComponent = (
                    <span className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-green-100 text-green-700 text-sm font-medium">
                      <span className="w-2 h-2 rounded-full bg-green-500" />
                      Paid
                    </span>
                  );
                }
                // Annual subscription in grace period and not yet paid
                else if (isAnnual && inGrace && isPaidMember && !isPaid) {
                  actionComponent = (
                    <div className="text-right">
                      <button
                        className="px-4 py-2 bg-orange-600 hover:bg-orange-700 text-white rounded-lg text-sm font-medium transition-colors"
                        onClick={() => {
                          setSelectedSubscription(sub);
                          setPaymentDialogOpen(true);
                        }}
                      >
                        Renew Now
                      </button>
                      <p className="text-xs text-orange-600 mt-1">Grace period until May 31</p>
                    </div>
                  );
                }
                // For any subscription that is paid (including annual) – but need‑based should NEVER be marked as paid
                else if (!sub.needBased && isPaid) {
                  actionComponent = (
                    <span
                      data-testid={`testid-paid-badge-${sub.id}`}
                      className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-green-100 text-green-700 text-sm font-medium"
                    >
                      <span className="w-2 h-2 rounded-full bg-green-500" />
                      Paid
                    </span>
                  );
                }
                // Default: show Pay Now button (always shown for need‑based, even if previously paid)
                else {
                  actionComponent = (
                    <button
                      data-testid={`testid-pay-now-${sub.id}-button`}
                      className="px-4 py-2 bg-green-700 hover:bg-green-800 text-white rounded-lg text-sm font-medium transition-colors"
                      onClick={() => {
                        setSelectedSubscription(sub);
                        setPaymentDialogOpen(true);
                      }}
                    >
                      Pay Now
                    </button>
                  );
                }

                return (
                  <div
                    key={sub.id}
                    className={`relative overflow-hidden rounded-xl border bg-white shadow-sm transition-shadow hover:shadow-md ${isAnnual ? "ring-2 ring-green-200 border-green-300" : "border-gray-200"
                      }`}
                  >
                    {/* Optional highlight banner for annual subscription */}
                    {isAnnual && (
                      <div className="absolute top-0 right-0 bg-green-600 text-white text-xs px-3 py-1 rounded-bl-lg">
                        Required
                      </div>
                    )}

                    <div className="p-5">
                      {/* Top row: Name, status & action */}
                      <div className="flex items-start justify-between gap-4">
                        <div className="flex-1">
                          <div className="flex items-center gap-3 mb-1">
                            <h3 className="text-base font-semibold text-gray-900">{sub.name}</h3>
                            {!isPaid && (
                              <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-gray-100 text-gray-600 text-xs">
                                <span className="w-1.5 h-1.5 rounded-full bg-gray-400" />
                                Not Paid
                              </span>
                            )}
                          </div>
                          <p className="text-sm text-gray-500 mb-3">{sub.description}</p>
                        </div>
                        <div className="shrink-0">{actionComponent}</div>
                      </div>

                      {/* Details row: Amount, Frequency, Next Due (only annual) */}
                      <div className="grid grid-cols-3 gap-4 text-sm border-t pt-4 mt-2">
                        <div>
                          <span className="text-gray-500">Amount</span>
                          <p className="font-medium text-gray-900">
                            {sub.allowCustomAmount ? "Flexible" : `₹ ${sub.amount}`}
                          </p>
                        </div>
                        <div>
                          <span className="text-gray-500">Frequency</span>
                          <p className="font-medium text-gray-900 capitalize">{sub.frequency}</p>
                        </div>
                        {isAnnual && sub.nextDueDate && (
                          <div>
                            <span className="text-gray-500">Next Due</span>
                            <p className="font-medium text-gray-900">{formatISODate(sub.nextDueDate)}</p>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </Card>

          {/* Payment info card (unchanged) */}
          <Card className="bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800">
            <CardContent className="pt-6">
              <div className="flex items-start space-x-3">
                <span className="text-blue-600 text-lg mt-0.5">ⓘ</span>
                <div className="text-sm">
                  <p className="font-semibold text-blue-900 dark:text-blue-100 mb-1">Payment Information</p>
                  <ul className="text-blue-800 dark:text-blue-200 space-y-1">
                    <li>• All payments are processed securely through Razorpay</li>
                    <li>• Annual Subscription period: April 1 - March 31</li>
                    <li>• Grace period: 2 months for renewal</li>
                    <li>• User IDs may be blocked after grace period</li>
                    <li>• Email confirmation after successful payment</li>
                    <li>• Keep receipts for future reference</li>
                  </ul>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* ANNOUNCEMENTS TAB */}
        <TabsContent value="announcements" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Megaphone className="w-5 h-5" />
                Announcements & Notifications
              </CardTitle>
              <CardDescription>
                Important updates and announcements from TAGA administration
              </CardDescription>
            </CardHeader>
            <CardContent>
              {notificationsLoading ? (
                <div className="text-center py-8 text-muted-foreground">Loading announcements...</div>
              ) : notifications.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  No announcements at this time
                </div>
              ) : (
                <div className="space-y-4">
                  {notifications.map((notification) => {
                    const styles = getPriorityStyles(notification.priority);
                    const IconComponent = styles.icon;
                    return (
                      <div
                        key={notification.id}
                        className={`p-4 rounded-lg border-l-4 ${styles.bg} ${styles.border} ${!notification.is_read ? 'cursor-pointer hover:shadow-md transition-shadow' : ''}`}
                        onClick={() => !notification.is_read && markAsRead(notification.id)}
                      >
                        <div className="flex items-start gap-3">
                          <IconComponent className={`w-5 h-5 mt-0.5 ${styles.color}`} />
                          <div className="flex-1">
                            <div className="flex items-center justify-between flex-wrap gap-2">
                              <h3 className="font-semibold flex items-center gap-2">
                                {notification.title}
                                {!notification.is_read && (
                                  <Badge variant="default" className="bg-green-500 text-white text-xs">
                                    New
                                  </Badge>
                                )}
                              </h3>
                              <span className="text-xs text-muted-foreground">
                                {new Date(notification.created_at).toLocaleDateString('en-GB')}
                              </span>
                            </div>
                            <p className="text-sm text-muted-foreground mt-2 whitespace-pre-wrap">
                              {notification.message}
                            </p>
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

      </Tabs>

      {/* PAYMENT DIALOG */}
      {paymentDialogOpen && selectedSubscription && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" data-testid="testid-membership-payment-modal-overlay">

          <div className="bg-white w-[95%] max-w-md rounded-2xl shadow-2xl" data-testid="testid-membership-payment-modal">

            {/* HEADER */}
            <div className="p-6 border-b">
              <h2 className="text-xl font-semibold">
                Make Payment
              </h2>
              <p className="text-sm text-gray-500">
                {selectedSubscription.name}
              </p>
            </div>

            {/* BODY */}
            <div className="p-6 space-y-4">
              <div className="bg-green-50 p-4 rounded-lg">
                <p className="text-sm text-gray-600 mb-2">
                  {selectedSubscription.description} (Grace period of 2 months for renewal.)
                </p>
                {selectedSubscription.allowCustomAmount ? (
                  <div className="space-y-2 mt-3">
                    <label className="text-sm">Enter Amount</label>
                    <input
                      type="number"
                      className="w-full border px-3 py-2 rounded-md"
                      data-testid="testid-membership-custom-amount-input"
                      value={customAmount}
                      onChange={(e) => setCustomAmount(e.target.value)}
                    />
                  </div>
                ) : (
                  <div className="mt-3">
                    <p className="text-sm text-gray-500">Payment Amount</p>
                    <p className="text-2xl font-bold text-green-700">
                      ₹ {selectedSubscription.amount}
                    </p>
                  </div>
                )}
              </div>

              {/* INFO */}
              <div className="bg-gray-100 p-3 rounded-md text-xs text-gray-600">
                Payment will be processed through Razorpay's secure payment gateway.
                You will receive a confirmation email upon successful payment.
              </div>
            </div>

            {/* FOOTER */}
            <div className="flex justify-end gap-3 p-4 border-t">
              <button
                onClick={() => setPaymentDialogOpen(false)}
                className="px-4 py-2 border rounded-md"
                data-testid="testid-membership-payment-cancel-button"
              >
                Cancel
              </button>

              <button
                onClick={handlePaymentSubmit}
                disabled={isProcessingPayment}
                className="px-4 py-2 bg-green-700 text-white rounded-md"
                data-testid="testid-membership-payment-submit-button"
              >
                {isProcessingPayment ? "Processing..." : "Proceed to Pay"}
              </button>
            </div>

          </div>
        </div>
      )}

      {/* EDIT REQUEST MODAL */}
      {editDialogOpen && (
        <div
          className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4"
          data-testid="testid-profile-edit-modal-overlay"
          onClick={(e) => e.target === e.currentTarget && setEditDialogOpen(false)}
        >
          <div
            className="bg-white w-full max-w-2xl rounded-2xl shadow-2xl max-h-[90vh] flex flex-col"
            data-testid="testid-profile-edit-modal"
          >

            {/* HEADER */}
            <div className="flex items-start justify-between p-5 border-b shrink-0">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-emerald-50">
                  <Edit className="w-5 h-5 text-emerald-700" />
                </div>
                <div>
                  <h2 className="text-lg font-semibold text-gray-900">Request Profile Edit</h2>
                  <p className="text-xs text-gray-500 mt-0.5">
                    Changes require admin approval before taking effect
                  </p>
                </div>
              </div>
              <button
                onClick={() => setEditDialogOpen(false)}
                className="text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg p-1.5 transition-colors"
              >
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            {/* SCROLLABLE BODY */}
            <div className="overflow-y-auto flex-1 p-5 space-y-5">

              {/* SECTION: Contact */}
              <div>
                <p className="text-[11px] font-semibold text-gray-400 uppercase tracking-widest mb-3">
                  Contact Details
                </p>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">

                  {/* Mobile */}
                  <div className="space-y-1">
                    <label className="text-xs font-medium text-gray-600">
                      Mobile Number
                    </label>
                    <input
                      className="w-full px-3 py-2 text-sm rounded-lg border border-gray-200 bg-[#F1F8E9] focus:outline-none focus:ring-2 focus:ring-emerald-500 placeholder:text-gray-400"
                      placeholder="e.g. 9876543210"
                      data-testid="testid-profile-edit-mobile-input"
                      value={editRequest.mobileNumber}
                      onChange={(e) =>
                        setEditRequest({ ...editRequest, mobileNumber: e.target.value })
                      }
                    />
                  </div>

                  {/* Email */}
                  <div className="space-y-1">
                    <label className="text-xs font-medium text-gray-600">
                      Email ID
                    </label>
                    <input
                      className="w-full px-3 py-2 text-sm rounded-lg border border-gray-200 bg-[#F1F8E9] focus:outline-none focus:ring-2 focus:ring-emerald-500 placeholder:text-gray-400"
                      placeholder="e.g. name@email.com"
                      data-testid="testid-profile-edit-email-input"
                      value={editRequest.mailId}
                      onChange={(e) =>
                        setEditRequest({ ...editRequest, mailId: e.target.value })
                      }
                    />
                  </div>
                </div>
              </div>

              <div className="border-t" />

              {/* SECTION: Professional */}
              <div>
                <p className="text-[11px] font-semibold text-gray-400 uppercase tracking-widest mb-3">
                  Professional Details
                </p>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">

                  {/* Designation */}
                  <div className="space-y-1">
                    <label className="text-xs font-medium text-gray-600">
                      Designation
                    </label>
                    <input
                      className="w-full px-3 py-2 text-sm rounded-lg border border-gray-200 bg-[#F1F8E9] focus:outline-none focus:ring-2 focus:ring-emerald-500 placeholder:text-gray-400"
                      placeholder="e.g. Agricultural Officer"
                      data-testid="testid-profile-edit-designation-input"
                      value={editRequest.designation}
                      onChange={(e) =>
                        setEditRequest({ ...editRequest, designation: e.target.value })
                      }
                    />
                  </div>

                  {/* Working District */}
                  <div className="space-y-1">
                    <label className="text-xs font-medium text-gray-600">
                      Working District
                    </label>
                    {districtsLoading ? (
                      <div className="flex items-center gap-2 h-9 px-3 rounded-lg border border-gray-200 bg-[#F1F8E9] text-sm text-gray-400">
                        <svg className="w-3.5 h-3.5 animate-spin shrink-0" fill="none" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z" />
                        </svg>
                        Loading districts…
                      </div>
                    ) : (
                      <Select
                        value={editRequest.workingDistrict}
                        onValueChange={(val: string) =>
                          setEditRequest({ ...editRequest, workingDistrict: val })
                        }
                      >
                        <SelectTrigger
                          className="w-full h-[38px] text-sm border border-gray-200 bg-[#F1F8E9] focus:ring-2 focus:ring-emerald-500 focus:ring-offset-0 rounded-lg"
                          data-testid="testid-profile-edit-working-district-input"
                        >
                          <SelectValue placeholder="Select district…" />
                        </SelectTrigger>
                        <SelectContent className="max-h-64">
                          {districtsList.map((d) => (
                            <SelectItem key={d} value={d} className="text-sm cursor-pointer">
                              {d}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    )}
                  </div>
                </div>
              </div>

              <div className="border-t" />

              {/* SECTION: Address */}
              <div>
                <p className="text-[11px] font-semibold text-gray-400 uppercase tracking-widest mb-3">
                  Address
                </p>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">

                  <div className="space-y-1">
                    <label className="text-xs font-medium text-gray-600">
                      Residential Address
                    </label>
                    <textarea
                      rows={3}
                      className="w-full px-3 py-2 text-sm rounded-lg border border-gray-200 bg-[#F1F8E9] focus:outline-none focus:ring-2 focus:ring-emerald-500 resize-none placeholder:text-gray-400"
                      placeholder="Current residential address"
                      data-testid="testid-profile-edit-residential-address-input"
                      value={editRequest.residentialAddress}
                      onChange={(e) =>
                        setEditRequest({ ...editRequest, residentialAddress: e.target.value })
                      }
                    />
                  </div>

                  <div className="space-y-1">
                    <label className="text-xs font-medium text-gray-600">
                      Permanent Address
                    </label>
                    <textarea
                      rows={3}
                      className="w-full px-3 py-2 text-sm rounded-lg border border-gray-200 bg-[#F1F8E9] focus:outline-none focus:ring-2 focus:ring-emerald-500 resize-none placeholder:text-gray-400"
                      placeholder="Permanent / native address"
                      data-testid="testid-profile-edit-permanent-address-input"
                      value={editRequest.permanentAddress}
                      onChange={(e) =>
                        setEditRequest({ ...editRequest, permanentAddress: e.target.value })
                      }
                    />
                  </div>
                </div>
              </div>

              <div className="border-t" />

              {/* SECTION: Remarks */}
              <div className="space-y-1">
                <label className="text-xs font-medium text-gray-600">
                  Remarks{" "}
                  <span className="text-gray-400 font-normal">(Optional)</span>
                </label>
                <textarea
                  rows={2}
                  className="w-full px-3 py-2 text-sm rounded-lg border border-gray-200 focus:outline-none focus:ring-2 focus:ring-emerald-500 resize-none placeholder:text-gray-400"
                  placeholder="Any additional notes for the admin…"
                  data-testid="testid-profile-edit-remarks-input"
                  value={editRequest.remarks}
                  onChange={(e) =>
                    setEditRequest({ ...editRequest, remarks: e.target.value })
                  }
                />
              </div>

              {/* INFO BANNER */}
              <div className="flex items-start gap-2.5 bg-amber-50 border border-amber-200 rounded-lg p-3">
                <svg className="w-4 h-4 text-amber-500 mt-0.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                    d="M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
                </svg>
                <p className="text-xs text-amber-800">
                  Only fill fields you want to change — leave others as-is. Your request goes to admin for review and you'll be notified once updated.
                </p>
              </div>

            </div>

            {/* FOOTER */}
            <div className="flex items-center justify-end gap-2 px-5 py-4 border-t bg-gray-50 rounded-b-2xl shrink-0">
              <button
                className="px-4 py-2 text-sm rounded-lg border border-gray-300 text-gray-700 hover:bg-gray-100 transition-colors"
                onClick={() => setEditDialogOpen(false)}
                data-testid="testid-profile-edit-cancel-button"
              >
                Cancel
              </button>
              <button
                className={`px-5 py-2 text-sm rounded-lg ${isSubmittingEdit ? 'bg-emerald-600 opacity-70 cursor-not-allowed' : 'bg-emerald-700 hover:bg-emerald-800'} text-white font-medium transition-colors flex items-center gap-2`}
                onClick={handleEditRequest}
                disabled={isSubmittingEdit}
                data-testid="testid-profile-edit-submit-button"
              >
                {isSubmittingEdit ? (
                  <>
                    <svg className="animate-spin w-4 h-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Sending...
                  </>
                ) : (
                  <>
                    <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                    Submit Request
                  </>
                )}
              </button>
            </div>

          </div>
        </div>
      )}

    </div>
  );
}

function Field({ label, value }: { label: string; value: any }) {
  return (
    <div>
      <p className="text-sm text-muted-foreground">{label}</p>
      <p className="font-medium">{value || "-"}</p>
    </div>
  );
}