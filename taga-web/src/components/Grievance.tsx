import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Textarea } from './ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { Alert, AlertDescription } from './ui/alert';
import { Badge } from './ui/badge';
import { CheckCircle, Send, Clock, MessageSquare, AlertCircle } from 'lucide-react';

import { createGrievance, getCategories, getPriorities } from '../api/Grievance';
import API_BASE_URL from '../config/api';

export function Grievance() {
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    subject: '',
    category: '',
    priority: 'medium',
    description: '',
    contactPhone: '',
    preferredResponse: 'email'
  });

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitSuccess, setSubmitSuccess] = useState(false);
 const bannerImage = `${API_BASE_URL}/images/banner-image.jpg`;
const [categories, setCategories] = useState<string[]>([]);
const [priorities, setPriorities] = useState<{ value: string; label: string }[]>([]);
const [isLoadingDropdowns, setIsLoadingDropdowns] = useState(true);
  const handleInputChange = (field: string, value: string) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);

    try {
      await createGrievance(formData);

      setSubmitSuccess(true);

       setFormData({
         name: '',
         email: '',
         subject: '',
         category: '',
         priority: '',
         description: '',
         contactPhone: '',
         preferredResponse: 'email'
       });

    } catch (error) {
      console.error("Submit error:", error);
    } finally {
      setIsSubmitting(false);
    }
  };
useEffect(() => {
  const loadData = async () => {
    try {
      const [catData, priData] = await Promise.all([
        getCategories(),
        getPriorities()
      ]);

      setCategories(catData.map((c: string) => c.trim()));
      setPriorities(priData);
    } catch (err) {
      console.error("Failed to load dropdown data", err);
      } finally {
      setIsLoadingDropdowns(false); // ✅ MUST
    }
  };

  loadData();
}, []);
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
            <MessageSquare className="w-3 h-3 mr-1" />
            Member Support
          </Badge>

          <h1 className="text-4xl font-bold mb-2">
            Submit Grievance
          </h1>

          <p>
            Submit your concerns and our team will assist you.
          </p>
        </div>
      </div>

      {/* Success */}
      {submitSuccess && (
        <Alert className="bg-green-50 border-green-200" data-testid="testid-success-message">
          <CheckCircle className="w-4 h-4 text-green-600" />
          <AlertDescription>
            Grievance submitted successfully.
          </AlertDescription>
        </Alert>
      )}

      {/* Form */}
      <Card>
        <CardHeader>
                  <CardTitle className="flex items-center gap-2">
          <MessageSquare className="w-5 h-5" />
          Submit New Grievance
        </CardTitle>
          <CardDescription> 
            Provide complete details for faster resolution.
          </CardDescription>
        </CardHeader>

        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-6" data-testid="testid-grievance-form">

            <div className="grid md:grid-cols-2 gap-6">

              <div className="space-y-2">
                <Label>Subject *</Label>
                <Input
                 placeholder="Brief description of your grievance"
                  value={formData.subject}
                  onChange={(e) => handleInputChange('subject', e.target.value)}
                  required
                  data-testid="testid-grievance-subject-input"
                />
              </div>

              {/* Category */}
              <div className="space-y-2">
<Label>Category *</Label>
<Select
  value={formData.category}
  onValueChange={(v: string) => handleInputChange('category', v)}
  disabled={isLoadingDropdowns}
>
  <SelectTrigger data-testid="testid-grievance-category-select">
    <SelectValue placeholder="Select category" />
  </SelectTrigger>

  <SelectContent>
    {categories.map((c) => (
      <SelectItem key={c} value={c}>
        {c}
      </SelectItem>
    ))}
  </SelectContent>
</Select>
              </div>

              {/* Priority */}
              <div className="space-y-2">
<Label>Priority *</Label>
<Select
  value={formData.priority}
  onValueChange={(v: string) => handleInputChange('priority', v)}
  disabled={isLoadingDropdowns}
>
  <SelectTrigger data-testid="testid-grievance-priority-select">
    <SelectValue placeholder="Select priority" />
  </SelectTrigger>

  <SelectContent>
    {priorities.map((p) => (
      <SelectItem key={p.value} value={p.value}>
        {p.label}
      </SelectItem>
    ))}
  </SelectContent>
</Select>
              </div>
              <div className="space-y-2">
                <Label>Phone</Label>
                <Input
                placeholder="Your phone number"
                  value={formData.contactPhone}
                  onChange={(e) => handleInputChange('contactPhone', e.target.value.replace(/\D/g, '').slice(0, 10))}
                  data-testid="testid-grievance-contact-phone-input"
                />
              </div>

            </div>
            <div className="space-y-2">
              <Label>Description *</Label>

              <Textarea
                placeholder="Please provide a detailed description of your grievance, including relevant dates, names, and any supporting information..."
                value={formData.description}
                onChange={(e) => handleInputChange('description', e.target.value)}
                rows={5}
                required
                data-testid="testid-grievance-description-input"
              />
              <p className="text-sm text-muted-foreground">
                Minimum 50 characters required. Be specific and include all relevant details.
              </p>
            </div>
            <div className="space-y-2">
              <Label>Preferred Response</Label>
              <Select defaultValue="email" onValueChange={(v: string) => handleInputChange('preferredResponse', v)}>
                <SelectTrigger data-testid="testid-grievance-preferred-response-select">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="email">Email</SelectItem>
                  <SelectItem value="phone">Phone</SelectItem>
                  <SelectItem value="letter">Letter</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <Alert>
              <AlertCircle className="w-4 h-4" />
              <AlertDescription>
                Your grievance will be reviewed and addressed promptly.
              </AlertDescription>
            </Alert>

            <Button type="submit" className="w-full" disabled={isSubmitting} data-testid="testid-submit-grievance-button">
              {isSubmitting ? (
                <>
                  <Clock className="w-4 h-4 mr-2 animate-spin" />
                  Submitting...
                </>
              ) : (
                <>
                  <Send className="w-4 h-4 mr-2" />
                  Submit Grievance
                </>
              )}
            </Button>

          </form>
        </CardContent>
      </Card>
    </div>
  );
}