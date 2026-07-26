import React, { useState, useEffect } from 'react';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '../ui/select';
import { Input } from '../ui/input';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
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
import { toast } from 'sonner';
import {
    Loader2,
    Save,
    RotateCcw,
    AlertTriangle,
    CheckCircle,
    User,
    Shield,
    Users as UsersIcon,
} from 'lucide-react';
import {
    getAllDistricts,
    getDistrictBearers,
    updateDistrictBearers,
    DistrictBearer,
} from '../../api/adminContent';

const POSITIONS = [
    'President',
    'Secretary',
    'Treasurer',
    'Joint Secretary (Women)',
    'Joint Secretary (Seed)',
    'Joint Secretary (Marketing)',
];

const POSITION_ICON: Record<string, React.ReactNode> = {
    'President': <Shield className="w-3.5 h-3.5 text-muted-foreground" />,
    'Secretary': <User className="w-3.5 h-3.5 text-muted-foreground" />,
    'Treasurer': <Shield className="w-3.5 h-3.5 text-muted-foreground" />,
    'Joint Secretary (Women)': <UsersIcon className="w-3.5 h-3.5 text-muted-foreground" />,
    'Joint Secretary (Seed)': <Shield className="w-3.5 h-3.5 text-muted-foreground" />,
    'Joint Secretary (Marketing)': <Shield className="w-3.5 h-3.5 text-muted-foreground" />,
};

const DistrictOfficeBearersManager: React.FC = () => {
    const [districts, setDistricts] = useState<string[]>([]);
    const [selectedDistrict, setSelectedDistrict] = useState<string>('');
    const [bearers, setBearers] = useState<DistrictBearer[]>([]);
    const [originalBearers, setOriginalBearers] = useState<DistrictBearer[]>([]);
    const [loading, setLoading] = useState(false);
    const [saving, setSaving] = useState(false);
    const [loadingDistricts, setLoadingDistricts] = useState(false);
    const [showSaveConfirm, setShowSaveConfirm] = useState(false);
    const [hasChanges, setHasChanges] = useState(false);
    const [validationErrors, setValidationErrors] = useState<Record<number, string>>({});

    useEffect(() => { loadDistricts(); }, []);

    useEffect(() => {
        if (selectedDistrict) loadBearers(selectedDistrict);
        else { setBearers([]); setOriginalBearers([]); setHasChanges(false); }
    }, [selectedDistrict]);

    useEffect(() => {
        setHasChanges(JSON.stringify(bearers) !== JSON.stringify(originalBearers));
    }, [bearers, originalBearers]);

    const normalize = (list: DistrictBearer[]): DistrictBearer[] =>
        POSITIONS.map(position => {
            const match = list.find(b => b.title === position);
            return match ?? { name: '', title: position, contact: '' };
        });

    const loadDistricts = async () => {
        setLoadingDistricts(true);
        try {
            const list = await getAllDistricts();
            setDistricts(list);
            if (list.length > 0) setSelectedDistrict(list[0]);
        } catch (err) {
            toast.error(err instanceof Error ? err.message : 'Failed to load districts');
        } finally {
            setLoadingDistricts(false);
        }
    };

    const loadBearers = async (district: string) => {
        setLoading(true);
        try {
            const list = await getDistrictBearers(district);
            const normalized = normalize(list);
            setBearers(normalized);
            setOriginalBearers(JSON.parse(JSON.stringify(normalized)));
            setValidationErrors({});
        } catch (err) {
            toast.error(err instanceof Error ? err.message : 'Failed to load bearers');
            const empty = normalize([]);
            setBearers(empty);
            setOriginalBearers(JSON.parse(JSON.stringify(empty)));
        } finally {
            setLoading(false);
        }
    };

    const handleBearerChange = (index: number, field: 'name' | 'contact', value: string) => {
        const updated = [...bearers];
        updated[index] = { ...updated[index], [field]: value };
        setBearers(updated);
        if (validationErrors[index]) {
            const errs = { ...validationErrors };
            delete errs[index];
            setValidationErrors(errs);
        }
    };

    const validateBearers = (): boolean => {
        const errors: Record<number, string> = {};
        bearers.forEach((b, i) => {
            if (b.contact && !/^[0-9]{10}$/.test(b.contact))
                errors[i] = 'Must be exactly 10 digits';
        });
        setValidationErrors(errors);
        return Object.keys(errors).length === 0;
    };

    const handleSave = () => {
        if (!validateBearers()) {
            toast.error('Fix validation errors before saving');
            return;
        }
        setShowSaveConfirm(true);
    };

    const confirmSave = async () => {
        setShowSaveConfirm(false);
        setSaving(true);
        try {
            await updateDistrictBearers(selectedDistrict, bearers);
            toast.success(
                <div className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    <span>Bearers updated successfully!</span>
                </div>
            );
            setOriginalBearers(JSON.parse(JSON.stringify(bearers)));
            setHasChanges(false);
        } catch (err) {
            toast.error(err instanceof Error ? err.message : 'Failed to update bearers');
        } finally {
            setSaving(false);
        }
    };

    const handleRevert = () => {
        setBearers(JSON.parse(JSON.stringify(originalBearers)));
        setValidationErrors({});
        toast.info('Changes reverted');
    };

    return (
        <div className="space-y-4">

            {/* ── Header ── */}
            <div className="flex items-center gap-2.5 pb-3 border-b">
                <div className="p-1.5 rounded-md bg-primary/10">
                    <UsersIcon className="w-4 h-4 text-primary" />
                </div>
                <div className="flex-1 min-w-0">
                    <h2 className="text-sm font-semibold leading-tight">District Office Bearers</h2>
                    <p className="text-xs text-muted-foreground">6 fixed positions · changes publish immediately</p>
                </div>
                {hasChanges && (
                    <span className="inline-flex items-center gap-1 text-[11px] font-medium text-amber-600 bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-800 px-2 py-0.5 rounded-full whitespace-nowrap">
                        <AlertTriangle className="w-3 h-3" />
                        Unsaved
                    </span>
                )}
            </div>

            {/* ── District selector ── */}
            <div className="flex items-center gap-2">
                <label className="text-xs font-medium text-muted-foreground whitespace-nowrap w-16 shrink-0">
                    District
                </label>
                {loadingDistricts ? (
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <Loader2 className="w-3.5 h-3.5 animate-spin" />
                        Loading…
                    </div>
                ) : (
                    <Select value={selectedDistrict} onValueChange={setSelectedDistrict} disabled={loading || saving}>
                        <SelectTrigger className="h-8 text-sm flex-1">
                            <SelectValue placeholder="Choose a district…" />
                        </SelectTrigger>
                        <SelectContent className="max-h-64">
                            {districts.map(d => (
                                <SelectItem key={d} value={d} className="text-sm">{d}</SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                )}
            </div>

            {/* ── Loading ── */}
            {loading && (
                <div className="flex items-center justify-center gap-2 py-8 text-sm text-muted-foreground">
                    <Loader2 className="w-5 h-5 animate-spin text-primary" />
                    Loading bearers…
                </div>
            )}

            {/* ── 2-column card grid ── */}
            {!loading && selectedDistrict && bearers.length > 0 && (
                <div className="grid grid-cols-2 gap-2">
                    {bearers.map((bearer, i) => (
                        <div
                            key={i}
                            className="rounded-lg border bg-card p-3 space-y-2 hover:border-primary/40 transition-colors"
                        >
                            {/* Card header — serial + position */}
                            <div className="flex items-center gap-1.5">
                                <span className="text-[10px] font-semibold text-muted-foreground bg-muted w-4 h-4 rounded flex items-center justify-center shrink-0">
                                    {i + 1}
                                </span>
                                {POSITION_ICON[bearer.title]}
                                <span className="text-xs font-semibold leading-tight flex-1 min-w-0">
                                    {bearer.title}
                                </span>
                            </div>

                            {/* Name field */}
                            <div className="space-y-0.5">
                                <label className="text-[10px] font-medium text-muted-foreground uppercase tracking-wide">
                                    Name
                                </label>
                                <Input
                                    type="text"
                                    value={bearer.name}
                                    onChange={e => handleBearerChange(i, 'name', e.target.value)}
                                    placeholder="Full name"
                                    className="h-7 text-xs px-2"
                                    data-testid={`testid-bearer-name-${i}`}
                                />
                            </div>

                            {/* Contact field */}
                            <div className="space-y-0.5">
                                <label className="text-[10px] font-medium text-muted-foreground uppercase tracking-wide">
                                    Mobile
                                </label>
                                <Input
                                    type="tel"
                                    value={bearer.contact}
                                    onChange={e => handleBearerChange(i, 'contact', e.target.value)}
                                    placeholder="10-digit number"
                                    maxLength={10}
                                    className={`h-7 text-xs px-2 ${validationErrors[i] ? 'border-red-500 focus-visible:ring-red-500' : ''}`}
                                    data-testid={`testid-bearer-contact-${i}`}
                                />
                                {validationErrors[i] && (
                                    <p className="flex items-center gap-1 text-[10px] text-red-500">
                                        <AlertTriangle className="w-2.5 h-2.5 shrink-0" />
                                        {validationErrors[i]}
                                    </p>
                                )}
                            </div>
                        </div>
                    ))}
                </div>
            )}

            {/* ── Action bar ── */}
            {!loading && selectedDistrict && (
                <div className="flex items-center justify-end gap-2 pt-1">
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={handleRevert}
                        disabled={!hasChanges || saving}
                        className="h-7 px-3 text-xs gap-1.5"
                        data-testid="testid-revert-button"
                    >
                        <RotateCcw className="w-3 h-3" />
                        Revert
                    </Button>
                    <Button
                        size="sm"
                        onClick={handleSave}
                        disabled={!hasChanges || saving}
                        className="h-7 px-3 text-xs gap-1.5 bg-primary hover:bg-primary/90 min-w-[100px]"
                        data-testid="testid-save-button"
                    >
                        {saving ? (
                            <><Loader2 className="w-3 h-3 animate-spin" /> Saving…</>
                        ) : (
                            <><Save className="w-3 h-3" /> Save Changes</>
                        )}
                    </Button>
                </div>
            )}

            {/* ── Empty state ── */}
            {!loading && !loadingDistricts && districts.length === 0 && (
                <div className="text-center py-10 rounded-lg border border-dashed bg-muted/20">
                    <UsersIcon className="w-8 h-8 text-muted-foreground mx-auto mb-2" />
                    <p className="text-sm font-medium">No districts available</p>
                    <p className="text-xs text-muted-foreground mt-0.5">
                        Ensure the district data file exists and reload.
                    </p>
                </div>
            )}

            {/* ── Confirm dialog ── */}
            <AlertDialog open={showSaveConfirm} onOpenChange={setShowSaveConfirm}>
                <AlertDialogContent className="max-w-sm">
                    <AlertDialogHeader>
                        <AlertDialogTitle className="flex items-center gap-2 text-base">
                            <AlertTriangle className="w-4 h-4 text-amber-500" />
                            Confirm Update
                        </AlertDialogTitle>
                        <AlertDialogDescription asChild>
                            <div className="space-y-2 text-sm">
                                <p>
                                    Update office bearers for{' '}
                                    <strong className="text-foreground">{selectedDistrict}</strong>?
                                </p>
                                <ul className="bg-muted/30 rounded-md p-2.5 space-y-1 text-muted-foreground text-xs list-disc list-inside">
                                    <li>Public page updates immediately</li>
                                    <li>Automatic backup created before save</li>
                                    <li>All existing bearer data for this district is replaced</li>
                                </ul>
                            </div>
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel disabled={saving} className="h-8 text-xs">Cancel</AlertDialogCancel>
                        <AlertDialogAction
                            onClick={confirmSave}
                            disabled={saving}
                            className="h-8 text-xs bg-primary hover:bg-primary/90 min-w-[100px]"
                        >
                            {saving ? (
                                <><Loader2 className="w-3.5 h-3.5 mr-1 animate-spin" /> Saving…</>
                            ) : (
                                'Confirm Save'
                            )}
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </div>
    );
};

export default DistrictOfficeBearersManager;