import API_BASE_URL from '../config/api';

const API_BASE = API_BASE_URL;

// ==================== TOKEN MANAGEMENT ====================
const getAuthToken = (): string | null => {
  return localStorage.getItem('admin_token');
};

export const getTokenExpiry = (): number | null => {
  const expiry = localStorage.getItem('admin_token_expiry');
  return expiry ? parseInt(expiry) : null;
};

export const isTokenExpired = (): boolean => {
  const expiry = getTokenExpiry();
  if (!expiry) return true;
  return Date.now() > expiry;
};

export const getTokenRemainingTime = (): number => {
  const expiry = getTokenExpiry();
  if (!expiry) return 0;
  const remaining = expiry - Date.now();
  return remaining > 0 ? Math.floor(remaining / 1000) : 0;
};

let tokenExpiryTimer: ReturnType<typeof setTimeout> | null = null;

export const setupTokenExpiryCheck = (expiresInSeconds: number, onExpiry: () => void): void => {
  if (tokenExpiryTimer) {
    clearTimeout(tokenExpiryTimer);
  }
  tokenExpiryTimer = setTimeout(() => {
    onExpiry();
  }, expiresInSeconds * 1000);
};

export const clearTokenExpiryCheck = (): void => {
  if (tokenExpiryTimer) {
    clearTimeout(tokenExpiryTimer);
    tokenExpiryTimer = null;
  }
};

const authFetch = async (url: string, options: RequestInit = {}) => {
  if (isTokenExpired()) {
    adminLogout();
    window.location.href = '/admin-login';
    throw new Error('Session expired. Please login again.');
  }

  const token = getAuthToken();
  if (!token) {
    window.location.href = '/admin-login';
    throw new Error('No authentication token found');
  }

  const headers = {
    ...options.headers,
    'Authorization': `Bearer ${token}`,
  };

  return fetch(url, { ...options, headers });
};

// ==================== Admin Login ====================
export interface AdminLoginRequest {
  username: string;
  password: string;
}

export interface AdminLoginResponse {
  token: string;
  role: string;
  expires_in: number;
}

export const adminLogin = async (credentials: AdminLoginRequest): Promise<AdminLoginResponse> => {
  const response = await fetch(`${API_BASE}/admin/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(credentials),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Login failed');
  }

  const data = await response.json();
  localStorage.setItem('admin_token', data.token);
  localStorage.setItem('admin_role', data.role);
  if (data.expires_in) {
    const expiresAt = Date.now() + (data.expires_in * 1000);
    localStorage.setItem('admin_token_expiry', expiresAt.toString());
  }

  return data;
};

export const adminLogout = (): void => {
  localStorage.removeItem('admin_token');
  localStorage.removeItem('admin_role');
  localStorage.removeItem('admin_token_expiry');
  clearTokenExpiryCheck();
};

export const isAdminLoggedIn = (): boolean => {
  const hasToken = !!localStorage.getItem('admin_token');
  const isExpired = isTokenExpired();
  if (hasToken && isExpired) {
    adminLogout();
    return false;
  }
  return hasToken && !isExpired;
};

// ==================== Events Management ====================
export interface CreateEventData {
  title: string;
  date: string;
  location?: string;
  description?: string;
  status?: string;
  image?: File;
}

export interface Event {
  id: string;
  title: string;
  date: string;
  location: string;
  description: string;
  attendees: number;
  status: string;
  imageUrl?: string;
}

export const createEvent = async (eventData: CreateEventData): Promise<Event> => {
  const formData = new FormData();
  formData.append('title', eventData.title);
  formData.append('date', eventData.date);
  if (eventData.location) formData.append('location', eventData.location);
  if (eventData.description) formData.append('description', eventData.description);
  if (eventData.status) formData.append('status', eventData.status);
  if (eventData.image) formData.append('image', eventData.image);

  const response = await authFetch(`${API_BASE}/admin/events/create`, {
    method: 'POST',
    body: formData,
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to create event');
  }

  const data = await response.json();
  return data.event;
};

export const updateEvent = async (id: string, eventData: Partial<CreateEventData>): Promise<void> => {
  const formData = new FormData();
  if (eventData.title) formData.append('title', eventData.title);
  if (eventData.date) formData.append('date', eventData.date);
  if (eventData.location) formData.append('location', eventData.location);
  if (eventData.description) formData.append('description', eventData.description);
  if (eventData.status) formData.append('status', eventData.status);
  if (eventData.image) formData.append('image', eventData.image);

  const response = await authFetch(`${API_BASE}/admin/events/${id}`, {
    method: 'PUT',
    body: formData,
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to update event');
  }
};

export const deleteEvent = async (id: string): Promise<void> => {
  const response = await authFetch(`${API_BASE}/admin/events/${id}`, {
    method: 'DELETE',
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to delete event');
  }
};

// ==================== Resources Management ====================
export interface UploadResourceData {
  categoryId: string;
  title: string;
  year: string;
  subcategory?: string;
  file: File;
}

export interface ResourceDocument {
  title: string;
  year: string;
  subcategory?: string;
  url?: string;
}

export interface ResourceCategoryWithDocs {
  id: string;
  name: string;
  documents: ResourceDocument[];
  subcategories?: string[];
}

export const uploadResource = async (resourceData: UploadResourceData): Promise<void> => {
  const formData = new FormData();
  formData.append('categoryId', resourceData.categoryId);
  formData.append('title', resourceData.title);
  formData.append('year', resourceData.year);
  if (resourceData.subcategory) formData.append('subcategory', resourceData.subcategory);
  formData.append('file', resourceData.file);

  const response = await authFetch(`${API_BASE}/admin/resources/upload`, {
    method: 'POST',
    body: formData,
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to upload resource');
  }
};

export const deleteResource = async (categoryId: string, documentTitle: string): Promise<void> => {
  const encodedTitle = encodeURIComponent(documentTitle);
  const response = await authFetch(`${API_BASE}/admin/resources/${categoryId}/${encodedTitle}`, {
    method: 'DELETE',
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to delete resource');
  }
};

export interface ExternalLinkItem {
  title: string;
  url: string;
}

export const getAdminExternalLinks = async (): Promise<ExternalLinkItem[]> => {
  const token = getAuthToken() || localStorage.getItem('member_token');
  const headers: Record<string, string> = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE}/resources/external-links`, { headers });
  if (!response.ok) {
    throw new Error('Failed to fetch external links');
  }
  return response.json();
};

export const addExternalLink = async (link: ExternalLinkItem): Promise<void> => {
  const response = await authFetch(`${API_BASE}/admin/resources/external-links`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(link),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to add external link');
  }
};

export const deleteExternalLink = async (title: string): Promise<void> => {
  const encodedTitle = encodeURIComponent(title);
  const response = await authFetch(`${API_BASE}/admin/resources/external-links/${encodedTitle}`, {
    method: 'DELETE',
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to delete external link');
  }
};

// Fetch all resource categories WITH their documents
export const getResourceCategoriesWithDocs = async (): Promise<ResourceCategoryWithDocs[]> => {
  const token = getAuthToken() || localStorage.getItem('member_token');
  const headers: Record<string, string> = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE}/resources/all`, { headers });
  if (!response.ok) {
    // Fallback: fetch categories then documents for each
    const catResponse = await fetch(`${API_BASE}/resources`, { headers });
    if (!catResponse.ok) throw new Error('Failed to fetch resource categories');
    const categories: { id: string; name: string }[] = await catResponse.json();

    const withDocs = await Promise.all(
      categories.map(async (cat) => {
        try {
          const docResponse = await fetch(`${API_BASE}/resources/${cat.id}`, { headers });
          const documents: ResourceDocument[] = docResponse.ok ? await docResponse.json() : [];
          return { ...cat, documents: documents || [] };
        } catch {
          return { ...cat, documents: [] };
        }
      })
    );
    return withDocs;
  }
  return response.json();
};

// ==================== Gallery Management ====================
export interface UploadGalleryData {
  title: string;
  event: string;
  date: string;
  year: number;
  image: File;
}

export interface GalleryImage {
  id: string;
  title: string;
  event: string;
  imageUrl: string;
  date: string;
  year: number;
}

export const uploadGalleryImage = async (galleryData: UploadGalleryData): Promise<GalleryImage> => {
  const formData = new FormData();
  formData.append('title', galleryData.title);
  formData.append('event', galleryData.event);
  formData.append('date', galleryData.date);
  formData.append('year', galleryData.year.toString());
  formData.append('image', galleryData.image);

  const response = await authFetch(`${API_BASE}/admin/gallery/upload`, {
    method: 'POST',
    body: formData,
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to upload gallery image');
  }

  const data = await response.json();
  return data.image;
};

export const deleteGalleryImage = async (id: string): Promise<void> => {
  const response = await authFetch(`${API_BASE}/admin/gallery/${id}`, {
    method: 'DELETE',
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to delete gallery image');
  }
};

// ==================== Public Get Methods (No Auth Required) ====================
export const getEvents = async (): Promise<Event[]> => {
  const response = await fetch(`${API_BASE}/events/upcoming`);
  if (!response.ok) {
    throw new Error('Failed to fetch events');
  }
  return response.json();
};

export const getGallery = async (): Promise<GalleryImage[]> => {
  const response = await fetch(`${API_BASE}/gallery`);
  if (!response.ok) {
    throw new Error('Failed to fetch gallery');
  }
  return response.json();
};

export const getResourceCategories = async (): Promise<any[]> => {
  const token = getAuthToken() || localStorage.getItem('member_token');
  const headers: Record<string, string> = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  const response = await fetch(`${API_BASE}/resources`, { headers });
  if (!response.ok) {
    throw new Error('Failed to fetch resources');
  }
  return response.json();
};

// ==================== Member Management ====================
export interface AddMemberData {
  tagaId: string;
  name: string;
  initial: string;
  gender: string;
  fatherName?: string;
  motherName?: string;
  educationalQualification: string;
  designation: string;
  workingDistrict: string;
  nativeDistrict: string;
  recruitmentBatch?: string;
  seniorityNumber?: string;
  residentialAddress?: string;
  permanentAddress?: string;
  dateOfBirth: string;
  mobileNumber: string;
  email: string;
  tbfNumber?: string;
  cpsGpfNumber?: string;
  paymentStatus?: string;
}

export interface AddMemberResponse {
  message: string;
  member_id: string;
  temp_password: string;
}

export const addMember = async (memberData: AddMemberData): Promise<AddMemberResponse> => {
  const response = await authFetch(`${API_BASE}/admin/members/add`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(memberData),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to add member');
  }

  return response.json();
};

export interface BulkUploadResponse {
  message: string;
  success_count: number;
  failed_count: number;
  failed: Array<{
    username: string;
    errors: string[];
  }>;
}

export const bulkUploadMembers = async (file: File): Promise<{ message: string; success_count: number; failed_count: number; failed?: any[] }> => {
  const formData = new FormData();
  formData.append('file', file);

  const response = await authFetch(`${API_BASE}/admin/members/bulk-upload`, {
    method: 'POST',
    body: formData,
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || error.message || 'Bulk upload failed');
  }

  return response.json();
};

// ==================== Reports ====================
export const generateMemberReport = async (reportType: string, period: string): Promise<Blob> => {
  const token = getAuthToken();
  if (!token) {
    throw new Error('No authentication token found');
  }

  const response = await fetch(`${API_BASE}/admin/reports/members?report_type=${reportType}&period=${period}`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    throw new Error('Failed to generate report');
  }

  return response.blob();
};

// ==================== Announcements ====================
export interface AnnouncementData {
  title: string;
  message: string;
  priority: 'normal' | 'high' | 'urgent';
  sendTo: 'all' | 'paid' | 'district';
}

export const sendAnnouncement = async (announcementData: AnnouncementData): Promise<void> => {
  const response = await authFetch(`${API_BASE}/admin/announcements/send`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(announcementData),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to send announcement');
  }
};

// ==================== Member List ====================
export interface MemberListItem {
  id: string;
  tagaId: string;
  name: string;
  initial: string;
  gender: string;
  district: string;
  designation: string;
  recruitment_batch: string;
  mobile_number: string;
  email: string;
  payment_status: string;
  membership_status: string;
}

export interface MemberListResponse {
  members: MemberListItem[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface DistrictOption {
  name: string;
  count: number;
}

export const getMembersList = async (
  page: number = 1,
  limit: number = 25,
  search: string = '',
  district: string = '',
  paymentStatus: string = ''
): Promise<MemberListResponse> => {
  const params = new URLSearchParams();
  params.append('page', page.toString());
  params.append('limit', limit.toString());
  if (search) params.append('search', search);
  if (district && district !== 'all') params.append('district', district);
  if (paymentStatus && paymentStatus !== 'all') params.append('payment_status', paymentStatus);

  const response = await authFetch(`${API_BASE}/admin/members?${params.toString()}`, {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('Failed to fetch members');
  }

  return response.json();
};

export const getMemberDistricts = async (): Promise<DistrictOption[]> => {
  const response = await authFetch(`${API_BASE}/admin/members/districts`, {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('Failed to fetch districts');
  }

  return response.json();
};

export const getMemberStats = async (): Promise<{
  totalMembers: number;
  activeMembers: number;
  unpaid: number;
  newThisMonth: number;
}> => {
  const response = await authFetch(`${API_BASE}/admin/members/stats`, {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error('Failed to fetch stats');
  }

  return response.json();
};

// ==================== OFFICE BEARERS TYPES ====================

export interface DistrictBearer {
  name: string;
  title: string;
  contact: string;
  backup_path?: string;
}

export interface UpdateDistrictBearersResponse {
  message: string;
  district: string;
  bearers: DistrictBearer[];
}

export interface BackupRestoreResponse {
  message: string;
  backup_file: string;
}

export interface BackupsListResponse {
  backups: string[];
}

// ==================== OFFICE BEARERS API FUNCTIONS ====================

/**
 * Get all districts for office bearers management
 * @returns Promise<string[]> - List of district names
 */
export const getAllDistricts = async (): Promise<string[]> => {
  const token = getAuthToken();

  if (!token) {
    throw new Error('No admin token found. Please login again.');
  }

  const response = await fetch(`${API_BASE}/admin/office-bearers/districts`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Failed to fetch districts' }));
    throw new Error(error.error || `HTTP ${response.status}: Failed to fetch districts`);
  }

  const data = await response.json();
  return data.districts;
};

/**
 * Get office bearers for a specific district
 * @param district - District name
 * @returns Promise<DistrictBearer[]> - Array of exactly 6 bearers
 */
export const getDistrictBearers = async (district: string): Promise<DistrictBearer[]> => {
  const token = getAuthToken();

  if (!token) {
    throw new Error('No admin token found. Please login again.');
  }

  if (!district) {
    throw new Error('District name is required');
  }

  const encodedDistrict = encodeURIComponent(district);
  const response = await fetch(`${API_BASE}/admin/office-bearers/district/${encodedDistrict}`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Failed to fetch district bearers' }));
    throw new Error(error.error || `HTTP ${response.status}: Failed to fetch district bearers`);
  }

  const data = await response.json();

  // Ensure we always return an array (even if empty)
  if (!data.bearers || !Array.isArray(data.bearers)) {
    console.warn('Invalid bearers data received:', data);
    return [];
  }

  return data.bearers;
};

/**
 * Update all office bearers for a specific district
 * @param district - District name
 * @param bearers - Array of exactly 6 bearers
 * @returns Promise<UpdateDistrictBearersResponse>
 */
export const updateDistrictBearers = async (
  district: string,
  bearers: DistrictBearer[]
): Promise<UpdateDistrictBearersResponse> => {
  const token = getAuthToken();

  if (!token) {
    throw new Error('No admin token found. Please login again.');
  }

  if (!district) {
    throw new Error('District name is required');
  }

  if (!bearers || bearers.length !== 6) {
    throw new Error('Exactly 6 bearers are required');
  }

  // Validate each bearer has required fields
  for (let i = 0; i < bearers.length; i++) {
    if (!bearers[i].title) {
      throw new Error(`Bearer at position ${i + 1} is missing a title`);
    }
  }

  const encodedDistrict = encodeURIComponent(district);
  const response = await fetch(`${API_BASE}/admin/office-bearers/district/${encodedDistrict}`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(bearers),
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Failed to update district bearers' }));
    throw new Error(error.error || `HTTP ${response.status}: Failed to update district bearers`);
  }

  return await response.json();
};

/**
 * Restore district bearers from a backup file
 * @param backupFile - Path to backup file
 * @returns Promise<BackupRestoreResponse>
 */
export const restoreBackup = async (backupFile: string): Promise<BackupRestoreResponse> => {
  const token = getAuthToken();

  if (!token) {
    throw new Error('No admin token found. Please login again.');
  }

  if (!backupFile) {
    throw new Error('Backup file path is required');
  }

  const response = await fetch(`${API_BASE}/admin/office-bearers/backup/restore`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ backup_file: backupFile }),
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Failed to restore backup' }));
    throw new Error(error.error || `HTTP ${response.status}: Failed to restore backup`);
  }

  return await response.json();
};

/**
 * List all available backup files
 * @returns Promise<string[]> - List of backup file paths
 */
export const listBackups = async (): Promise<string[]> => {
  const token = getAuthToken();

  if (!token) {
    throw new Error('No admin token found. Please login again.');
  }

  const response = await fetch(`${API_BASE}/admin/office-bearers/backups`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Failed to list backups' }));
    throw new Error(error.error || `HTTP ${response.status}: Failed to list backups`);
  }

  const data = await response.json();
  return data.backups || [];
};
// ==================== Member Export to Excel ====================
export const exportMembersToExcel = async (
  district: string = '',
  paymentStatus: string = ''
): Promise<Blob> => {
  const token = getAuthToken();
  if (!token) {
    throw new Error('No authentication token found');
  }

  const params = new URLSearchParams();
  if (district && district !== 'all') params.append('district', district);
  if (paymentStatus && paymentStatus !== 'all') params.append('payment_status', paymentStatus);

  // Debug: Log what's being sent
  console.log('=== Export Excel Debug ===');
  console.log('District filter:', district);
  console.log('Payment Status filter:', paymentStatus);
  console.log('Params string:', params.toString());

  const url = `${API_BASE}/admin/members/export${params.toString() ? `?${params.toString()}` : ''}`;
  console.log('Export URL:', url);

  const response = await fetch(url, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Failed to export members' }));
    console.error('Export error response:', error);
    throw new Error(error.error || 'Failed to export members');
  }

  console.log('Export successful, blob size:', response.headers.get('content-length'));
  return response.blob();
};