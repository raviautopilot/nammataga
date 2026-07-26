import React, { useState, useEffect } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "./ui/card";
import { Button } from "./ui/button";
import { Badge } from "./ui/badge";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "./ui/tabs";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "./ui/dialog";
import {
  Calendar,
  Image as ImageIcon,
  MapPin,
  Clock,
  ExternalLink,
} from "lucide-react";
import { format } from "date-fns";

interface EventsProps {
  isLoggedIn: boolean;
}

interface Event {
  id: string;
  title: string;
  date: Date;
  hasTime?: boolean;
  location: string;
  description: string;
  attendees?: number;
  imageUrl?: string;
  status: "upcoming" | "completed";
}

interface GalleryImage {
  id: number;
  title: string;
  date: Date;
  event: string;
  imageUrl: string | null;
  year: number;
}

import { API_BASE_URL } from "../config/api";

export function Events({ isLoggedIn }: EventsProps) {

  const API_BASE = API_BASE_URL;
  const BASE_URL = API_BASE.replace('/api', '');
  const [selectedYear, setSelectedYear] = useState<number>(0);
  const [years, setYears] = useState<number[]>([]);
  const [galleryByYear, setGalleryByYear] = useState<{ [key: number]: GalleryImage[] }>({});
  const [upcomingEvents, setUpcomingEvents] = useState<Event[]>([]);
  const [selectedImage, setSelectedImage] = useState<GalleryImage | null>(null);
  const bannerImage = `${API_BASE}/images/banner-image.jpg`;
  // Banner image (use local or external)
//  const bannerImage = `${API_BASE}/images/events-banner.jpg`; // Adjust path as needed

  // 🔥 FETCH YEARS
  useEffect(() => {
    fetch(`${API_BASE}/gallery/years`)
      .then(res => {
        if (!res.ok) throw new Error("API ERROR");
        return res.json();
      })
      .then(data => {
        let existingYears: number[] = data || [];

        const futureYears = [2025, 2026];
        futureYears.forEach(year => {
          if (!existingYears.includes(year)) {
            existingYears.push(year);
          }
        });

        existingYears.sort((a: number, b: number) => b - a);

        setYears(existingYears);
        
        if (existingYears && existingYears.length > 0) {
          if (existingYears.includes(2026)) {
            setSelectedYear(2026);
          } else {
            setSelectedYear(existingYears[0]);
          }
        }
      })
      .catch(err => console.error("YEAR ERROR:", err));
  }, [API_BASE]);

  // 🔥 FETCH GALLERY
  useEffect(() => {
    if (!selectedYear) return;

    fetch(`${API_BASE}/gallery?year=${selectedYear}`)
      .then(res => {
        if (!res.ok) throw new Error("API ERROR");
        return res.json();
      })
      .then(data => {
        const formatted: GalleryImage[] = (data || []).map((img: any, index: number) => {
          let imageUrl = null;
          const imageField = img.imageUrl || img.ImageURL || img.image;

          if (imageField) {
            if (imageField.startsWith('http')) {
              imageUrl = imageField;
            } else if (imageField.startsWith('/api/')) {
              imageUrl = `${API_BASE}${imageField.replace(/^\/api/, '')}`;
            } else if (imageField.startsWith('/')) {
              imageUrl = `${API_BASE}${imageField}`;
            } else {
              imageUrl = `${API_BASE}/images/${imageField}`;
            }
          }
          return {
            id: index,
            title: img.title,
            date: new Date(img.date),
            event: img.event,
            imageUrl: imageUrl,
            year: selectedYear
          };
        });

        setGalleryByYear(prev => ({
          ...prev,
          [selectedYear]: formatted
        }));
      })
      .catch(err => console.error("GALLERY ERROR:", err));
  }, [selectedYear, API_BASE]);

  // 🔥 FETCH EVENTS
  useEffect(() => {
    fetch(`${API_BASE}/events/upcoming`)
      .then(res => {
        if (!res.ok) throw new Error("API ERROR");
        return res.json();
      })
      .then(data => {
        const today = new Date();
        today.setHours(0, 0, 0, 0);

        const mapped: Event[] = (data || [])
          .map((e: any, index: number) => {
            const hasTime = e.date && e.date.trim().split(/\s+/).length > 1;
            return {
              id: String(index),
              title: e.title || "No Title",
              date: new Date(e.date),
              hasTime,
              location: e.location || "N/A",
              description: e.description || "",
              status: "upcoming"
            };
          })
          .filter((e: Event) => !isNaN(e.date.getTime()) && e.date >= today);
        setUpcomingEvents(mapped);
      })
      .catch(err => console.error("EVENT ERROR:", err));
  }, [API_BASE]);

  const handlePreviousImage = () => {
    if (!selectedImage) return;
    const currentYearImages = galleryByYear[selectedYear] || [];
    if (currentYearImages.length === 0) return;
    
    const currentIndex = currentYearImages.findIndex((img: GalleryImage) => img.id === selectedImage.id);
    const prevIndex = currentIndex === 0 ? currentYearImages.length - 1 : currentIndex - 1;
    setSelectedImage(currentYearImages[prevIndex]);
  };

  const handleNextImage = () => {
    if (!selectedImage) return;
    const currentYearImages = galleryByYear[selectedYear] || [];
    if (currentYearImages.length === 0) return;
    
    const currentIndex = currentYearImages.findIndex((img: GalleryImage) => img.id === selectedImage.id);
    const nextIndex = currentIndex === currentYearImages.length - 1 ? 0 : currentIndex + 1;
    setSelectedImage(currentYearImages[nextIndex]);
  };

  return (
    <div className="space-y-8">

           {/* Header */}
      <div className="relative overflow-hidden rounded-2xl shadow-2xl">
        <div className="absolute inset-0">
          <img 
            src={bannerImage}
            className="w-full h-full object-cover"
          />
          <div className="absolute inset-0 bg-green-900/80" />
        </div>

        <div className="relative p-10 text-white">
          <Badge className="mb-4 bg-green-600">
            <Calendar className="w-3 h-3 mr-1" />
            Events & Gallery
          </Badge>

          <h1 className="text-4xl font-bold mb-2">
            TAGA Events
          </h1>

          <p>
            Stay updated with our latest events, workshops, and community activities
          </p>
        </div>
      </div>

      <Tabs defaultValue="gallery" className="space-y-6" data-testid="testid-events-form">
        <TabsList className="grid w-full grid-cols-2" data-testid="testid-events-tabs-list">
          <TabsTrigger value="gallery" data-testid="testid-gallery-button">Gallery</TabsTrigger>
          <TabsTrigger value="upcoming" data-testid="testid-upcoming-events-button">Upcoming Events</TabsTrigger>
        </TabsList>

        {/* GALLERY */}
        <TabsContent value="gallery" className="space-y-6" data-testid="testid-gallery-tab-content">
          <Card>
            <CardHeader>
              <CardTitle>Photo Gallery</CardTitle>
              <CardDescription>
                Browse through photos from our events and programs
              </CardDescription>
            </CardHeader>
            <CardContent>

              <div className="flex space-x-2 mb-6">
                {years.map((year) => (
                  <Button
                    key={year}
                    variant={selectedYear === year ? "default" : "outline"}
                    onClick={() => setSelectedYear(year)}
                    data-testid={`testid-gallery-year-${year}-button`}
                  >
                    {year}
                  </Button>
                ))}
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                {(galleryByYear[selectedYear] || []).map((image: GalleryImage) => (
                  <Card key={image.id} className="overflow-hidden group p-0">
                    <div className="aspect-video bg-muted relative">
                      {image.imageUrl ? (
                        <img src={image.imageUrl} className="w-full h-full object-cover" alt={image.title} />
                      ) : (
                        <div className="flex items-center justify-center h-full">
                          <ImageIcon className="w-8 h-8 text-muted-foreground" />
                        </div>
                      )}
                      <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 flex items-center justify-center transition">
                        <Button size="sm" onClick={() => setSelectedImage(image)} data-testid={`testid-view-gallery-image-${image.id}-button`}>
                          <ExternalLink className="w-4 h-4 mr-2" />
                          View
                        </Button>
                      </div>
                    </div>

                    <CardContent className="pt-0 pb-2 px-2">
                      <h4 className="font-bold text-center">{image.title}</h4>
                    </CardContent>
                  </Card>
                ))}
              </div>

              {(galleryByYear[selectedYear] || []).length === 0 && (
                <div className="text-center py-12">
                  <ImageIcon className="w-16 h-16 mx-auto text-muted-foreground mb-4" />
                  <p>No photos available</p>
                </div>
              )}

            </CardContent>
          </Card>
        </TabsContent>

        {/* EVENTS */}
        <TabsContent value="upcoming" className="space-y-6" data-testid="testid-upcoming-events-tab-content">
          <Card>
            <CardHeader>
              <CardTitle>Upcoming Events</CardTitle>
              <CardDescription>
                Mark your calendar for these upcoming TAGA events
              </CardDescription>
            </CardHeader>

            <CardContent>
              <div className="space-y-4">
                {upcomingEvents.map((event: Event) => (
                  <Card key={event.id}>
                    <CardContent className="pt-6">
                      <h3 className="text-xl font-bold">{event.title}</h3>
                      <p className="text-muted-foreground">{event.description}</p>
                      <div className="mt-2 text-sm text-muted-foreground">
                        <p><MapPin className="inline w-4 h-4 mr-1" /> {event.location}</p>
                        <p>
                          <Clock className="inline w-4 h-4 mr-1" />
                          {format(event.date, "dd MMM yyyy")}
                          {event.hasTime && ` at ${format(event.date, "h:mm a")}`}
                        </p>
                      </div>
                    </CardContent>
                  </Card>
                ))}
                {upcomingEvents.length === 0 && (
                  <div className="text-center py-12 text-muted-foreground">
                    <Calendar className="w-12 h-12 mx-auto mb-3 opacity-60" />
                    <p className="text-base font-medium">No upcoming events scheduled</p>
                    <p className="text-sm">Please check back later for updates on association programs.</p>
                  </div>
                )}
              </div>
            </CardContent>

          </Card>
        </TabsContent>
      </Tabs>
      
      {/* Image Full View Dialog */}
      <Dialog open={!!selectedImage} onOpenChange={() => setSelectedImage(null)}>
        <DialogContent
          className="p-0 overflow-hidden flex flex-col"
          data-testid="testid-gallery-image-modal"
          style={{
            width: "90vw",
            maxWidth: "90vw",
            maxHeight: "92vh",
            borderRadius: "1.5rem",
            background: "rgba(255,255,255,0.85)",
            backdropFilter: "blur(16px)",
            WebkitBackdropFilter: "blur(16px)",
            border: "1px solid rgba(255,255,255,0.4)",
            boxShadow: "0 8px 40px rgba(0,0,0,0.18)",
          }}
        >
          <DialogHeader className="px-6 pt-5 pb-3 flex-shrink-0">
            <DialogTitle className="text-lg font-semibold">{selectedImage?.title}</DialogTitle>
          </DialogHeader>
          
          <div className="relative px-6 pb-6 overflow-y-auto flex flex-col gap-4">
            <div className="relative flex items-center justify-center">
              <button
                onClick={handlePreviousImage}
                className="absolute left-0 z-10 bg-black/50 hover:bg-black/70 text-white rounded-full p-2 transition-all duration-200"
                style={{ left: '-20px' }}
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="15 18 9 12 15 6"></polyline>
                </svg>
              </button>

              {selectedImage?.imageUrl && (
                <img
                  src={selectedImage.imageUrl}
                  alt={selectedImage.title}
                  style={{
                    maxHeight: "68vh",
                    maxWidth: "100%",
                    width: "auto",
                    height: "auto",
                    display: "block",
                    borderRadius: "1.25rem",
                    margin: "0 auto",
                  }}
                />
              )}

              <button
                onClick={handleNextImage}
                className="absolute right-0 z-10 bg-black/50 hover:bg-black/70 text-white rounded-full p-2 transition-all duration-200"
                style={{ right: '-20px' }}
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="9 18 15 12 9 6"></polyline>
                </svg>
              </button>
            </div>

            <div className="flex-shrink-0 space-y-1">
            </div>
          </div>
        </DialogContent>
      </Dialog>

    </div>
  );
}