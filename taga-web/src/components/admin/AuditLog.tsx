import React, { useState, useEffect, useCallback, useRef } from 'react';
import { Button } from '../ui/button';
import {
    Search,
    Filter,
    ChevronLeft,
    ChevronRight,
    X,
    Eye,
    Shield,
    RefreshCw,
    Clock,
    User,
    Activity,
    FileText,
    AlertTriangle,
    Play,
    Pause,
    Database,
    AlertOctagon,
    Settings,
    Users
} from 'lucide-react';
import API_BASE_URL from '../../config/api';

// ─── Types ────────────────────────────────────────────────────────────────────

interface AuditRecord {
    audit_id: string;
    user_id: string;
    username: string;
    action: string;
    module: string;
    resource_type?: string;
    resource_id?: string;
    description: string;
    old_data?: unknown;
    new_data?: unknown;
    ip_address?: string;
    user_agent?: string;
    timestamp: string;
}

interface AuditResponse {
    data: AuditRecord[];
    page: number;
    limit: number;
    total: number;
}

// ─── Constants ────────────────────────────────────────────────────────────────

const ACTIONS = ['', 'LOGIN', 'LOGIN_FAILED', 'LOGOUT', 'PASSWORD_CHANGED',
    'CREATE', 'UPDATE', 'DELETE', 'ROLE_CHANGED', 'PERMISSION_CHANGED',
    'BOOKING_CREATED', 'BOOKING_CANCELLED', 'PAYMENT_CONFIRMED'];

const MODULES = ['', 'AUTH', 'MEMBER', 'BOOKING', 'PAYMENT', 'RESOURCE', 'EVENT', 'GALLERY'];

// Accent badge styling matching custom action alerts
const ACTION_COLORS: Record<string, string> = {
    LOGIN: 'bg-green-500/10 text-green-700 border-green-500/20 dark:text-green-400',
    LOGIN_FAILED: 'bg-destructive/10 text-destructive border-destructive/20',
    LOGOUT: 'bg-slate-500/10 text-slate-700 border-slate-500/20 dark:text-slate-400',
    CREATE: 'bg-emerald-600/10 text-emerald-700 border-emerald-500/20 dark:text-emerald-400',
    UPDATE: 'bg-amber-500/10 text-amber-700 border-amber-500/20 dark:text-amber-400',
    DELETE: 'bg-destructive/10 text-destructive border-destructive/20',
    PASSWORD_CHANGED: 'bg-purple-500/10 text-purple-700 border-purple-500/20 dark:text-purple-400',
    ROLE_CHANGED: 'bg-orange-500/10 text-orange-700 border-orange-500/20 dark:text-orange-400',
    PERMISSION_CHANGED: 'bg-orange-500/10 text-orange-700 border-orange-500/20 dark:text-orange-400',
    BOOKING_CREATED: 'bg-cyan-500/10 text-cyan-700 border-cyan-500/20 dark:text-cyan-400',
    BOOKING_CANCELLED: 'bg-destructive/10 text-destructive border-destructive/20',
    PAYMENT_CONFIRMED: 'bg-green-600/10 text-green-700 border-green-500/20 dark:text-green-400',
};

const MODULE_COLORS: Record<string, { bg: string, text: string, dot: string }> = {
    AUTH: { bg: 'bg-purple-500/10', text: 'text-purple-700 dark:text-purple-400', dot: 'bg-purple-500' },
    MEMBER: { bg: 'bg-blue-500/10', text: 'text-blue-700 dark:text-blue-400', dot: 'bg-blue-500' },
    BOOKING: { bg: 'bg-cyan-500/10', text: 'text-cyan-700 dark:text-cyan-400', dot: 'bg-cyan-500' },
    PAYMENT: { bg: 'bg-green-600/10', text: 'text-green-700 dark:text-green-400', dot: 'bg-green-600' },
    RESOURCE: { bg: 'bg-amber-500/10', text: 'text-amber-700 dark:text-amber-400', dot: 'bg-amber-500' },
    EVENT: { bg: 'bg-emerald-600/10', text: 'text-emerald-700 dark:text-emerald-400', dot: 'bg-emerald-600' },
    GALLERY: { bg: 'bg-pink-500/10', text: 'text-pink-700 dark:text-pink-400', dot: 'bg-pink-500' },
};

function getActionColor(action: string): string {
    return ACTION_COLORS[action] ?? 'bg-slate-500/10 text-slate-700 border border-slate-500/20 dark:text-slate-400';
}

function getModuleStyles(mod: string) {
    return MODULE_COLORS[mod] ?? { bg: 'bg-slate-500/10', text: 'text-slate-700 dark:text-slate-400', dot: 'bg-slate-500' };
}

function getAuthToken(): string {
    return localStorage.getItem('admin_token') ?? '';
}

function formatTimestamp(ts: string): string {
    try {
        const d = new Date(ts);
        return d.toLocaleString('en-IN', {
            year: 'numeric', month: 'short', day: '2-digit',
            hour: '2-digit', minute: '2-digit', second: '2-digit',
            hour12: false,
        });
    } catch {
        return ts;
    }
}

// ─── Structured Comparative Diff Component ───────────────────────────────────

function DetailedDiffDisplay({ oldData, newData }: { oldData: unknown; newData: unknown }) {
    if (oldData == null || newData == null) return null;

    const oldObj = oldData as Record<string, unknown>;
    const newObj = newData as Record<string, unknown>;

    const allKeys = Array.from(new Set([...Object.keys(oldObj || {}), ...Object.keys(newObj || {})]));
    const changes: Array<{ key: string; oldVal: string; newVal: string; isDifferent: boolean }> = [];

    allKeys.forEach(k => {
        if (k === 'updated_at' || k === 'password') return;
        const oVal = JSON.stringify(oldObj[k]);
        const nVal = JSON.stringify(newObj[k]);
        if (oVal !== nVal) {
            changes.push({
                key: k,
                oldVal: oldObj[k] !== undefined ? String(oldObj[k]) : '—',
                newVal: newObj[k] !== undefined ? String(newObj[k]) : '—',
                isDifferent: true
            });
        }
    });

    if (changes.length === 0) {
        return (
            <div className="rounded-xl border border-border bg-muted/30 p-4 text-center text-xs text-muted-foreground">
                No differences found in tracked payload keys.
            </div>
        );
    }

    return (
        <div className="rounded-xl border border-border bg-card overflow-hidden text-xs">
            <div className="grid grid-cols-3 bg-muted px-4 py-2 font-semibold text-foreground border-b border-border">
                <span>Field</span>
                <span>Before</span>
                <span>After</span>
            </div>
            <div className="divide-y divide-border">
                {changes.map(ch => (
                    <div key={ch.key} className="grid grid-cols-3 px-4 py-2.5 items-center font-mono leading-relaxed">
                        <span className="text-muted-foreground font-sans font-medium">{ch.key}</span>
                        <span className="text-destructive line-through bg-destructive/10 px-1.5 py-0.5 rounded truncate mr-2">{ch.oldVal}</span>
                        <span className="text-primary bg-primary/10 px-1.5 py-0.5 rounded truncate font-semibold">{ch.newVal}</span>
                    </div>
                ))}
            </div>
        </div>
    );
}

// ─── JSON Diff Display (Raw Fallback) ─────────────────────────────────────────

function JsonDisplay({ label, data, colorClass }: {
    label: string;
    data: unknown;
    colorClass: string;
}) {
    if (data == null) return null;
    return (
        <div className={`rounded-xl border p-4 ${colorClass}`}>
            <p className="text-xs font-semibold uppercase tracking-wider mb-2 opacity-70">{label}</p>
            <pre className="text-[11px] font-mono whitespace-pre-wrap break-all leading-relaxed">
                {JSON.stringify(data, null, 2)}
            </pre>
        </div>
    );
}

// ─── Detail Modal ─────────────────────────────────────────────────────────────

function DetailModal({ record, onClose }: { record: AuditRecord; onClose: () => void }) {
    const isUpdate = record.action === 'UPDATE';
    const hasDiff = isUpdate && record.old_data != null && record.new_data != null;

    return (
        <div
            className="fixed inset-0 z-50 flex items-center justify-center p-4 animate-fadeIn"
            style={{ backgroundColor: 'rgba(0,0,0,0.5)', backdropFilter: 'blur(3px)' }}
            onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
        >
            <div
                className="relative w-full max-w-2xl max-h-[90vh] overflow-y-auto rounded-2xl shadow-2xl transition-all duration-300 bg-card text-card-foreground border border-border"
            >
                {/* Header */}
                <div className="sticky top-0 flex items-center justify-between px-6 py-4 border-b border-border bg-card">
                    <div className="flex items-center gap-3">
                        <div className="p-2 rounded-xl bg-primary/10 text-primary">
                            <Shield className="w-5 h-5" />
                        </div>
                        <div>
                            <h3 className="font-bold text-lg text-foreground">Audit Record Details</h3>
                            <p className="text-muted-foreground text-xs font-mono">{record.audit_id}</p>
                        </div>
                    </div>
                    <button
                        onClick={onClose}
                        className="p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
                    >
                        <X className="w-5 h-5" />
                    </button>
                </div>

                {/* Body */}
                <div className="p-6 space-y-5">
                    {/* Action badge */}
                    <div className="flex items-center gap-2">
                        <span className={`px-3 py-1 rounded-full text-xs font-extrabold tracking-wider border ${getActionColor(record.action)}`}>
                            {record.action}
                        </span>
                        <span className={`px-3 py-1 rounded-full text-xs font-semibold border border-border ${getModuleStyles(record.module).bg} ${getModuleStyles(record.module).text}`}>
                            {record.module}
                        </span>
                    </div>

                    {/* Info grid */}
                    <div className="grid grid-cols-2 gap-3">
                        {[
                            { icon: User, label: 'Actor (User ID)', value: record.user_id },
                            { icon: User, label: 'Actor Name/Email', value: record.username },
                            { icon: Activity, label: 'Target Resource ID', value: record.resource_id || '—' },
                            { icon: FileText, label: 'Target Resource Type', value: record.resource_type || '—' },
                            { icon: Clock, label: 'Timestamp (Local)', value: formatTimestamp(record.timestamp) },
                            { icon: Shield, label: 'Client IP Address', value: record.ip_address || '—' },
                        ].map(({ icon: Icon, label, value }) => (
                            <div key={label} className="rounded-xl bg-muted/30 border border-border/50 p-3">
                                <div className="flex items-center gap-2 mb-1">
                                    <Icon className="w-3.5 h-3.5 text-muted-foreground" />
                                    <span className="text-muted-foreground text-xs">{label}</span>
                                </div>
                                <p className="text-foreground text-sm font-medium break-all">{value}</p>
                            </div>
                        ))}
                    </div>

                    {/* Description */}
                    <div className="rounded-xl bg-muted/30 border border-border/50 p-4">
                        <p className="text-muted-foreground text-xs mb-1">Description</p>
                        <p className="text-foreground text-sm leading-relaxed">{record.description}</p>
                    </div>

                    {/* User Agent */}
                    {record.user_agent && (
                        <div className="rounded-xl bg-muted/30 border border-border/50 p-3">
                            <p className="text-muted-foreground text-xs mb-1">Client User Agent</p>
                            <p className="text-muted-foreground text-xs font-mono break-all leading-normal">{record.user_agent}</p>
                        </div>
                    )}

                    {/* Diff comparison section */}
                    {hasDiff && (
                        <div className="space-y-3">
                            <h4 className="text-foreground text-sm font-bold flex items-center gap-2">
                                <Activity className="w-4 h-4 text-primary" />
                                Comparison Diff (Before → After)
                            </h4>
                            <DetailedDiffDisplay oldData={record.old_data} newData={record.new_data} />
                        </div>
                    )}

                    {/* Raw payload comparison fallback */}
                    {(record.old_data != null || record.new_data != null) && !hasDiff && (
                        <div className="space-y-3">
                            <h4 className="text-foreground text-sm font-bold flex items-center gap-2">
                                <Database className="w-4 h-4 text-primary" />
                                Payload Data
                            </h4>
                            <div className="grid gap-3 grid-cols-1">
                                {record.old_data != null && (
                                    <JsonDisplay
                                        label="Deleted / Old State Data"
                                        data={record.old_data}
                                        colorClass="bg-destructive/5 border-destructive/20 text-destructive"
                                    />
                                )}
                                {record.new_data != null && (
                                    <JsonDisplay
                                        label="Created / Updated State Data"
                                        data={record.new_data}
                                        colorClass="bg-primary/5 border-primary/20 text-primary"
                                    />
                                )}
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}

// ─── Main Component ───────────────────────────────────────────────────────────

interface AuditLogProps {
    isAdmin: boolean;
}

export function AuditLog({ isAdmin }: AuditLogProps) {
    const [records, setRecords] = useState<AuditRecord[]>([]);
    const [total, setTotal] = useState(0);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [selected, setSelected] = useState<AuditRecord | null>(null);
    const [users, setUsers] = useState<string[]>([]);

    // Auto-Refresh state
    const [autoRefresh, setAutoRefresh] = useState(false);
    const autoRefreshTimerRef = useRef<NodeJS.Timeout | null>(null);

    // Filter panel state
    const [tagaId, setTagaId] = useState('');
    const [year, setYear] = useState('');
    const [month, setMonth] = useState('');
    const [action, setAction] = useState('');
    const [module, setModule] = useState('');
    const [search, setSearch] = useState('');
    const [page, setPage] = useState(1);
    const limit = 20;

    const currentYear = new Date().getFullYear();
    const years = Array.from({ length: 5 }, (_, i) => String(currentYear - i));
    const months = [
        { value: '01', label: 'January' }, { value: '02', label: 'February' },
        { value: '03', label: 'March' }, { value: '04', label: 'April' },
        { value: '05', label: 'May' }, { value: '06', label: 'June' },
        { value: '07', label: 'July' }, { value: '08', label: 'August' },
        { value: '09', label: 'September' }, { value: '10', label: 'October' },
        { value: '11', label: 'November' }, { value: '12', label: 'December' },
    ];

    const fetchUsers = useCallback(async () => {
        try {
            const params = new URLSearchParams();
            if (year) params.set('year', year);
            if (month) params.set('month', month);
            const res = await fetch(`${API_BASE_URL}/admin/audit/users?${params}`, {
                headers: { Authorization: `Bearer ${getAuthToken()}` },
            });
            if (res.ok) {
                const data = await res.json();
                setUsers(data.users ?? []);
            }
        } catch { /* silent */ }
    }, [year, month]);

    const fetchAudit = useCallback(async (pg = 1, showSpinner = true) => {
        if (showSpinner) setLoading(true);
        setError('');
        try {
            const params = new URLSearchParams();
            if (tagaId) params.set('taga_id', tagaId);
            if (year) params.set('year', year);
            if (month) params.set('month', month);
            if (action) params.set('action', action);
            if (module) params.set('module', module);
            if (search) params.set('search', search);
            params.set('page', String(pg));
            params.set('limit', String(limit));

            const res = await fetch(`${API_BASE_URL}/admin/audit?${params}`, {
                headers: { Authorization: `Bearer ${getAuthToken()}` },
            });

            if (!res.ok) {
                const body = await res.json().catch(() => ({}));
                throw new Error(body.error ?? `HTTP ${res.status}`);
            }

            const data: AuditResponse = await res.json();
            setRecords(data.data ?? []);
            setTotal(data.total ?? 0);
            setPage(pg);
        } catch (e: unknown) {
            setError(e instanceof Error ? e.message : 'Failed to load audit logs');
        } finally {
            if (showSpinner) setLoading(false);
        }
    }, [tagaId, year, month, action, module, search]);

    useEffect(() => { fetchUsers(); }, [fetchUsers]);
    useEffect(() => { fetchAudit(1, true); }, []); // eslint-disable-line react-hooks/exhaustive-deps

    useEffect(() => {
        if (autoRefresh) {
            autoRefreshTimerRef.current = setInterval(() => {
                fetchAudit(page, false);
            }, 15000);
        } else {
            if (autoRefreshTimerRef.current) {
                clearInterval(autoRefreshTimerRef.current);
            }
        }
        return () => {
            if (autoRefreshTimerRef.current) {
                clearInterval(autoRefreshTimerRef.current);
            }
        };
    }, [autoRefresh, page, fetchAudit]);

    const handleSearch = () => { fetchAudit(1, true); };
    const handleReset = () => {
        setTagaId(''); setYear(''); setMonth('');
        setAction(''); setModule(''); setSearch('');
        setTimeout(() => fetchAudit(1, true), 0);
    };

    const totalPages = Math.ceil(total / limit);

    const failedLoginsCount = records.filter(r => r.action === 'LOGIN_FAILED').length;
    const updateActionsCount = records.filter(r => r.action === 'UPDATE').length;
    const activeOperators = new Set(records.map(r => r.user_id)).size;

    if (!isAdmin) {
        return (
            <div className="flex items-center justify-center h-96">
                <div className="text-center">
                    <AlertTriangle className="w-12 h-12 text-destructive mx-auto mb-4" />
                    <h2 className="text-xl font-semibold text-foreground mb-2">Access Denied</h2>
                    <p className="text-muted-foreground">Administrator credentials are required to view audit files.</p>
                </div>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* ── Header ── */}
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                <div className="flex items-center gap-4">
                    <div className="p-3 rounded-2xl bg-primary text-primary-foreground shadow-md">
                        <Shield className="w-6 h-6" />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold text-foreground">Audit Dashboard</h1>
                        <p className="text-muted-foreground text-xs font-semibold uppercase tracking-wider mt-0.5">Secure Log Management System</p>
                    </div>
                </div>

                <div className="flex items-center gap-3 bg-card border border-border px-4 py-2 rounded-2xl shadow-sm">
                    <button
                        onClick={() => setAutoRefresh(!autoRefresh)}
                        className={`flex items-center gap-2 text-xs font-semibold px-2.5 py-1.5 rounded-xl transition-all ${
                            autoRefresh ? 'bg-primary/25 text-primary' : 'bg-transparent text-muted-foreground hover:text-foreground'
                        }`}
                    >
                        {autoRefresh ? <Pause className="w-3.5 h-3.5" /> : <Play className="w-3.5 h-3.5" />}
                        <span>{autoRefresh ? 'Auto-Refresh Active' : 'Enable Live Feed'}</span>
                    </button>
                    <div className="w-[1px] h-6 bg-border" />
                    <button
                        onClick={() => fetchAudit(page, true)}
                        className="p-1.5 rounded-xl hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                        title="Reload logs"
                    >
                        <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin text-primary' : ''}`} />
                    </button>
                </div>
            </div>

            {/* ── Metric Stats Cards ── */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                {[
                    { title: 'Total Log entries', val: total, desc: 'Across selected period', icon: Database, color: 'text-primary', bg: 'bg-primary/10' },
                    { title: 'Failed login attempts', val: failedLoginsCount, desc: 'Potential threats caught', icon: AlertOctagon, color: 'text-destructive', bg: 'bg-destructive/10' },
                    { title: 'Write/Update modifications', val: updateActionsCount, desc: 'Mutating records logged', icon: Settings, color: 'text-amber-500', bg: 'bg-amber-500/10' },
                    { title: 'Distinct user ID actors', val: activeOperators, desc: 'Unique operating identities', icon: Users, color: 'text-blue-500', bg: 'bg-blue-500/10' },
                ].map((st, i) => {
                    const Icon = st.icon;
                    return (
                        <div key={i} className="rounded-2xl border border-border bg-card text-card-foreground p-4 flex items-center justify-between shadow-sm">
                            <div className="space-y-1">
                                <p className="text-muted-foreground text-xs font-semibold">{st.title}</p>
                                <p className="text-2xl font-black text-foreground">{st.val}</p>
                                <p className="text-slate-500 text-[10px]">{st.desc}</p>
                            </div>
                            <div className={`p-3 rounded-xl ${st.bg} ${st.color}`}>
                                <Icon className="w-5 h-5" />
                            </div>
                        </div>
                    );
                })}
            </div>

            {/* ── Filter Panel ── */}
            <div
                className="rounded-2xl p-5 bg-card border border-border shadow-sm"
            >
                <div className="flex items-center gap-2 mb-4">
                    <Filter className="w-4 h-4 text-muted-foreground" />
                    <span className="text-foreground text-sm font-bold">Search & Filters</span>
                </div>
                <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
                    {/* User dropdown filter */}
                    <select
                        value={tagaId}
                        onChange={e => setTagaId(e.target.value)}
                        className="rounded-xl px-3 py-2.5 text-sm text-foreground bg-card border border-border focus:ring-1 focus:ring-primary focus:outline-none"
                    >
                        <option value="" className="bg-card text-foreground">All Users</option>
                        {users.map(u => <option key={u} value={u} className="bg-card text-foreground">{u}</option>)}
                    </select>

                    {/* Year selection */}
                    <select
                        value={year}
                        onChange={e => setYear(e.target.value)}
                        className="rounded-xl px-3 py-2.5 text-sm text-foreground bg-card border border-border focus:ring-1 focus:ring-primary focus:outline-none"
                    >
                        <option value="" className="bg-card text-foreground">All Years</option>
                        {years.map(y => <option key={y} value={y} className="bg-card text-foreground">{y}</option>)}
                    </select>

                    {/* Month selection */}
                    <select
                        value={month}
                        onChange={e => setMonth(e.target.value)}
                        className="rounded-xl px-3 py-2.5 text-sm text-foreground bg-card border border-border focus:ring-1 focus:ring-primary focus:outline-none"
                    >
                        <option value="" className="bg-card text-foreground">All Months</option>
                        {months.map(m => <option key={m.value} value={m.value} className="bg-card text-foreground">{m.label}</option>)}
                    </select>

                    {/* Action selection */}
                    <select
                        value={action}
                        onChange={e => setAction(e.target.value)}
                        className="rounded-xl px-3 py-2.5 text-sm text-foreground bg-card border border-border focus:ring-1 focus:ring-primary focus:outline-none"
                    >
                        <option value="" className="bg-card text-foreground">All Actions</option>
                        {ACTIONS.filter(Boolean).map(a => <option key={a} value={a} className="bg-card text-foreground">{a}</option>)}
                    </select>

                    {/* Module selection */}
                    <select
                        value={module}
                        onChange={e => setModule(e.target.value)}
                        className="rounded-xl px-3 py-2.5 text-sm text-foreground bg-card border border-border focus:ring-1 focus:ring-primary focus:outline-none"
                    >
                        <option value="" className="bg-card text-foreground">All Modules</option>
                        {MODULES.filter(Boolean).map(m => <option key={m} value={m} className="bg-card text-foreground">{m}</option>)}
                    </select>

                    {/* Search bar with clear button */}
                    <div className="relative">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                        <input
                            type="text"
                            value={search}
                            onChange={e => setSearch(e.target.value)}
                            onKeyDown={e => e.key === 'Enter' && handleSearch()}
                            placeholder="Type search terms..."
                            className="w-full rounded-xl pl-9 pr-8 py-2.5 text-sm text-foreground bg-card border border-border focus:ring-1 focus:ring-primary focus:outline-none placeholder-slate-400"
                        />
                        {search && (
                            <button
                                onClick={() => { setSearch(''); setTimeout(() => fetchAudit(1, true), 0); }}
                                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground p-0.5 rounded-full"
                            >
                                <X className="w-3.5 h-3.5" />
                            </button>
                        )}
                    </div>
                </div>

                <div className="flex gap-2 mt-4">
                    <Button
                        onClick={handleSearch}
                        className="text-sm px-4 py-2 rounded-xl font-semibold shadow-md bg-primary hover:bg-primary/95 text-primary-foreground border-none"
                    >
                        Apply Filters
                    </Button>
                    <Button
                        onClick={handleReset}
                        variant="ghost"
                        className="text-sm px-4 py-2 rounded-xl text-muted-foreground hover:text-foreground"
                    >
                        Reset Configuration
                    </Button>
                </div>
            </div>

            {/* ── Table Grid ── */}
            <div
                className="rounded-2xl overflow-hidden bg-card border border-border shadow-sm"
            >
                {loading && (
                    <div className="flex items-center justify-center py-24">
                        <div className="w-8 h-8 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
                        <span className="ml-3 text-muted-foreground text-sm font-medium">Reading audit database files...</span>
                    </div>
                )}

                {error && !loading && (
                    <div className="flex items-center justify-center py-24 text-destructive text-sm font-medium">
                        <AlertTriangle className="w-5 h-5 mr-2" />
                        <span>{error}</span>
                    </div>
                )}

                {!loading && !error && records.length === 0 && (
                    <div className="flex flex-col items-center justify-center py-24 text-center">
                        <FileText className="w-12 h-12 text-muted-foreground/30 mb-3" />
                        <p className="text-muted-foreground text-sm font-semibold">No audit matches found</p>
                        <p className="text-slate-400 text-xs mt-1">Try broadening your filter criteria or changing the search keyword.</p>
                    </div>
                )}

                {!loading && !error && records.length > 0 && (
                    <>
                        {/* Table Header */}
                        <div
                            className="grid text-xs font-bold uppercase tracking-widest text-muted-foreground px-5 py-3.5 bg-muted/40 border-b border-border"
                            style={{
                                gridTemplateColumns: '170px 1.2fr 130px 110px 130px 2fr 48px',
                            }}
                        >
                            <span>Time</span>
                            <span>Actor identity</span>
                            <span>Action type</span>
                            <span>Module</span>
                            <span>Affected resource</span>
                            <span>Brief Description</span>
                            <span />
                        </div>

                        {/* Rows */}
                        <div className="divide-y divide-border">
                            {records.map((r) => {
                                const modStyles = getModuleStyles(r.module);
                                return (
                                    <div
                                        key={r.audit_id}
                                        className="grid items-center px-5 py-3.5 cursor-pointer transition-all duration-200 hover:bg-muted/30"
                                        style={{
                                            gridTemplateColumns: '170px 1.2fr 130px 110px 130px 2fr 48px',
                                        }}
                                        onClick={() => setSelected(r)}
                                    >
                                        {/* Timestamp */}
                                        <span className="text-muted-foreground text-xs font-mono">
                                            {formatTimestamp(r.timestamp)}
                                        </span>

                                        {/* User Identity */}
                                        <div className="flex items-center gap-2">
                                            <div className="p-1 rounded-lg bg-muted text-muted-foreground">
                                                <User className="w-3.5 h-3.5" />
                                            </div>
                                            <div className="truncate">
                                                <p className="text-foreground text-xs font-bold truncate leading-tight">{r.user_id}</p>
                                                <p className="text-muted-foreground text-[10px] truncate">{r.username}</p>
                                            </div>
                                        </div>

                                        {/* Action badge */}
                                        <span className={`inline-flex items-center justify-center px-2 py-0.5 rounded-full text-[9px] tracking-wider uppercase font-extrabold whitespace-nowrap truncate min-w-[100px] max-w-[120px] border ${getActionColor(r.action)}`}>
                                            {r.action}
                                        </span>

                                        {/* Module style tag with small colored dot */}
                                        <div className="flex items-center gap-1.5">
                                            <div className={`w-2 h-2 rounded-full ${modStyles.dot}`} />
                                            <span className={`text-[11px] font-semibold ${modStyles.text}`}>{r.module}</span>
                                        </div>

                                        {/* Resource details */}
                                        <span className="text-muted-foreground text-xs font-mono truncate max-w-[110px]">{r.resource_id || '—'}</span>

                                        {/* Description */}
                                        <span className="text-foreground text-xs font-medium truncate pr-2">{r.description}</span>

                                        {/* View detailed record */}
                                        <button
                                            className="p-1.5 rounded-lg text-muted-foreground hover:text-primary hover:bg-primary/10 transition-colors"
                                            onClick={(e) => { e.stopPropagation(); setSelected(r); }}
                                        >
                                            <Eye className="w-4 h-4" />
                                        </button>
                                    </div>
                                );
                            })}
                        </div>
                    </>
                )}
            </div>

            {/* ── Pagination ── */}
            {totalPages > 1 && (
                <div className="flex items-center justify-between mt-4">
                    <span className="text-muted-foreground text-xs font-medium">
                        Showing page {page} of {totalPages} (Total {total} entries matching)
                    </span>
                    <div className="flex gap-2">
                        <Button
                            onClick={() => fetchAudit(page - 1, true)}
                            disabled={page <= 1}
                            variant="ghost"
                            className="rounded-xl text-muted-foreground disabled:opacity-30 px-3 py-1 text-xs border border-border hover:bg-muted"
                        >
                            <ChevronLeft className="w-3.5 h-3.5 mr-1" /> Prev
                        </Button>
                        <Button
                            onClick={() => fetchAudit(page + 1, true)}
                            disabled={page >= totalPages}
                            variant="ghost"
                            className="rounded-xl text-muted-foreground disabled:opacity-30 px-3 py-1 text-xs border border-border hover:bg-muted"
                        >
                            Next <ChevronRight className="w-3.5 h-3.5 ml-1" />
                        </Button>
                    </div>
                </div>
            )}

            {/* ── Detail Modal ── */}
            {selected && <DetailModal record={selected} onClose={() => setSelected(null)} />}
        </div>
    );
}
