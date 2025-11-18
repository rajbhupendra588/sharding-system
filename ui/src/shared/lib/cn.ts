/**
 * Class Name Utility
 * Combines class names with clsx
 */

import { type ClassValue, clsx } from 'clsx';

export function cn(...inputs: ClassValue[]) {
  return clsx(inputs);
}

