import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

// Склеивает классы Tailwind: условные части через clsx, конфликтующие (p-3 и p-4) — побеждает последний.
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
