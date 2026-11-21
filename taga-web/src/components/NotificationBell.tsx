import React, { useState, useEffect, useCallback } from 'react';
import { Button } from './ui/button';
import { Bell } from 'lucide-react';
import API_BASE_URL from '../config/api';

export interface NotificationBellProps {
  memberId?: string;
  isLoggedIn?: boolean;
  onClick?: () => void;
  className?: string;
}

export function NotificationBell({
  memberId,
  isLoggedIn,
  onClick,
  className,
}: NotificationBellProps) {
  const [hasNotifications, setHasNotifications] = useState(false);

  const fetchNotifications = useCallback(async () => {
    try {
      const token = localStorage.getItem('member_token') || localStorage.getItem('token');
      if (!token || !isLoggedIn) {
        setHasNotifications(false);
        return;
      }

      const user = JSON.parse(localStorage.getItem('user') || '{}');
      const targetMemberId = memberId || user.id || user.emailId;

      let url = `${API_BASE_URL}/member/notifications`;
      if (targetMemberId) {
        url += `?member_id=${encodeURIComponent(targetMemberId)}`;
      }

      const headers: Record<string, string> = {
        'Authorization': `Bearer ${token}`
      };

      const response = await fetch(url, { headers });
      if (response.ok) {
        const data = await response.json();
        if (Array.isArray(data)) {
          setHasNotifications(data.length > 0);
        } else {
          setHasNotifications(false);
        }
      } else {
        setHasNotifications(false);
      }
    } catch (error) {
      console.error('Failed to fetch notifications:', error);
      setHasNotifications(false);
    }
  }, [memberId, isLoggedIn]);

  useEffect(() => {
    if (!isLoggedIn) {
      setHasNotifications(false);
      return;
    }
    const token = localStorage.getItem('member_token') || localStorage.getItem('token');
    if (!token) {
      setHasNotifications(false);
      return;
    }

    fetchNotifications();
    const interval = setInterval(fetchNotifications, 30000); // Poll every 30 seconds
    return () => clearInterval(interval);
  }, [fetchNotifications, isLoggedIn]);

  if (!isLoggedIn) {
    return null;
  }

  const handleClick = () => {
    if (onClick) {
      onClick();
      return;
    }
    const token = localStorage.getItem('member_token') || localStorage.getItem('token');
    if (token) {
      window.location.hash = '#membership';
    } else {
      window.location.hash = '#member-login';
    }
  };

  return (
    <Button
      variant="ghost"
      size="icon"
      className={`relative h-9 w-9 rounded-full hover:bg-accent/80 text-foreground transition-colors cursor-pointer ${className || ''}`}
      onClick={handleClick}
      aria-label="Official Announcements"
      title={hasNotifications ? "Active Announcements Available" : "Official Announcements"}
      data-testid="testid-notification-bell-button"
    >
      <Bell className="w-5 h-5 text-foreground" />
      {hasNotifications && (
        <span
          className="absolute top-1.5 right-1.5 w-2.5 h-2.5 rounded-full bg-emerald-500 ring-2 ring-background animate-pulse"
          data-testid="testid-notification-bell-indicator"
        />
      )}
    </Button>
  );
}
