import React, { useState, useEffect } from 'react';
import API_BASE_URL from '../config/api';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from './ui/dialog';
import { Textarea } from './ui/textarea';
import { Search, Eye, RefreshCw, Pencil, Trash2, AlertTriangle, Loader2, Save, X, Download } from 'lucide-react';
import { toast } from 'sonner';
import {
  getMembersList,
  getMemberDistricts,
  MemberListItem,
  DistrictOption,
  exportMembersToExcel,
} from '../api/adminContent';


interface MemberListTableProps {
  onUpdateStats?: (stats: any) => void;
}

const districts = [
  'Ariyalur', 'Chengalpattu', 'Chennai', 'Coimbatore', 'Cuddalore', 'Dharmapuri',
  'Dindigul', 'Erode', 'Kallakurichi', 'Kanchipuram', 'Kanyakumari', 'Karur',
  'Krishnagiri', 'Madurai', 'Mayiladuthurai', 'Nagapattinam', 'Namakkal', 'Nilgiris',
  'Perambalur', 'Pudukkottai', 'Ramanathapuram', 'Ranipet', 'Salem', 'Sivaganga',
  'Tenkasi', 'Thanjavur', 'Theni', 'Thoothukudi', 'Tiruchirappalli', 'Tirunelveli',
  'Tirupathur', 'Tiruppur', 'Tiruvallur', 'Tiruvannamalai', 'Tiruvarur', 'Vellore',
  'Viluppuram', 'Virudhunagar'
];

// ---- Editable fields interface ----
interface EditMemberForm {
  name: string;
  initial: string;
  gender: string;
  father_name: string;
  mother_name: string;
  educational_qualification: string;
  designation: string;
  working_district: string;
  native_district: string;
  recruitment_batch: string;
  seniority_number: string;
  residential_address: string;
  permanent_address: string;
  date_of_birth: string;
  mobile_number: string;
  emailId: string;
  tbf_number: string;
  cps_gpf_number: string;
}

// ---- Confirm Delete Dialog ----
interface ConfirmDeleteProps {
  open: boolean;
  memberName: string;
  onConfirm: () => void;
  onCancel: () => void;
  isLoading?: boolean;
}

function ConfirmDeleteDialog({ open, memberName, onConfirm, onCancel, isLoading }: ConfirmDeleteProps) {
  return (
    <Dialog open={open} onOpenChange={(v: boolean) => !v && onCancel()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-destructive">
            <AlertTriangle className="w-5 h-5" />
            Delete Member
          </DialogTitle>
          <DialogDescription>
            Are you sure you want to permanently delete <strong>{memberName}</strong>? This action
            cannot be undone and will remove all their data from the system.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onCancel} disabled={isLoading}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={onConfirm} disabled={isLoading}>
            {isLoading ? (
              <><Loader2 className="w-4 h-4 mr-2 animate-spin" />Deleting...</>
            ) : (
              <><Trash2 className="w-4 h-4 mr-2" />Delete Permanently</>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export function MemberListTable({ onUpdateStats }: MemberListTableProps) {
  const [members, setMembers] = useState<MemberListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedMember, setSelectedMember] = useState<MemberListItem | null>(null);
  const [viewDetailsOpen, setViewDetailsOpen] = useState(false);
  const [districtFilter, setDistrictFilter] = useState('all');
  const [paymentStatusFilter, setPaymentStatusFilter] = useState('all');
  const [districtsList, setDistrictsList] = useState<DistrictOption[]>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalMembers, setTotalMembers] = useState(0);
  const itemsPerPage = 10;

  // Edit state
  const [isEditing, setIsEditing] = useState(false);
  const [editForm, setEditForm] = useState<EditMemberForm | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  // Delete state
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [memberToDelete, setMemberToDelete] = useState<MemberListItem | null>(null);

  // Export state
  const [isExporting, setIsExporting] = useState(false);

  const fetchMembers = async () => {
    setLoading(true);
    try {
      const response = await getMembersList(
        currentPage,
        itemsPerPage,
        searchQuery,
        districtFilter,
        paymentStatusFilter
      );
      setMembers(response.members || []);
      setTotalPages(response.total_pages || 1);
      setTotalMembers(response.total || 0);

      if (onUpdateStats) {
        onUpdateStats({
          totalMembers: response.total || 0,
          activeMembers: response.members?.filter((m) => m.membership_status === 'Active').length || 0,
          unpaid: response.members?.filter((m) => m.payment_status === 'Unpaid').length || 0,
        });
      }
    } catch (error) {
      console.error('Error fetching members:', error);
      toast.error('Failed to load members');
      setMembers([]);
    } finally {
      setLoading(false);
    }
  };

  const fetchDistricts = async () => {
    try {
      const data = await getMemberDistricts();
      setDistrictsList(data || []);
    } catch (error) {
      console.error('Error fetching districts:', error);
      setDistrictsList([]);
    }
  };

  useEffect(() => {
    fetchMembers();
  }, [currentPage, districtFilter, paymentStatusFilter, searchQuery]);

  useEffect(() => {
    fetchDistricts();
  }, []);

  // ---- Open View Panel ----
  const handleViewDetails = (member: MemberListItem) => {
    setSelectedMember(member);
    setIsEditing(false);
    setEditForm(null);
    setViewDetailsOpen(true);
  };

  // ---- Enter Edit Mode ----
  const handleStartEdit = () => {
    if (!selectedMember) return;
    setEditForm({
      name: selectedMember.name || '',
      initial: selectedMember.initial || '',
      gender: selectedMember.gender || '',
      father_name: (selectedMember as any).father_name || '',
      mother_name: (selectedMember as any).mother_name || '',
      educational_qualification: (selectedMember as any).educational_qualification || '',
      designation: selectedMember.designation || '',
      working_district: selectedMember.district || '',
      native_district: (selectedMember as any).native_district || '',
      recruitment_batch: selectedMember.recruitment_batch || '',
      seniority_number: (selectedMember as any).seniority_number || '',
      residential_address: (selectedMember as any).residential_address || '',
      permanent_address: (selectedMember as any).permanent_address || '',
      date_of_birth: (selectedMember as any).date_of_birth || '',
      mobile_number: selectedMember.mobile_number || '',
      emailId: selectedMember.email || '',
      tbf_number: (selectedMember as any).tbf_number || '',
      cps_gpf_number: (selectedMember as any).cps_gpf_number || '',
    });
    setIsEditing(true);
  };

  // ---- Save Edit ----
  const handleSaveEdit = async () => {
    if (!selectedMember || !editForm) return;
    setIsSaving(true);
    try {
      const token = localStorage.getItem('admin_token');
      const response = await fetch(
        `${API_BASE_URL}/admin/members/${selectedMember.id}`,
        {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify(editForm),
        }
      );

      const data = await response.json();
      if (!response.ok) {
        throw new Error(data.error || 'Failed to update member');
      }

      toast.success('Member updated successfully');
      setIsEditing(false);
      setViewDetailsOpen(false);
      await fetchMembers();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to update member');
    } finally {
      setIsSaving(false);
    }
  };

  // ---- Cancel Edit ----
  const handleCancelEdit = () => {
    setIsEditing(false);
    setEditForm(null);
  };

  // ---- Trigger Delete Confirm ----
  const handleDeleteClick = (member: MemberListItem) => {
    setMemberToDelete(member);
    setViewDetailsOpen(false);
    setDeleteConfirmOpen(true);
  };

  // ---- Confirm Delete ----
  const handleConfirmDelete = async () => {
    if (!memberToDelete) return;
    setIsDeleting(true);
    try {
      const token = localStorage.getItem('admin_token');
      const response = await fetch(
        `${API_BASE_URL}/admin/members/${memberToDelete.id}`,
        {
          method: 'DELETE',
          headers: { Authorization: `Bearer ${token}` },
        }
      );

      const data = await response.json();
      if (!response.ok) {
        throw new Error(data.error || 'Failed to delete member');
      }

      toast.success(`Member "${memberToDelete.name}" deleted successfully`);
      setDeleteConfirmOpen(false);
      setMemberToDelete(null);
      await fetchMembers();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to delete member');
    } finally {
      setIsDeleting(false);
    }
  };

  const handleRefresh = () => {
    fetchMembers();
    toast.success('Refreshed');
  };

  // ---- Export to Excel ----
  const handleExportExcel = async () => {
    setIsExporting(true);
    try {
      toast.info('Generating Excel report...');

      const blob = await exportMembersToExcel(districtFilter, paymentStatusFilter);

      // Create download link
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      const date = new Date().toISOString().slice(0, 19).replace(/:/g, '-');
      const filterText = districtFilter !== 'all' ? `_${districtFilter}` : '';
      const paymentText = paymentStatusFilter !== 'all' ? `_${paymentStatusFilter}` : '';
      a.download = `TAGA_Members${filterText}${paymentText}_${date}.xlsx`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);

      toast.success('Excel report downloaded successfully');
    } catch (error) {
      console.error('Export error:', error);
      toast.error(error instanceof Error ? error.message : 'Failed to export members');
    } finally {
      setIsExporting(false);
    }
  };

  const getPaymentBadge = (status: string) => {
    const normalized = status?.toLowerCase() || 'unpaid';
    if (normalized === 'paid')
      return <Badge className="bg-green-100 text-green-800 px-2 py-1">Paid</Badge>;
    if (normalized === 'unpaid')
      return <Badge className="bg-red-100 text-red-800 px-2 py-1">Unpaid</Badge>;
    return <Badge variant="secondary" className="px-2 py-1">Unknown</Badge>;
  };

  const getMembershipBadge = (status: string) => {
    const normalized = status?.toLowerCase() || 'inactive';
    if (normalized === 'active')
      return <Badge className="bg-green-600 text-white px-2 py-1">Active</Badge>;
    if (normalized === 'inactive')
      return <Badge className="bg-gray-400 text-white px-2 py-1">Inactive</Badge>;
    return <Badge variant="secondary" className="px-2 py-1">Unknown</Badge>;
  };

  const getDistrictOptions = () => {
    if (!districtsList || districtsList.length === 0) return [];
    return districtsList.map((d) => d.name).sort();
  };

  const displayMembers = members || [];

  const getDisplayValue = (value: string | undefined) => (value && value !== '' ? value : '-');

  // ---- Edit form field renderer ----
  const EditField = ({
    label,
    fieldKey,
    type = 'text',
  }: {
    label: string;
    fieldKey: keyof EditMemberForm;
    type?: string;
  }) => (
    <div className="space-y-1">
      <Label className="text-xs text-muted-foreground">{label}</Label>
      <Input
        type={type}
        value={editForm?.[fieldKey] || ''}
        onChange={(e) => setEditForm((prev) => prev ? { ...prev, [fieldKey]: e.target.value } : prev)}
        className="h-8 text-sm"
      />
    </div>
  );

  const EditSelect = ({
    label,
    fieldKey,
    options,
  }: {
    label: string;
    fieldKey: keyof EditMemberForm;
    options: string[];
  }) => (
    <div className="space-y-1">
      <Label className="text-xs text-muted-foreground">{label}</Label>
      <Select
        value={editForm?.[fieldKey] || ''}
        onValueChange={(val: string) => setEditForm((prev) => prev ? { ...prev, [fieldKey]: val } : prev)}
      >
        <SelectTrigger className="h-8 text-sm">
          <SelectValue placeholder={`Select ${label}`} />
        </SelectTrigger>
        <SelectContent>
          {options.map((opt) => (
            <SelectItem key={opt} value={opt}>
              {opt}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );

  const EditTextarea = ({ label, fieldKey }: { label: string; fieldKey: keyof EditMemberForm }) => (
    <div className="space-y-1 col-span-2">
      <Label className="text-xs text-muted-foreground">{label}</Label>
      <Textarea
        value={editForm?.[fieldKey] || ''}
        onChange={(e) => setEditForm((prev) => prev ? { ...prev, [fieldKey]: e.target.value } : prev)}
        rows={2}
        className="text-sm resize-none"
      />
    </div>
  );

  // ---- Read-only display ----
  const Field = ({ label, value }: { label: string; value: string | undefined }) => (
    <div className="space-y-1">
      <p className="text-xs font-medium text-muted-foreground">{label}</p>
      <p className="text-sm font-medium">{getDisplayValue(value)}</p>
    </div>
  );

  return (
    <>
      <div className="space-y-4">
        {/* Header */}
        <div className="flex justify-between items-center">
          <div>
            <h3 className="text-lg font-semibold">Member Management</h3>
            <p className="text-sm text-muted-foreground">
              Manage member accounts, subscriptions, and access rights
            </p>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handleExportExcel}
              disabled={isExporting}
              data-testid="testid-export-excel-button"
            >
              {isExporting ? (
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <Download className="w-4 h-4 mr-2" />
              )}
              {isExporting ? 'Exporting...' : 'Export Excel'}
            </Button>
            <Button variant="outline" size="sm" onClick={handleRefresh}>
              <RefreshCw className="w-4 h-4 mr-2" />
              Refresh
            </Button>
          </div>
        </div>

        {/* Filters */}
        <div className="bg-gradient-to-r from-green-50 to-green-100 dark:from-green-900/20 dark:to-green-800/20 border border-green-200 dark:border-green-800 rounded-lg p-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="relative md:col-span-1">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
              <Input
                placeholder="Search by name, email or mobile..."
                value={searchQuery}
                onChange={(e) => { setSearchQuery(e.target.value); setCurrentPage(1); }}
                className="pl-10"
              />
            </div>
            <div>
              <Select value={districtFilter} onValueChange={(value: string) => { setDistrictFilter(value); setCurrentPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All Districts" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Districts</SelectItem>
                  {getDistrictOptions().map((district) => (
                    <SelectItem key={district} value={district}>{district}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <Select value={paymentStatusFilter} onValueChange={(value: string) => { setPaymentStatusFilter(value); setCurrentPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All Payment Status" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Payment Status</SelectItem>
                  <SelectItem value="paid">Paid</SelectItem>
                  <SelectItem value="unpaid">Unpaid</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>

        {/* Count */}
        <div className="text-sm text-muted-foreground">
          Showing {displayMembers.length} of {totalMembers} members
        </div>

        {/* Table */}
        <div className="border rounded-lg overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow className="bg-gray-100">
                <TableHead className="font-semibold w-[120px]">TAGA ID</TableHead>
                <TableHead className="font-semibold">Name</TableHead>
                <TableHead className="font-semibold">Initial</TableHead>
                <TableHead className="font-semibold">Gender</TableHead>
                <TableHead className="font-semibold">District</TableHead>
                <TableHead className="font-semibold">Designation</TableHead>
                <TableHead className="font-semibold">Batch</TableHead>
                <TableHead className="font-semibold">Mobile</TableHead>
                <TableHead className="font-semibold">Email</TableHead>
                <TableHead className="font-semibold">Payment Status</TableHead>
                <TableHead className="font-semibold">Membership Status</TableHead>
                <TableHead className="font-semibold text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={12} className="text-center py-8">
                    <div className="flex justify-center items-center gap-2">
                      <RefreshCw className="w-4 h-4 animate-spin" />
                      Loading members...
                    </div>
                  </TableCell>
                </TableRow>
              ) : displayMembers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={12} className="text-center py-8 text-muted-foreground">
                    No members found
                  </TableCell>
                </TableRow>
              ) : (
                displayMembers.map((member) => (
                  <TableRow key={member.id} className="hover:bg-gray-50">
                    <TableCell className="font-mono text-sm font-semibold text-primary">
                      {getDisplayValue(member.tagaId)}
                    </TableCell>
                    <TableCell className="font-medium">{getDisplayValue(member.name)}</TableCell>
                    <TableCell>{getDisplayValue(member.initial)}</TableCell>
                    <TableCell>{getDisplayValue(member.gender)}</TableCell>
                    <TableCell>{getDisplayValue(member.district)}</TableCell>
                    <TableCell className="max-w-[200px] truncate">{getDisplayValue(member.designation)}</TableCell>
                    <TableCell>{getDisplayValue(member.recruitment_batch)}</TableCell>
                    <TableCell>{getDisplayValue(member.mobile_number)}</TableCell>
                    <TableCell className="max-w-[200px] truncate">{getDisplayValue(member.email)}</TableCell>
                    <TableCell>{getPaymentBadge(member.payment_status)}</TableCell>
                    <TableCell>{getMembershipBadge(member.membership_status)}</TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleViewDetails(member)}
                      >
                        <Eye className="w-4 h-4 mr-1" />
                        View
                      </Button>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex justify-between items-center mt-4">
            <div className="text-sm text-muted-foreground">
              Page {currentPage} of {totalPages}
            </div>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                disabled={currentPage === 1}
              >
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                disabled={currentPage === totalPages}
              >
                Next
              </Button>
            </div>
          </div>
        )}
      </div>

      {/* ===================== VIEW / EDIT DIALOG ===================== */}
      <Dialog
        open={viewDetailsOpen}
        onOpenChange={(v: boolean) => {
          if (!v) {
            setViewDetailsOpen(false);
            setIsEditing(false);
            setEditForm(null);
          }
        }}
      >
        <DialogContent className="max-w-3xl max-h-[85vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center justify-between">
              <span>{isEditing ? 'Edit Member Details' : 'Member Details'}</span>
              {!isEditing && selectedMember && (
                <div className="flex items-center gap-2 mr-6">
                  <Button
                    variant="outline"
                    size="sm"
                    className="text-blue-600 border-blue-300 hover:bg-blue-50"
                    onClick={handleStartEdit}
                  >
                    <Pencil className="w-3.5 h-3.5 mr-1" />
                    Edit
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    className="text-destructive border-destructive/30 hover:bg-destructive/10"
                    onClick={() => selectedMember && handleDeleteClick(selectedMember)}
                  >
                    <Trash2 className="w-3.5 h-3.5 mr-1" />
                    Delete
                  </Button>
                </div>
              )}
            </DialogTitle>
            <DialogDescription>
              {isEditing
                ? 'Update the member information below. TAGA ID and password cannot be changed.'
                : `Complete information for ${selectedMember?.name}`}
            </DialogDescription>
          </DialogHeader>

          {selectedMember && !isEditing && (
            <div className="grid grid-cols-2 gap-4 py-4">
              <Field label="TAGA ID" value={selectedMember.tagaId} />
              <Field label="Name" value={selectedMember.name} />
              <Field label="Initial" value={selectedMember.initial} />
              <Field label="Gender" value={selectedMember.gender} />
              <Field label="District" value={selectedMember.district} />
              <Field label="Designation" value={selectedMember.designation} />
              <Field label="Recruitment Batch" value={selectedMember.recruitment_batch} />
              <Field label="Mobile Number" value={selectedMember.mobile_number} />
              <Field label="Email ID" value={selectedMember.email} />
              <div className="space-y-1">
                <p className="text-xs font-medium text-muted-foreground">Payment Status</p>
                {getPaymentBadge(selectedMember.payment_status)}
              </div>
              <div className="space-y-1">
                <p className="text-xs font-medium text-muted-foreground">Membership Status</p>
                {getMembershipBadge(selectedMember.membership_status)}
              </div>
            </div>
          )}

          {isEditing && editForm && (
            <div className="grid grid-cols-2 gap-3 py-4">
              <div className="space-y-1" key="field-name">
                <Label className="text-xs text-muted-foreground">Full Name *</Label>
                <Input
                  type="text"
                  value={editForm.name}
                  onChange={(e) => setEditForm((prev) => prev ? { ...prev, name: e.target.value } : prev)}
                  className="h-8 text-sm"
                />
              </div>
              <div className="space-y-1" key="field-initial">
                <Label className="text-xs text-muted-foreground">Initial *</Label>
                <Input
                  type="text"
                  value={editForm.initial}
                  onChange={(e) => setEditForm((prev) => prev ? { ...prev, initial: e.target.value } : prev)}
                  className="h-8 text-sm"
                />
              </div>
              <div className="space-y-1" key="field-gender">
                <Label className="text-xs text-muted-foreground">Gender *</Label>
                <Select
                  value={editForm.gender}
                  onValueChange={(val: string) => setEditForm((prev) => prev ? { ...prev, gender: val } : prev)}
                >
                  <SelectTrigger className="h-8 text-sm">
                    <SelectValue placeholder="Select Gender" />
                  </SelectTrigger>
                  <SelectContent>
                    {['Male', 'Female', 'Other'].map((opt) => (
                      <SelectItem key={opt} value={opt}>{opt}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1" key="field-dob">
                <Label className="text-xs text-muted-foreground">Date of Birth</Label>
                <Input
                  type="date"
                  value={editForm.date_of_birth}
                  onChange={(e) => setEditForm((prev) => prev ? { ...prev, date_of_birth: e.target.value } : prev)}
                  className="h-8 text-sm"
                />
              </div>
              <div className="space-y-1" key="field-father">
                <Label className="text-xs text-muted-foreground">Father Name</Label>
                <Input
                  type="text"
                  value={editForm.father_name}
                  onChange={(e) => setEditForm((prev) => prev ? { ...prev, father_name: e.target.value } : prev)}
                  className="h-8 text-sm"
                />
              </div>
              <div className="space-y-1" key="field-mother">
                <Label className="text-xs text-muted-foreground">Mother Name</Label>
                <Input
                  type="text"
                  value={editForm.mother_name}
                  onChange={(e) => setEditForm((prev) => prev ? { ...prev, mother_name: e.target.value } : prev)}
                  className="h-8 text-sm"
                />
              </div>
              <div className="space-y-1" key="field-edu">
                <Label className="text-xs text-muted-foreground">Educational Qualification</Label>
                <Input
                  type="text"
                  value={editForm.educational_qualification}
                  onChange={(e) => setEditForm((prev) => prev ? { ...prev, educational_qualification: e.target.value } : prev)}
                  className="h-8 text-sm"
                />
              </div>
              <div className="space-y-1" key="field-designation">
                <Label className="text-xs text-muted-foreground">Designation</Label>
                <Input
                  type="text"
                  value={editForm.designation}
                  onChange={(e) => setEditForm((prev) => prev ? { ...prev, designation: e.target.value } : prev)}
                  className="h-8 text-sm"
                />
              </div>
              <div className="space-y-1" key="field-working-district">
                <Label className="text-xs text-muted-foreground">Working District</Label>
                <Select
                  value={editForm.working_district}
                  onValueChange={(val: string) => setEditForm((prev) => prev ? { ...prev, working_district: val } : prev)}
                >
                  <SelectTrigger className="h-8 text-sm">
                    <SelectValue placeholder="Select Working District" />
                  </SelectTrigger>
                  <SelectContent>
                    {districts.map((district) => (
                      <SelectItem key={district} value={district}>{district}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1" key="field-native-district">
                <Label className="text-xs text-muted-foreground">Native District</Label>
                <Select
                  value={editForm.native_district}
                  onValueChange={(val: string) => setEditForm((prev) => prev ? { ...prev, native_district: val } : prev)}
                >
                  <SelectTrigger className="h-8 text-sm">
                    <SelectValue placeholder="Select Native District" />
                  </SelectTrigger>
                  <SelectContent>
                    {districts.map((district) => (
                      <SelectItem key={district} value={district}>{district}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1" key="field-batch">
                <Label className="text-xs text-muted-foreground">Recruitment Batch</Label>
                <Input
                  type="text"
                  value={editForm.recruitment_batch}
                  onChange={(e) => setEditForm((prev) => prev ? { ...prev, recruitment_batch: e.target.value } : prev)}
                  className="h-8 text-sm"
                />
              </div>
              <div className="space-y-1" key="field-seniority">
                <Label className="text-xs text-muted-foreground">Seniority Number</Label>
                <Input
                  type="text"
                  value={editForm.seniority_number}
                  onChange={(e) => setEditForm((prev) => prev ? { ...prev, seniority_number: e.target.value } : prev)}
                  className="h-8 text-sm"
                />
              </div>
              <div className="space-y-1" key="field-mobile">
                <Label className="text-xs text-muted-foreground">Mobile Number *</Label>
                <Input
                  type="tel"
                  value={editForm.mobile_number}
                  onChange={(e) => setEditForm((prev) => prev ? { ...prev, mobile_number: e.target.value } : prev)}
                  className="h-8 text-sm"
                />
              </div>
              <div className="space-y-1" key="field-email">
                <Label className="text-xs text-muted-foreground">Email ID *</Label>
                <Input
                  type="email"
                  value={editForm.emailId}
                  onChange={(e) => setEditForm((prev) => prev ? { ...prev, emailId: e.target.value } : prev)}
                  className="h-8 text-sm"
                />
              </div>
              <div className="space-y-1" key="field-tbf">
                <Label className="text-xs text-muted-foreground">TBF Number</Label>
                <Input
                  type="text"
                  value={editForm.tbf_number}
                  onChange={(e) => setEditForm((prev) => prev ? { ...prev, tbf_number: e.target.value } : prev)}
                  className="h-8 text-sm"
                />
              </div>
              <div className="space-y-1" key="field-cps">
                <Label className="text-xs text-muted-foreground">CPS / GPF Number</Label>
                <Input
                  type="text"
                  value={editForm.cps_gpf_number}
                  onChange={(e) => setEditForm((prev) => prev ? { ...prev, cps_gpf_number: e.target.value } : prev)}
                  className="h-8 text-sm"
                />
              </div>
              <div className="space-y-1 col-span-2" key="field-residential">
                <Label className="text-xs text-muted-foreground">Residential Address</Label>
                <Textarea
                  value={editForm.residential_address}
                  onChange={(e) => setEditForm((prev) => prev ? { ...prev, residential_address: e.target.value } : prev)}
                  rows={2}
                  className="text-sm resize-none"
                />
              </div>
              <div className="space-y-1 col-span-2" key="field-permanent">
                <Label className="text-xs text-muted-foreground">Permanent Address</Label>
                <Textarea
                  value={editForm.permanent_address}
                  onChange={(e) => setEditForm((prev) => prev ? { ...prev, permanent_address: e.target.value } : prev)}
                  rows={2}
                  className="text-sm resize-none"
                />
              </div>
            </div>
          )}
          <DialogFooter className="gap-2">
            {!isEditing ? (
              <Button variant="outline" onClick={() => setViewDetailsOpen(false)}>
                Close
              </Button>
            ) : (
              <>
                <Button variant="outline" onClick={handleCancelEdit} disabled={isSaving}>
                  <X className="w-4 h-4 mr-1" />
                  Cancel
                </Button>
                <Button onClick={handleSaveEdit} disabled={isSaving}>
                  {isSaving ? (
                    <><Loader2 className="w-4 h-4 mr-2 animate-spin" />Saving...</>
                  ) : (
                    <><Save className="w-4 h-4 mr-1" />Save Changes</>
                  )}
                </Button>
              </>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* ===================== CONFIRM DELETE DIALOG ===================== */}
      <ConfirmDeleteDialog
        open={deleteConfirmOpen}
        memberName={memberToDelete?.name || ''}
        onConfirm={handleConfirmDelete}
        onCancel={() => { setDeleteConfirmOpen(false); setMemberToDelete(null); }}
        isLoading={isDeleting}
      />
    </>
  );
}