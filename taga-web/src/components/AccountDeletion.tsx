import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Textarea } from './ui/textarea';
import { Checkbox } from './ui/checkbox';
import { Alert, AlertDescription, AlertTitle } from './ui/alert';
import { Badge } from './ui/badge';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from './ui/alert-dialog';
import { toast } from 'sonner';
import {
  ShieldAlert,
  AlertTriangle,
  Clock,
  XCircle,
  Eye,
  EyeOff,
  LogOut,
  ArrowLeft,
  Loader2,
  FileText,
  Building2,
  Lock
} from 'lucide-react';
import {
  getAccountDeletionEligibility,
  getAccountDeletionStatus,
  submitAccountDeletionRequest,
  AccountDeletionStatus,
  AccountDeletionEligibility
} from '../api/member';

interface AccountDeletionProps {
  onNavigate?: (page: string) => void;
  onLogout?: () => void;
}

export function AccountDeletion({ onNavigate, onLogout }: AccountDeletionProps) {
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [statusData, setStatusData] = useState<AccountDeletionStatus | null>(null);
  const [eligibility, setEligibility] = useState<AccountDeletionEligibility | null>(null);

  // Form states
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [reason, setReason] = useState('');
  const [confirmedCheck, setConfirmedCheck] = useState(false);
  const [confirmDialogOpen, setConfirmDialogOpen] = useState(false);
  const [postDeletionDialogOpen, setPostDeletionDialogOpen] = useState(false);
  const [countdown, setCountdown] = useState(5);

  // Auto-logout countdown timer when post-deletion dialog is open
  useEffect(() => {
    let timer: NodeJS.Timeout | undefined;
    if (postDeletionDialogOpen && countdown > 0) {
      timer = setTimeout(() => setCountdown(c => c - 1), 1000);
    } else if (postDeletionDialogOpen && countdown === 0) {
      handleLogoutAction();
    }
    return () => {
      if (timer) clearTimeout(timer);
    };
  }, [postDeletionDialogOpen, countdown]);

  // Re-request state if rejected
  const [isReRequesting, setIsReRequesting] = useState(false);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [statusRes, eligRes] = await Promise.all([
        getAccountDeletionStatus().catch(() => ({ status: 'none' as const })),
        getAccountDeletionEligibility().catch((err: any) => ({
          eligible: false,
          message: err.response?.data?.message || err.response?.data?.error || 'Failed to verify eligibility',
          reason: err.response?.data?.reason
        }))
      ]);

      setStatusData(statusRes);
      setEligibility(eligRes);
    } catch (err) {
      console.error('Failed to load account deletion data', err);
      toast.error('Failed to verify account deletion status');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleLogoutAction = () => {
    try {
      localStorage.removeItem('member_token');
      localStorage.removeItem('member_token_expiry');
      localStorage.removeItem('member_role');
      localStorage.removeItem('user');
      localStorage.removeItem('lastPage');
      sessionStorage.removeItem('lastPage');
      sessionStorage.removeItem('post_login_redirect');
    } catch {
      // safe ignore
    }

    if (onLogout) {
      onLogout();
    }
    if (onNavigate) {
      onNavigate('member-login');
    } else {
      window.location.href = '/#/member-login';
      window.location.reload();
    }
  };

  const handleSubmit = async () => {
    if (!password.trim()) {
      toast.error('Password is required');
      return;
    }
    if (!reason.trim()) {
      toast.error('Please enter a reason for account deletion');
      return;
    }
    if (!confirmedCheck) {
      toast.error('Please confirm that you have read and accepted the disclosures');
      return;
    }

    setSubmitting(true);
    try {
      const fullReason = reason.trim();

      const res = await submitAccountDeletionRequest({
        password,
        reason: fullReason
      });

      toast.success(res.message || 'Account deletion request submitted successfully');

      // Clear tokens upon successful submission per requirements
      try {
        localStorage.removeItem('member_token');
        localStorage.removeItem('member_token_expiry');
        localStorage.removeItem('member_role');
        localStorage.removeItem('user');
        localStorage.removeItem('lastPage');
        sessionStorage.removeItem('lastPage');
        sessionStorage.removeItem('post_login_redirect');
      } catch {
        // safe ignore
      }

      // Close submit confirmation dialog and open post-deletion logout dialog
      setConfirmDialogOpen(false);
      setPostDeletionDialogOpen(true);
      setCountdown(5);
    } catch (err: any) {
      const msg = err.response?.data?.error || err.response?.data?.message || 'Failed to submit account deletion request';
      toast.error(msg);
      setConfirmDialogOpen(false);
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center py-20 space-y-4">
        <Loader2 className="w-10 h-10 animate-spin text-primary" />
        <p className="text-muted-foreground text-sm">Checking account deletion eligibility...</p>
      </div>
    );
  }

  // Current effective status
  const currentStatus = isReRequesting ? 'none' : statusData?.status || 'none';
  const hasActiveStay = eligibility?.reason === 'active_stay' ||
    (!eligibility?.eligible && eligibility?.message?.includes('staying at TAGA Towers'));

  return (
    <div className="max-w-4xl mx-auto space-y-8 pb-12" data-testid="account-deletion-page">
      {/* Back Button */}
      {onNavigate && (
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onNavigate('membership')}
          className="text-muted-foreground hover:text-foreground -ml-2"
          data-testid="testid-back-to-profile-button"
        >
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back to Member Profile
        </Button>
      )}

      {/* Header */}
      <div className="border-b pb-6">
        <div className="flex items-center gap-3 mb-2">
          <div className="p-2.5 bg-red-100 rounded-xl">
            <ShieldAlert className="w-7 h-7 text-red-600" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-foreground">
              Account Deletion
            </h1>
            <p className="text-sm text-muted-foreground">
              Submit a request to close your account and conclude association services.
            </p>
          </div>
        </div>
      </div>

      {/* ==================== STATE 1: PENDING STATE ==================== */}
      {currentStatus === 'pending' && (
        <Card className="border-amber-200 bg-amber-50/40 shadow-sm" data-testid="testid-deletion-pending-card">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Clock className="w-6 h-6 text-amber-600 animate-pulse" />
                <CardTitle className="text-xl text-amber-900">
                  Request Under Review
                </CardTitle>
              </div>
              <Badge variant="outline" className="bg-amber-100 text-amber-800 border-amber-300 font-medium">
                Pending Review
              </Badge>
            </div>
            <CardDescription className="text-amber-800/90 text-sm mt-1">
              Your account deletion request is currently awaiting administrator review. Login access remains paused during this period. You will be notified once reviewed.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <Alert className="bg-white border-amber-200">
              <AlertTriangle className="h-4 w-4 text-amber-600" />
              <AlertTitle className="text-amber-900 font-medium">Request Details</AlertTitle>
              <AlertDescription className="text-sm text-slate-700 mt-2 space-y-1">
                {statusData?.requested_at && (
                  <div>
                    <span className="font-semibold text-slate-900">Submitted On: </span>
                    {new Date(statusData.requested_at).toLocaleString('en-IN', {
                      dateStyle: 'medium',
                      timeStyle: 'short'
                    })}
                  </div>
                )}
                {statusData?.reason && (
                  <div>
                    <span className="font-semibold text-slate-900">Reason: </span>
                    {statusData.reason}
                  </div>
                )}
              </AlertDescription>
            </Alert>

            <div className="p-4 bg-amber-100/60 rounded-lg text-xs text-amber-900 space-y-1">
              <p className="font-semibold">Important Policy Note:</p>
              <p>Per association policy, deletion requests are permanent and have no withdrawal option once submitted. An administrator will review your request shortly.</p>
            </div>
          </CardContent>
          <CardFooter className="flex justify-between items-center border-t border-amber-200 pt-4">
            <p className="text-xs text-muted-foreground">Session credentials have been cleared.</p>
            <Button
              variant="outline"
              onClick={handleLogoutAction}
              className="border-amber-300 text-amber-900 hover:bg-amber-100 font-medium"
              data-testid="testid-pending-logout-button"
            >
              <LogOut className="w-4 h-4 mr-2" />
              Log Out & Refresh
            </Button>
          </CardFooter>
        </Card>
      )}

      {/* ==================== STATE 2: REJECTED STATE ==================== */}
      {currentStatus === 'rejected' && (
        <Card className="border-red-200 bg-red-50/30 shadow-sm" data-testid="testid-deletion-rejected-card">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <XCircle className="w-6 h-6 text-red-600" />
                <CardTitle className="text-xl text-red-900">
                  Account Deletion Request Rejected
                </CardTitle>
              </div>
              <Badge variant="destructive" className="bg-red-600 text-white font-medium">
                Rejected
              </Badge>
            </div>
            <CardDescription className="text-red-800/90 text-sm mt-1">
              An administrator has reviewed and rejected your deletion request. Your account remains active and in good standing.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <Alert className="bg-white border-red-200">
              <AlertTriangle className="h-4 w-4 text-red-600" />
              <AlertTitle className="text-red-900 font-medium">Administrator Remarks</AlertTitle>
              <AlertDescription className="text-sm text-slate-800 mt-2 space-y-1">
                <p className="p-3 bg-red-50/60 rounded border border-red-100 font-medium text-red-950">
                  {statusData?.remarks || 'No specific remarks provided.'}
                </p>
                {statusData?.processed_at && (
                  <p className="text-xs text-muted-foreground mt-2">
                    Processed Date: {new Date(statusData.processed_at).toLocaleString('en-IN', {
                      dateStyle: 'medium',
                      timeStyle: 'short'
                    })}
                  </p>
                )}
              </AlertDescription>
            </Alert>
          </CardContent>
          <CardFooter className="flex justify-between items-center border-t border-red-200 pt-4">
            <Button
              variant="outline"
              onClick={() => onNavigate ? onNavigate('membership') : undefined}
            >
              Return to Profile
            </Button>
            <Button
              variant="default"
              className="bg-red-600 hover:bg-red-700 text-white"
              onClick={() => setIsReRequesting(true)}
              data-testid="testid-submit-new-deletion-request-button"
            >
              Submit New Request
            </Button>
          </CardFooter>
        </Card>
      )}

      {/* ==================== STATE 3 & 4: NORMAL OR ACTIVE STAY BLOCKED ==================== */}
      {currentStatus === 'none' && (
        <div className="space-y-8">
          {/* ACTIVE STAY WARNING BLOCK */}
          {hasActiveStay && (
            <Alert variant="destructive" className="bg-red-50 border-red-300 text-red-950" data-testid="testid-active-stay-alert">
              <Building2 className="h-5 w-5 text-red-600" />
              <AlertTitle className="text-base font-semibold text-red-900 mb-1">
                Active TAGA Towers Stay Detected
              </AlertTitle>
              <AlertDescription className="text-sm text-red-800">
                You are currently staying at TAGA Towers. You can request account deletion after completing your checkout.
              </AlertDescription>
            </Alert>
          )}

          {/* DISCLOSURES CHECKLIST CARD */}
          <Card className="border-slate-200 shadow-sm overflow-hidden">
            <CardHeader className="bg-slate-50/80 border-b pb-4">
              <div className="flex items-center gap-2">
                <FileText className="w-5 h-5 text-slate-700" />
                <CardTitle className="text-lg text-slate-900">
                  Important Information Before You Proceed
                </CardTitle>
              </div>
              <CardDescription className="text-xs text-slate-500">
                Please review the operational details below before submitting your request.
              </CardDescription>
            </CardHeader>
            <CardContent className="p-6">
              <ul className="space-y-3.5 text-sm text-slate-700">
                <li className="flex items-start gap-3">
                  <Lock className="w-5 h-5 text-amber-600 shrink-0 mt-0.5" />
                  <div>
                    <span className="font-semibold text-slate-900">Account Access During Review: </span>
                    Once your request is submitted, your account login will be disabled while under review. You will not be able to log into your account unless the request is rejected by an administrator.
                  </div>
                </li>
                <li className="flex items-start gap-3">
                  <Building2 className="w-5 h-5 text-blue-600 shrink-0 mt-0.5" />
                  <div>
                    <span className="font-semibold text-slate-900">TAGA Towers Check-in Status: </span>
                    Account deletion cannot be submitted if you are currently checked in at TAGA Towers. Please complete your checkout before requesting deletion.
                  </div>
                </li>
                <li className="flex items-start gap-3">
                  <ShieldAlert className="w-5 h-5 text-indigo-600 shrink-0 mt-0.5" />
                  <div>
                    <span className="font-semibold text-slate-900">Subscriber &amp; Association Services: </span>
                    Access to TAGA subscriber benefits, active subscription services, and association features will end once the deletion is processed.
                  </div>
                </li>
                <li className="flex items-start gap-3">
                  <Clock className="w-5 h-5 text-purple-600 shrink-0 mt-0.5" />
                  <div>
                    <span className="font-semibold text-slate-900">Administrator Review: </span>
                    Requests are reviewed and verified by association administrators before final processing.
                  </div>
                </li>
                <li className="flex items-start gap-3">
                  <AlertTriangle className="w-5 h-5 text-red-600 shrink-0 mt-0.5" />
                  <div>
                    <span className="font-semibold text-slate-900">Request Finality: </span>
                    Once submitted, an account deletion request cannot be cancelled or withdrawn.
                  </div>
                </li>
              </ul>
            </CardContent>
          </Card>

          {/* DELETION FORM CARD */}
          <Card className="border-red-200/80 shadow-md">
            <CardHeader className="bg-red-50/30 border-b pb-4">
              <CardTitle className="text-lg text-red-900 flex items-center gap-2">
                <AlertTriangle className="w-5 h-5 text-red-600" />
                Submit Deletion Request
              </CardTitle>
              <CardDescription className="text-xs text-red-700/80">
                Authentication and reason verification are required to process account deletion.
              </CardDescription>
            </CardHeader>
            <CardContent className="p-6 space-y-6">
              {/* Reason Input (Dynamic Textarea) */}
              <div className="space-y-2">
                <Label htmlFor="reason-input" className="text-sm font-semibold text-slate-800">
                  Reason for Account Deletion <span className="text-red-500">*</span>
                </Label>
                <Textarea
                  id="reason-input"
                  placeholder="Please write your reason for requesting account deletion..."
                  rows={4}
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  disabled={hasActiveStay || submitting}
                  data-testid="testid-reason-input"
                />
              </div>

              {/* Password Confirmation */}
              <div className="space-y-2">
                <Label htmlFor="password-input" className="text-sm font-semibold text-slate-800">
                  Confirm Current Password <span className="text-red-500">*</span>
                </Label>
                <div className="relative">
                  <Input
                    id="password-input"
                    type={showPassword ? 'text' : 'password'}
                    placeholder="Enter your current password to authorize"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    disabled={hasActiveStay || submitting}
                    className="pr-10"
                    data-testid="testid-password-input"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    tabIndex={-1}
                  >
                    {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>
                <p className="text-xs text-muted-foreground">
                  Your password is verified securely against association records.
                </p>
              </div>

              {/* Confirmation Checkbox */}
              <div className="flex items-start space-x-3 pt-2">
                <Checkbox
                  id="confirm-terms"
                  checked={confirmedCheck}
                  onCheckedChange={(checked) => setConfirmedCheck(checked === true)}
                  disabled={hasActiveStay || submitting}
                  data-testid="testid-confirm-checkbox"
                />
                <Label
                  htmlFor="confirm-terms"
                  className="text-xs text-slate-700 leading-snug cursor-pointer select-none"
                >
                  I understand that submitting this request will initiate account closure, and once submitted, this request cannot be cancelled or withdrawn.
                </Label>
              </div>
            </CardContent>

            <CardFooter className="bg-slate-50/50 border-t p-6 flex flex-col sm:flex-row items-center justify-between gap-4">
              <p className="text-xs text-muted-foreground order-2 sm:order-1">
                {hasActiveStay
                  ? 'Form disabled due to active stay at TAGA Towers.'
                  : 'Requests require administrator verification.'}
              </p>

              {/* Submit Confirmation Dialog */}
              <AlertDialog open={confirmDialogOpen} onOpenChange={setConfirmDialogOpen}>
                <AlertDialogTrigger asChild>
                  <Button
                    variant="destructive"
                    className="bg-red-600 hover:bg-red-700 text-white w-full sm:w-auto order-1 sm:order-2"
                    disabled={hasActiveStay || submitting || !password.trim() || !reason.trim() || !confirmedCheck}
                    data-testid="testid-submit-deletion-button"
                  >
                    {submitting ? (
                      <>
                        <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                        Submitting Request...
                      </>
                    ) : (
                      'Request Account Deletion'
                    )}
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle className="text-red-600 flex items-center gap-2">
                      <AlertTriangle className="w-5 h-5" />
                      Confirm Account Deletion Request
                    </AlertDialogTitle>
                    <AlertDialogDescription className="text-slate-700 space-y-2">
                      <p>
                        Are you sure you want to proceed with requesting account deletion?
                      </p>
                      <p className="text-xs text-red-600 font-semibold">
                        Once submitted, your session will be ended immediately, your request will be queued for administrator review, and there is no withdrawal option.
                      </p>
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel disabled={submitting}>Cancel</AlertDialogCancel>
                    <AlertDialogAction
                      onClick={handleSubmit}
                      disabled={submitting}
                      className="bg-red-600 hover:bg-red-700 text-white"
                      data-testid="testid-confirm-dialog-submit-button"
                    >
                      {submitting ? (
                        <>
                          <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                          Submitting...
                        </>
                      ) : (
                        'Yes, Submit Deletion Request'
                      )}
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </CardFooter>
          </Card>
        </div>
      )}

      {/* Post-Deletion Logout Modal Popup */}
      <AlertDialog open={postDeletionDialogOpen} onOpenChange={() => {}}>
        <AlertDialogContent className="sm:max-w-md bg-white border border-amber-200 shadow-xl" data-testid="testid-post-deletion-dialog">
          <AlertDialogHeader className="items-center text-center">
            <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-amber-100 text-amber-600 mb-2">
              <Clock className="h-7 w-7 animate-pulse" />
            </div>
            <AlertDialogTitle className="text-xl font-bold text-amber-950">
              Account Deletion Requested
            </AlertDialogTitle>
            <AlertDialogDescription className="text-sm text-slate-700 space-y-3 pt-2 text-center">
              <p>
                Your account deletion request has been submitted successfully and is awaiting administrator verification.
              </p>
              <div className="p-3 bg-amber-50/80 rounded-lg border border-amber-200/80 text-xs text-amber-900 text-left space-y-1">
                <p className="font-semibold">Security & Session Notice:</p>
                <p>
                  Per association policy, your account login credentials are restricted while your request is pending review. For your protection, you are now logged out.
                </p>
              </div>
              <p className="text-xs text-muted-foreground font-medium">
                {countdown > 0 ? `Refreshing session to login in ${countdown}s...` : 'Session closed.'}
              </p>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter className="sm:justify-center mt-3">
            <Button
              onClick={handleLogoutAction}
              className="bg-amber-600 hover:bg-amber-700 text-white w-full sm:w-auto font-medium px-6 py-2.5 shadow-sm"
              data-testid="testid-post-deletion-logout-button"
            >
              <LogOut className="w-4 h-4 mr-2" />
              Log Out & Refresh
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

export default AccountDeletion;
