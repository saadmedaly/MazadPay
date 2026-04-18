import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getRatings, deleteRating, type RatingsParams } from '@/api/ratings'

export function useRatings(params?: RatingsParams) {
  return useQuery({
    queryKey: ['ratings', params],
    queryFn: () => getRatings(params),
  })
}

export function useDeleteRating() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: deleteRating,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['ratings'] }),
  })
}