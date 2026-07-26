import React, { useEffect, useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "./ui/card";
import { Badge } from "./ui/badge";
import { Target, Heart, Mail, Phone, MapPin, Globe } from "lucide-react";
import { API_BASE_URL } from "../config/api";
import objectivesImage from "../assets/tagatowerabout.jpg";
/* ---------------- TYPES ---------------- */

interface ContactData {
  primary_email: string;
  secondary_email?: string;
  primary_phone: string;
  website?: string;
  headquarters?: {
    address: string;
  };
}

interface ObjectiveData {
  id: number;
  title?: string;
  name: string;
  description: string;
}

interface StatData {
  id: number;
  label: string;
  value: number;
}

interface AboutData {
  name: string;
  established_year: number;
  tagline: string;
  hero_image_url: string;
  mission: string;
  vision: string;
  description: string;
}

/* ---------------- COMPONENT ---------------- */

export function AboutUs() {
  const [about, setAbout] = useState<AboutData | null>(null);
  const [contact, setContact] = useState<ContactData | null>(null);
  const [objectives, setObjectives] = useState<ObjectiveData[]>([]);
  const [stats, setStats] = useState<StatData[]>([]);
  const [loading, setLoading] = useState(true);
  const bannerImage = `${API_BASE_URL}/images/about.jpg`;

  useEffect(() => {
    const fetchAll = async () => {
      try {
        const base = `${API_BASE_URL}/public/about`;

        const [a, c, o, s] = await Promise.all([
          fetch(base),
          fetch(`${base}/contact`),
          fetch(`${base}/objectives`),
          fetch(`${base}/stats`),
        ]);

        const aboutData = await a.json();
        const contactData = await c.json();
        const objectivesData = await o.json();
        const statsData = await s.json();

        setAbout(aboutData);
        setContact(contactData);
        setObjectives(objectivesData);

        // ✅ ensure stats is always array
        setStats(Array.isArray(statsData) ? statsData : []);
      } catch (err) {
        console.error("API ERROR:", err);
      } finally {
        setLoading(false);
      }
    };

    fetchAll();
  }, []);

  if (loading) return <div className="h-64 flex items-center justify-center">Loading...</div>;

  return (
    <div className="space-y-10">
      {/* ✅ HERO (same design) */}
      <div className="relative overflow-hidden rounded-2xl shadow-2xl">
        <div className="absolute inset-0">
          <img
            src={bannerImage}
            alt="Hero"
            className="w-full h-full object-cover"
          />
          <div className="absolute inset-0 bg-gradient-to-r from-green-900/95 via-green-800/90 to-green-900/80" />
        </div>

        <div className="relative p-12 md:p-16">
          <div className="max-w-3xl">
            <Badge className="mb-4 bg-green-600 text-white">
              Established {about?.established_year}
            </Badge>

            <h1 className="text-5xl md:text-6xl font-bold text-white mb-6">
              {about?.name}
            </h1>

            <p className="text-xl text-green-50 mb-8">
              {about?.tagline}
            </p>

            {/* ✅ STATS */}
            <div className="flex flex-wrap gap-4">
              {stats.map((s) => (
                <div
                  key={s.id}
                  className="bg-white/10 backdrop-blur-sm rounded-lg px-6 py-3 border border-white/20"
                >
                  <p className="text-3xl font-bold text-white">
                    {Number(s.value).toLocaleString()}+
                  </p>
                  <p className="text-sm text-green-100">{s.label}</p>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* ✅ ABOUT */}
      <Card className="border-green-100 bg-gradient-to-br from-background to-secondary/20 shadow-md">
        <CardHeader>
          <CardTitle className="text-primary text-xl font-bold">About Us</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground leading-relaxed">
            {about?.description}
          </p>
        </CardContent>
      </Card>

      {/* ✅ MISSION & VISION (same UI) */}
      <div className="grid md:grid-cols-2 gap-8">
        <Card className="border-green-100 bg-gradient-to-br from-background to-secondary/20 shadow-md">
          <CardHeader>
            <CardTitle className="text-primary text-xl font-bold">Our Mission
              {/* <Target className="w-5 h-5 text-primary" /> Our Mission */}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground leading-relaxed">
              {about?.mission}
            </p>
          </CardContent>
        </Card>

        <Card className="border-green-100 bg-gradient-to-br from-background to-secondary/30 shadow-md">
          <CardHeader>
            <CardTitle className="text-primary text-xl font-bold">Our Vision
              {/* <Heart className="w-5 h-5 text-primary" /> Our Vision */}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground leading-relaxed">
              {about?.vision}
            </p>
          </CardContent>
        </Card>
      </div>

      <Card className="border border-[#e2e4db] shadow-md">
        <CardHeader>
          <CardTitle className="text-primary text-xl font-bold">Our Objectives</CardTitle>
          <CardDescription> Key goals that drive our association's activities and initiatives </CardDescription>
        </CardHeader>

        <CardContent>
          {/* Change the grid from md:grid-cols-2 to this: */}
          <div className="grid md:grid-cols-3 gap-8">

            {/* LEFT SIDE - takes 2 columns */}
            <div className="md:col-span-2 space-y-4">
              {objectives.map((obj) => (
                <div key={obj.id} className="flex items-start space-x-4">
                  <div className="w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center flex-shrink-0 mt-1">
                    <div className="w-3 h-3 bg-primary rounded-full" />
                  </div>
                  <div>
                    <h4 className="font-semibold text-foreground mb-1 text-base">{obj.title}</h4>
                    <p className="text-muted-foreground text-base">{obj.description}</p>
                  </div>
                </div>
              ))}
            </div>

            {/* RIGHT SIDE - takes 1 column, natural square */}
            <div className="hidden md:flex items-center justify-center">
              <img
                src={objectivesImage}
                alt="TAGA Tower"
                className="rounded-xl shadow-md w-full aspect-square object-cover"
                style={{ objectPosition: '50% 30%' }}
              />
            </div>

          </div>
        </CardContent>
      </Card>

      {/* ✅ CONTACT - SAME AS HARDCODED UI */}
      {contact && (
        <Card className="bg-gradient-to-r from-primary/5 to-accent/20 border border-[#e2e4db] shadow-md">
          <CardHeader>
            <CardTitle className="text-primary text-xl font-bold">Contact Us</CardTitle>
            <CardDescription>
              Get in touch with TAGA for any inquiries or support
            </CardDescription>
          </CardHeader>

          <CardContent>
            <div className="grid md:grid-cols-2 gap-8">

              {/* LEFT SIDE */}
              <div className="space-y-6">

                {/* EMAIL */}
                <div className="flex items-start space-x-3">
                  <Mail className="w-5 h-5 text-primary mt-1" />
                  <div>
                    <h4 className="font-semibold mb-1">Email</h4>

                    {/* Primary */}
                    <p className="text-sm text-muted-foreground">
                      <span className="font-medium text-foreground">General id: </span>
                      <a
                        href={`mailto:${contact.primary_email}`}
                        className="text-primary hover:underline"
                      >
                        {contact.primary_email}
                      </a>
                    </p>

                    {/* Secondary */}
                    {contact.secondary_email && (
                      <p className="text-sm text-muted-foreground">
                        <span className="font-medium text-foreground">Seithikkathir: </span>
                        <a
                          href={`mailto:${contact.secondary_email}`}
                          className="text-primary hover:underline"
                        >
                          {contact.secondary_email}
                        </a>
                      </p>
                    )}
                  </div>
                </div>

                {/* PHONE */}
                <div className="flex items-start space-x-3">
                  <Phone className="w-5 h-5 text-primary mt-1" />
                  <div>
                    <h4 className="font-semibold mb-1">Phone</h4>
                    <p className="text-sm text-muted-foreground">
                      {contact.primary_phone}
                    </p>
                  </div>
                </div>

                {/* WEBSITE */}
                {contact.website && (
                  <div className="flex items-start space-x-3">
                    <Globe className="w-5 h-5 text-primary mt-1" />
                    <div>
                      <h4 className="font-semibold mb-1">Website</h4>
                      <p className="text-sm text-muted-foreground">
                        <a
                          href={`https://${contact.website}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary hover:underline"
                        >
                          {contact.website}
                        </a>
                      </p>
                    </div>
                  </div>
                )}
              </div>

              {/* RIGHT SIDE (ADDRESS) */}
              <div className="space-y-6">
                <div className="flex items-start space-x-3">
                  <MapPin className="w-5 h-5 text-primary mt-1" />
                  <div>
                    <h4 className="font-semibold mb-2">Address</h4>

                    <div className="bg-white/50 dark:bg-gray-800/50 p-4 rounded-lg border border-primary/20 max-w-sm">
                      <p className="text-sm text-muted-foreground leading-relaxed">
                        {contact.headquarters?.address}
                      </p>
                    </div>

                  </div>
                </div>
              </div>

            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}