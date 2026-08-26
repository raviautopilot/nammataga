import React, { useState, useEffect } from 'react';
import API_BASE_URL from '../../config/api';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogTrigger } from '../ui/dialog';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
import { toast } from 'sonner';
import { FileEdit, Check, X, ChevronDown, ChevronUp, Save, Loader2 } from 'lucide-react';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '../ui/collapsible';
import { Input } from '../ui/input';

interface FieldEditRequest {
  id: string;
  requestGroupId: string;
  memberId: string;
  email: string;
  memberName: string;
  field: string;
  oldValue: string;
  newValue: string;
  memberRemarks: string;
  status: string;
  createdAt: string;
}

export default function AdminEditRequestsButton() {
  const [requests, setRequests] = useState<FieldEditRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [openGroups, setOpenGroups] = useState<Record<string, boolean>>({});
  
  // Local state to track which fields the admin is approving/rejecting before saving
  const [pendingDecisions, setPendingDecisions] = useState<Record<string, 'approved' | 'rejected'>>({});
  const [adminRemarks, setAdminRemarks] = useState<Record<string, string>>({});
  const [savingMemberId, setSavingMemberId] = useState<string | null>(null);

  const fetchRequests = async () => {
    try {
      const token = localStorage.getItem('admin_token');
      if (!token) return;
      const res = await fetch(`${API_BASE_URL}/admin/edit-requests`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      const data = await res.json();
      setRequests(data || []);
      setPendingDecisions({});
      setAdminRemarks({});
    } catch (err) {
      console.error(err);
      toast.error("Failed to load edit requests");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRequests();
  }, [open]);

  const toggleGroup = (memberId: string) => {
    setOpenGroups(prev => ({ ...prev, [memberId]: !prev[memberId] }));
  };

  const handleDecision = (id: string, decision: 'approved' | 'rejected') => {
    setPendingDecisions(prev => ({ ...prev, [id]: decision }));
  };

  const handleBulkSave = async (memberId: string, groupFields: FieldEditRequest[]) => {
    // Only process fields that have a decision
    const itemsToProcess = groupFields
      .filter(f => pendingDecisions[f.id])
      .map(f => ({
        id: f.id,
        status: pendingDecisions[f.id],
        adminRemarks: adminRemarks[f.id] || ''
      }));

    if (itemsToProcess.length === 0) {
      toast.info("No decisions made for this member yet.");
      return;
    }

    try {
      setSavingMemberId(memberId);
      const token = localStorage.getItem('admin_token');
      const res = await fetch(`${API_BASE_URL}/admin/edit-requests/bulk-process`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({ items: itemsToProcess })
      });
      
      if (!res.ok) throw new Error("Failed to process");
      
      toast.success("Changes saved and email sent to member successfully");
      fetchRequests(); // Reload
    } catch (err) {
      toast.error("Failed to save changes");
    } finally {
      setSavingMemberId(null);
    }
  };

  // Group requests by Member
  const groupedRequests = requests.reduce((acc, req) => {
    if (!acc[req.memberId]) {
      acc[req.memberId] = {
        memberName: req.memberName,
        email: req.email,
        fields: []
      };
    }
    acc[req.memberId].fields.push(req);
    return acc;
  }, {} as Record<string, { memberName: string; email: string; fields: FieldEditRequest[] }>);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" className="w-full justify-start relative">
          <FileEdit className="w-4 h-4 mr-2" />
          Pending Edit Requests
          {requests.length > 0 && (
            <Badge variant="destructive" className="absolute right-2 top-1/2 -translate-y-1/2 rounded-full px-2 py-0.5 text-xs">
              {requests.length}
            </Badge>
          )}
        </Button>
      </DialogTrigger>
      
      <DialogContent className="sm:max-w-5xl md:max-w-6xl max-h-[85vh] overflow-hidden flex flex-col p-0">
        <DialogHeader className="p-6 pb-4 border-b bg-emerald-50/50 shrink-0">
          <div className="flex items-center gap-2">
            <div className="p-2 bg-emerald-100 rounded-lg">
              <FileEdit className="w-5 h-5 text-emerald-700" />
            </div>
            <div>
              <DialogTitle className="text-xl">Pending Edit Requests</DialogTitle>
              <DialogDescription>Review and approve member profile changes</DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div className="overflow-y-auto p-6 bg-slate-50 flex-1 space-y-4">
          {loading ? (
            <div className="text-center py-10 text-muted-foreground">Loading requests...</div>
          ) : Object.keys(groupedRequests).length === 0 ? (
            <div className="text-center py-10 bg-white border rounded-xl shadow-sm">
              <Check className="w-10 h-10 text-emerald-500 mx-auto mb-3 opacity-50" />
              <p className="text-muted-foreground">No pending edit requests.</p>
            </div>
          ) : (
            Object.entries(groupedRequests).map(([memberId, group]) => {
              const isGroupOpen = openGroups[memberId];
              const decisionsMade = group.fields.filter(f => pendingDecisions[f.id]).length;
              const isSaving = savingMemberId === memberId;

              return (
                <Collapsible 
                  key={memberId} 
                  open={isGroupOpen} 
                  onOpenChange={() => toggleGroup(memberId)}
                  className="bg-white border rounded-xl shadow-sm overflow-hidden"
                >
                  <div className="bg-slate-50 hover:bg-slate-100 transition-colors border-b px-4 py-3 flex items-center justify-between cursor-pointer" onClick={() => toggleGroup(memberId)}>
                    <div className="flex items-center gap-3">
                      <CollapsibleTrigger asChild>
                        <Button variant="ghost" size="sm" className="p-0 h-8 w-8 hover:bg-slate-200">
                          {isGroupOpen ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
                        </Button>
                      </CollapsibleTrigger>
                      <div className="w-10 h-10 bg-blue-100 text-blue-700 rounded-full flex items-center justify-center font-bold">
                        {group.memberName.charAt(0).toUpperCase()}
                      </div>
                      <div>
                        <h4 className="font-semibold text-base">{group.memberName}</h4>
                        <p className="text-sm text-muted-foreground">{group.email}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-4">
                      <Badge variant="secondary" className="bg-emerald-100 text-emerald-800">
                        {group.fields.length} pending change{group.fields.length !== 1 ? 's' : ''}
                      </Badge>
                      
                      {/* Save Button only shows when expanded */}
                      {isGroupOpen && (
                        <Button 
                          size="sm" 
                          onClick={(e: React.MouseEvent) => { e.stopPropagation(); handleBulkSave(memberId, group.fields); }}
                          disabled={decisionsMade === 0 || isSaving}
                          className={`min-w-[140px] text-white ${isSaving ? 'bg-blue-500 opacity-80 cursor-not-allowed' : 'bg-blue-600 hover:bg-blue-700'}`}
                        >
                          {isSaving ? (
                            <>
                              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                              Saving...
                            </>
                          ) : (
                            <>
                              <Save className="w-4 h-4 mr-2" />
                              Save & Publish
                            </>
                          )}
                        </Button>
                      )}
                    </div>
                  </div>
                  
                  <CollapsibleContent>
                    <div className="p-0 overflow-x-auto">
                      <table className="w-full text-sm min-w-[900px]">
                        <thead>
                          <tr className="bg-slate-50/80 text-slate-500 border-b text-left text-xs uppercase tracking-wider">
                            <th className="p-4 font-semibold w-1/5">Field</th>
                            <th className="p-4 font-semibold w-[20%]">Old Value</th>
                            <th className="p-4 font-semibold w-[20%]">Requested Value</th>
                            <th className="p-4 font-semibold w-1/5">Member Remarks</th>
                            <th className="p-4 font-semibold w-1/4">Decision & Remarks</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y">
                          {group.fields.map(req => {
                            const decision = pendingDecisions[req.id];
                            return (
                              <tr key={req.id} className={`transition-colors ${decision === 'approved' ? 'bg-emerald-50/30' : decision === 'rejected' ? 'bg-red-50/30' : 'hover:bg-slate-50'}`}>
                                <td className="p-4">
                                  <Badge variant="outline" className="capitalize bg-white font-medium text-slate-700 border-slate-300">
                                    {req.field.replace(/([A-Z])/g, ' $1').replace(/_/g, ' ').trim()}
                                  </Badge>
                                </td>
                                <td className="p-4 text-amber-700 line-through opacity-80 break-words font-mono text-xs max-w-[200px]">
                                  {req.oldValue || <span className="italic text-muted-foreground">Empty</span>}
                                </td>
                                <td className="p-4 text-emerald-700 font-semibold break-words font-mono text-xs max-w-[200px]">
                                  {req.newValue}
                                </td>
                                <td className="p-4 text-slate-600 text-sm whitespace-pre-wrap max-w-[200px]">
                                  {req.memberRemarks || <span className="italic text-muted-foreground opacity-70">No remarks</span>}
                                </td>
                                <td className="p-4">
                                  <div className="space-y-2">
                                    <div className="flex items-center gap-2">
                                      <Button 
                                        size="sm" 
                                        variant={decision === 'approved' ? 'default' : 'outline'} 
                                        className={`h-8 px-3 flex-1 ${decision === 'approved' ? 'bg-emerald-600 hover:bg-emerald-700 text-white' : 'hover:bg-emerald-50 hover:text-emerald-700'}`} 
                                        onClick={() => handleDecision(req.id, 'approved')}
                                      >
                                        <Check className="w-4 h-4 mr-1" /> Approve
                                      </Button>
                                      <Button 
                                        size="sm" 
                                        variant={decision === 'rejected' ? 'default' : 'outline'} 
                                        className={`h-8 px-3 flex-1 ${decision === 'rejected' ? 'bg-red-600 hover:bg-red-700 text-white' : 'hover:bg-red-50 hover:text-red-700'}`} 
                                        onClick={() => handleDecision(req.id, 'rejected')}
                                      >
                                        <X className="w-4 h-4 mr-1" /> Reject
                                      </Button>
                                    </div>
                                    <Input 
                                      placeholder="Admin remarks (optional)" 
                                      className="h-8 text-xs bg-white"
                                      value={adminRemarks[req.id] || ''}
                                      onChange={(e) => setAdminRemarks(prev => ({...prev, [req.id]: e.target.value}))}
                                    />
                                  </div>
                                </td>
                              </tr>
                            );
                          })}
                        </tbody>
                      </table>
                    </div>
                  </CollapsibleContent>
                </Collapsible>
              );
            })
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
