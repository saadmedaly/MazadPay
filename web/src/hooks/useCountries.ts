import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getCountries, createCountry, updateCountry, deleteCountry } from '@/api/countries'
import type { Country } from '@/types/api'

export function useCountries() {
  return useQuery({
    queryKey: ['countries'],
    queryFn: getCountries,
  })
}

export function useCreateCountry() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: createCountry,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['countries'] }),
  })
}

export function useUpdateCountry() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: Partial<Country> }) =>
      updateCountry(id, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['countries'] }),
  })
}

export function useDeleteCountry() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: deleteCountry,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['countries'] }),
  })
}
