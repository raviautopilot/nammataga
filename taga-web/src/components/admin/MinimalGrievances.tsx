import React, { useState, useEffect, useMemo } from 'react';
import API_BASE_URL from '../../config/api';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogTrigger } from '../ui/dialog';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
import { ChevronDown, ChevronUp, AlertCircle, MessageSquare, ChevronLeft, ChevronRight, User, Mail, Phone, Calendar, Tag, Flag, MessageCircle } from 'lucide-react';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '../ui/collapsible';

interface Grievance {
  id: string;
  memberName?: string;
  memberEmail?: string;
  subject: string;
  category: string;
  priority: string;
  description: string;
  contactPhone: string;
  preferredResponse: string;
  status: string;
  submittedDate: string;
}

const ITEMS_PER_PAGE = 8;

export default function MinimalGrievances() {
  const [open, setOpen] = useState(false);
  const [grievances, setGrievances] = useState<Grievance[]>([]);
  const [openItems, setOpenItems] = useState<Record<string, boolean>>({});
  const [currentPage, setCurrentPage] = useState(1);

  const fetchGrievances = async () => {
    try {
      const token = localStorage.getItem('admin_token');
      const res = await fetch(`${API_BASE_URL}/grievances`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (res.ok) {
        const data = await res.json();
        setGrievances(data || []);
      }
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    fetchGrievances();
  }, []);

  const markAsRead = async (g: Grievance) => {
    if (g.status === 'Read') return;
    try {
      const token = localStorage.getItem('admin_token');
      await fetch(`${API_BASE_URL}/grievances/${g.id}`, {
        method: 'PUT',
        headers: { 
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({ ...g, status: 'Read' })
      });
      setGrievances(prev => prev.map(item => item.id === g.id ? { ...item, status: 'Read' } : item));
    } catch (err) {
      console.error('Failed to mark read', err);
    }
  };

  const toggleOpen = (g: Grievance) => {
    const isNowOpen = !openItems[g.id];
    setOpenItems(prev => ({ ...prev, [g.id]: isNowOpen }));
    if (isNowOpen) markAsRead(g);
  };

  const unreadCount = grievances.filter(g => g.status === 'Pending').length;

  const sorted = useMemo(
    () => [...grievances].sort((a, b) => new Date(b.submittedDate).getTime() - new Date(a.submittedDate).getTime()),
    [grievances]
  );

  const totalPages = Math.max(1, Math.ceil(sorted.length / ITEMS_PER_PAGE));
  const paginated = sorted.slice((currentPage - 1) * ITEMS_PER_PAGE, currentPage * ITEMS_PER_PAGE);

  // Reset to page 1 when grievances change
  useEffect(() => {
    setCurrentPage(1);
  }, [grievances.length]);

  const priorityColor = (p: string) => {
    switch (p.toLowerCase()) {
      case 'high': return 'bg-red-50 text-red-700 border-red-200';
      case 'medium': return 'bg-amber-50 text-amber-700 border-amber-200';
      default: return 'bg-emerald-50 text-emerald-700 border-emerald-200';
    }
  };

  return (
    <Dialog open={open} onOpenChange={(val: boolean) => { setOpen(val); if (val) { fetchGrievances(); setOpenItems({}); } }}>
      <DialogTrigger asChild>
        <Button variant="outline" className="w-full justify-start relative shadow-sm">
          <MessageSquare className="w-4 h-4 mr-2" />
          Member Grievances
          {unreadCount > 0 && (
            <Badge variant="destructive" className="ml-auto flex h-5 w-5 items-center justify-center rounded-full p-0 text-[10px]">
              {unreadCount}
            </Badge>
          )}
        </Button>
      </DialogTrigger>
      
      <DialogContent className="max-w-2xl max-h-[85vh] flex flex-col bg-white border-slate-200 shadow-2xl p-0 gap-0">
        {/* Header */}
        <div className="px-5 pt-5 pb-4 border-b border-slate-100">
          <DialogHeader className="space-y-1">
            <DialogTitle className="flex items-center gap-2 text-lg font-semibold text-slate-800">
              <div className="w-8 h-8 rounded-lg bg-blue-50 flex items-center justify-center">
                <AlertCircle className="w-4 h-4 text-blue-600" />
              </div>
              Member Grievances
              {grievances.length > 0 && (
                <span className="text-xs font-normal text-slate-400 ml-1">({grievances.length})</span>
              )}
            </DialogTitle>
            <DialogDescription className="text-xs text-slate-500">
              Click on a grievance to expand details. Unread items are highlighted.
            </DialogDescription>
          </DialogHeader>

          {/* Stats bar */}
          {grievances.length > 0 && (
            <div className="flex gap-4 mt-3">
              <div className="flex items-center gap-1.5 text-xs text-slate-500">
                <div className="w-2 h-2 rounded-full bg-blue-500" />
                <span>{unreadCount} unread</span>
              </div>
              <div className="flex items-center gap-1.5 text-xs text-slate-500">
                <div className="w-2 h-2 rounded-full bg-slate-300" />
                <span>{grievances.length - unreadCount} read</span>
              </div>
            </div>
          )}
        </div>

        {/* List */}
        <div className="flex-1 overflow-y-auto px-5 py-3">
          {grievances.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-slate-400">
              <MessageSquare className="w-10 h-10 mb-3 opacity-40" />
              <p className="text-sm font-medium">No grievances yet</p>
              <p className="text-xs mt-1">When members submit issues, they'll appear here.</p>
            </div>
          ) : (
            <div className="space-y-2">
              {paginated.map(g => (
                <Collapsible
                  key={g.id}
                  open={openItems[g.id]}
                  onOpenChange={() => toggleOpen(g)}
                  className="group"
                >
                  {/* Compact trigger row */}
                  <CollapsibleTrigger className={`w-full flex items-center gap-3 px-3.5 py-2.5 text-left rounded-lg transition-all cursor-pointer
                    ${g.status === 'Pending'
                      ? 'bg-blue-50/60 hover:bg-blue-50 border border-blue-100'
                      : 'bg-slate-50/60 hover:bg-slate-50 border border-slate-100'
                    }
                    ${openItems[g.id] ? 'rounded-b-none border-b-0' : ''}
                  `}>
                    {/* Unread dot */}
                    <div className="w-4 flex-shrink-0 flex justify-center">
                      {g.status === 'Pending' && <div className="w-2 h-2 rounded-full bg-blue-500 animate-pulse" />}
                    </div>

                    {/* Main info */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className={`text-sm truncate ${g.status === 'Pending' ? 'font-semibold text-slate-900' : 'font-medium text-slate-600'}`}>
                          {g.subject}
                        </span>
                      </div>
                      <div className="flex items-center gap-2 mt-0.5">
                        <span className="text-[11px] text-slate-400">{g.memberName || 'Unknown'}</span>
                        <span className="text-[11px] text-slate-300">·</span>
                        <span className="text-[11px] text-slate-400">{new Date(g.submittedDate).toLocaleDateString('en-IN', { day: 'numeric', month: 'short' })}</span>
                      </div>
                    </div>

                    {/* Priority + chevron */}
                    <Badge variant="outline" className={`text-[10px] px-1.5 py-0 h-5 font-semibold border ${priorityColor(g.priority)}`}>
                      {g.priority}
                    </Badge>
                    {openItems[g.id]
                      ? <ChevronUp className="w-3.5 h-3.5 text-slate-400 flex-shrink-0" />
                      : <ChevronDown className="w-3.5 h-3.5 text-slate-400 flex-shrink-0" />
                    }
                  </CollapsibleTrigger>

                  {/* Expanded detail card */}
                  <CollapsibleContent className={`border border-t-0 rounded-b-lg px-4 pb-4 pt-3
                    ${g.status === 'Pending' ? 'border-blue-100 bg-white' : 'border-slate-100 bg-white'}
                  `}>
                    {/* Subject label */}
                    <div className="mb-3">
                      <div className="flex items-center gap-1.5 mb-1">
                        <Tag className="w-3 h-3 text-slate-400" />
                        <span className="text-[10px] font-semibold text-slate-400 uppercase tracking-wider">Subject</span>
                      </div>
                      <p className="text-sm font-semibold text-slate-800">{g.subject}</p>
                    </div>

                    {/* Description */}
                    <div className="mb-3">
                      <div className="flex items-center gap-1.5 mb-1">
                        <MessageCircle className="w-3 h-3 text-slate-400" />
                        <span className="text-[10px] font-semibold text-slate-400 uppercase tracking-wider">Description</span>
                      </div>
                      <p className="text-xs text-slate-700 leading-relaxed whitespace-pre-wrap bg-slate-50 rounded-md px-3 py-2.5 border border-slate-100">
                        {g.description}
                      </p>
                    </div>

                    {/* Metadata grid */}
                    <div className="grid grid-cols-2 gap-x-6 gap-y-2.5 py-2.5 px-3 bg-slate-50/50 rounded-md border border-slate-100">
                      <div className="flex items-start gap-1.5 min-w-0">
                        <User className="w-3 h-3 text-slate-400 mt-0.5 flex-shrink-0" />
                        <div className="min-w-0">
                          <span className="block text-[10px] font-semibold text-slate-400 uppercase tracking-wider">Member</span>
                          <span className="text-xs text-slate-700 font-medium truncate block">{g.memberName || 'Unknown'}</span>
                        </div>
                      </div>
                      <div className="flex items-start gap-1.5 min-w-0">
                        <Phone className="w-3 h-3 text-slate-400 mt-0.5 flex-shrink-0" />
                        <div className="min-w-0">
                          <span className="block text-[10px] font-semibold text-slate-400 uppercase tracking-wider">Phone</span>
                          <span className="text-xs text-slate-700 font-medium">{g.contactPhone}</span>
                        </div>
                      </div>
                      <div className="flex items-start gap-1.5 min-w-0 col-span-2">
                        <Mail className="w-3 h-3 text-slate-400 mt-0.5 flex-shrink-0" />
                        <div className="min-w-0">
                          <span className="block text-[10px] font-semibold text-slate-400 uppercase tracking-wider">Email</span>
                          <span className="text-xs text-slate-700 font-medium truncate block">{g.memberEmail || 'N/A'}</span>
                        </div>
                      </div>
                      <div className="flex items-start gap-1.5 min-w-0">
                        <Calendar className="w-3 h-3 text-slate-400 mt-0.5 flex-shrink-0" />
                        <div className="min-w-0">
                          <span className="block text-[10px] font-semibold text-slate-400 uppercase tracking-wider">Submitted</span>
                          <span className="text-xs text-slate-700 font-medium">{new Date(g.submittedDate).toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' })}</span>
                        </div>
                      </div>
                      <div className="flex items-start gap-1.5 min-w-0">
                        <Flag className="w-3 h-3 text-slate-400 mt-0.5 flex-shrink-0" />
                        <div className="min-w-0">
                          <span className="block text-[10px] font-semibold text-slate-400 uppercase tracking-wider">Priority</span>
                          <span className={`text-xs font-bold ${g.priority.toLowerCase() === 'high' ? 'text-red-600' : g.priority.toLowerCase() === 'medium' ? 'text-amber-600' : 'text-emerald-600'}`}>
                            {g.priority}
                          </span>
                        </div>
                      </div>
                      <div className="flex items-start gap-1.5 min-w-0 col-span-2">
                        <Tag className="w-3 h-3 text-slate-400 mt-0.5 flex-shrink-0" />
                        <div className="min-w-0">
                          <span className="block text-[10px] font-semibold text-slate-400 uppercase tracking-wider">Category</span>
                          <span className="text-xs text-slate-700 font-medium">{g.category}</span>
                        </div>
                      </div>
                    </div>

                    {/* Preferred Response */}
                    {g.preferredResponse && (
                      <div className="mt-3">
                        <div className="flex items-center gap-1.5 mb-1">
                          <MessageSquare className="w-3 h-3 text-indigo-400" />
                          <span className="text-[10px] font-semibold text-slate-400 uppercase tracking-wider">Preferred Response</span>
                        </div>
                        <p className="text-xs text-indigo-700 font-medium bg-indigo-50/60 px-3 py-2 rounded-md border border-indigo-100">
                          {g.preferredResponse}
                        </p>
                      </div>
                    )}
                  </CollapsibleContent>
                </Collapsible>
              ))}
            </div>
          )}
        </div>

        {/* Pagination footer */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between px-5 py-3 border-t border-slate-100 bg-slate-50/50">
            <span className="text-xs text-slate-400">
              Page {currentPage} of {totalPages}
            </span>
            <div className="flex items-center gap-1">
              <Button
                variant="ghost"
                size="sm"
                disabled={currentPage === 1}
                onClick={() => setCurrentPage(p => p - 1)}
                className="h-7 w-7 p-0"
              >
                <ChevronLeft className="w-4 h-4" />
              </Button>
              {Array.from({ length: totalPages }, (_, i) => i + 1).map(p => (
                <Button
                  key={p}
                  variant={p === currentPage ? 'default' : 'ghost'}
                  size="sm"
                  onClick={() => setCurrentPage(p)}
                  className={`h-7 w-7 p-0 text-xs ${p === currentPage ? 'bg-blue-600 text-white hover:bg-blue-700' : 'text-slate-600'}`}
                >
                  {p}
                </Button>
              ))}
              <Button
                variant="ghost"
                size="sm"
                disabled={currentPage === totalPages}
                onClick={() => setCurrentPage(p => p + 1)}
                className="h-7 w-7 p-0"
              >
                <ChevronRight className="w-4 h-4" />
              </Button>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
