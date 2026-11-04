export interface Tour {
  id: string;
  slug: string;
  title: string;
  description: string;
  coverImage: string;
  duration: string;
  distance: string;
  price: number;
  isOnlineRoute: boolean;
  guideName?: string;
  guideRole?: string;
  maxGroupSize?: number;
  nextDate?: string;
  stopsCount: number;
  stops: {
    title: string;
    address: string;
    description: string;
  }[];
}
