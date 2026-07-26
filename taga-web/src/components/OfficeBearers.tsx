import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Badge } from './ui/badge';
import { Separator } from './ui/separator';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table';
import { Phone, MapPin, Calendar, Users, Award, GraduationCap, Briefcase, UserCircle } from 'lucide-react';
import API_BASE from "../config/api";
export function OfficeBearers() {


  const [selectedDistrict, setSelectedDistrict] = useState('');

  // 🔥 API STATES (same variable names — IMPORTANT)
  const [stateExecutiveCommittee, setStateExecutiveCommittee] = useState<any[]>([]);
  const [districts, setDistricts] = useState<string[]>([]);
  const [districtOfficeBearers, setDistrictOfficeBearers] = useState<{ [key: string]: any[] }>({});
  const bannerImage = `${API_BASE}/images/banner-image.jpg`;
  // ✅ STATE EXECUTIVE
  useEffect(() => {
    fetch(`${API_BASE}/office-bearers/state-executive`)
      .then(res => res.json())
      .then(data => {
        const mapped = data.map((item: any) => {
          // Build image URL - only if backend provides one
          let imageUrl = null;
          if (item.image) {
            if (item.image.startsWith('http')) {
              // Already a full URL
              imageUrl = item.image;
            } else if (item.image.startsWith('/api/')) {
              imageUrl = `${API_BASE}${item.image.replace(/^\/api/, '')}`;
            } else if (item.image.startsWith('/')) {
              imageUrl = `${API_BASE}${item.image}`;
            } else {
              imageUrl = `${API_BASE}/images/${item.image.replace(/^\/+/, '')}`;
            }
          }

          return {
            name: item.name,
            position: item.designation,
            department: '',
            location: item.location,
            tenure: '2023-2025',
            email: '',
            phone: item.phone,
            experience: item.experience,
            education: item.qualification,
            image: imageUrl
          };
        });

        setStateExecutiveCommittee(mapped);
      });
  }, []);

  // ✅ DISTRICTS
  useEffect(() => {
    fetch(`${API_BASE}/office-bearers/districts`)
      .then(res => res.json())
      .then(data => setDistricts(data || []));
  }, []);

  // ✅ DISTRICT BEARERS (IMPORTANT FIX)
  useEffect(() => {
    if (!selectedDistrict) return;

    fetch(`${API_BASE}/office-bearers/district-office-bearers?district=${selectedDistrict}`)
      .then(res => res.json())
      .then(data => {
        const finalData = data[selectedDistrict] || data;

        setDistrictOfficeBearers(prev => ({
          ...prev,
          [selectedDistrict]: finalData
        }));
      });

  }, [selectedDistrict]);

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
            Term: 2025-2027
          </Badge>

          <h1 className="text-4xl font-bold mb-2">
            TAGA Leadership
          </h1>

          <p>
            Meet our dedicated team of agriculture graduates leading Tamil Nadu's agricultural transformation and professional development
          </p>
        </div>
      </div>

      {/* State Executive Committee */}
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-3xl font-bold text-foreground flex items-center space-x-3">
              <div className="p-2 bg-primary/10 rounded-lg">
                <Users className="w-6 h-6 text-primary" />
              </div>
              <span> State Office Bearers </span>
            </h2>
            <p className="text-muted-foreground mt-2">
              The core leadership team elected by association members across Tamil Nadu
            </p>
          </div>
        </div>

        <div className="grid md:grid-cols-2 gap-6">
          {stateExecutiveCommittee.map((member, index) => (
            <Card
              key={index}
              className="overflow-hidden border-2 hover:shadow-2xl transition-all duration-300 hover:border-primary/50 group"
            >
              <div className="flex items-stretch">

                <div className="relative bg-gradient-to-br from-primary/10 to-primary/20 p-4 flex items-center justify-center">
                  <div className="relative">
                    {member.image ? (
                      <>
                        <img
                          src={member.image}
                          alt={member.name}
                          className="w-32 h-40 object-cover rounded-lg border-4 border-white shadow-lg"
                        />
                        <div className="absolute -bottom-2 -right-2 bg-primary text-primary-foreground rounded-full p-2 shadow-lg">
                          <Award className="w-4 h-4" />
                        </div>
                      </>
                    ) : (
                      <div className="flex flex-col items-center justify-center p-6 bg-muted rounded-lg border-4 border-white shadow-lg text-center">
                        <UserCircle className="w-16 h-20 text-muted-foreground mb-2" />
                        <p className="text-sm font-medium text-foreground">{member.name}</p>
                      </div>
                    )}
                  </div>
                </div>

                <div className="flex-1 p-4 bg-gradient-to-br from-background to-secondary/10">
                  <div className="space-y-3">
                    <div>
                      <h3 className="text-lg font-bold text-foreground">{member.name}</h3>
                      <Badge className="bg-primary mt-1">{member.position}</Badge>
                    </div>

                    <Separator />

                    <div className="space-y-2 text-sm">
                      <div className="flex items-start space-x-2">
                        <GraduationCap className="w-4 h-4 text-muted-foreground mt-0.5 flex-shrink-0" />
                        <div>
                          <p className="font-medium text-foreground">Qualification</p>
                          <p className="text-muted-foreground">{member.education}</p>
                        </div>
                      </div>

                      <div className="flex items-start space-x-2">
                        <Briefcase className="w-4 h-4 text-muted-foreground mt-0.5 flex-shrink-0" />
                        <div>
                          <p className="font-medium text-foreground">Experience</p>
                          <p className="text-muted-foreground">{member.experience}</p>
                        </div>
                      </div>

                      <div className="flex items-start space-x-2">
                        <Phone className="w-4 h-4 text-muted-foreground mt-0.5 flex-shrink-0" />
                        <div>
                          <p className="font-medium text-foreground">Phone</p>
                          <p className="text-muted-foreground">{member.phone}</p>
                        </div>
                      </div>

                      <div className="flex items-start space-x-2">
                        <MapPin className="w-4 h-4 text-muted-foreground mt-0.5 flex-shrink-0" />
                        <div>
                          <p className="font-medium text-foreground">Location</p>
                          <p className="text-muted-foreground">{member.location}</p>
                        </div>
                      </div>
                    </div>

                  </div>
                </div>

              </div>
            </Card>
          ))}
        </div>
      </div>

      {/* District UI COMPLETELY SAME */}
      {/* (no change below) */}

      <div className="space-y-6">
        <div>
          <h2 className="text-3xl font-bold text-foreground flex items-center space-x-3 mb-4">
            <div className="p-2 bg-primary/10 rounded-lg">
              <MapPin className="w-6 h-6 text-primary" />
            </div>
            <span>District Office Bearers</span>
          </h2>
          <p className="text-muted-foreground">
            Select a district to view the office bearers serving that region
          </p>
        </div>

        <Card className="bg-gradient-to-br from-primary/5 via-transparent to-accent/10 border-2 border-primary/20 shadow-lg">
          <CardHeader>
            <CardTitle className="text-xl">Select District</CardTitle>
            <CardDescription>
              Choose from any of the 36 districts in Tamil Nadu to view their office bearers
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Select value={selectedDistrict} onValueChange={setSelectedDistrict}>
              <SelectTrigger className="w-full max-w-md h-12 text-lg border-2" data-testid="testid-office-bearers-district-select">
                <SelectValue placeholder="Select a district..." />
              </SelectTrigger>
              <SelectContent>
                {districts.map((district) => (
                  <SelectItem key={district} value={district} className="text-base">
                    {district}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </CardContent>
        </Card>

        {selectedDistrict && (
          <Card className="border-2 shadow-lg">
            <CardHeader className="bg-gradient-to-r from-primary/5 to-primary/10">
              <div className="flex items-center space-x-3">
                <div className="p-2 bg-primary/10 rounded-lg">
                  <MapPin className="w-5 h-5 text-primary" />
                </div>
                <div>
                  <CardTitle className="text-2xl">{selectedDistrict} District Office Bearers</CardTitle>
                  <CardDescription className="text-base mt-1">
                    Office bearers serving the agricultural community in {selectedDistrict} district
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="pt-6">
              {districtOfficeBearers[selectedDistrict] ? (
                <div className="rounded-lg border overflow-hidden">
                  <Table className="w-full table-fixed">
                    <TableHeader>
                      <TableRow className="bg-muted/50">
                        <TableHead className="w-1/3 font-bold text-left">Name</TableHead>
                        <TableHead className="w-1/3 font-bold text-left">Title</TableHead>
                        <TableHead className="w-1/3 font-bold text-left">Contact</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {districtOfficeBearers[selectedDistrict].map((bearer, index) => (
                        <TableRow key={index} className="hover:bg-primary/5 transition-colors">
                          <TableCell className="font-semibold text-left align-middle">{bearer.name}</TableCell>
                          <TableCell className="text-left align-middle">
                            <Badge variant="outline" className="text-primary border-primary/30 font-medium">
                              {bearer.title}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-left align-middle">
                            <a href={`tel:${bearer.contact}`} className="inline-flex items-center space-x-2 text-sm hover:text-primary transition-colors group">
                              <Phone className="w-4 h-4 text-muted-foreground group-hover:text-primary" />
                              <span>{bearer.contact}</span>
                            </a>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              ) : (
                <div className="text-center py-12 bg-muted/20 rounded-lg">
                  <div className="inline-flex items-center justify-center p-4 bg-primary/10 rounded-full mb-4">
                    <Users className="w-12 h-12 text-primary" />
                  </div>
                  <h3 className="text-lg font-semibold text-foreground mb-2">
                    Data Coming Soon
                  </h3>
                  <p className="text-muted-foreground mb-1">
                    Office bearers data for {selectedDistrict} district will be updated soon.
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}