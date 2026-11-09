'use client';

import { useQuery } from '@tanstack/react-query';
import { objectApi, objectKeys } from './objectApi';

export const useObjects = (all = false) =>
  useQuery({ queryKey: objectKeys.list(all), queryFn: () => objectApi.list(all) });

export const useObject = (slug: string | null | undefined) =>
  useQuery({
    queryKey: objectKeys.card(slug ?? ''),
    queryFn: () => objectApi.card(slug!),
    enabled: !!slug,
  });

export const useCategories = () =>
  useQuery({ queryKey: objectKeys.categories, queryFn: objectApi.categories, staleTime: 5 * 60_000 });
