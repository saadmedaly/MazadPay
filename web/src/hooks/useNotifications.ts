import { useEffect, useState, useRef } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { messaging, requestNotificationPermission } from '../lib/firebase';
import { onMessage } from 'firebase/messaging';
import { useAuthStore } from '../stores/authStore';
import api from '../api/client';
import { toast } from 'sonner';

export const useNotifications = () => {
  const { user, isAuthenticated } = useAuthStore();
  const [permission, setPermission] = useState<NotificationPermission | null>(null);
  const [fcmToken, setFcmToken] = useState<string | null>(null);
  const initialized = useRef(false);

  useEffect(() => {
    const initPermissions = async () => {
      if (initialized.current === true) return;
      if (!isAuthenticated || !user) return;
      initialized.current = true;

      const token = await requestNotificationPermission();
      if (token) {
        setFcmToken(token);
        setPermission('granted');

        try {
          // Augmenter le timeout pour ce endpoint spécifique (30s)
          await api.post('/v1/api/notifications/push-tokens', {
            fcm_token: token,
            device_id: navigator.userAgent,
            platform: 'web'
          }, { timeout: 30000 });
        } catch (err: any) {
          // Ignorer silencieusement les timeouts - le token sera resynchronisé plus tard
          if (err.code === 'ECONNABORTED' || err.message?.includes('timeout')) {
            console.warn('Push token sync timeout - will retry later');
          } else {
            console.error('Failed to sync FCM token:', err);
          }
        }
      }
    };

    initPermissions();
  }, [isAuthenticated, user]);

  useEffect(() => {
    if (!messaging) return;

    const unsubscribe = onMessage(messaging, (payload) => {
      console.log('Message received in foreground: ', payload);
      // On peut ajouter une notification locale ou mettre à jour un store ici
      if (Notification.permission === 'granted') {
        new Notification(payload.notification?.title || 'Notification', {
          body: payload.notification?.body,
          icon: '/logo192.png'
        });
      }
    });

    return () => unsubscribe();
  }, []);

  const requestPermission = async () => {
    if (initialized.current === true) return;
    if (!isAuthenticated || !user) return;
    initialized.current = true;

    const token = await requestNotificationPermission();
    if (token) {
      setFcmToken(token);
      setPermission('granted');

      try {
        await api.post('/v1/api/notifications/push-tokens', {
          fcm_token: token,
          device_id: navigator.userAgent,
          platform: 'web'
        });
      } catch (err) {
        console.error('Failed to sync FCM token:', err);
      }
    }
  };

  return { permission, fcmToken, requestPermission };
};

// Hook for sending notifications (admin functionality)
export const useSendNotification = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: {
      user_id?: string;
      title: string;
      body: string;
      type?: string;
      data?: Record<string, string>;
      broadcast?: boolean;
    }) => {
      const response = await api.post('/v1/api/admin/notifications/send', data);
      return response.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-notifications'] });
      toast.success('Notification envoyée avec succès');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Erreur lors de l\'envoi de la notification');
    },
  });
};

// Hook for fetching notifications (user)
export const useFetchNotifications = () => {
  return useMutation({
    mutationFn: async () => {
      const response = await api.get('/v1/api/notifications');
      return response.data.data || [];
    },
  });
};

export const useFetchAdminNotifications = (status: string = 'all', limit: number = 50) => {
  return useQuery({
    queryKey: ['admin-notifications', status, limit],
    queryFn: async () => {
      const response = await api.get('/v1/api/admin/notifications', {
        params: { status, limit }
      });
      const data = response.data.data || [];
      const uniqueIds = new Set();
      return data.filter((notif: any) => {
        if (uniqueIds.has(notif.id)) return false;
        uniqueIds.add(notif.id);
        return true;
      });
    },
  });
};

// Hook for deleting notifications
export const useDeleteNotification = () => {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/v1/api/admin/notifications/${id}`);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-notifications'] });
      toast.success('Notification supprimée avec succès');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Erreur lors de la suppression');
    },
  });
};

// Hook for marking notification as read (admin)
export const useMarkNotificationAsRead = () => {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      await api.put(`/v1/api/admin/notifications/${id}/read`);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-notifications'] });
      toast.success('Notification marquée comme lue');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Erreur lors du marquage');
    },
  });
};

// Hook for marking all notifications as read (admin)
export const useMarkAllAsReadAdmin = () => {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      await api.put(`/v1/api/admin/notifications/read-all`);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-notifications'] });
      toast.success('Toutes les notifications marquées comme lues');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Erreur lors du marquage');
    },
  });
};