import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../api/client';
import { toast } from 'sonner';

export interface Setting {
  id: number;
  key: string;
  value: string;
  type: 'boolean' | 'number' | 'text' | 'select';
  description?: string;
  group: string;
  created_at: string;
  updated_at: string;
}

export interface CreateSettingData {
  key: string;
  value: string;
  type: 'boolean' | 'number' | 'text' | 'select';
  description?: string;
  group: string;
  options?: { value: string; label: string }[];
}

// Hook for fetching all settings
export const useSettings = () => {
  return useQuery({
    queryKey: ['settings'],
    queryFn: async () => {
      const { data } = await api.get('/v1/api/admin/settings');
      return data.data as Setting[];
    },
  });
};

// Hook for fetching settings by group
export const useSettingsByGroup = (group: string) => {
  return useQuery({
    queryKey: ['settings', group],
    queryFn: async () => {
      const { data } = await api.get(`/v1/api/admin/settings`, { params: { group } });
      return data.data as Setting[];
    },
  });
};

// Hook for creating a new setting
export const useCreateSetting = () => {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: async (settingData: CreateSettingData) => {
      const response = await api.post('/v1/api/admin/settings', settingData);
      return response.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['settings'] });
      toast.success('Paramètre créé avec succès');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Erreur lors de la création du paramètre');
    },
  });
};

// Hook for updating a setting
export const useUpdateSetting = () => {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: async ({ key, data }: { key: string; data: Partial<CreateSettingData> }) => {
      const response = await api.put(`/v1/api/admin/settings/${key}`, data);
      return response.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['settings'] });
      toast.success('Paramètre mis à jour avec succès');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Erreur lors de la mise à jour du paramètre');
    },
  });
};

// Hook for deleting a setting
export const useDeleteSetting = () => {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: async (key: string) => {
      await api.delete(`/v1/api/admin/settings/${key}`);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['settings'] });
      toast.success('Paramètre supprimé avec succès');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Erreur lors de la suppression du paramètre');
    },
  });
};

// Hook for bulk updating settings
export const useBulkUpdateSettings = () => {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: async (settings: Array<{ key: string; value: string }>) => {
      const response = await api.put('/v1/api/admin/settings/bulk', { settings });
      return response.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['settings'] });
      toast.success('Paramètres mis à jour en masse');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Erreur lors de la mise à jour des paramètres');
    },
  });
};
