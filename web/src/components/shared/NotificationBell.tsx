import { useState, useEffect } from 'react';
import { Bell, AlertCircle } from 'lucide-react';
import { Link } from 'react-router-dom';
import api from '../../api/client';
import { toast } from 'sonner';
import { cn } from '@/lib/utils';

export const NotificationBell = () => {
  const [unreadCount, setUnreadCount] = useState(0);
  const [hasError, setHasError] = useState(false);

  const fetchUnreadCount = async () => {
    try {
      const res = await api.get('/v1/api/notifications?limit=0');
      const notifications = res.data.data || [];
      const count = notifications.filter((n: any) => !n.is_read).length;
      setUnreadCount(count);
      setHasError(false);
    } catch (err: any) {
      console.error('Failed to fetch unread count:', err);
      if (!hasError) {
         setHasError(true);
         toast.error(err.message || 'خطأ في الاتصال بخدمة الإشعارات');
      }
    }
  };

  useEffect(() => {
    fetchUnreadCount();
    // Rafraîchir toutes les 60 secondes
    const interval = setInterval(fetchUnreadCount, 60000);
    return () => clearInterval(interval);
  }, [hasError]); // Adding hasError dependency to ensure toast doesn't duplicate but logic works

  return (
    <Link 
      to="/notifications" 
      title={hasError ? "يوجد خطأ في الإشعارات" : "الإشعارات"}
      className={cn(
        "relative p-2 transition-colors rounded-full border shadow-sm",
        hasError 
          ? "bg-red-500/10 text-red-500 border-red-500/50 hover:bg-red-500/20" 
          : "text-gray-400 hover:text-primary-600 bg-[#1b4fd8] border-transparent"
      )}
    >
      {hasError ? <AlertCircle className="w-6 h-6" /> : <Bell className="w-6 h-6" />}
      
      {!hasError && unreadCount > 0 && (
        <span className="absolute top-1.5 right-1.5 bg-red-500 text-white text-[10px] font-bold px-1.5 py-0.5 rounded-full border-2 border-white ring-1 ring-red-200">
          {unreadCount > 9 ? '9+' : unreadCount}
        </span>
      )}
    </Link>
  );
};
