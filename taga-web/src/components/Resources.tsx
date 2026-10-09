import React, { useEffect, useState } from 'react';
import { Card, CardContent } from './ui/card';
import { Button } from './ui/button';
import { ScrollArea } from './ui/scroll-area';
import { FileText, ExternalLink, Calendar, AlertCircle } from 'lucide-react';
import {
  getCategories,
  getDocumentsByCategory,
  getExternalLinks,
  getFileUrl,
  sortDocuments,
  sortExternalLinks
} from "../api/resources";
import { Badge } from './ui/badge';
import API_BASE_URL from '../config/api';

interface ResourcesProps {
  isLoggedIn: boolean;
  isPaidMember?: boolean;
  onNavigateToMembership?: () => void;
}

interface Category {
  id: string;
  name: string;
}

interface Document {
  title: string;
  year: string;
  subcategory?: string;
  url?: string;
}

export function Resources({ isLoggedIn, isPaidMember = true, onNavigateToMembership }: ResourcesProps) {
  const [categories, setCategories] = useState<Category[]>([]);
  const [documents, setDocuments] = useState<Document[]>([]);
  const [selectedCategory, setSelectedCategory] = useState<string>('');
  const [selectedSubcategory, setSelectedSubcategory] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [externalLinks, setExternalLinks] = useState<any[]>([]);
  const bannerImage = `${API_BASE_URL}/images/banner-image.jpg`;

  /* -------------------------------
     ✅ FIX: OPEN LINK SAFELY
  -------------------------------- */
  const openLink = (url: string) => {
    let finalUrl = url;
    if (!url.startsWith("http://") && !url.startsWith("https://")) {
      finalUrl = "https://" + url;
    }
    window.open(finalUrl, "_blank", "noopener,noreferrer");
  };

  /* -------------------------------
     Load Categories
  -------------------------------- */
  useEffect(() => {
    const load = async () => {
      try {
        const data = await getCategories();
        setCategories(data);
        if (data.length > 0) {
          setSelectedCategory(data[0].id);
        }
      } catch (err) {
        console.error("Error loading categories:", err);
      }
    };
    load();
  }, []);

  /* -------------------------------
     Load Documents with Sorting
  -------------------------------- */
  useEffect(() => {
    if (!selectedCategory) return;

    const loadDocs = async () => {
      try {
        setLoading(true);
        const data = await getDocumentsByCategory(
          selectedCategory,
          selectedCategory === "scheme-gos" ? selectedSubcategory : undefined
        );

        // ✅ Apply sorting: newest year first, then alphabetical
        const sortedData = sortDocuments(data || []);
        setDocuments(sortedData);
      } catch (err) {
        console.error("Error loading documents:", err);
      } finally {
        setLoading(false);
      }
    };

    loadDocs();
  }, [selectedCategory, selectedSubcategory]);

  /* -------------------------------
     Load External Links with Sorting
  -------------------------------- */
  useEffect(() => {
    if (selectedCategory !== "links") return;

    const loadLinks = async () => {
      try {
        setLoading(true);
        const data = await getExternalLinks();

        // ✅ Apply sorting: alphabetical by title
        const sortedLinks = sortExternalLinks(data || []);
        setExternalLinks(sortedLinks);
      } catch (err) {
        console.error("Error loading links:", err);
      } finally {
        setLoading(false);
      }
    };

    loadLinks();
  }, [selectedCategory]);

  if (!isPaidMember) {
    return (
      <div className="space-y-8" data-testid="testid-resources-locked-page">
        <div className="relative overflow-hidden rounded-2xl shadow-2xl">
          <div className="absolute inset-0">
            <img
              src={bannerImage}
              alt="Resources Banner"
              className="w-full h-full object-cover"
            />
            <div className="absolute inset-0 bg-green-900/80" />
          </div>
          <div className="relative p-10 text-white">
            <Badge className="mb-4 bg-green-600">
              <FileText className="w-3 h-3 mr-1" />
              For Paid Members Only
            </Badge>
            <h1 className="text-4xl font-bold mb-2">
              Resources Center
            </h1>
            <p>
              Official documents, government orders, and agricultural circulars
            </p>
          </div>
        </div>

        <Card className="max-w-2xl mx-auto bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800">
          <CardContent className="pt-6 text-center py-12">
            <AlertCircle className="w-16 h-16 mx-auto text-yellow-600 mb-4" />
            <h3 className="text-xl font-semibold text-yellow-900 dark:text-yellow-100 mb-2">
              Subscriber Access Only
            </h3>
            <p className="text-yellow-800 dark:text-yellow-200 mb-4">
              Access to TAGA official resources and government orders is reserved for active paid subscribers. Please complete or renew your annual subscription to view resources.
            </p>
            {onNavigateToMembership && (
              <Button onClick={onNavigateToMembership} data-testid="testid-manage-subscription-button">
                Manage Subscription
              </Button>
            )}
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="relative overflow-hidden rounded-2xl shadow-2xl">
        <div className="absolute inset-0">
          <img
            src={bannerImage}
            alt="Resources Banner"
            className="w-full h-full object-cover"
          />
          <div className="absolute inset-0 bg-green-900/80" />
        </div>
        <div className="relative p-10 text-white">
          <Badge className="mb-4 bg-green-600">
            <FileText className="w-3 h-3 mr-1" />
            Resources
          </Badge>
          <h1 className="text-4xl font-bold mb-2">
            Resources Center
          </h1>
          <p>
            Access all documents dynamically from backend
          </p>
        </div>
      </div>

      {/* MAIN - Updated layout with better sizing */}
      <div className="flex gap-6 flex-col md:flex-row">
        {/* SIDEBAR */}
        <Card className="w-full md:w-[380px] lg:w-[420px] flex-shrink-0">
          <CardContent className="p-4">
            <ScrollArea className="h-[500px]">
              {categories.map((cat) => (
                <div key={cat.id}>
                  {/* CATEGORY */}
                  <Button
                    variant={selectedCategory === cat.id ? 'default' : 'ghost'}
                    className="w-full justify-start mb-2 text-left whitespace-nowrap overflow-hidden text-ellipsis"
                    onClick={() => {
                      setSelectedCategory(cat.id);
                      setSelectedSubcategory('');
                    }}
                    data-testid={`testid-resource-category-${cat.id}-button`}
                    title={cat.name}
                  >
                    <FileText className="w-4 h-4 mr-2 flex-shrink-0" />
                    <span className="truncate">{cat.name}</span>
                  </Button>

                  {/* SUBCATEGORY ONLY FOR Scheme G.Os */}
                  {cat.id === "scheme-gos" && selectedCategory === cat.id && (
                    <div className="ml-6 space-y-1 mb-2">
                      <Button
                        size="sm"
                        variant={selectedSubcategory === "Central" ? "default" : "outline"}
                        className="w-full justify-start text-xs whitespace-nowrap"
                        onClick={() => setSelectedSubcategory("Central")}
                        data-testid="testid-resource-central-button"
                      >
                        Central
                      </Button>
                      <Button
                        size="sm"
                        variant={selectedSubcategory === "State" ? "default" : "outline"}
                        className="w-full justify-start text-xs whitespace-nowrap"
                        onClick={() => setSelectedSubcategory("State")}
                        data-testid="testid-resource-state-button"
                      >
                        State
                      </Button>
                    </div>
                  )}
                </div>
              ))}
            </ScrollArea>
          </CardContent>
        </Card>

        {/* CONTENT */}
        <Card className="flex-1 min-w-0">
          <CardContent className="p-6 space-y-4">
            {/* LOADING */}
            {loading && <p>Loading documents...</p>}

            {/* DOCUMENT LIST (or external links) - NOW SORTED */}
            {!loading && (
              <ScrollArea className="h-[500px]">
                <div className="space-y-3">
                  {/* External links when category is "links" - SORTED */}
                  {selectedCategory === "links" && externalLinks.length === 0 && (
                    <div className="text-center py-12 text-muted-foreground">
                      <ExternalLink className="w-8 h-8 mx-auto mb-2 opacity-40" />
                      <p>No external links found.</p>
                    </div>
                  )}

                  {selectedCategory === "links" && externalLinks.length > 0 && (
                    externalLinks.map((link: any, i: number) => (
                      <div
                        key={i}
                        className="flex justify-between p-3 border rounded hover:bg-accent"
                      >
                        <div className="flex items-center gap-2">
                          <ExternalLink className="w-4 h-4" />
                          <button
                            className="text-left hover:underline"
                            onClick={() => {
                              if (!isLoggedIn) {
                                alert('Please login');
                                return;
                              }
                              if (!link.url) return;
                              openLink(link.url);
                            }}
                          >
                            {link.title || link.name || link.url}
                          </button>
                        </div>
                      </div>
                    ))
                  )}

                  {/* Normal documents for other categories - SORTED */}
                  {selectedCategory !== "links" && documents.map((doc, i) => (
                    <div
                      key={i}
                      className="flex justify-between p-3 border rounded hover:bg-accent"
                    >
                      <div className="flex items-center gap-2">
                        {/* ICON */}
                        {doc.url?.startsWith("http") ? (
                          <ExternalLink className="w-4 h-4" />
                        ) : (
                          <FileText className="w-4 h-4" />
                        )}

                        {/* TITLE */}
                        <button
                          className="text-left hover:underline"
                          onClick={() => {
                            if (!isLoggedIn) {
                              alert('Please login');
                              return;
                            }
                            if (!doc.url) return;
                            const fileUrl = doc.url.startsWith("http")
                              ? doc.url
                              : getFileUrl(doc.url);
                            window.open(fileUrl, "_blank");
                          }}
                          data-testid={`testid-resource-${i}-link`}
                        >
                          {doc.title}
                        </button>
                      </div>

                      <div className="flex items-center gap-3">
                        <span className="text-sm flex items-center">
                          <Calendar className="w-3 h-3 mr-1" />
                          {doc.year}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            )}

            {/* LOGIN NOTICE */}
            {!isLoggedIn && (
              <div className="p-3 bg-yellow-100 rounded" data-testid="testid-warning-message">
                Please login to access documents
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}