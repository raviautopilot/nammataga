import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Alert, AlertDescription } from './ui/alert';
import { Badge } from './ui/badge';
import { Shield, Eye, EyeOff, AlertTriangle, Lock } from 'lucide-react';
import { adminLogin } from '../api/adminContent';

interface AdminLoginProps {
  onLogin: () => void;
}

export function AdminLogin({ onLogin }: AdminLoginProps) {
  const [formData, setFormData] = useState({
    username: '',
    password: ''
  });
  const [showPassword, setShowPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');
    setSuccess('');

    try {
      const response = await adminLogin({
        username: formData.username,
        password: formData.password
      });
      
      if (response.token) {
        setSuccess('Login successful! Redirecting...');
        setTimeout(() => {
          onLogin();
        }, 1000);
      }
    } catch (err: any) {
      console.error('Login error:', err);
      setError(err.message || 'Invalid credentials. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleInputChange = (field: string, value: string) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
    setError('');
    setSuccess('');
  };

  return (
    <div className="max-w-md mx-auto space-y-6">
      {/* Header */}
      <div className="text-center space-y-2">
        <div className="flex items-center justify-center space-x-2">
          <Shield className="w-8 h-8 text-primary" />
          <h1 className="text-3xl font-bold text-foreground">Admin Portal</h1>
        </div>
        <p className="text-muted-foreground">
          Secure access for association administrators
        </p>
        <Badge variant="secondary" className="bg-yellow-100 text-yellow-800">
          Restricted Access
        </Badge>
      </div>

      {/* Security Notice */}
      <Alert>
        <AlertTriangle className="w-4 h-4" />
        <AlertDescription>
          This portal is for authorized administrators only. All access attempts are logged and monitored.
        </AlertDescription>
      </Alert>

      {/* Login Form */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Lock className="w-5 h-5" />
            <span>Administrator Login</span>
          </CardTitle>
          <CardDescription>
            Enter your administrator credentials
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4" data-testid="testid-admin-login-form">
            {error && (
              <Alert variant="destructive" data-testid="testid-error-message">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
            {success && (
              <Alert className="border-green-500 bg-green-50">
                <AlertDescription className="text-green-700">{success}</AlertDescription>
              </Alert>
            )}

            <div className="space-y-2">
              <Label htmlFor="username">Administrator Email</Label>
              <Input
                id="username"
                type="email"
                placeholder="Enter your admin email"
                value={formData.username}
                onChange={(e) => handleInputChange('username', e.target.value)}
                required
                data-testid="testid-admin-login-username-input"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <div className="relative">
                <Input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  placeholder="Enter your password"
                  value={formData.password}
                  onChange={(e) => handleInputChange('password', e.target.value)}
                  required
                  data-testid="testid-admin-login-password-input"
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="absolute right-0 top-0 h-full px-3 hover:bg-transparent"
                  onClick={() => setShowPassword(!showPassword)}
                  data-testid="testid-admin-login-password-visibility-button"
                >
                  {showPassword ? (
                    <EyeOff className="w-4 h-4 text-muted-foreground" />
                  ) : (
                    <Eye className="w-4 h-4 text-muted-foreground" />
                  )}
                </Button>
              </div>
            </div>

            <Button
              type="submit"
              className="w-full"
              disabled={isLoading}
              data-testid="testid-admin-login-submit-button"
            >
              {isLoading ? 'Verifying...' : 'Login'}
            </Button>
          </form>
        </CardContent>
      </Card>

      {/* Admin Features Preview */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Administrator Features</CardTitle>
          <CardDescription>Available after successful login</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            <div className="flex items-center space-x-2">
              <div className="w-2 h-2 bg-primary rounded-full" />
              <span className="text-sm text-muted-foreground">Manage member accounts and subscriptions</span>
            </div>
            <div className="flex items-center space-x-2">
              <div className="w-2 h-2 bg-primary rounded-full" />
              <span className="text-sm text-muted-foreground">Review and respond to grievances</span>
            </div>
            <div className="flex items-center space-x-2">
              <div className="w-2 h-2 bg-primary rounded-full" />
              <span className="text-sm text-muted-foreground">Publish resources and announcements</span>
            </div>
            <div className="flex items-center space-x-2">
              <div className="w-2 h-2 bg-primary rounded-full" />
              <span className="text-sm text-muted-foreground">Generate reports and analytics</span>
            </div>
            <div className="flex items-center space-x-2">
              <div className="w-2 h-2 bg-primary rounded-full" />
              <span className="text-sm text-muted-foreground">System configuration and settings</span>
            </div>
          </div>
        </CardContent>
      </Card>

      
    </div>
  );
}