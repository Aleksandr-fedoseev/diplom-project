import { hasPanorama, type HeritageObject } from '@/entities/heritage-object';

// Текст из данных (а потом из API и заявок пользователей) нельзя вставлять в HTML как есть.
export function escapeHtml(s: string): string {
  return s.replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c]!);
}

// HTML значка объекта на карте Leaflet.
export function markerHtml(object: HeritageObject, selected: boolean): string {
  const isLost = object.category === 'lost';
  const classes = ['custom-osm-pin', selected && 'is-active', isLost && 'is-lost'].filter(Boolean).join(' ');
  return `
    <div class="${classes}">
      <div class="pin-badge">
        ${hasPanorama(object) ? '<span class="vr-badge">360°</span>' : ''}
        <span class="pin-symbol">${isLost ? '⚑' : '🏛'}</span>
      </div>
      <div class="pin-title">${escapeHtml(object.title)}</div>
    </div>`;
}
