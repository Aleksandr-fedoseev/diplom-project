import type { HeritageObject } from '../model/types';

export const hasPanorama = (obj: HeritageObject) => (obj.panoramas?.length ?? 0) > 0;
