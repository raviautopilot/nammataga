import React, { useState } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from './ui/dialog';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Badge } from './ui/badge';
import { Button } from './ui/button';
import { Separator } from './ui/separator';
import { ScrollArea } from './ui/scroll-area';
import { 
  UserPlus, 
  CreditCard, 
  MessageSquare, 
  Calendar, 
  Phone, 
  Mail, 
  MapPin,
  Clock,
  CheckCircle,
  AlertCircle,
  Eye,
  Download,
  Reply
} from 'lucide-react';

interface ActivityDetailsProps {
  type: 'registrations' | 'payments' | 'grievances' | null;
  onClose: () => void;
}

export function ActivityDetails({ type, onClose }: ActivityDetailsProps) {
  // Mock data for new registrations
  const newRegistrations = [
    {
      id: 'AG2024156',
      name: 'Dr. Priya Krishnan',
      email: 'priya.krishnan@agri.tn.gov.in',
      phone: '+91 98765 43210',
      district: 'Coimbatore',
      department: 'Agriculture Department',
      designation: 'Agricultural Officer',
      registrationTime: '09:30 AM',
      status: 'pending_verification'
    },
    {
      id: 'AG2024157',
      name: 'Mr. Rajesh Kumar',
      email: 'rajesh.kumar@hort.tn.gov.in',
      phone: '+91 98765 43211',
      district: 'Madurai',
      department: 'Horticulture Department',
      designation: 'Horticulture Officer',
      registrationTime: '11:15 AM',
      status: 'verified'
    },
    {
      id: 'AG2024158',
      name: 'Ms. Lakshmi Devi',
      email: 'lakshmi.devi@agri.tn.gov.in',
      phone: '+91 98765 43212',
      district: 'Salem',
      department: 'Agriculture Department',
      designation: 'Extension Officer',
      registrationTime: '02:45 PM',
      status: 'verified'
    }
  ];

  // Mock data for payments
  const recentPayments = [
    {
      id: 'PAY2024089',
      memberId: 'AG2023045',
      memberName: 'Dr. Suresh Babu',
      amount: 500,
      paymentMethod: 'UPI',
      transactionId: 'TXN789456123',
      timestamp: '10:30 AM',
      status: 'completed',
      type: 'membership_renewal'
    },
    {
      id: 'PAY2024090',
      memberId: 'AG2023078',
      memberName: 'Ms. Anitha Rajesh',
      amount: 500,
      paymentMethod: 'Net Banking',
      transactionId: 'TXN789456124',
      timestamp: '11:45 AM',
      status: 'completed',
      type: 'membership_renewal'
    },
    {
      id: 'PAY2024091',
      memberId: 'AG2024156',
      memberName: 'Dr. Priya Krishnan',
      amount: 500,
      paymentMethod: 'Credit Card',
      transactionId: 'TXN789456125',
      timestamp: '09:45 AM',
      status: 'completed',
      type: 'new_membership'
    },
    {
      id: 'PAY2024092',
      memberId: 'AG2023112',
      memberName: 'Mr. Venkat Raman',
      amount: 500,
      paymentMethod: 'Debit Card',
      transactionId: 'TXN789456126',
      timestamp: '01:20 PM',
      status: 'pending',
      type: 'membership_renewal'
    },
    {
      id: 'PAY2024093',
      memberId: 'AG2023089',
      memberName: 'Dr. Meera Krishnan',
      amount: 500,
      paymentMethod: 'UPI',
      transactionId: 'TXN789456127',
      timestamp: '02:15 PM',
      status: 'completed',
      type: 'membership_renewal'
    }
  ];

  // Mock data for grievances
  const recentGrievances = [
    {
      id: 'GRV2024045',
      memberId: 'AG2023067',
      memberName: 'Mr. Kumaran Pillai',
      subject: 'Salary increment delay for Agricultural Officers',
      priority: 'high',
      category: 'Service Matter',
      submittedTime: '09:15 AM',
      status: 'under_review',
      description: 'The salary increment for Agricultural Officers as per the new pay scale has been delayed for the past 3 months. Request urgent intervention.'
    },
    {
      id: 'GRV2024046',
      memberId: 'AG2023091',
      memberName: 'Dr. Saranya Devi',
      subject: 'Transfer request not processed',
      priority: 'medium',
      category: 'Transfer',
      submittedTime: '02:30 PM',
      status: 'new',
      description: 'Applied for transfer from Thanjavur to Chennai due to family circumstances. Application submitted 2 months ago but no response received.'
    }
  ];

  const getTitle = () => {
    switch (type) {
      case 'registrations':
        return 'New Registrations Today';
      case 'payments':
        return 'Payments Received Today';
      case 'grievances':
        return 'Grievances Submitted Today';
      default:
        return '';
    }
  };

  const getIcon = () => {
    switch (type) {
      case 'registrations':
        return <UserPlus className="w-5 h-5 text-blue-600" />;
      case 'payments':
        return <CreditCard className="w-5 h-5 text-green-600" />;
      case 'grievances':
        return <MessageSquare className="w-5 h-5 text-orange-600" />;
      default:
        return null;
    }
  };

  const renderRegistrations = () => (
    <ScrollArea className="h-96" data-testid="testid-registration-activities-scroll-area">
      <div className="space-y-4" data-testid="testid-registration-activities-list">
        {newRegistrations.map((registration) => (
          <Card key={registration.id} className="p-4" data-testid={`testid-registration-${registration.id}-card`}>
            <div className="flex justify-between items-start mb-3">
              <div>
                <h4 className="font-medium">{registration.name}</h4>
                <p className="text-sm text-muted-foreground">ID: {registration.id}</p>
              </div>
              <div className="flex items-center space-x-2">
                <Badge variant={registration.status === 'verified' ? 'default' : 'secondary'}>
                  {registration.status === 'verified' ? 'Verified' : 'Pending Verification'}
                </Badge>
                <Badge variant="outline">
                  <Clock className="w-3 h-3 mr-1" />
                  {registration.registrationTime}
                </Badge>
              </div>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
              <div className="flex items-center space-x-2">
                <Mail className="w-4 h-4 text-muted-foreground" />
                <span>{registration.email}</span>
              </div>
              <div className="flex items-center space-x-2">
                <Phone className="w-4 h-4 text-muted-foreground" />
                <span>{registration.phone}</span>
              </div>
              <div className="flex items-center space-x-2">
                <MapPin className="w-4 h-4 text-muted-foreground" />
                <span>{registration.district}</span>
              </div>
              <div className="flex items-center space-x-2">
                <span className="text-muted-foreground">{registration.designation}</span>
              </div>
            </div>
            
            <Separator className="my-3" />
            
            <div className="flex justify-between items-center">
              <span className="text-xs text-muted-foreground">{registration.department}</span>
              <div className="space-x-2">
                <Button size="sm" variant="outline" data-testid={`testid-registration-${registration.id}-view-details-button`}>
                  <Eye className="w-3 h-3 mr-1" />
                  View Details
                </Button>
                {registration.status === 'pending_verification' && (
                  <Button size="sm" data-testid={`testid-registration-${registration.id}-verify-button`}>
                    <CheckCircle className="w-3 h-3 mr-1" />
                    Verify
                  </Button>
                )}
              </div>
            </div>
          </Card>
        ))}
      </div>
    </ScrollArea>
  );

  const renderPayments = () => (
    <ScrollArea className="h-96" data-testid="testid-payment-activities-scroll-area">
      <div className="space-y-4" data-testid="testid-payment-activities-list">
        {recentPayments.map((payment) => (
          <Card key={payment.id} className="p-4" data-testid={`testid-payment-${payment.id}-card`}>
            <div className="flex justify-between items-start mb-3">
              <div>
                <h4 className="font-medium">{payment.memberName}</h4>
                <p className="text-sm text-muted-foreground">Member ID: {payment.memberId}</p>
              </div>
              <div className="flex items-center space-x-2">
                <Badge variant={payment.status === 'completed' ? 'default' : 'secondary'}>
                  {payment.status === 'completed' ? 'Completed' : 'Pending'}
                </Badge>
                <Badge variant="outline">
                  <Clock className="w-3 h-3 mr-1" />
                  {payment.timestamp}
                </Badge>
              </div>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
              <div>
                <span className="text-muted-foreground">Amount: </span>
                <span className="font-medium text-green-600">₹{payment.amount}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Method: </span>
                <span>{payment.paymentMethod}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Transaction ID: </span>
                <span className="font-mono text-xs">{payment.transactionId}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Type: </span>
                <span className="capitalize">{payment.type.replace('_', ' ')}</span>
              </div>
            </div>
            
            <Separator className="my-3" />
            
            <div className="flex justify-between items-center">
              <span className="text-xs text-muted-foreground">Payment ID: {payment.id}</span>
              <div className="space-x-2">
                <Button size="sm" variant="outline" data-testid={`testid-payment-${payment.id}-receipt-button`}>
                  <Download className="w-3 h-3 mr-1" />
                  Receipt
                </Button>
                <Button size="sm" variant="outline" data-testid={`testid-payment-${payment.id}-view-details-button`}>
                  <Eye className="w-3 h-3 mr-1" />
                  View Details
                </Button>
              </div>
            </div>
          </Card>
        ))}
      </div>
    </ScrollArea>
  );

  const renderGrievances = () => (
    <ScrollArea className="h-96" data-testid="testid-grievance-activities-scroll-area">
      <div className="space-y-4" data-testid="testid-grievance-activities-list">
        {recentGrievances.map((grievance) => (
          <Card key={grievance.id} className="p-4" data-testid={`testid-grievance-${grievance.id}-card`}>
            <div className="flex justify-between items-start mb-3">
              <div>
                <h4 className="font-medium">{grievance.subject}</h4>
                <p className="text-sm text-muted-foreground">
                  By: {grievance.memberName} (ID: {grievance.memberId})
                </p>
              </div>
              <div className="flex items-center space-x-2">
                <Badge variant={
                  grievance.priority === 'high' ? 'destructive' : 
                  grievance.priority === 'medium' ? 'default' : 'secondary'
                }>
                  {grievance.priority} Priority
                </Badge>
                <Badge variant="outline">
                  <Clock className="w-3 h-3 mr-1" />
                  {grievance.submittedTime}
                </Badge>
              </div>
            </div>
            
            <div className="mb-3">
              <Badge variant="outline" className="mb-2">
                {grievance.category}
              </Badge>
              <Badge variant={
                grievance.status === 'under_review' ? 'default' : 'secondary'
              } className="ml-2">
                {grievance.status.replace('_', ' ').toUpperCase()}
              </Badge>
            </div>
            
            <p className="text-sm text-muted-foreground mb-3 line-clamp-2">
              {grievance.description}
            </p>
            
            <Separator className="my-3" />
            
            <div className="flex justify-between items-center">
              <span className="text-xs text-muted-foreground">Grievance ID: {grievance.id}</span>
              <div className="space-x-2">
                <Button size="sm" variant="outline" data-testid={`testid-grievance-${grievance.id}-view-full-button`}>
                  <Eye className="w-3 h-3 mr-1" />
                  View Full
                </Button>
                <Button size="sm" data-testid={`testid-grievance-${grievance.id}-respond-button`}>
                  <Reply className="w-3 h-3 mr-1" />
                  Respond
                </Button>
              </div>
            </div>
          </Card>
        ))}
      </div>
    </ScrollArea>
  );

  const renderContent = () => {
    switch (type) {
      case 'registrations':
        return renderRegistrations();
      case 'payments':
        return renderPayments();
      case 'grievances':
        return renderGrievances();
      default:
        return null;
    }
  };

  if (!type) return null;

  return (
    <Dialog open={!!type} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[80vh]" data-testid="testid-activity-details-modal">
        <DialogHeader>
          <DialogTitle className="flex items-center space-x-2">
            {getIcon()}
            <span>{getTitle()}</span>
          </DialogTitle>
          <DialogDescription>
            View and manage recent {type === 'registrations' ? 'member registrations' : 
                                  type === 'payments' ? 'payment transactions' : 
                                  'grievance submissions'} from today.
          </DialogDescription>
        </DialogHeader>
        
        <div className="mt-4">
          {renderContent()}
        </div>
      </DialogContent>
    </Dialog>
  );
}