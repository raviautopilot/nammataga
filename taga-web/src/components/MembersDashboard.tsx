import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import { Separator } from './ui/separator';
import { Alert, AlertDescription } from './ui/alert';
import { AdminActions } from './AdminActions';
import { ActivityDetails } from './ActivityDetails';
import { NotificationBell } from './NotificationBell';
import { getMemberStats } from '../api/adminContent';
import {
    CreditCard,
    Users,
    Calendar,
    TrendingUp,
    Bell,
    CheckCircle,
    AlertCircle,
    Shield,
    UserPlus,
} from 'lucide-react';
import { toast } from 'sonner';

interface MembersDashboardProps {
    isAdmin: boolean;
}

// Member data type
interface Member {
    id: string;
    name: string;
    initial: string;
    gender: string;
    fatherName: string;
    designation: string;
    district?: string;
    workingDistrict?: string;
    recruitmentBatch: string;
    seniorityNumber: number;
    dateOfBirth: string;
    mobileNumber: string;
    email: string;
    tbfNumber: string;
    cpsGpfNumber: string;
    paymentStatus: 'Paid' | 'Unpaid';
    paymentDate: string;
    membershipStatus: 'Active' | 'Unpaid';
}

export function MembersDashboard({ isAdmin }: MembersDashboardProps) {
    const [membershipStatus] = useState({
        id: 'AG2023001',
        status: 'Active',
        joinDate: '2018-03-15',
        expiryDate: '2025-03-14',
        lastPayment: '2024-03-15',
        nextDue: '2025-03-15'
    });

    const [selectedActivity, setSelectedActivity] = useState<'registrations' | 'payments' | 'grievances' | null>(null);

    // Mock member data with state management
    const [members] = useState<Member[]>([
        {
            id: 'AG2023001',
            name: 'Rajkumar S',
            initial: 'S',
            gender: 'Male',
            fatherName: 'Selvam R',
            designation: 'Agriculture Officer',
            workingDistrict: 'Chennai',
            recruitmentBatch: '2018',
            seniorityNumber: 1234,
            dateOfBirth: '1990-05-15',
            mobileNumber: '9876543210',
            email: 'rajkumar@taga.org',
            tbfNumber: 'TBF123456',
            cpsGpfNumber: 'CPS789012',
            paymentStatus: 'Paid',
            paymentDate: '2024-03-15',
            membershipStatus: 'Active'
        },
        {
            id: 'AG2023002',
            name: 'Lakshmi P',
            initial: 'P',
            gender: 'Female',
            fatherName: 'Prakash M',
            designation: 'Assistant Agriculture Officer',
            workingDistrict: 'Coimbatore',
            recruitmentBatch: '2019',
            seniorityNumber: 1567,
            dateOfBirth: '1992-08-20',
            mobileNumber: '9876543211',
            email: 'lakshmi@taga.org',
            tbfNumber: 'TBF123457',
            cpsGpfNumber: 'GPF456123',
            paymentStatus: 'Unpaid',
            paymentDate: '2023-03-10',
            membershipStatus: 'Unpaid'
        },
        {
            id: 'AG2023003',
            name: 'Kumar V',
            initial: 'V',
            gender: 'Male',
            fatherName: 'Vijayan K',
            designation: 'Senior Agriculture Officer',
            workingDistrict: 'Madurai',
            recruitmentBatch: '2015',
            seniorityNumber: 987,
            dateOfBirth: '1988-12-05',
            mobileNumber: '9876543212',
            email: 'kumar@taga.org',
            tbfNumber: 'TBF123458',
            cpsGpfNumber: 'CPS321654',
            paymentStatus: 'Paid',
            paymentDate: '2024-04-01',
            membershipStatus: 'Active'
        },
        {
            id: 'AG2023004',
            name: 'Priya M',
            initial: 'M',
            gender: 'Female',
            fatherName: 'Murthy S',
            designation: 'Agriculture Inspector',
            workingDistrict: 'Salem',
            recruitmentBatch: '2020',
            seniorityNumber: 2105,
            dateOfBirth: '1994-03-12',
            mobileNumber: '9876543213',
            email: 'priya@taga.org',
            tbfNumber: 'TBF123459',
            cpsGpfNumber: 'GPF987654',
            paymentStatus: 'Paid',
            paymentDate: '2024-02-28',
            membershipStatus: 'Active'
        },
        {
            id: 'AG2023005',
            name: 'Arjun R',
            initial: 'R',
            gender: 'Male',
            fatherName: 'Raman K',
            designation: 'Agriculture Extension Officer',
            workingDistrict: 'Trichy',
            recruitmentBatch: '2021',
            seniorityNumber: 2389,
            dateOfBirth: '1995-07-18',
            mobileNumber: '9876543214',
            email: 'arjun@taga.org',
            tbfNumber: 'TBF123460',
            cpsGpfNumber: 'CPS654321',
            paymentStatus: 'Unpaid',
            paymentDate: '2023-01-20',
            membershipStatus: 'Unpaid'
        }
    ]);

    // Real-time member stats
    const [memberStats, setMemberStats] = useState({
        totalMembers: 0,
        activeMembers: 0,
        unpaid: 0,
        newThisMonth: 0
    });
    const [statsLoading, setStatsLoading] = useState(true);

    useEffect(() => {
        const fetchStats = async () => {
            try {
                const stats = await getMemberStats();
                setMemberStats(stats);
            } catch (error) {
                console.error('Failed to fetch member stats:', error);
                toast.error('Could not load member statistics');
            } finally {
                setStatsLoading(false);
            }
        };
        fetchStats();
    }, []);

    const recentActivities = [
        {
            type: 'payment',
            title: 'Membership fee paid',
            date: '2024-03-15',
            status: 'completed'
        },
        {
            type: 'document',
            title: 'Downloaded salary certificate',
            date: '2024-03-10',
            status: 'completed'
        },
        {
            type: 'grievance',
            title: 'Submitted transfer request',
            date: '2024-03-08',
            status: 'pending'
        }
    ];

    const subscriptionBenefits = [
        'Access to all government orders and circulars',
        'Priority support for grievance resolution',
        'Participation in professional development programs',
        'Health insurance scheme eligibility',
        'Legal assistance for service matters',
        'Networking events and conferences',
        'Career guidance and mentorship',
        'Retirement planning workshops'
    ];

    if (isAdmin) {
        return (
            <div className="space-y-8">
                {/* Admin Header */}
                <div className="flex items-center justify-between">
                    <div>
                        <h1 className="text-3xl font-bold text-foreground flex items-center space-x-2">
                            <Shield className="w-8 h-8 text-primary" />
                            <span>Admin Dashboard</span>
                        </h1>
                        <p className="text-muted-foreground mt-1">
                            Manage association members and operations
                        </p>
                    </div>
                    <Badge variant="secondary" className="bg-blue-100 text-blue-800">
                        Administrator
                    </Badge>
                </div>

                {/* Admin Stats - real data */}
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <Card>
                        <CardContent className="pt-6">
                            <div className="flex items-center space-x-2">
                                <Users className="w-8 h-8 text-blue-600" />
                                <div>
                                    <p className="text-2xl font-bold">{statsLoading ? '...' : memberStats.totalMembers}</p>
                                    <p className="text-sm text-muted-foreground">Total Members</p>
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardContent className="pt-6">
                            <div className="flex items-center space-x-2">
                                <CheckCircle className="w-8 h-8 text-green-600" />
                                <div>
                                    <p className="text-2xl font-bold">{statsLoading ? '...' : memberStats.activeMembers}</p>
                                    <p className="text-sm text-muted-foreground">Active Members</p>
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardContent className="pt-6">
                            <div className="flex items-center space-x-2">
                                <AlertCircle className="w-8 h-8 text-orange-600" />
                                <div>
                                    <p className="text-2xl font-bold">{statsLoading ? '...' : memberStats.unpaid}</p>
                                    <p className="text-sm text-muted-foreground">Unpaid Members</p>
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardContent className="pt-6">
                            <div className="flex items-center space-x-2">
                                <UserPlus className="w-8 h-8 text-purple-600" />
                                <div>
                                    <p className="text-2xl font-bold">{statsLoading ? '...' : memberStats.newThisMonth}</p>
                                    <p className="text-sm text-muted-foreground">New This Month</p>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </div>

                {/* Manage Members Content (previously the default tab) */}
                <div className="space-y-6">
                    <AdminActions memberStats={memberStats} />
                </div>

                <ActivityDetails
                    type={selectedActivity}
                    onClose={() => setSelectedActivity(null)}
                />
            </div>
        );
    }

    return (
        <div className="space-y-8">
            {/* Member Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold text-foreground">Welcome Back!</h1>
                    <p className="text-muted-foreground mt-1">
                        Membership ID: {membershipStatus.id}
                    </p>
                </div>
                <div className="flex items-center gap-4">
                    <NotificationBell memberId={membershipStatus.id} />
                    <Badge variant="default" className="bg-green-100 text-green-800">
                        {membershipStatus.status}
                    </Badge>
                </div>
            </div>

            {/* Membership Status */}
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center space-x-2">
                        <CreditCard className="w-5 h-5" />
                        <span>Membership Status</span>
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="grid md:grid-cols-4 gap-4">
                        <div>
                            <p className="text-sm text-muted-foreground mb-1">Join Date</p>
                            <p className="font-medium">{new Date(membershipStatus.joinDate).toLocaleDateString()}</p>
                        </div>
                        <div>
                            <p className="text-sm text-muted-foreground mb-1">Last Payment</p>
                            <p className="font-medium">{new Date(membershipStatus.lastPayment).toLocaleDateString()}</p>
                        </div>
                        <div>
                            <p className="text-sm text-muted-foreground mb-1">Next Due</p>
                            <p className="font-medium">{new Date(membershipStatus.nextDue).toLocaleDateString()}</p>
                        </div>
                        <div>
                            <p className="text-sm text-muted-foreground mb-1">Membership Expires</p>
                            <p className="font-medium">{new Date(membershipStatus.expiryDate).toLocaleDateString()}</p>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Subscription Benefits */}
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center space-x-2">
                        <TrendingUp className="w-5 h-5" />
                        <span>Membership Benefits</span>
                    </CardTitle>
                    <CardDescription>
                        Enjoy these exclusive benefits as an active member
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="grid md:grid-cols-2 gap-4">
                        {subscriptionBenefits.map((benefit, index) => (
                            <div key={index} className="flex items-start space-x-3">
                                <CheckCircle className="w-5 h-5 text-green-600 mt-0.5 flex-shrink-0" />
                                <span className="text-sm text-muted-foreground">{benefit}</span>
                            </div>
                        ))}
                    </div>
                </CardContent>
            </Card>

            {/* How to Subscribe/Renew */}
            <Card>
                <CardHeader>
                    <CardTitle>Subscription Management</CardTitle>
                    <CardDescription>
                        Keep your membership active to continue enjoying all benefits
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <Alert>
                        <Bell className="w-4 h-4" />
                        <AlertDescription>
                            Your membership will expire on {new Date(membershipStatus.expiryDate).toLocaleDateString()}.
                            Renew before this date to avoid service interruption.
                        </AlertDescription>
                    </Alert>

                    <div className="space-y-4">
                        <div>
                            <h4 className="font-semibold text-foreground mb-2">Annual Membership Fee</h4>
                            <p className="text-2xl font-bold text-primary">₹500</p>
                            <p className="text-sm text-muted-foreground">Valid for 12 months from payment date</p>
                        </div>

                        <Separator />

                        <div>
                            <h4 className="font-semibold text-foreground mb-2">Payment Methods</h4>
                            <div className="space-y-2">
                                <div className="flex items-center space-x-2">
                                    <div className="w-2 h-2 bg-primary rounded-full" />
                                    <span className="text-sm text-muted-foreground">
                                        Online payment through government portal (UPI, Net Banking, Cards)
                                    </span>
                                </div>
                                <div className="flex items-center space-x-2">
                                    <div className="w-2 h-2 bg-primary rounded-full" />
                                    <span className="text-sm text-muted-foreground">
                                        Bank transfer to Association account
                                    </span>
                                </div>
                                <div className="flex items-center space-x-2">
                                    <div className="w-2 h-2 bg-primary rounded-full" />
                                    <span className="text-sm text-muted-foreground">
                                        Cash payment at district offices
                                    </span>
                                </div>
                            </div>
                        </div>

                        <div className="flex space-x-4">
                            <Button className="flex-1" data-testid="testid-renew-membership-button">
                                Renew Membership
                            </Button>
                            <Button variant="outline" className="flex-1" data-testid="testid-payment-history-button">
                                Payment History
                            </Button>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Recent Activities */}
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center space-x-2">
                        <Calendar className="w-5 h-5" />
                        <span>Recent Activities</span>
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="space-y-4">
                        {recentActivities.map((activity, index) => (
                            <div key={index} className="flex items-center justify-between p-3 rounded-lg border">
                                <div className="flex items-center space-x-3">
                                    <div className={`w-2 h-2 rounded-full ${activity.status === 'completed' ? 'bg-green-600' : 'bg-orange-600'
                                        }`} />
                                    <div>
                                        <p className="font-medium text-foreground">{activity.title}</p>
                                        <p className="text-sm text-muted-foreground">{activity.date}</p>
                                    </div>
                                </div>
                                <Badge variant={activity.status === 'completed' ? 'default' : 'secondary'}>
                                    {activity.status}
                                </Badge>
                            </div>
                        ))}
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}