'use client';

import React, { useEffect, useRef } from 'react';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import { HeritageObject } from '@/types/heritage';
import { ASTRAKHAN_BOUNDS } from '@/data/heritage';

interface AstrakhanMapProps {
  objects: HeritageObject[];
  activeObject: HeritageObject | null;
  onSelectObject: (obj: HeritageObject) => void;
}

export default function AstrakhanMap({
  objects,
  activeObject,
  onSelectObject,
}: AstrakhanMapProps) {
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const mapInstanceRef = useRef<L.Map | null>(null);
  const markersLayerRef = useRef<L.LayerGroup | null>(null);

  // 1. Initialize Leaflet Map locked strictly to Astrakhan Oblast
  useEffect(() => {
    if (!mapContainerRef.current || mapInstanceRef.current) return;

    const southWest = L.latLng(ASTRAKHAN_BOUNDS.bounds[0][0], ASTRAKHAN_BOUNDS.bounds[0][1]);
    const northEast = L.latLng(ASTRAKHAN_BOUNDS.bounds[1][0], ASTRAKHAN_BOUNDS.bounds[1][1]);
    const astrakhanBounds = L.latLngBounds(southWest, northEast);

    const map = L.map(mapContainerRef.current, {
      center: ASTRAKHAN_BOUNDS.center,
      zoom: ASTRAKHAN_BOUNDS.defaultZoom,
      minZoom: ASTRAKHAN_BOUNDS.minZoom,
      maxZoom: ASTRAKHAN_BOUNDS.maxZoom,
      maxBounds: astrakhanBounds,
      maxBoundsViscosity: 1.0, // Жесткое удержание камеры внутри Астраханской области
      zoomControl: false,
      attributionControl: false, // Отключаем стандартный блок атрибуции Leaflet с флагом
    });

    // Real OpenStreetMap Standard Tile Layer
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      maxZoom: 19,
    }).addTo(map);

    // Add controls in top right
    L.control.zoom({ position: 'topright' }).addTo(map);
    L.control.scale({ position: 'bottomright', imperial: false, metric: true }).addTo(map);

    markersLayerRef.current = L.layerGroup().addTo(map);
    mapInstanceRef.current = map;

    return () => {
      map.remove();
      mapInstanceRef.current = null;
    };
  }, []);

  // 2. Render Markers for Astrakhan Monuments
  useEffect(() => {
    const map = mapInstanceRef.current;
    const markersLayer = markersLayerRef.current;
    if (!map || !markersLayer) return;

    markersLayer.clearLayers();

    objects.forEach((obj) => {
      const isSelected = activeObject?.id === obj.id;
      const isLost = obj.category === 'lost';

      const iconHtml = `
        <div class="custom-osm-pin ${isSelected ? 'is-active' : ''} ${isLost ? 'is-lost' : ''}">
          <div class="pin-badge">
            ${obj.hasPanorama360 ? '<span class="vr-badge">360°</span>' : ''}
            <span class="pin-symbol">${isLost ? '⚑' : '🏛'}</span>
          </div>
          <div class="pin-title">${obj.title}</div>
        </div>
      `;

      const customIcon = L.divIcon({
        html: iconHtml,
        className: 'astrakhan-marker-container',
        iconSize: [120, 48],
        iconAnchor: [60, 48],
      });

      const marker = L.marker([obj.coordinates.lat, obj.coordinates.lng], {
        icon: customIcon,
      });

      marker.on('click', () => {
        onSelectObject(obj);
        map.flyTo([obj.coordinates.lat, obj.coordinates.lng], 16, {
          duration: 0.8,
        });
      });

      markersLayer.addLayer(marker);
    });
  }, [objects, activeObject, onSelectObject]);

  // 3. Pan to active object if changed from outside
  useEffect(() => {
    const map = mapInstanceRef.current;
    if (!map || !activeObject) return;

    map.flyTo([activeObject.coordinates.lat, activeObject.coordinates.lng], 16, {
      duration: 0.8,
    });
  }, [activeObject]);

  return (
    <div className="w-full h-full relative">
      <div ref={mapContainerRef} className="w-full h-full z-0" />

      <style jsx global>{`
        .astrakhan-marker-container {
          background: transparent;
          border: none;
        }
        .custom-osm-pin {
          display: flex;
          flex-direction: column;
          align-items: center;
          cursor: pointer;
          transform: translateY(0);
          transition: transform 0.15s ease;
        }
        .custom-osm-pin:hover {
          transform: translateY(-3px) scale(1.05);
          z-index: 999;
        }
        .custom-osm-pin.is-active {
          transform: translateY(-4px) scale(1.1);
          z-index: 1000;
        }
        .pin-badge {
          display: inline-flex;
          align-items: center;
          justify-content: center;
          gap: 3px;
          background: #0F172A;
          color: #FFFFFF;
          border: 2px solid #FFFFFF;
          box-shadow: 0 2px 8px rgba(0,0,0,0.3);
          border-radius: 20px;
          padding: 3px 8px;
          font-size: 11px;
          font-weight: 700;
          position: relative;
        }
        .custom-osm-pin.is-active .pin-badge {
          background: #2563EB;
          border-color: #DBEAFE;
          box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.4);
        }
        .custom-osm-pin.is-lost .pin-badge {
          background: #991B1B;
          border-style: dashed;
        }
        .vr-badge {
          background: #2563EB;
          color: #FFF;
          font-size: 9px;
          padding: 1px 4px;
          border-radius: 4px;
          font-weight: 800;
        }
        .pin-title {
          background: rgba(255, 255, 255, 0.95);
          border: 1px solid #E2E8F0;
          border-radius: 4px;
          padding: 1px 5px;
          font-size: 10px;
          font-weight: 600;
          color: #1E293B;
          white-space: nowrap;
          max-width: 140px;
          overflow: hidden;
          text-overflow: ellipsis;
          box-shadow: 0 1px 4px rgba(0,0,0,0.1);
          margin-top: 2px;
        }
        .custom-osm-pin.is-active .pin-title {
          background: #0F172A;
          color: #FFFFFF;
          border-color: #0F172A;
        }
        .leaflet-control-attribution,
        .leaflet-attribution-flag {
          display: none !important;
        }
      `}</style>
    </div>
  );
}
