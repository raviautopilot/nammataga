import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogTrigger,
  DialogFooter
} from '../ui/dialog';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '../ui/alert-dialog';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
import { Textarea } from '../ui/textarea';
import { Label } from '../ui/label';
import { toast } from 'sonner';
import {
  UserX,
  Check,
  X,
  Clock,
  AlertTriangle,
  Loader2,
  ShieldAlert,
  Phone,
  Mail,
  CreditCard,
  Calendar,
  User,
  FileText
} from 'lucide-react';
import {
  getAdminDeletionRequests,
  approveDeletionRequest,
  rejectDeletionRequest,
  AdminDeletionRequest,
  getMembersList,
  MemberListItem
} from '../../api/adminContent';

export default function AdminDeletionRequestsButton() {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [requests, setRequests] = useState<AdminDeletionRequest[]>([]);
  const [memberMap, setMemberMap] = useState<Record<string, MemberListItem>>({});
  const [statusFilter, setStatusFilter] = useState<'all' | 'pending' | 'approved' | 'rejected'>('pending');

  // Approval state
  const [approvingReq, setApprovingReq] = useState<AdminDeletionRequest | null>(null);
  const [isApproving, setIsApproving] = useState(false);

  // Rejection state
  const [rejectingReq, setRejectingReq] = useState<AdminDeletionRequest | null>(null);
  const [rejectRemarks, setRejectRemarks] = useState('');
  const [rejectError, setRejectError] = useState('');
  const [isRejecting, setIsRejecting] = useState(false);

  const fetchRequestsData = async () => {
    setLoading(true);
    try {
      const [delRes, membersRes] = await Promise.all([
        getAdminDeletionRequests().catch((err) => {
          console.error('Failed to fetch deletion requests', err);
          return { requests: [], total: 0 };
        }),
        getMembersList(1, 200).catch((err) => {
          console.error('Failed to fetch members list for enrichment', err);
          return { members: [], total: 0, page: 1, limit: 200, total_pages: 0 };
        })
      ]);

      setRequests(delRes.requests || []);

      // Build quick lookup map by member ID, tagaId, and email
      const map: Record<string, MemberListItem> = {};
      (membersRes.members || []).forEach((m) => {
        if (m.id) map[m.id.toLowerCase()] = m;
        if (m.tagaId) map[m.tagaId.toLowerCase()] = m;
        if ((m as any).taga_id) map[String((m as any).taga_id).toLowerCase()] = m;
        if (m.email) map[m.email.toLowerCase()] = m;
      });
      setMemberMap(map);
    } catch (err) {
      console.error(err);
      toast.error('Failed to load account deletion requests');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRequestsData();
  }, [open]);

  const pendingRequests = requests.filter((r) => r.status === 'pending');
  const pendingCount = pendingRequests.length;

  const filteredRequests = requests.filter((r) => {
    if (statusFilter === 'all') return true;
    return r.status === statusFilter;
  });

  const handleApproveConfirm = async () => {
    if (!approvingReq) return;
    setIsApproving(true);
    try {
      const res = await approveDeletionRequest(approvingReq.id);
      toast.success(res.message || 'Account deletion approved and processed successfully');
      setApprovingReq(null);
      await fetchRequestsData();
    } catch (err: any) {
      const msg = err.message || 'Failed to approve account deletion';
      toast.error(msg);
    } finally {
      setIsApproving(false);
    }
  };

  const handleRejectSubmit = async () => {
    if (!rejectingReq) return;
    const trimmed = rejectRemarks.trim();
    if (!trimmed) {
      setRejectError('Remarks are mandatory for rejecting a deletion request');
      return;
    }

    setIsRejecting(true);
    try {
      const res = await rejectDeletionRequest(rejectingReq.id, trimmed);
      toast.success(res.message || 'Account deletion request rejected');
      setRejectingReq(null);
      setRejectRemarks('');
      setRejectError('');
      await fetchRequestsData();
    } catch (err: any) {
      const msg = err.message || 'Failed to reject account deletion request';
      toast.error(msg);
    } finally {
      setIsRejecting(false);
    }
  };

  return (
    <>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogTrigger asChild>
          <Button
            variant="outline"
            className="w-full justify-start relative text-red-700 hover:text-red-800 hover:bg-red-50 border-red-200"
            data-testid="testid-pending-deletion-requests-button"
          >
            <UserX className="w-4 h-4 mr-2 text-red-600" />
            Account Deletion Requests
            {pendingCount > 0 && (
              <Badge
                variant="destructive"
                className="absolute right-2 top-1/2 -translate-y-1/2 rounded-full px-2 py-0.5 text-xs bg-red-600"
                data-testid="testid-deletion-requests-count-badge"
              >
                {pendingCount}
              </Badge>
            )}
          </Button>
        </DialogTrigger>

        <DialogContent className="sm:max-w-4xl md:max-w-5xl max-h-[88vh] overflow-hidden flex flex-col p-0">
          {/* Header */}
          <DialogHeader className="p-6 pb-4 border-b bg-red-50/50 shrink-0">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="p-2.5 bg-red-100 rounded-lg">
                  <UserX className="w-6 h-6 text-red-600" />
                </div>
                <div>
                  <DialogTitle className="text-xl text-red-950">
                    Account Deletion Requests
                  </DialogTitle>
                  <DialogDescription className="text-xs text-red-800/80">
                    Review, approve, or reject member account retirement and permanent deletion requests.
                  </DialogDescription>
                </div>
              </div>
              {pendingCount > 0 && (
                <Badge variant="destructive" className="bg-red-600">
                  {pendingCount} Pending
                </Badge>
              )}
            </div>

            {/* Filter Tabs */}
            <div className="flex gap-2 pt-3">
              {(['pending', 'all', 'approved', 'rejected'] as const).map((tab) => (
                <Button
                  key={tab}
                  variant={statusFilter === tab ? 'default' : 'outline'}
                  size="sm"
                  className={`text-xs h-8 capitalize ${
                    statusFilter === tab
                      ? 'bg-red-600 hover:bg-red-700 text-white'
                      : 'border-slate-200 text-slate-700 hover:bg-slate-100'
                  }`}
                  onClick={() => setStatusFilter(tab)}
                >
                  {tab}
                  {tab === 'pending' && pendingCount > 0 && (
                    <span className="ml-1.5 px-1.5 py-0.2 rounded-full bg-white/30 text-white text-[10px]">
                      {pendingCount}
                    </span>
                  )}
                </Button>
              ))}
            </div>
          </DialogHeader>

          {/* Body */}
          <div className="overflow-y-auto p-6 bg-slate-50 flex-1 space-y-4">
            {loading ? (
              <div className="text-center py-12 text-muted-foreground flex flex-col items-center justify-center space-y-2">
                <Loader2 className="w-8 h-8 animate-spin text-red-600" />
                <p className="text-sm">Loading deletion requests...</p>
              </div>
            ) : filteredRequests.length === 0 ? (
              <div className="text-center py-12 bg-white border rounded-xl shadow-xs space-y-2">
                <Check className="w-10 h-10 text-emerald-500 mx-auto opacity-70" />
                <p className="text-sm font-semibold text-slate-800">
                  No {statusFilter === 'all' ? '' : statusFilter} deletion requests found
                </p>
                <p className="text-xs text-muted-foreground">
                  All member accounts are processed or in regular standing.
                </p>
              </div>
            ) : (
              <div className="space-y-4">
                {filteredRequests.map((req) => {
                  const tagaId = req.tagaId || req.taga_id || '';
                  const enrichedMember =
                    memberMap[req.member_id?.toLowerCase()] ||
                    memberMap[tagaId.toLowerCase()] ||
                    memberMap[req.email?.toLowerCase()];

                  const mobile =
                    req.mobile_number ||
                    req.mobile ||
                    enrichedMember?.mobile_number ||
                    '-';

                  const paymentStatus =
                    req.subscription_status ||
                    req.payment_status ||
                    enrichedMember?.payment_status ||
                    '';

                  const isSubscriptionPaid =
                    req.is_paid === true ||
                    req.subscription_active === true ||
                    paymentStatus.toLowerCase() === 'paid' ||
                    paymentStatus.toLowerCase() === 'active' ||
                    enrichedMember?.payment_status?.toLowerCase() === 'paid' ||
                    enrichedMember?.membership_status?.toLowerCase() === 'active';

                  return (
                    <div
                      key={req.id}
                      className="bg-white border rounded-xl p-5 shadow-xs transition-shadow hover:shadow-sm space-y-4"
                      data-testid={`deletion-request-card-${req.id}`}
                    >
                      {/* Top Row: Name, TAGA ID, Status */}
                      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b pb-3">
                        <div className="flex items-center gap-2.5">
                          <div className="w-9 h-9 rounded-full bg-slate-100 flex items-center justify-center text-slate-700 font-semibold text-sm">
                            {req.member_name ? req.member_name.charAt(0).toUpperCase() : 'M'}
                          </div>
                          <div>
                            <div className="font-semibold text-slate-900 flex items-center gap-2">
                              <span>{req.member_name || 'Unknown Member'}</span>
                              <Badge variant="outline" className="text-xs bg-slate-50 font-mono">
                                TAGA ID: {tagaId || '-'}
                              </Badge>
                            </div>
                            <div className="text-xs text-muted-foreground">
                              Requested ID: {req.id}
                            </div>
                          </div>
                        </div>

                        <div className="flex items-center gap-2">
                          {req.status === 'pending' && (
                            <Badge className="bg-amber-100 text-amber-800 border-amber-300 font-medium">
                              <Clock className="w-3 h-3 mr-1" /> Pending Review
                            </Badge>
                          )}
                          {req.status === 'approved' && (
                            <Badge className="bg-emerald-100 text-emerald-800 border-emerald-300 font-medium">
                              <Check className="w-3 h-3 mr-1" /> Approved
                            </Badge>
                          )}
                          {req.status === 'rejected' && (
                            <Badge className="bg-red-100 text-red-800 border-red-300 font-medium">
                              <X className="w-3 h-3 mr-1" /> Rejected
                            </Badge>
                          )}
                        </div>
                      </div>

                      {/* Detail Grid: Email, Mobile, Submission Date, Annual Subscription Status */}
                      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-3 text-xs">
                        <div className="flex items-center gap-2 text-slate-600">
                          <Mail className="w-4 h-4 text-slate-400 shrink-0" />
                          <span className="truncate" title={req.email}>
                            {req.email || '-'}
                          </span>
                        </div>

                        <div className="flex items-center gap-2 text-slate-600">
                          <Phone className="w-4 h-4 text-slate-400 shrink-0" />
                          <span>{mobile}</span>
                        </div>

                        <div className="flex items-center gap-2 text-slate-600">
                          <Calendar className="w-4 h-4 text-slate-400 shrink-0" />
                          <span>
                            {req.requested_at
                              ? new Date(req.requested_at).toLocaleString('en-IN', {
                                  dateStyle: 'medium',
                                  timeStyle: 'short'
                                })
                              : '-'}
                          </span>
                        </div>

                        <div className="flex items-center gap-2">
                          <CreditCard className="w-4 h-4 text-slate-400 shrink-0" />
                          <span className="text-slate-500 mr-1">Subscription:</span>
                          {isSubscriptionPaid ? (
                            <Badge variant="outline" className="bg-emerald-50 text-emerald-700 border-emerald-200 text-[11px] py-0 font-medium">
                              Active / Paid
                            </Badge>
                          ) : (
                            <Badge variant="outline" className="bg-amber-50 text-amber-700 border-amber-200 text-[11px] py-0 font-medium">
                              Inactive / Unpaid
                            </Badge>
                          )}
                        </div>
                      </div>

                      {/* Reason Box */}
                      {req.reason && (
                        <div className="bg-slate-50 rounded-lg p-3 text-xs text-slate-700 border border-slate-100">
                          <span className="font-semibold text-slate-900 block mb-1">
                            Reason for Deletion:
                          </span>
                          <p className="whitespace-pre-line">{req.reason}</p>
                        </div>
                      )}

                      {/* Admin Remarks if rejected or processed */}
                      {req.remarks && (
                        <div className="bg-red-50/60 rounded-lg p-3 text-xs text-red-900 border border-red-100">
                          <span className="font-semibold text-red-950 block mb-1">
                            Rejection Remarks:
                          </span>
                          <p className="whitespace-pre-line">{req.remarks}</p>
                          {req.processed_at && (
                            <span className="text-[10px] text-muted-foreground block mt-1">
                              Processed on: {new Date(req.processed_at).toLocaleString('en-IN')} by {req.processed_by || 'Admin'}
                            </span>
                          )}
                        </div>
                      )}

                      {/* Action Buttons for Pending Requests */}
                      {req.status === 'pending' && (
                        <div className="flex justify-end gap-2 pt-2 border-t border-slate-100">
                          <Button
                            variant="outline"
                            size="sm"
                            className="text-red-700 border-red-200 hover:bg-red-50 hover:text-red-800 text-xs h-8"
                            onClick={() => {
                              setRejectingReq(req);
                              setRejectRemarks('');
                              setRejectError('');
                            }}
                            data-testid={`reject-button-${req.id}`}
                          >
                            <X className="w-3.5 h-3.5 mr-1 text-red-600" />
                            Reject Request
                          </Button>

                          <Button
                            variant="default"
                            size="sm"
                            className="bg-emerald-600 hover:bg-emerald-700 text-white text-xs h-8"
                            onClick={() => setApprovingReq(req)}
                            data-testid={`approve-button-${req.id}`}
                          >
                            <Check className="w-3.5 h-3.5 mr-1" />
                            Approve Deletion
                          </Button>
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>

      {/* ===================== APPROVE CONFIRMATION ALERT DIALOG ===================== */}
      <AlertDialog open={!!approvingReq} onOpenChange={(open) => !open && setApprovingReq(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="text-red-600 flex items-center gap-2">
              <ShieldAlert className="w-5 h-5" />
              Approve Account Deletion
            </AlertDialogTitle>
            <AlertDialogDescription className="space-y-2 text-slate-700 text-sm">
              <p>
                Are you sure you want to approve account deletion for{' '}
                <span className="font-semibold text-slate-900">
                  {approvingReq?.member_name}
                </span>{' '}
                ({approvingReq?.tagaId || approvingReq?.taga_id})?
              </p>
              <div className="bg-red-50 border border-red-200 rounded p-3 text-xs text-red-900 space-y-1">
                <p className="font-semibold">This action will immediately:</p>
                <ul className="list-disc pl-4 space-y-0.5">
                  <li>Retire the member's UID / TAGA ID per policy</li>
                  <li>Permanently anonymize credentials &amp; end active sessions</li>
                  <li>Cancel upcoming TAGA Towers bookings and release room inventory</li>
                  <li>Safely archive member record for historical &amp; audit compliance</li>
                </ul>
              </div>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isApproving}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleApproveConfirm}
              disabled={isApproving}
              className="bg-red-600 hover:bg-red-700 text-white"
              data-testid="testid-confirm-approve-deletion-button"
            >
              {isApproving ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Approving...
                </>
              ) : (
                'Approve & Delete Account'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* ===================== REJECT MODAL WITH MANDATORY REMARKS ===================== */}
      <Dialog open={!!rejectingReq} onOpenChange={(open) => !open && setRejectingReq(null)}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-red-700">
              <AlertTriangle className="w-5 h-5 text-red-600" />
              Reject Deletion Request
            </DialogTitle>
            <DialogDescription className="text-xs text-muted-foreground">
              Provide mandatory administrative remarks explaining why this deletion request is rejected.
              An official email notification will be dispatched to the member.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-3 py-2">
            <div>
              <p className="text-xs text-slate-700 mb-1">
                <span className="font-semibold">Member: </span>
                {rejectingReq?.member_name} ({rejectingReq?.tagaId || rejectingReq?.taga_id})
              </p>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="reject-remarks" className="text-xs font-semibold text-slate-800">
                Administrative Remarks <span className="text-red-500">*</span>
              </Label>
              <Textarea
                id="reject-remarks"
                placeholder="Enter mandatory reason for rejection (e.g. pending dues, active stay, policy requirement)..."
                rows={4}
                value={rejectRemarks}
                onChange={(e) => {
                  setRejectRemarks(e.target.value);
                  if (rejectError) setRejectError('');
                }}
                disabled={isRejecting}
                className={rejectError ? 'border-red-500 focus-visible:ring-red-400' : ''}
                data-testid="testid-rejection-remarks-textarea"
              />
              {rejectError && (
                <p className="text-xs text-red-600 font-medium">{rejectError}</p>
              )}
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setRejectingReq(null)}
              disabled={isRejecting}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              className="bg-red-600 hover:bg-red-700 text-white"
              onClick={handleRejectSubmit}
              disabled={isRejecting || !rejectRemarks.trim()}
              data-testid="testid-confirm-reject-deletion-button"
            >
              {isRejecting ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Rejecting...
                </>
              ) : (
                'Reject Deletion Request'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
