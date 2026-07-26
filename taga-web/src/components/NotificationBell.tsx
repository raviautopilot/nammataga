import React, { useState, useEffect } from 'react';
import { Button } from './ui/button';
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover';
import { Bell, BellRing, CheckCircle } from 'lucide-react';
import { Badge } from './ui/badge';
import { toast } from 'sonner';

interface Notification {
  id: string;
  title: string;
  message: string;
  priority: string;
  is_read: boolean;
  created_at: string;
}

interface NotificationBellProps {
  memberId: string;
}

export function NotificationBell({ memberId }: NotificationBellProps) {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [isOpen, setIsOpen] = useState(false);

  const fetchNotifications = async () => {
    try {
      const token = localStorage.getItem('token'); // or your auth token
      const response = await fetch(`/api/member/notifications?member_id=${memberId}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      const data = await response.json();
      setNotifications(data);
      setUnreadCount(data.filter((n: Notification) => !n.is_read).length);
    } catch (error) {
      console.error('Failed to fetch notifications:', error);
    }
  };

  const markAsRead = async (notificationId: string) => {
    try {
      const token = localStorage.getItem('token');
      await fetch(`/api/member/notifications/${notificationId}/read`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      fetchNotifications();
    } catch (error) {
      console.error('Failed to mark as read:', error);
    }
  };

  useEffect(() => {
    fetchNotifications();
    const interval = setInterval(fetchNotifications, 30000); // Poll every 30 seconds
    return () => clearInterval(interval);
  }, [memberId]);

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'urgent': return 'bg-red-500';
      case 'high': return 'bg-orange-500';
      default: return 'bg-blue-500';
    }
  };

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className="relative">
          {unreadCount > 0 ? (
            <>
              <BellRing className="w-5 h-5" />
              <Badge className="absolute -top-1 -right-1 h-5 w-5 p-0 flex items-center justify-center bg-red-500">
                {unreadCount}
              </Badge>
            </>
          ) : (
            <Bell className="w-5 h-5" />
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80 p-0" align="end">
        <div className="p-3 border-b">
          <h4 className="font-semibold">Notifications</h4>
          {unreadCount > 0 && (
            <p className="text-xs text-muted-foreground">{unreadCount} unread</p>
          )}
        </div>
        <div className="max-h-96 overflow-y-auto">
          {notifications.length === 0 ? (
            <div className="p-4 text-center text-muted-foreground">
              No notifications
            </div>
          ) : (
            notifications.map((notif) => (
              <div
                key={notif.id}
                className={`p-3 border-b hover:bg-muted/50 cursor-pointer ${!notif.is_read ? 'bg-muted/30' : ''}`}
                onClick={() => markAsRead(notif.id)}
              >
                <div className="flex items-start gap-2">
                  <div className={`w-2 h-2 rounded-full mt-2 ${getPriorityColor(notif.priority)}`} />
                  <div className="flex-1">
                    <div className="flex items-center justify-between">
                      <p className="text-sm font-medium">{notif.title}</p>
                      {!notif.is_read && (
                        <CheckCircle className="w-3 h-3 text-green-500" />
                      )}
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">{notif.message}</p>
                    <p className="text-xs text-muted-foreground mt-1">
                      {new Date(notif.created_at).toLocaleDateString()}
                    </p>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}