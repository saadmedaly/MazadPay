import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../api/client';
import { toast } from 'sonner';

export interface DashboardWidget {
  id: string;
  title: string;
  type: 'metric' | 'chart' | 'table';
  position: { x: number; y: number; w: number; h: number };
  config: Record<string, any>;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateWidgetData {
  title: string;
  type: 'metric' | 'chart' | 'table';
  position: { x: number; y: number; w: number; h: number };
  config: Record<string, any>;
}

// Hook for fetching dashboard widgets
export const useDashboardWidgets = () => {
  return useQuery({
    queryKey: ['dashboard-widgets'],
    queryFn: async () => {
      const { data } = await api.get('/v1/api/admin/dashboard/widgets');
      return data.data as DashboardWidget[];
    },
  });
};

// Hook for creating a dashboard widget
export const useCreateDashboardWidget = () => {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: async (widgetData: CreateWidgetData) => {
      const response = await api.post('/v1/api/admin/dashboard/widgets', widgetData);
      return response.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['dashboard-widgets'] });
      toast.success('Widget créé avec succès');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Erreur lors de la création du widget');
    },
  });
};

// Hook for updating a dashboard widget
export const useUpdateDashboardWidget = () => {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: Partial<CreateWidgetData> }) => {
      const response = await api.put(`/v1/api/admin/dashboard/widgets/${id}`, data);
      return response.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['dashboard-widgets'] });
      toast.success('Widget mis à jour avec succès');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Erreur lors de la mise à jour du widget');
    },
  });
};

// Hook for deleting a dashboard widget
export const useDeleteDashboardWidget = () => {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/v1/api/admin/dashboard/widgets/${id}`);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['dashboard-widgets'] });
      toast.success('Widget supprimé avec succès');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Erreur lors de la suppression du widget');
    },
  });
};

// Hook for repositioning widgets
export const useRepositionWidgets = () => {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: async (widgets: Array<{ id: string; position: { x: number; y: number; w: number; h: number } }>) => {
      const response = await api.put('/v1/api/admin/dashboard/widgets/reposition', { widgets });
      return response.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['dashboard-widgets'] });
      toast.success('Position des widgets mise à jour');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Erreur lors de la reposition des widgets');
    },
  });
};
