import React, { useEffect, useState } from 'react';
import { Button } from './components/ui/button';
import { Separator } from './components/ui/separator';
import { Toaster } from './components/ui/sonner';
import { toast } from 'sonner';
import {
    Home,
    Users,
    FileText,
    LogIn,
    UserCheck,
    MessageSquare,
    Menu,
    X,
    Building2,
    Calendar,
    CreditCard,
    Shield
} from 'lucide-react';
import { AboutUs } from './components/AboutUs';
import { OfficeBearers } from "./components/OfficeBearers";
import { MemberLogin } from './components/MemberLogin';
import { AdminLogin } from './components/AdminLogin';
import { MembersDashboard } from './components/MembersDashboard';
import { Resources } from './components/Resources';
import { Grievance } from './components/Grievance';
import { TAGATowers } from './components/TAGATowers';
import { Membership } from './components/Membership';
import { Events } from './components/Events';
import { AuditLog } from './components/admin/AuditLog';
import { getLogo } from './api/logo';
import API_BASE_URL from "./config/api";

type Page =
    'home'
    | 'office-bearers'
    | 'member-login'
    | 'admin-login'
    | 'members'
    | 'resources'
    | 'grievance'
    | 'taga-towers'
    | 'membership'
    | 'events'
    | 'audit-log';

type UserType = 'general' | 'member' | 'subscriber';

const authenticatedPages: Page[] = ['members', 'resources', 'grievance', 'taga-towers', 'membership'];
const publicPages: Page[] = ['home', 'office-bearers', 'events', 'member-login', 'admin-login'];

// 🔥 Helper: Read isPaid directly from localStorage (synchronous, no state lag)
function getIsPaidFromStorage(): boolean {
    try {
        const user = JSON.parse(localStorage.getItem('user') || '{}');
        // 🔥 ONLY check isPaid from backend (which only sets true for Annual Subscription)
        return user.isPaid === true || user.subscription_active === true;
    } catch {
        return false;
    }
}

export default function App() {
    const API_BASE = API_BASE_URL;
    const [currentPage, setCurrentPage] = useState<Page>('home');
    const [isLoggedIn, setIsLoggedIn] = useState(false);
    const [isAdmin, setIsAdmin] = useState(false);
    const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
    const [userType, setUserType] = useState<UserType>('general');
    const [logoUrl, setLogoUrl] = useState<string | null>(null);
    const [isInitialized, setIsInitialized] = useState(false);
    const [copyrightClickCount, setCopyrightClickCount] = useState(0);

    const handleCopyrightClick = () => {
        if (!isLoggedIn || !isAdmin) {
            return;
        }
        setCopyrightClickCount(prev => {
            const next = prev + 1;
            if (next >= 5) {
                setCurrentPage('audit-log');
                toast.success("Secret Access Granted: Opening Audit Logs");
                return 0;
            }
            return next;
        });
    };

    // ==================== RESTORE SESSION ON PAGE LOAD ====================
    useEffect(() => {
        const params = new URLSearchParams(window.location.search);
        const pageParam = params.get('page') as Page | null;
        const validQueryPages: Page[] = ['member-login', 'admin-login', 'home'];

        // Support /internal/audit direct URL
        if (window.location.pathname === '/internal/audit') {
            window.history.replaceState({}, '', window.location.pathname);
            // Will set to audit-log after auth is checked below
        }

        if (pageParam && validQueryPages.includes(pageParam)) {
            setCurrentPage(pageParam);
            window.history.replaceState({}, '', window.location.pathname);
            setIsInitialized(true);
            return;
        }

        // 5-minute grace buffer to avoid race conditions on slow mobile networks
        const GRACE_MS = 5 * 60 * 1000;

        try {
            // Use localStorage for lastPage (sessionStorage is cleared when mobile browsers kill tabs)
            const savedPage = localStorage.getItem('lastPage') as Page;

            const adminToken = localStorage.getItem('admin_token');
            const adminExpiry = localStorage.getItem('admin_token_expiry');
            const isAdminValid = adminToken && adminExpiry && (Date.now() - GRACE_MS) < parseInt(adminExpiry);

            const memberToken = localStorage.getItem('member_token');
            const memberExpiry = localStorage.getItem('member_token_expiry');
            const isMemberValid = memberToken && memberExpiry && (Date.now() - GRACE_MS) < parseInt(memberExpiry);

            let restoredPage: Page | null = null;

            if (isAdminValid) {
                setIsLoggedIn(true);
                setIsAdmin(true);
                setUserType('subscriber');

                if (savedPage && savedPage !== 'member-login' && savedPage !== 'admin-login') {
                    restoredPage = savedPage;
                } else if (window.location.pathname === '/internal/audit') {
                    restoredPage = 'audit-log';
                } else {
                    restoredPage = 'members';
                }
            }
            else if (isMemberValid) {
                setIsLoggedIn(true);
                setIsAdmin(false);

                // 🔥 Determine userType from stored data
                const isPaid = getIsPaidFromStorage();
                setUserType(isPaid ? 'subscriber' : 'member');

                if (savedPage && savedPage !== 'member-login' && savedPage !== 'admin-login') {
                    // 🔥 FIX: If saved page is 'members', redirect to home instead
                    if (savedPage === 'members') {
                        restoredPage = 'home';
                    }
                    // If non-paid member, only restore public pages
                    else if (!isPaid && !publicPages.includes(savedPage)) {
                        restoredPage = 'membership'; // Go to profile to pay
                    } else {
                        restoredPage = savedPage;
                    }
                } else {
                    restoredPage = 'home';
                }
            }
            else {
                localStorage.removeItem('admin_token');
                localStorage.removeItem('admin_token_expiry');
                localStorage.removeItem('admin_role');
                localStorage.removeItem('member_token');
                localStorage.removeItem('member_token_expiry');
                localStorage.removeItem('member_role');
                localStorage.removeItem('user');

                if (savedPage && publicPages.includes(savedPage)) {
                    restoredPage = savedPage;
                }
            }

            if (restoredPage) {
                setCurrentPage(restoredPage);
            }
        } catch {
            // localStorage may throw in private browsing on some mobile browsers — treat as logged out
        }

        setIsInitialized(true);
    }, []);

    // ==================== SAVE CURRENT PAGE TO LOCAL STORAGE ====================
    // Using localStorage instead of sessionStorage — sessionStorage is cleared when mobile browsers kill tabs
    useEffect(() => {
        if (!isInitialized) return;
        if (currentPage === 'member-login' || currentPage === 'admin-login' || currentPage === 'home') return;
        try {
            localStorage.setItem('lastPage', currentPage);
        } catch {
            // localStorage may throw in private browsing on some mobile browsers
        }
    }, [currentPage]);

    // ==================== FETCH LOGO ====================
    useEffect(() => {
        const fetchLogo = async () => {
            try {
                const data = await getLogo();
                setLogoUrl(API_BASE_URL + data.url);
            } catch (error) {
                console.error("Failed to load logo", error);
            }
        };
        fetchLogo();
    }, []);

    // ==================== LOGIN HANDLER ====================
    const handleLogin = (isAdminLogin: boolean = false, isPaid: boolean = false) => {
        setIsLoggedIn(true);
        setIsAdmin(isAdminLogin);
        setUserType(isPaid ? 'subscriber' : 'member');

        const lastPage = localStorage.getItem('lastPage') as Page;

        if (lastPage && lastPage !== 'member-login' && lastPage !== 'admin-login' && (isAdminLogin || (lastPage !== 'members' && lastPage !== 'audit-log'))) {
            setCurrentPage(lastPage);
        } else {
            setCurrentPage('home');
        }

        toast.success(isAdminLogin ? 'Admin logged in successfully' : 'Member logged in successfully');
    };

    // ==================== LOGOUT HANDLER ====================
    const handleLogout = () => {
        localStorage.removeItem('lastPage');
        sessionStorage.removeItem('lastPage'); // clean up legacy key if present

        localStorage.removeItem('admin_token');
        localStorage.removeItem('admin_token_expiry');
        localStorage.removeItem('admin_role');
        localStorage.removeItem('member_token');
        localStorage.removeItem('member_token_expiry');
        localStorage.removeItem('member_role');
        localStorage.removeItem('user');

        setIsLoggedIn(false);
        setIsAdmin(false);
        setUserType('general');
        setCurrentPage('home');

        toast.success('Logged out successfully');
    };

    // ==================== ACCESS CONTROL ====================
    const hasAccess = (page: string): boolean => {
        // 🔥 Admin bypass - can see everything except profile
        if (isAdmin) {
            if (page === 'membership') return false;
            return true;
        }

        // 🔥 Check payment status directly from localStorage
        const isPaid = getIsPaidFromStorage();

        const accessControl: Record<string, boolean> = {
            'home': true,
            'office-bearers': true,
            'resources': isPaid,
            'taga-towers': isPaid,
            'membership': isLoggedIn, // All logged-in members can see profile
            'events': true,
            'grievance': isPaid,
            'members': false, // 🔥 FIX: Members cannot access the 'members' page at all
            'member-login': !isLoggedIn,
            'audit-log': isAdmin, // Only admins can access audit log
        };

        return accessControl[page] ?? true;
    };

    // ==================== NAVIGATION ====================
    const navigation = [
        { id: 'home', label: 'Home', icon: Home },
        { id: 'office-bearers', label: 'Office Bearers', icon: Users },
        ...(hasAccess('resources') ? [{ id: 'resources', label: 'Resources', icon: FileText }] : []),
        ...(hasAccess('taga-towers') ? [{ id: 'taga-towers', label: 'TAGA Towers', icon: Building2 }] : []),
        ...(hasAccess('membership') && !isAdmin ? [{ id: 'membership', label: 'Profile', icon: CreditCard }] : []),
        { id: 'events', label: 'Events', icon: Calendar },
        ...(hasAccess('grievance') ? [{ id: 'grievance', label: 'Grievance', icon: MessageSquare }] : []),
        ...(!isLoggedIn ? [{ id: 'member-login', label: 'Member Login', icon: LogIn }] : []),
        ...(isLoggedIn && isAdmin ? [{ id: 'members', label: 'Admin Panel', icon: UserCheck }] : []) // 🔥 Only shows for admins
    ].filter(item => hasAccess(item.id));

    // ==================== PAGE RENDERER ====================
    if (!isInitialized) {
        return (
            <div className="min-h-screen bg-background flex items-center justify-center">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
                    <p className="text-muted-foreground">Loading...</p>
                </div>
            </div>
        );
    }

    const renderPage = () => {
        // 🔥 Read isPaid directly from localStorage for immediate access
        const isPaid = getIsPaidFromStorage();

        // 🔥 FIX: Redirect members away from 'members' page
        if (currentPage === 'members' && !isAdmin) {
            setTimeout(() => setCurrentPage('home'), 0);
            return <AboutUs />;
        }

        if (!hasAccess(currentPage)) {
            if (currentPage !== 'home') {
                setTimeout(() => setCurrentPage('home'), 0);
            }
            return <AboutUs />;
        }

        switch (currentPage) {
            case 'home':
                return <AboutUs />;
            case 'office-bearers':
                return <OfficeBearers />;
            case 'member-login':
                return <MemberLogin onLogin={(paid) => handleLogin(false, paid !== undefined ? paid : getIsPaidFromStorage())} />;
            case 'admin-login':
                return <AdminLogin onLogin={() => handleLogin(true)} />;
            case 'members':
                // 🔥 Only admins can see this page
                return isAdmin ? <MembersDashboard isAdmin={isAdmin} /> : <AboutUs />;
            case 'audit-log':
                return <AuditLog isAdmin={isAdmin} />;
            case 'resources':
                return <Resources isLoggedIn={isLoggedIn} />;
            case 'grievance':
                return <Grievance />;
            case 'taga-towers':
                return <TAGATowers isLoggedIn={isLoggedIn} isPaidMember={isPaid || userType === 'subscriber'} isAdmin={isAdmin} />;
            case 'membership':
                return <Membership isLoggedIn={isLoggedIn} isPaidMember={isPaid || userType === 'subscriber'} />;
            case 'events':
                return <Events isLoggedIn={isLoggedIn} />;
            default:
                return <AboutUs />;
        }
    };

    return (
        <div className="min-h-screen bg-background">
            <header className="border-b bg-card shadow-md sticky top-0 z-50">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="flex justify-between items-center h-16">
                        <div className="flex items-center space-x-3 cursor-pointer"
                            onClick={() => setCurrentPage('home')}
                            data-testid="home-link">
                            <div className="w-12 h-12 rounded-xl overflow-hidden shadow-lg bg-white flex items-center justify-center">
                                {logoUrl ? (
                                    <img src={logoUrl} alt="TAGA Logo" className="w-full h-full object-contain" />
                                ) : (
                                    <span className="text-primary font-bold text-lg">T</span>
                                )}
                            </div>
                            <div>
                                <h1 className="text-lg font-bold text-foreground">TAGA</h1>
                                {currentPage === 'home' && (
                                    <p className="text-sm text-muted-foreground font-medium">Empowering Agricultural Professionals</p>
                                )}
                            </div>
                        </div>

                        <nav className="hidden lg:flex items-center space-x-1">
                            {navigation.map((item) => {
                                const Icon = item.icon;
                                return (
                                    <Button
                                        key={item.id}
                                        variant={currentPage === item.id ? "default" : "ghost"}
                                        onClick={() => setCurrentPage(item.id as Page)}
                                        className="flex items-center space-x-2"
                                        size="sm"
                                        data-testid={`testid-${item.id}-button`}
                                    >
                                        <Icon className="w-4 h-4" />
                                        <span>{item.label}</span>
                                    </Button>
                                );
                            })}
                            {isLoggedIn && (
                                <Button variant="outline" onClick={handleLogout} className="ml-4" size="sm" data-testid="testid-logout-button">
                                    Logout
                                </Button>
                            )}
                        </nav>

                        <Button
                            variant="ghost"
                            size="sm"
                            className="lg:hidden"
                            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
                            data-testid="testid-mobile-menu-button"
                        >
                            {mobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
                        </Button>
                    </div>

                    {mobileMenuOpen && (
                        <div className="lg:hidden border-t py-2">
                            <nav className="flex flex-col space-y-1">
                                {navigation.map((item) => {
                                    const Icon = item.icon;
                                    return (
                                        <Button
                                            key={item.id}
                                            variant={currentPage === item.id ? "default" : "ghost"}
                                            onClick={() => {
                                                setCurrentPage(item.id as Page);
                                                setMobileMenuOpen(false);
                                            }}
                                            className="flex items-center space-x-2 justify-start"
                                            size="sm"
                                            data-testid={`testid-mobile-${item.id}-button`}
                                        >
                                            <Icon className="w-4 h-4" />
                                            <span>{item.label}</span>
                                        </Button>
                                    );
                                })}
                                {isLoggedIn && (
                                    <Button
                                        variant="outline"
                                        onClick={() => {
                                            handleLogout();
                                            setMobileMenuOpen(false);
                                        }}
                                        className="justify-start"
                                        size="sm"
                                        data-testid="testid-mobile-logout-button"
                                    >
                                        Logout
                                    </Button>
                                )}
                            </nav>
                        </div>
                    )}
                </div>
            </header>

            <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                {renderPage()}
            </main>

            <footer className="border-t bg-card mt-16">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                    <div className="text-center text-muted-foreground">
                        <p>
                            <span
                                onClick={handleCopyrightClick}
                                className={`select-none ${isLoggedIn && isAdmin ? 'cursor-pointer hover:text-primary transition-colors' : 'cursor-default'}`}
                                data-testid="copyright-symbol"
                            >
                                &copy;
                            </span>
                            {" "}2026 Tamil Nadu Agriculture Graduates Association (TAGA). All rights reserved.
                        </p>
                        <p className="mt-2">Empowering agricultural graduates of the Tamil Nadu agriculture department with knowledge, unity, and progress.</p>
                        {!isLoggedIn && (
                            <div className="mt-4">
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={() => setCurrentPage('admin-login')}
                                    className="text-xs text-muted-foreground hover:text-foreground"
                                    data-testid="testid-admin-login-button"
                                >
                                    Administrative Access
                                </Button>
                            </div>
                        )}
                    </div>
                </div>
            </footer>
            <Toaster />
        </div>
    );
}