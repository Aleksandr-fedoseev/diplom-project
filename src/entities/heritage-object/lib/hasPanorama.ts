import type { HeritageObject } from '../model/types';

// В списке сервер присылает признак hasPanorama, в карточке есть и сами панорамы.
export const hasPanorama = (obj: Pick<HeritageObject, 'hasPanorama' | 'panoramas'>) =>
  obj.hasPanorama || (obj.panoramas?.length ?? 0) > 0;
