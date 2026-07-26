import React, { useState, useEffect } from 'react';
import API_BASE_URL from '../config/api';
import { Button } from './ui/button';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from './ui/dialog';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Textarea } from './ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from './ui/tabs';
import { Badge } from './ui/badge';
import { UserPlus, FileText, MessageSquare, Download, Send, Plus, Upload, Calendar, Image as ImageIcon, Link as LinkIcon, Trash2, Pencil, ChevronDown, ChevronRight, Loader2, AlertTriangle, Users } from 'lucide-react';
import { toast } from 'sonner';
import {
  createEvent,
  updateEvent,
  deleteEvent,
  uploadResource,
  deleteResource,
  uploadGalleryImage,
  deleteGalleryImage,
  getResourceCategoriesWithDocs,
  getEvents,
  getGallery,
  CreateEventData,
  UploadGalleryData,
  ResourceCategoryWithDocs,
  ResourceDocument,
  Event,
  GalleryImage,
  addMember,
  bulkUploadMembers,
  generateMemberReport,
  sendAnnouncement
} from '../api/adminContent';
import { sortDocuments } from '../api/resources';
import { MemberListTable } from './MemberListTable';
import DistrictOfficeBearersManager from './admin/DistrictOfficeBearersManager';

interface AdminActionsProps {
  memberStats: {
    totalMembers: number;
    activeMembers: number;
    unpaid: number;
    newThisMonth: number;
  };
}

// Map display category names to backend IDs
const categoryToIdMap: Record<string, string> = {
  'Establishment': 'establishment',
  'Leave Forms & Other Applications': 'leave-forms',
  'Miscellaneous': 'miscellaneous',
  'Office Address & Contacts': 'office-contacts',
  'Pay Related G.Os': 'pay-gos',
  'Scheme G.Os': 'scheme-gos',
  'TAGA Membership & TBF Application': 'taga-membership',
  'Technical': 'technical',
  'Links': 'links'
};

// ===================== Confirm Delete Dialog =====================
interface ConfirmDeleteProps {
  open: boolean;
  title: string;
  description: string;
  onConfirm: () => void;
  onCancel: () => void;
  isLoading?: boolean;
}

function ConfirmDeleteDialog({ open, title, description, onConfirm, onCancel, isLoading }: ConfirmDeleteProps) {
  return (
    <Dialog open={open} onOpenChange={(v: boolean) => !v && onCancel()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-destructive">
            <AlertTriangle className="w-5 h-5" />
            {title}
          </DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onCancel} disabled={isLoading}>Cancel</Button>
          <Button variant="destructive" onClick={onConfirm} disabled={isLoading}>
            {isLoading ? <><Loader2 className="w-4 h-4 mr-2 animate-spin" />Deleting...</> : <><Trash2 className="w-4 h-4 mr-2" />Delete</>}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export function AdminActions({ memberStats }: AdminActionsProps) {
  const [addMemberOpen, setAddMemberOpen] = useState(false);
  const [bulkUploadOpen, setBulkUploadOpen] = useState(false);
  const [contentManagementOpen, setContentManagementOpen] = useState(false);
  const [announcementOpen, setAnnouncementOpen] = useState(false);
  const [reportOpen, setReportOpen] = useState(false);
  const [officeBearersOpen, setOfficeBearersOpen] = useState(false);

  // Loading states
  const [isUploadingResource, setIsUploadingResource] = useState(false);
  const [isPublishingEvent, setIsPublishingEvent] = useState(false);
  const [isUploadingPhoto, setIsUploadingPhoto] = useState(false);
  const [isAddingMember, setIsAddingMember] = useState(false);
  const [isBulkUploading, setIsBulkUploading] = useState(false);
  const [isGeneratingReport, setIsGeneratingReport] = useState(false);

  // ==================== MANAGE DATA STATE ====================
  // Resources
  const [resourceCategories, setResourceCategories] = useState<ResourceCategoryWithDocs[]>([]);
  const [isLoadingResources, setIsLoadingResources] = useState(false);
  const [expandedCategories, setExpandedCategories] = useState<Record<string, boolean>>({});
  const [deleteResourceConfirm, setDeleteResourceConfirm] = useState<{ open: boolean; categoryId: string; title: string } | null>(null);
  const [isDeletingResource, setIsDeletingResource] = useState(false);

  // Events
  const [eventsList, setEventsList] = useState<Event[]>([]);
  const [isLoadingEvents, setIsLoadingEvents] = useState(false);
  const [deleteEventConfirm, setDeleteEventConfirm] = useState<{ open: boolean; id: string; title: string } | null>(null);
  const [isDeletingEvent, setIsDeletingEvent] = useState(false);
  const [editEventOpen, setEditEventOpen] = useState(false);
  const [editingEvent, setEditingEvent] = useState<Event | null>(null);
  const [isUpdatingEvent, setIsUpdatingEvent] = useState(false);
  const [editEventForm, setEditEventForm] = useState({
    title: '',
    date: '',
    time: '',
    location: '',
    description: '',
    status: 'upcoming',
    image: null as File | null
  });

  // Gallery
  const [galleryList, setGalleryList] = useState<GalleryImage[]>([]);
  const [isLoadingGallery, setIsLoadingGallery] = useState(false);
  const [deleteGalleryConfirm, setDeleteGalleryConfirm] = useState<{ open: boolean; id: string; title: string } | null>(null);
  const [isDeletingGallery, setIsDeletingGallery] = useState(false);

  // ==================== ADD MEMBER FORM ====================
  const [memberForm, setMemberForm] = useState({
    tagaId: '',
    name: '',
    initial: '',
    gender: '',
    fatherName: '',
    motherName: '',
    educationalQualification: '',
    designation: '',
    workingDistrict: '',
    nativeDistrict: '',
    recruitmentBatch: '',
    seniorityNumber: '',
    residentialAddress: '',
    permanentAddress: '',
    dateOfBirth: '',
    mobileNumber: '',
    email: '',
    tbfNumber: '',
    cpsGpfNumber: ''
  });

  const [bulkUploadFile, setBulkUploadFile] = useState<File | null>(null);
  const [contentTab, setContentTab] = useState('resources');

  // Resources Form
  const [resourceForm, setResourceForm] = useState({
    category: '',
    year: new Date().getFullYear().toString(),
    subcategory: '',
    file: null as File | null
  });

  // Events Form
  const [eventForm, setEventForm] = useState({
    title: '',
    date: '',
    time: '',
    location: '',
    description: '',
    image: null as File | null
  });

  // Gallery Form
  const [galleryForm, setGalleryForm] = useState({
    date: '',
    description: '',
    event: '',
    photo: null as File | null
  });

  // Announcement Form
  const [announcementForm, setAnnouncementForm] = useState({
    title: '',
    message: '',
    priority: 'normal',
    sendTo: 'all'
  });

  // Report Form
  const [reportForm, setReportForm] = useState({
    type: 'membership',
    period: 'current_month'
  });



  const resourceCategoriess = [
    'Establishment',
    'Leave Forms & Other Applications',
    'Miscellaneous',
    'Office Address & Contacts',
    'Pay Related G.Os',
    'Scheme G.Os',
    'TAGA Membership & TBF Application',
    'Technical',
    'Links'
  ];

  const districts = [
    'Ariyalur', 'Chengalpattu', 'Chennai', 'Coimbatore', 'Cuddalore', 'Dharmapuri',
    'Dindigul', 'Erode', 'Kallakurichi', 'Kanchipuram', 'Kanyakumari', 'Karur',
    'Krishnagiri', 'Madurai', 'Mayiladuthurai', 'Nagapattinam', 'Namakkal', 'Nilgiris',
    'Perambalur', 'Pudukkottai', 'Ramanathapuram', 'Ranipet', 'Salem', 'Sivaganga',
    'Tenkasi', 'Thanjavur', 'Theni', 'Thoothukudi', 'Tiruchirappalli', 'Tirunelveli',
    'Tirupathur', 'Tiruppur', 'Tiruvallur', 'Tiruvannamalai', 'Tiruvarur', 'Vellore',
    'Viluppuram', 'Virudhunagar'
  ];

  // ==================== LOAD DATA WHEN TAB CHANGES ====================
  useEffect(() => {
    if (contentManagementOpen) {
      if (contentTab === 'resources') loadResources();
      if (contentTab === 'events') loadEvents();
      if (contentTab === 'gallery') loadGallery();
    }
  }, [contentTab, contentManagementOpen]);

  const loadResources = async () => {
    setIsLoadingResources(true);
    try {
      const data = await getResourceCategoriesWithDocs();
      // ✅ Sort documents within each category: newest year first, then alphabetical
      const sortedData = data.map(category => ({
        ...category,
        documents: sortDocuments(category.documents || [])
      }));
      setResourceCategories(sortedData);
    } catch (error) {
      toast.error('Failed to load resources');
      console.error(error);
    } finally {
      setIsLoadingResources(false);
    }
  };

  const loadEvents = async () => {
    setIsLoadingEvents(true);
    try {
      const data = await getEvents();
      setEventsList(Array.isArray(data) ? data : []);
    } catch (error) {
      toast.error('Failed to load events');
      console.error(error);
    } finally {
      setIsLoadingEvents(false);
    }
  };

  const loadGallery = async () => {
    setIsLoadingGallery(true);
    try {
      const data = await getGallery();
      setGalleryList(Array.isArray(data) ? data : []);
    } catch (error) {
      toast.error('Failed to load gallery');
      console.error(error);
    } finally {
      setIsLoadingGallery(false);
    }
  };

  // ==================== RESOURCE ACTIONS ====================
  const toggleCategory = (id: string) => {
    setExpandedCategories(prev => ({ ...prev, [id]: !prev[id] }));
  };

  const handleDeleteResource = async () => {
    if (!deleteResourceConfirm) return;
    setIsDeletingResource(true);
    try {
      await deleteResource(deleteResourceConfirm.categoryId, deleteResourceConfirm.title);
      toast.success('Resource deleted successfully');
      setDeleteResourceConfirm(null);
      await loadResources();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to delete resource');
    } finally {
      setIsDeletingResource(false);
    }
  };

  // ==================== EVENT ACTIONS ====================
  const openEditEvent = (event: Event) => {
    setEditingEvent(event);
    // Parse datetime
    const dateParts = event.date ? event.date.split(' ') : ['', ''];
    setEditEventForm({
      title: event.title || '',
      date: dateParts[0] || '',
      time: dateParts[1] || '',
      location: event.location || '',
      description: event.description || '',
      status: event.status || 'upcoming',
      image: null
    });
    setEditEventOpen(true);
  };

  const handleUpdateEvent = async () => {
    if (!editingEvent || !editEventForm.title || !editEventForm.date) {
      toast.error('Please fill all required fields');
      return;
    }
    setIsUpdatingEvent(true);
    try {
      let eventDateTime = editEventForm.date;
      if (editEventForm.time) eventDateTime = `${editEventForm.date} ${editEventForm.time}`;

      const updateData: Partial<CreateEventData> = {
        title: editEventForm.title,
        date: eventDateTime,
        location: editEventForm.location,
        description: editEventForm.description,
        status: editEventForm.status,
      };
      if (editEventForm.image) updateData.image = editEventForm.image;

      await updateEvent(editingEvent.id, updateData);
      toast.success('Event updated successfully');
      setEditEventOpen(false);
      setEditingEvent(null);
      await loadEvents();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to update event');
    } finally {
      setIsUpdatingEvent(false);
    }
  };

  const handleDeleteEvent = async () => {
    if (!deleteEventConfirm) return;
    setIsDeletingEvent(true);
    try {
      await deleteEvent(deleteEventConfirm.id);
      toast.success('Event deleted successfully');
      setDeleteEventConfirm(null);
      await loadEvents();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to delete event');
    } finally {
      setIsDeletingEvent(false);
    }
  };

  // ==================== GALLERY ACTIONS ====================
  const handleDeleteGallery = async () => {
    if (!deleteGalleryConfirm) return;
    setIsDeletingGallery(true);
    try {
      await deleteGalleryImage(deleteGalleryConfirm.id);
      toast.success('Gallery photo deleted successfully');
      setDeleteGalleryConfirm(null);
      await loadGallery();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to delete gallery photo');
    } finally {
      setIsDeletingGallery(false);
    }
  };

  // ==================== ADD MEMBER ====================
  const handleAddMember = async () => {
    if (!memberForm.tagaId || !memberForm.name || !memberForm.email || !memberForm.mobileNumber) {
      toast.error('Please fill all required fields (TAGA ID, Name, Email, Mobile)');
      return;
    }
    setIsAddingMember(true);
    try {
      const response = await addMember({
        tagaId: memberForm.tagaId,
        name: memberForm.name,
        initial: memberForm.initial,
        gender: memberForm.gender,
        fatherName: memberForm.fatherName,
        motherName: memberForm.motherName,
        educationalQualification: memberForm.educationalQualification,
        designation: memberForm.designation,
        workingDistrict: memberForm.workingDistrict,
        nativeDistrict: memberForm.nativeDistrict,
        recruitmentBatch: memberForm.recruitmentBatch,
        seniorityNumber: memberForm.seniorityNumber,
        residentialAddress: memberForm.residentialAddress,
        permanentAddress: memberForm.permanentAddress,
        dateOfBirth: memberForm.dateOfBirth,
        mobileNumber: memberForm.mobileNumber,
        email: memberForm.email,
        tbfNumber: memberForm.tbfNumber,
        cpsGpfNumber: memberForm.cpsGpfNumber,
      });
      toast.success(response.message);
      toast.info(`Temporary password: ${response.temp_password} - Please share with the member`);
      setMemberForm({
        tagaId: '', name: '', initial: '', gender: '', fatherName: '', motherName: '',
        educationalQualification: '', designation: '', workingDistrict: '', nativeDistrict: '',
        recruitmentBatch: '', seniorityNumber: '', residentialAddress: '', permanentAddress: '',
        dateOfBirth: '', mobileNumber: '', email: '', tbfNumber: '', cpsGpfNumber: ''
      });
      setAddMemberOpen(false);
    } catch (error) {
      console.error('Add member error:', error);
      toast.error(error instanceof Error ? error.message : 'Failed to add member');
    } finally {
      setIsAddingMember(false);
    }
  };

  // ==================== BULK UPLOAD ====================
  const handleBulkUpload = async () => {
    if (!bulkUploadFile) {
      toast.error('Please select a file to upload');
      return;
    }
    setIsBulkUploading(true);
    try {
      const response = await bulkUploadMembers(bulkUploadFile);
      toast.success(response.message);
      toast.success(`Successfully added: ${response.success_count} members`);
      if (response.failed_count > 0) {
        toast.error(`Failed: ${response.failed_count} members`);
        console.error('Failed uploads:', response.failed);
      }
      setBulkUploadFile(null);
      setBulkUploadOpen(false);
    } catch (error) {
      console.error('Bulk upload error:', error);
      toast.error(error instanceof Error ? error.message : 'Failed to upload members');
    } finally {
      setIsBulkUploading(false);
    }
  };

  // ==================== RESOURCE UPLOAD ====================
  const handleResourceUpload = async () => {
    if (!resourceForm.category || !resourceForm.file) {
      toast.error('Please fill all required fields');
      return;
    }
    setIsUploadingResource(true);
    try {
      const categoryId = categoryToIdMap[resourceForm.category] || resourceForm.category.toLowerCase().trim().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
      if (!categoryId) {
        toast.error('Invalid category selected');
        setIsUploadingResource(false);
        return;
      }
      await uploadResource({
        categoryId: categoryId,
        title: resourceForm.file.name.replace(/\.[^/.]+$/, ''),
        year: resourceForm.year,
        subcategory: resourceForm.subcategory || undefined,
        file: resourceForm.file
      });
      toast.success(`Resource uploaded to ${resourceForm.category} for year ${resourceForm.year}`);
      toast.info('Members have been notified about the new resource');
      setResourceForm({ category: '', year: new Date().getFullYear().toString(), subcategory: '', file: null });
      // Reload resources after upload
      await loadResources();
    } catch (error) {
      console.error('Upload error:', error);
      toast.error(error instanceof Error ? error.message : 'Failed to upload resource');
    } finally {
      setIsUploadingResource(false);
    }
  };

  // ==================== EVENT UPLOAD ====================
  const handleEventUpload = async () => {
    if (!eventForm.title || !eventForm.date) {
      toast.error('Please fill all required fields');
      return;
    }
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const selectedDate = new Date(eventForm.date);
    if (selectedDate < today) {
      toast.error('Cannot create an event for a past date');
      return;
    }
    setIsPublishingEvent(true);
    try {
      let eventDateTime = eventForm.date;
      if (eventForm.time) eventDateTime = `${eventForm.date} ${eventForm.time}`;

      const eventData: CreateEventData = {
        title: eventForm.title,
        date: eventDateTime,
        location: eventForm.location,
        description: eventForm.description,
        status: 'upcoming'
      };
      if (eventForm.image) eventData.image = eventForm.image;

      await createEvent(eventData);
      toast.success(`Event "${eventForm.title}" has been published`);
      toast.info('Members have been notified about the new event');
      setEventForm({ title: '', date: '', time: '', location: '', description: '', image: null });
      await loadEvents();
    } catch (error) {
      console.error('Event creation error:', error);
      toast.error(error instanceof Error ? error.message : 'Failed to create event');
    } finally {
      setIsPublishingEvent(false);
    }
  };

  // ==================== GALLERY UPLOAD ====================
  const handleGalleryUpload = async () => {
    if (!galleryForm.date || !galleryForm.photo || !galleryForm.description) {
      toast.error('Please fill all required fields');
      return;
    }
    setIsUploadingPhoto(true);
    try {
      const year = new Date(galleryForm.date).getFullYear();
      const galleryData: UploadGalleryData = {
        title: galleryForm.description,
        event: galleryForm.event || 'TAGA Event',
        date: galleryForm.date,
        year: year,
        image: galleryForm.photo
      };
      await uploadGalleryImage(galleryData);
      toast.success('Photo uploaded to gallery successfully');
      setGalleryForm({ date: '', description: '', event: '', photo: null });
      await loadGallery();
    } catch (error) {
      console.error('Gallery upload error:', error);
      toast.error(error instanceof Error ? error.message : 'Failed to upload photo');
    } finally {
      setIsUploadingPhoto(false);
    }
  };

  // ==================== SEND ANNOUNCEMENT ====================
  const handleSendAnnouncement = async () => {
    if (!announcementForm.title || !announcementForm.message) {
      toast.error('Please fill title and message');
      return;
    }
    try {
      const token = localStorage.getItem('admin_token');
      const response = await fetch(`${API_BASE_URL}/admin/announcements/send`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          title: announcementForm.title,
          message: announcementForm.message,
          priority: announcementForm.priority,
          sendTo: announcementForm.sendTo
        })
      });
      const data = await response.json();
      if (response.ok) {
        toast.success(data.message || `Announcement sent to ${data.recipients} members`);
        setAnnouncementForm({ title: '', message: '', priority: 'normal', sendTo: 'all' });
        setAnnouncementOpen(false);
      } else {
        toast.error(data.error || 'Failed to send announcement');
      }
    } catch (error) {
      console.error('Error sending announcement:', error);
      toast.error('Failed to send announcement');
    }
  };

  // ==================== GENERATE REPORT ====================
  const handleGenerateReport = async () => {
    setIsGeneratingReport(true);
    try {
      const blob = await generateMemberReport(reportForm.type, reportForm.period);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `TAGA_${reportForm.type}_report_${reportForm.period}_${new Date().toISOString().split('T')[0]}.csv`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
      toast.success('Report downloaded successfully');
      setReportOpen(false);
    } catch (error) {
      console.error('Report generation error:', error);
      toast.error(error instanceof Error ? error.message : 'Failed to generate report');
    } finally {
      setIsGeneratingReport(false);
    }
  };

  // ==================== STATUS BADGE ====================
  const getStatusBadge = (status: string) => {
    const map: Record<string, string> = {
      upcoming: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300',
      ongoing: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300',
      completed: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
      cancelled: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300',
    };
    return map[status] || map.upcoming;
  };

  return (
    <div className="space-y-6">
      {/* ===================== CONTENT MANAGEMENT ===================== */}
      <Card>
        <CardHeader>
          <CardTitle>Content Management</CardTitle>
          <CardDescription>Manage resources, events, and gallery content</CardDescription>
        </CardHeader>
        <CardContent>
          <Dialog open={contentManagementOpen} onOpenChange={setContentManagementOpen}>
            <DialogTrigger asChild>
              <Button variant="outline" className="w-full justify-start" data-testid="testid-manage-content-button">
                <FileText className="w-4 h-4 mr-2" />
                Manage Content
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-4xl max-h-[85vh] overflow-y-auto" data-testid="testid-content-management-modal">
              <DialogHeader>
                <DialogTitle>Content Management</DialogTitle>
                <DialogDescription>Upload, edit, and manage resources, events, and gallery content</DialogDescription>
              </DialogHeader>

              <Tabs value={contentTab} onValueChange={setContentTab} className="w-full" data-testid="testid-content-management-tabs-form">
                <TabsList className="grid w-full grid-cols-3" data-testid="testid-content-management-tabs-list">
                  <TabsTrigger value="resources" data-testid="testid-resources-button">Resources</TabsTrigger>
                  <TabsTrigger value="events" data-testid="testid-events-button">Events</TabsTrigger>
                  <TabsTrigger value="gallery" data-testid="testid-gallery-button">Gallery</TabsTrigger>
                </TabsList>

                {/* ============ RESOURCES TAB ============ */}
                <TabsContent value="resources" className="space-y-4 mt-4">
                  {/* Upload Form */}
                  <div className="border rounded-lg p-4 space-y-3 bg-muted/30">
                    <h3 className="font-semibold text-sm">Upload New Resource</h3>
                    <div className="space-y-2">
                      <Label htmlFor="resourceCategory">Resource Category *</Label>
                      <Select value={resourceForm.category} onValueChange={(value: string) => setResourceForm({ ...resourceForm, category: value })}>
                        <SelectTrigger data-testid="testid-resource-category-select">
                          <SelectValue placeholder="Select category" />
                        </SelectTrigger>
                        <SelectContent>
                          {resourceCategoriess.map((category) => (
                            <SelectItem key={category} value={category}>{category}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="resourceYear">Year *</Label>
                      <Input
                        id="resourceYear"
                        type="text"
                        value={resourceForm.year}
                        onChange={(e) => setResourceForm({ ...resourceForm, year: e.target.value })}
                        placeholder="2025"
                        data-testid="testid-resource-year-input"
                      />
                    </div>
                    {resourceForm.category === 'Scheme G.Os' && (
                      <div className="space-y-2">
                        <Label>Subcategory</Label>
                        <Select value={resourceForm.subcategory} onValueChange={(value: string) => setResourceForm({ ...resourceForm, subcategory: value })}>
                          <SelectTrigger>
                            <SelectValue placeholder="Select subcategory" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="Central">Central</SelectItem>
                            <SelectItem value="State">State</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                    )}
                    <div className="space-y-2">
                      <Label htmlFor="resourceFile">Upload Document *</Label>
                      <Input
                        id="resourceFile"
                        type="file"
                        accept=".pdf"
                        onChange={(e) => setResourceForm({ ...resourceForm, file: e.target.files?.[0] || null })}
                        data-testid="testid-resource-file-input"
                      />
                      {resourceForm.file && (
                        <p className="text-sm text-muted-foreground">Selected: {resourceForm.file.name}</p>
                      )}
                    </div>
                    <Button onClick={handleResourceUpload} className="w-full" disabled={isUploadingResource} data-testid="testid-upload-resource-button">
                      <Upload className="w-4 h-4 mr-2" />
                      {isUploadingResource ? 'Uploading...' : 'Upload Resource'}
                    </Button>
                  </div>

                  {/* Existing Resources List */}
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <h3 className="font-semibold text-sm">Existing Resources</h3>
                      <Button variant="ghost" size="sm" onClick={loadResources} disabled={isLoadingResources}>
                        {isLoadingResources ? <Loader2 className="w-3 h-3 animate-spin" /> : 'Refresh'}
                      </Button>
                    </div>
                    {isLoadingResources ? (
                      <div className="flex items-center justify-center py-8 text-muted-foreground">
                        <Loader2 className="w-5 h-5 animate-spin mr-2" />
                        Loading resources...
                      </div>
                    ) : resourceCategories.length === 0 ? (
                      <p className="text-sm text-muted-foreground text-center py-4">No resources found.</p>
                    ) : (
                      <div className="space-y-2">
                        {resourceCategories.map((cat) => (
                          <div key={cat.id} className="border rounded-lg overflow-hidden">
                            {/* Category Header */}
                            <button
                              className="w-full flex items-center justify-between px-4 py-3 bg-muted/40 hover:bg-muted/60 transition-colors text-left"
                              onClick={() => toggleCategory(cat.id)}
                            >
                              <div className="flex items-center gap-2">
                                {expandedCategories[cat.id] ? (
                                  <ChevronDown className="w-4 h-4 text-muted-foreground" />
                                ) : (
                                  <ChevronRight className="w-4 h-4 text-muted-foreground" />
                                )}
                                <span className="font-medium text-sm">{cat.name}</span>
                                <Badge variant="secondary" className="text-xs">
                                  {cat.documents?.length ?? 0} docs
                                </Badge>
                              </div>
                            </button>

                            {/* Documents List */}
                            {expandedCategories[cat.id] && (
                              <div className="divide-y">
                                {!cat.documents || cat.documents.length === 0 ? (
                                  <p className="text-sm text-muted-foreground px-4 py-3">No documents in this category.</p>
                                ) : (
                                  cat.documents.map((doc, idx) => (
                                    <div key={`${cat.id}-${idx}`} className="flex items-center justify-between px-4 py-2.5 hover:bg-muted/20">
                                      <div className="flex-1 min-w-0">
                                        <p className="text-sm font-medium truncate">{doc.title}</p>
                                        <div className="flex items-center gap-2 mt-0.5">
                                          <span className="text-xs text-muted-foreground">Year: {doc.year}</span>
                                          {doc.subcategory && (
                                            <Badge variant="outline" className="text-xs px-1 py-0">{doc.subcategory}</Badge>
                                          )}
                                        </div>
                                      </div>
                                      <Button
                                        variant="ghost"
                                        size="sm"
                                        className="text-destructive hover:text-destructive hover:bg-destructive/10 ml-2 shrink-0"
                                        onClick={() => setDeleteResourceConfirm({ open: true, categoryId: cat.id, title: doc.title })}
                                      >
                                        <Trash2 className="w-3.5 h-3.5" />
                                      </Button>
                                    </div>
                                  ))
                                )}
                              </div>
                            )}
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </TabsContent>

                {/* ============ EVENTS TAB ============ */}
                <TabsContent value="events" className="space-y-4 mt-4">
                  {/* Upload Form */}
                  <div className="border rounded-lg p-4 space-y-3 bg-muted/30">
                    <h3 className="font-semibold text-sm">Create New Event</h3>
                    <div className="space-y-2">
                      <Label htmlFor="eventTitle">Event Title *</Label>
                      <Input
                        id="eventTitle"
                        value={eventForm.title}
                        onChange={(e) => setEventForm({ ...eventForm, title: e.target.value })}
                        placeholder="Enter event title"
                        data-testid="testid-event-title-input"
                      />
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor="eventDate">Date *</Label>
                        <Input id="eventDate" type="date" min={new Date().toISOString().split('T')[0]} value={eventForm.date} onChange={(e) => setEventForm({ ...eventForm, date: e.target.value })} data-testid="testid-event-date-input" />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="eventTime">Time</Label>
                        <Input id="eventTime" type="time" value={eventForm.time} onChange={(e) => setEventForm({ ...eventForm, time: e.target.value })} data-testid="testid-event-time-input" />
                      </div>
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="eventLocation">Location</Label>
                      <Input id="eventLocation" value={eventForm.location} onChange={(e) => setEventForm({ ...eventForm, location: e.target.value })} placeholder="Event location" data-testid="testid-event-location-input" />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="eventDescription">Description</Label>
                      <Textarea id="eventDescription" value={eventForm.description} onChange={(e) => setEventForm({ ...eventForm, description: e.target.value })} placeholder="Event description" rows={3} data-testid="testid-event-description-input" />
                    </div>
                    <Button onClick={handleEventUpload} className="w-full" disabled={isPublishingEvent} data-testid="testid-publish-event-button">
                      <Calendar className="w-4 h-4 mr-2" />
                      {isPublishingEvent ? 'Publishing...' : 'Publish Event'}
                    </Button>
                  </div>

                  {/* Existing Events List */}
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <h3 className="font-semibold text-sm">Existing Events</h3>
                      <Button variant="ghost" size="sm" onClick={loadEvents} disabled={isLoadingEvents}>
                        {isLoadingEvents ? <Loader2 className="w-3 h-3 animate-spin" /> : 'Refresh'}
                      </Button>
                    </div>
                    {isLoadingEvents ? (
                      <div className="flex items-center justify-center py-8 text-muted-foreground">
                        <Loader2 className="w-5 h-5 animate-spin mr-2" />
                        Loading events...
                      </div>
                    ) : eventsList.length === 0 ? (
                      <p className="text-sm text-muted-foreground text-center py-4">No events found.</p>
                    ) : (
                      <div className="space-y-2">
                        {eventsList.map((event) => (
                          <div key={event.id} className="border rounded-lg p-3 hover:bg-muted/20 transition-colors">
                            <div className="flex items-start justify-between gap-2">
                              <div className="flex-1 min-w-0">
                                <div className="flex items-center gap-2 flex-wrap">
                                  <p className="font-medium text-sm">{event.title}</p>
                                  <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${getStatusBadge(event.status)}`}>
                                    {event.status}
                                  </span>
                                </div>
                                <p className="text-xs text-muted-foreground mt-1">
                                  📅 {event.date}
                                  {event.location && <span> &bull; 📍 {event.location}</span>}
                                </p>
                                {event.description && (
                                  <p className="text-xs text-muted-foreground mt-1 line-clamp-2">{event.description}</p>
                                )}
                              </div>
                              <div className="flex items-center gap-1 shrink-0">
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  className="text-blue-600 hover:text-blue-700 hover:bg-blue-50 dark:hover:bg-blue-900/20"
                                  onClick={() => openEditEvent(event)}
                                >
                                  <Pencil className="w-3.5 h-3.5" />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  className="text-destructive hover:text-destructive hover:bg-destructive/10"
                                  onClick={() => setDeleteEventConfirm({ open: true, id: event.id, title: event.title })}
                                >
                                  <Trash2 className="w-3.5 h-3.5" />
                                </Button>
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </TabsContent>

                {/* ============ GALLERY TAB ============ */}
                <TabsContent value="gallery" className="space-y-3 mt-4">
                  {/* Upload Form */}
                  <div className="border rounded-lg p-4 space-y-2 bg-muted/30">
                    <h3 className="font-semibold text-sm">Upload New Photo</h3>
                    <div className="space-y-1">
                      <Label htmlFor="galleryDescription" className="text-xs">Description *</Label>
                      <Textarea id="galleryDescription" value={galleryForm.description} onChange={(e) => setGalleryForm({ ...galleryForm, description: e.target.value })} placeholder="Photo description" rows={2} className="text-sm resize-none" data-testid="testid-gallery-description-input" />
                    </div>
                    <div className="space-y-1">
                      <Label htmlFor="galleryDate" className="text-xs">Date *</Label>
                      <Input id="galleryDate" type="date" value={galleryForm.date} onChange={(e) => setGalleryForm({ ...galleryForm, date: e.target.value })} className="h-8 text-sm" data-testid="testid-gallery-date-input" />
                    </div>
                    <div className="space-y-1">
                      <Label htmlFor="galleryPhoto" className="text-xs">Upload Photo *</Label>
                      <Input id="galleryPhoto" type="file" accept="image/*" onChange={(e) => setGalleryForm({ ...galleryForm, photo: e.target.files?.[0] || null })} className="h-8 text-sm" data-testid="testid-gallery-photo-input" />
                      {galleryForm.photo && <p className="text-xs text-muted-foreground">Selected: {galleryForm.photo.name}</p>}
                    </div>
                    <Button onClick={handleGalleryUpload} className="w-full h-8 text-sm" disabled={isUploadingPhoto} data-testid="testid-upload-photo-button">
                      <ImageIcon className="w-3.5 h-3.5 mr-2" />
                      {isUploadingPhoto ? 'Uploading...' : 'Upload Photo'}
                    </Button>
                  </div>

                  {/* Existing Gallery List */}
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <h3 className="font-semibold text-sm">Existing Gallery Photos</h3>
                      <Button variant="ghost" size="sm" onClick={loadGallery} disabled={isLoadingGallery}>
                        {isLoadingGallery ? <Loader2 className="w-3 h-3 animate-spin" /> : 'Refresh'}
                      </Button>
                    </div>
                    {isLoadingGallery ? (
                      <div className="flex items-center justify-center py-6 text-muted-foreground">
                        <Loader2 className="w-4 h-4 animate-spin mr-2" />
                        <span className="text-sm">Loading gallery...</span>
                      </div>
                    ) : galleryList.length === 0 ? (
                      <p className="text-sm text-muted-foreground text-center py-3">No gallery photos found.</p>
                    ) : (
                      <div className="space-y-1.5">
                        {galleryList.map((img) => (
                          <div key={img.id} className="border rounded-lg px-3 py-2 flex items-start justify-between gap-2 hover:bg-muted/20 transition-colors">
                            <p className="text-sm flex-1 min-w-0 break-words leading-snug">{img.title}</p>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="text-destructive hover:text-destructive hover:bg-destructive/10 shrink-0 h-7 w-7 p-0 mt-0.5"
                              onClick={() => setDeleteGalleryConfirm({ open: true, id: img.id, title: img.title })}
                            >
                              <Trash2 className="w-3.5 h-3.5" />
                            </Button>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </TabsContent>
              </Tabs>
            </DialogContent>
          </Dialog>
        </CardContent>
      </Card>

      {/* ===================== QUICK ACTIONS ===================== */}
      <Card>
        <CardHeader>
          <CardTitle>Quick Actions</CardTitle>
          <CardDescription>Manage members and communications</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">

          {/* Add New Member */}
          <Dialog open={addMemberOpen} onOpenChange={setAddMemberOpen}>
            <DialogTrigger asChild>
              <Button variant="outline" className="w-full justify-start" data-testid="testid-add-member-button">
                <UserPlus className="w-4 h-4 mr-2" />
                Add New Member
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto" data-testid="testid-add-member-modal">
              <DialogHeader>
                <DialogTitle>Add New Member</DialogTitle>
                <DialogDescription>Register a new member with complete details</DialogDescription>
              </DialogHeader>
              <div className="grid grid-cols-2 gap-4 py-4">
                <div className="space-y-2">
                  <Label htmlFor="tagaId">Taga ID *</Label>
                  <Input id="tagaId" value={memberForm.tagaId} onChange={(e) => setMemberForm({ ...memberForm, tagaId: e.target.value })} placeholder="Taga ID" data-testid="testid-add-member-tagaid-input" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="name">Name *</Label>
                  <Input id="name" value={memberForm.name} onChange={(e) => setMemberForm({ ...memberForm, name: e.target.value })} placeholder="Full name" data-testid="testid-add-member-name-input" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="initial">Initial *</Label>
                  <Input id="initial" value={memberForm.initial} onChange={(e) => setMemberForm({ ...memberForm, initial: e.target.value })} placeholder="Initial" data-testid="testid-add-member-initial-input" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="gender">Gender *</Label>
                  <Select value={memberForm.gender} onValueChange={(value: string) => setMemberForm({ ...memberForm, gender: value })}>
                    <SelectTrigger data-testid="testid-add-member-gender-select">
                      <SelectValue placeholder="Select gender" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="male">Male</SelectItem>
                      <SelectItem value="female">Female</SelectItem>
                      <SelectItem value="other">Other</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="fatherName">Father Name</Label>
                  <Input id="fatherName" value={memberForm.fatherName} onChange={(e) => setMemberForm({ ...memberForm, fatherName: e.target.value })} placeholder="Father's name" data-testid="testid-add-member-father-name-input" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="motherName">Mother Name</Label>
                  <Input id="motherName" value={memberForm.motherName} onChange={(e) => setMemberForm({ ...memberForm, motherName: e.target.value })} placeholder="Mother's name" data-testid="testid-add-member-mother-name-input" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="educationalQualification">Educational Qualification *</Label>
                  <Input id="educationalQualification" value={memberForm.educationalQualification} onChange={(e) => setMemberForm({ ...memberForm, educationalQualification: e.target.value })} placeholder="e.g., B.Sc Agriculture" data-testid="testid-add-member-educational-qualification-input" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="designation">Designation *</Label>
                  <Input id="designation" value={memberForm.designation} onChange={(e) => setMemberForm({ ...memberForm, designation: e.target.value })} placeholder="e.g., Agriculture Officer" data-testid="testid-add-member-designation-input" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="workingDistrict">Working District *</Label>
                  <Select value={memberForm.workingDistrict} onValueChange={(value: string) => setMemberForm({ ...memberForm, workingDistrict: value })}>
                    <SelectTrigger data-testid="testid-add-member-working-district-select">
                      <SelectValue placeholder="Select district" />
                    </SelectTrigger>
                    <SelectContent>
                      {districts.map((district) => (
                        <SelectItem key={district} value={district}>{district}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="nativeDistrict">Native District *</Label>
                  <Select value={memberForm.nativeDistrict} onValueChange={(value: string) => setMemberForm({ ...memberForm, nativeDistrict: value })}>
                    <SelectTrigger data-testid="testid-add-member-native-district-select">
                      <SelectValue placeholder="Select district" />
                    </SelectTrigger>
                    <SelectContent>
                      {districts.map((district) => (
                        <SelectItem key={district} value={district}>{district}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="recruitmentBatch">Recruitment Batch</Label>
                  <Input id="recruitmentBatch" value={memberForm.recruitmentBatch} onChange={(e) => setMemberForm({ ...memberForm, recruitmentBatch: e.target.value })} placeholder="e.g., 2020" data-testid="testid-add-member-recruitment-batch-input" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="seniorityNumber">Seniority Number</Label>
                  <Input id="seniorityNumber" value={memberForm.seniorityNumber} onChange={(e) => setMemberForm({ ...memberForm, seniorityNumber: e.target.value })} placeholder="Seniority number" data-testid="testid-add-member-seniority-number-input" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="dateOfBirth">Date of Birth *</Label>
                  <Input id="dateOfBirth" type="date" value={memberForm.dateOfBirth} onChange={(e) => setMemberForm({ ...memberForm, dateOfBirth: e.target.value })} data-testid="testid-add-member-date-of-birth-input" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="mobileNumber">Mobile Number *</Label>
                  <Input id="mobileNumber" value={memberForm.mobileNumber} onChange={(e) => setMemberForm({ ...memberForm, mobileNumber: e.target.value })} placeholder="10-digit mobile number" data-testid="testid-add-member-mobile-number-input" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="email">Email ID *</Label>
                  <Input id="email" type="email" value={memberForm.email} onChange={(e) => setMemberForm({ ...memberForm, email: e.target.value })} placeholder="email@example.com" data-testid="testid-add-member-email-input" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="tbfNumber">TBF Number</Label>
                  <Input id="tbfNumber" value={memberForm.tbfNumber} onChange={(e) => setMemberForm({ ...memberForm, tbfNumber: e.target.value })} placeholder="TBF number" data-testid="testid-add-member-tbf-number-input" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="cpsGpfNumber">CPS / GPF Number</Label>
                  <Input id="cpsGpfNumber" value={memberForm.cpsGpfNumber} onChange={(e) => setMemberForm({ ...memberForm, cpsGpfNumber: e.target.value })} placeholder="CPS or GPF number" data-testid="testid-add-member-cps-gpf-number-input" />
                </div>
                <div className="space-y-2 col-span-2">
                  <Label htmlFor="residentialAddress">Residential Address</Label>
                  <Textarea id="residentialAddress" value={memberForm.residentialAddress} onChange={(e) => setMemberForm({ ...memberForm, residentialAddress: e.target.value })} placeholder="Current residential address" rows={2} data-testid="testid-add-member-residential-address-input" />
                </div>
                <div className="space-y-2 col-span-2">
                  <Label htmlFor="permanentAddress">Permanent Address</Label>
                  <Textarea id="permanentAddress" value={memberForm.permanentAddress} onChange={(e) => setMemberForm({ ...memberForm, permanentAddress: e.target.value })} placeholder="Permanent address" rows={2} data-testid="testid-add-member-permanent-address-input" />
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setAddMemberOpen(false)} data-testid="testid-cancel-button">Cancel</Button>
                <Button onClick={handleAddMember} disabled={!memberForm.name || !memberForm.email || !memberForm.mobileNumber || isAddingMember} data-testid="testid-add-member-submit-button">
                  {isAddingMember ? <><Loader2 className="w-4 h-4 mr-2 animate-spin" />Adding...</> : <><Plus className="w-4 h-4 mr-2" />Add Member</>}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* Bulk Member Upload */}
          <Dialog open={bulkUploadOpen} onOpenChange={setBulkUploadOpen}>
            <DialogTrigger asChild>
              <Button variant="outline" className="w-full justify-start" data-testid="testid-bulk-member-upload-button">
                <Upload className="w-4 h-4 mr-2" />
                Bulk Member Upload
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl" data-testid="testid-bulk-upload-modal">
              <DialogHeader>
                <DialogTitle>Bulk Member Upload</DialogTitle>
                <DialogDescription>Upload multiple members at once using an Excel file</DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-4">
                <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
                  <h4 className="font-semibold text-sm text-blue-900 dark:text-blue-100 mb-2">Excel File Format Requirements:</h4>
                  <ul className="text-sm text-blue-800 dark:text-blue-200 space-y-1 list-disc list-inside">
                    <li>Column headers: TAGA ID, Name, Initial, Gender, Father Name, Mother Name, Educational Qualification, Designation, Working District, Native District, Recruitment Batch, Seniority Number, Residential Address, Permanent Address, Date of Birth, Mobile Number, Email ID, TBF Number, CPS/GPF Number</li>
                    <li>Supported formats: .xlsx, .csv</li>
                    <li>Maximum 500 members per upload</li>
                  </ul>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="bulkUploadFile">Select Excel File</Label>
                  <Input id="bulkUploadFile" type="file" accept=".xlsx,.xls,.csv" onChange={(e) => setBulkUploadFile(e.target.files?.[0] || null)} data-testid="testid-bulk-upload-file-input" />
                  {bulkUploadFile && <p className="text-sm text-muted-foreground">Selected: {bulkUploadFile.name}</p>}
                </div>
                <Button variant="outline" className="w-full" data-testid="testid-download-sample-template-button" onClick={() => {
                  const csvContent = 'Taga ID,Name,Initial,Gender,Father Name,Mother Name,Educational Qualification,Designation,Working District,Native District,Recruitment Batch,Seniority Number,Residential Address,Permanent Address,Date of Birth,Mobile Number,Email ID,TBF Number,CPS/GPF Number\nTAGA001,John,S,Male,Father Name,Mother Name,B.Sc Agriculture,Agriculture Officer,Chennai,Coimbatore,2020,1001,123 Main St,456 Oak Ave,1990-01-01,9876543210,john@example.com,TBF001,GPF001';
                  const blob = new Blob([csvContent], { type: 'text/csv' });
                  const url = URL.createObjectURL(blob);
                  const a = document.createElement('a');
                  a.href = url;
                  a.download = 'TAGA_Member_Upload_Template.csv';
                  document.body.appendChild(a);
                  a.click();
                  document.body.removeChild(a);
                  URL.revokeObjectURL(url);
                  toast.success('Template downloaded successfully');
                }}>
                  <Download className="w-4 h-4 mr-2" />
                  Download Sample Template
                </Button>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setBulkUploadOpen(false)} data-testid="testid-bulk-upload-cancel-button">Cancel</Button>
                <Button onClick={handleBulkUpload} disabled={!bulkUploadFile || isBulkUploading} data-testid="testid-bulk-upload-submit-button">
                  {isBulkUploading ? <><Loader2 className="w-4 h-4 mr-2 animate-spin" />Uploading...</> : <><Upload className="w-4 h-4 mr-2" />Upload Members</>}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* Send Announcement */}
          <Dialog open={announcementOpen} onOpenChange={setAnnouncementOpen}>
            <DialogTrigger asChild>
              <Button variant="outline" className="w-full justify-start" data-testid="testid-send-announcement-button">
                <MessageSquare className="w-4 h-4 mr-2" />
                Send Announcement
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl" data-testid="testid-send-announcement-modal">
              <DialogHeader>
                <DialogTitle>Send Announcement</DialogTitle>
                <DialogDescription>Send important announcements to association members. Members will be notified in their profile.</DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-4">
                <div className="space-y-2">
                  <Label htmlFor="title">Announcement Title</Label>
                  <Input id="title" value={announcementForm.title} onChange={(e) => setAnnouncementForm({ ...announcementForm, title: e.target.value })} placeholder="Enter announcement title" data-testid="testid-announcement-title-input" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="message">Message</Label>
                  <Textarea id="message" value={announcementForm.message} onChange={(e) => setAnnouncementForm({ ...announcementForm, message: e.target.value })} placeholder="Enter your announcement message" rows={5} data-testid="testid-announcement-message-input" />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="priority">Priority</Label>
                    <Select value={announcementForm.priority} onValueChange={(value: string) => setAnnouncementForm({ ...announcementForm, priority: value })}>
                      <SelectTrigger data-testid="testid-announcement-priority-select"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="normal">Normal</SelectItem>
                        <SelectItem value="high">High</SelectItem>
                        <SelectItem value="urgent">Urgent</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="sendTo">Send To</Label>
                    <Select value={announcementForm.sendTo} onValueChange={(value: string) => setAnnouncementForm({ ...announcementForm, sendTo: value })}>
                      <SelectTrigger data-testid="testid-announcement-send-to-select"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="all">All Members</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setAnnouncementOpen(false)} data-testid="testid-announcement-cancel-button">Cancel</Button>
                <Button onClick={handleSendAnnouncement} disabled={!announcementForm.title || !announcementForm.message} data-testid="testid-announcement-submit-button">
                  <Send className="w-4 h-4 mr-2" />
                  Send Announcement
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* Generate Report */}
          <Dialog open={reportOpen} onOpenChange={setReportOpen}>
            <DialogTrigger asChild>
              <Button variant="outline" className="w-full justify-start" data-testid="testid-generate-excel-report-button">
                <Download className="w-4 h-4 mr-2" />
                Generate Excel Report
              </Button>
            </DialogTrigger>
            <DialogContent data-testid="testid-generate-report-modal">
              <DialogHeader>
                <DialogTitle>Generate Excel Report</DialogTitle>
                <DialogDescription>Export member data and statistics in Excel format</DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-4">
                <div className="space-y-2">
                  <Label htmlFor="reportType">Report Type</Label>
                  <Select value={reportForm.type} onValueChange={(value: string) => setReportForm({ ...reportForm, type: value })}>
                    <SelectTrigger data-testid="testid-report-type-select"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="membership">Membership Report</SelectItem>
                      {/* <SelectItem value="financial">Financial Report</SelectItem>
                      <SelectItem value="activities">Activities Report</SelectItem>
                      <SelectItem value="district">District-wise Report</SelectItem> */}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="period">Period</Label>
                  <Select value={reportForm.period} onValueChange={(value: string) => setReportForm({ ...reportForm, period: value })}>
                    <SelectTrigger data-testid="testid-report-period-select"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="current_month">Current Month</SelectItem>
                      <SelectItem value="last_month">Last Month</SelectItem>
                      <SelectItem value="current_quarter">Current Quarter</SelectItem>
                      <SelectItem value="current_year">Current Year</SelectItem>
                      <SelectItem value="all_time">All Time</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setReportOpen(false)} data-testid="testid-generate-report-cancel-button">Cancel</Button>
                <Button onClick={handleGenerateReport} disabled={isGeneratingReport} data-testid="testid-download-excel-report-button">
                  {isGeneratingReport ? <><Loader2 className="w-4 h-4 mr-2 animate-spin" />Generating...</> : <><Download className="w-4 h-4 mr-2" />Download Excel Report</>}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* Manage District Office Bearers */}
          <Dialog open={officeBearersOpen} onOpenChange={setOfficeBearersOpen}>
            <DialogTrigger asChild>
              <Button variant="outline" className="w-full justify-start" data-testid="testid-manage-office-bearers-button">
                <Users className="w-4 h-4 mr-2" />
                Manage District Office Bearers
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-5xl max-h-[85vh] overflow-y-auto" data-testid="testid-office-bearers-modal">
              <DialogHeader>
                <DialogTitle>District Office Bearers Management</DialogTitle>
                <DialogDescription>
                  Edit district-level office bearers. Changes will be visible immediately on the public page.
                  Each district has exactly 6 fixed positions. Only name and contact fields can be edited.
                </DialogDescription>
              </DialogHeader>
              <DistrictOfficeBearersManager />
            </DialogContent>
          </Dialog>
        </CardContent>
      </Card>

      {/* ===================== MEMBER TABLE ===================== */}
      <div className="mt-6">
        <MemberListTable
          onUpdateStats={(newStats) => {
            console.log('Stats updated:', newStats);
          }}
        />
      </div>

      {/* ===================== CONFIRM DIALOGS ===================== */}
      <ConfirmDeleteDialog
        open={!!deleteResourceConfirm?.open}
        title="Delete Resource"
        description={`Are you sure you want to delete "${deleteResourceConfirm?.title}"? This action cannot be undone.`}
        onConfirm={handleDeleteResource}
        onCancel={() => setDeleteResourceConfirm(null)}
        isLoading={isDeletingResource}
      />

      <ConfirmDeleteDialog
        open={!!deleteEventConfirm?.open}
        title="Delete Event"
        description={`Are you sure you want to delete the event "${deleteEventConfirm?.title}"? This action cannot be undone.`}
        onConfirm={handleDeleteEvent}
        onCancel={() => setDeleteEventConfirm(null)}
        isLoading={isDeletingEvent}
      />

      <ConfirmDeleteDialog
        open={!!deleteGalleryConfirm?.open}
        title="Delete Gallery Photo"
        description="Are you sure you want to delete this photo? This action cannot be undone."
        onConfirm={handleDeleteGallery}
        onCancel={() => setDeleteGalleryConfirm(null)}
        isLoading={isDeletingGallery}
      />

      {/* ===================== EDIT EVENT DIALOG ===================== */}
      <Dialog open={editEventOpen} onOpenChange={(v: boolean) => { if (!v) { setEditEventOpen(false); setEditingEvent(null); } }}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Edit Event</DialogTitle>
            <DialogDescription>Update the event details below.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label>Event Title *</Label>
              <Input value={editEventForm.title} onChange={(e) => setEditEventForm({ ...editEventForm, title: e.target.value })} placeholder="Enter event title" />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Date *</Label>
                <Input type="date" value={editEventForm.date} onChange={(e) => setEditEventForm({ ...editEventForm, date: e.target.value })} />
              </div>
              <div className="space-y-2">
                <Label>Time</Label>
                <Input type="time" value={editEventForm.time} onChange={(e) => setEditEventForm({ ...editEventForm, time: e.target.value })} />
              </div>
            </div>
            <div className="space-y-2">
              <Label>Location</Label>
              <Input value={editEventForm.location} onChange={(e) => setEditEventForm({ ...editEventForm, location: e.target.value })} placeholder="Event location" />
            </div>
            <div className="space-y-2">
              <Label>Description</Label>
              <Textarea value={editEventForm.description} onChange={(e) => setEditEventForm({ ...editEventForm, description: e.target.value })} placeholder="Event description" rows={3} />
            </div>
            <div className="space-y-2">
              <Label>Status</Label>
              <Select value={editEventForm.status} onValueChange={(value: string) => setEditEventForm({ ...editEventForm, status: value })}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="upcoming">Upcoming</SelectItem>
                  <SelectItem value="ongoing">Ongoing</SelectItem>
                  <SelectItem value="completed">Completed</SelectItem>
                  <SelectItem value="cancelled">Cancelled</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => { setEditEventOpen(false); setEditingEvent(null); }} disabled={isUpdatingEvent}>Cancel</Button>
            <Button onClick={handleUpdateEvent} disabled={isUpdatingEvent || !editEventForm.title || !editEventForm.date}>
              {isUpdatingEvent ? <><Loader2 className="w-4 h-4 mr-2 animate-spin" />Updating...</> : <><Pencil className="w-4 h-4 mr-2" />Update Event</>}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

