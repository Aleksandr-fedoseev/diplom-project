export interface StorySource {
  type: 'archive' | 'book' | 'museum' | 'interview' | 'link';
  typeLabel: string;
  citation: string;
}

export interface StoryImage {
  url: string;
  caption: string;
  source?: string;
}

export interface StoryContentBlock {
  type: 'paragraph' | 'quote' | 'image' | 'gallery' | 'audio' | 'video';
  text?: string;
  author?: string; // для цитат
  images?: StoryImage[]; // для фото и галерей
  audioTitle?: string;
  audioDuration?: string;
  videoUrl?: string;
  videoTitle?: string;
}

export interface Story {
  id: string;
  slug: string;
  title: string;
  lead: string;
  category: string;
  categoryLabel: string;
  publishedAt: string;
  readTime: string;
  coverImage: string;
  coverCaption?: string;
  author: {
    name: string;
    role: string;
    avatar?: string;
  };
  blocks: StoryContentBlock[];
  sources: StorySource[]; // Развернутый блок источников по ТЗ 5.2
  mapCoordinates?: {
    lat: number;
    lng: number;
    title: string;
    address: string;
  };
  relatedObjectSlugs?: string[]; // Связанные объекты на карте
  relatedStorySlugs?: string[]; // Связанные статьи («Читайте также»)
  isLeadArticle?: boolean;
  isPopular?: boolean;
}
