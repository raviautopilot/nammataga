import React, { useEffect, useState } from 'react';
import {
  Card, CardContent, CardDescription, CardHeader, CardTitle
} from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Alert, AlertDescription } from './ui/alert';
import { Separator } from './ui/separator';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger
} from './ui/dialog';
import {
  Select, SelectContent, SelectItem,
  SelectTrigger, SelectValue
} from './ui/select';
import { toast } from 'sonner';
import { Eye, EyeOff, Lock } from 'lucide-react';
import { memberLoginAPI } from '../api/member';

// API for membership and forgot password (keep these)
import {
  forgotPassword,
  applyMembership,
  getDistricts,
  changePassword 
} from '../api/memberLoginApi';

interface MemberLoginProps {
  onLogin: (isPaid?: boolean) => void;
}

export function MemberLogin({ onLogin }: MemberLoginProps) {

  // ✅ STATES
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    rememberMe: false
  });

const [membershipForm, setMembershipForm] = useState({
  firstName: '',
  lastName: '',
  email: '',
  phone: '',
  dateOfBirth: '',
  gender: '',
  address: '',
  district: '',
  qualification: '',
  graduationYear: '',
  currentEmployment: '',
  organization: '',
  experience: '',
  specialization: '',
  references: ''
});

  const [forgotPasswordForm, setForgotPasswordForm] = useState({
    membershipId: '',
    email: '',
    securityQuestion: '',
    securityAnswer: ''
  });

  const [districts, setDistricts] = useState<string[]>([]);

  const [showPassword, setShowPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const [showMembershipDialog, setShowMembershipDialog] = useState(false);
  const [showForgotPasswordDialog, setShowForgotPasswordDialog] = useState(false);
  const [showChangePasswordDialog, setShowChangePasswordDialog] = useState(false);
  const [showOldPassword, setShowOldPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [changePasswordForm, setChangePasswordForm] = useState({
  email: '',
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
});
  // ✅ HANDLERS (RESTORED PROPERLY)
  const handleInputChange = (field: string, value: string | boolean) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    setError('');
  };

  const handleMembershipFormChange = (field: string, value: string) => {
    setMembershipForm(prev => ({ ...prev, [field]: value }));
  };

  const handleForgotPasswordFormChange = (field: string, value: string) => {
    setForgotPasswordForm(prev => ({ ...prev, [field]: value }));
  };
  const handleChangePasswordFormChange = (field: string, value: string) => {
  setChangePasswordForm(prev => ({ ...prev, [field]: value }));
};


  // ✅ FETCH DISTRICTS (NO HARDCODE)
useEffect(() => {
  const fetchDistricts = async () => {
    try {
      const data = await getDistricts(); // ✅ already array

      console.log("District API Response:", data); // ✅ correct

      setDistricts(data); // ✅ correct

    } catch (err) {
      console.error("Failed to fetch districts", err);
    }
  };

  fetchDistricts();
}, []);
  // ✅ LOGIN
const handleSubmit = async (e: React.FormEvent) => {
  e.preventDefault();

  setIsLoading(true);
  setError('');

  try {
    const res = await memberLoginAPI({
      email: formData.email,
      password: formData.password
    });

    // FIRST LOGIN → FORCE PASSWORD CHANGE
    if (res.forceChangePassword) {

      toast.info(
        "Temporary password detected. Change password before login."
      );

      setChangePasswordForm({
        email: res.email || formData.email,
        oldPassword: formData.password,
        newPassword: '',
        confirmPassword: ''
      });

      setShowChangePasswordDialog(true);
setIsLoading(false);
      return;
    }

    //toast.success(res.message || "Login successful");

    onLogin(true);

  } catch (err: any) {

    setError(
      err.response?.data?.error ||
      err.response?.data?.message ||
      "Login failed"
    );

    toast.error(
      err.response?.data?.error ||
      err.response?.data?.message ||
      "Login failed"
    );

  } finally {
    setIsLoading(false);
  }
};
  // ✅ MEMBERSHIP
  const handleMembershipSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      await applyMembership(membershipForm);
      toast.success("Membership applied successfully!");
      setShowMembershipDialog(false);

      // reset form
setMembershipForm({
  firstName: '',
  lastName: '',
  email: '',
  phone: '',
  dateOfBirth: '',
  gender: '',
  address: '',
  district: '',
  qualification: '',
  graduationYear: '',
  currentEmployment: '',
  organization: '',
  experience: '',
  specialization: '',
  references: ''
});

    } catch (err: any) {
      toast.error(err.response?.data?.error || "Submission failed");
    }
  };
const handleChangePasswordSubmit = async (
  e: React.FormEvent
) => {

  e.preventDefault();

  if (
    changePasswordForm.newPassword !==
    changePasswordForm.confirmPassword
  ) {
    toast.error("Passwords do not match");
    return;
  }

  try {

    const res = await changePassword(
      changePasswordForm
    );

    toast.success(
      res.message ||
      "Password changed successfully. Please login again."
    );

    setShowChangePasswordDialog(false);

    // clear login form
    setFormData({
      email: '',
      password: '',
      rememberMe: false
    });

    // clear password form
    setChangePasswordForm({
      email: '',
      oldPassword: '',
      newPassword: '',
      confirmPassword: ''
    });

  } catch (err: any) {

    toast.error(
      err.response?.data?.error ||
      "Failed to change password"
    );

  }
};
// ✅ FORGOT PASSWORD
const handleForgotPasswordSubmit = async (
  e: React.FormEvent
) => {

  e.preventDefault();

  try {

    const res = await forgotPassword(
      forgotPasswordForm
    );

    toast.success(
      res.message ||
      "Reset successful"
    );

    setShowForgotPasswordDialog(false);

    setForgotPasswordForm({
      membershipId: '',
      email: '',
      securityQuestion: '',
      securityAnswer: ''
    });

  } catch (err: any) {

    toast.error(
      err.response?.data?.error ||
      "Failed to reset password"
    );

  }
};
  return (
    <div className="max-w-md mx-auto space-y-6">

      {/* LOGIN */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Lock className="w-5 h-5" />
            Sign In
          </CardTitle>
          <CardDescription>Enter your credentials</CardDescription>
        </CardHeader>

        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4" data-testid="testid-login-form">

            {error && (
              <Alert variant="destructive" data-testid="testid-error-message">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            <div>
              <Label>Email</Label>
              <Input
                value={formData.email}
                onChange={(e) =>
                  handleInputChange('email', e.target.value)
                }
                required
                data-testid="testid-login-identifier-input"
              />
            </div>

            <div>
              <Label>Password</Label>
              <div className="relative">
                <Input
                  type={showPassword ? 'text' : 'password'}
                  value={formData.password}
                  onChange={(e) =>
                    handleInputChange('password', e.target.value)
                  }
                  required
                  data-testid="testid-login-password-input"
                />
                <Button
                  type="button"
                  variant="ghost"
                  className="absolute right-0 top-0"
                  onClick={() => setShowPassword(!showPassword)}
                  data-testid="testid-login-password-visibility-button"
                >
                  {showPassword ? <EyeOff /> : <Eye />}
                </Button>
              </div>
            </div>

            <Button className="w-full" disabled={isLoading} data-testid="testid-login-submit-button">
              {isLoading ? 'Signing In...' : 'Sign In'}
            </Button>
          </form>

          <Separator className="my-4" />

          {/* FORGOT PASSWORD */}
          <Dialog open={showForgotPasswordDialog} onOpenChange={setShowForgotPasswordDialog}>
            <DialogTrigger asChild>
              <Button variant="link" data-testid="testid-forgot-password-button">Forgot Password?</Button>
            </DialogTrigger>

            <DialogContent data-testid="testid-forgot-password-modal">
              <DialogHeader>
                <DialogTitle>Reset Password</DialogTitle>
              </DialogHeader>

              <form onSubmit={handleForgotPasswordSubmit} className="space-y-3" data-testid="testid-forgot-password-form">

                <Input placeholder="Membership ID"
                  value={forgotPasswordForm.membershipId}
                  onChange={(e) =>
                    handleForgotPasswordFormChange('membershipId', e.target.value)
                  }
                  data-testid="testid-forgot-password-membership-id-input"
                />

                <Input placeholder="Email"
                  value={forgotPasswordForm.email}
                  onChange={(e) =>
                    handleForgotPasswordFormChange('email', e.target.value)
                  }
                  data-testid="testid-forgot-password-email-input"
                />

                <Input placeholder="Security Question"
                  value={forgotPasswordForm.securityQuestion}
                  onChange={(e) =>
                    handleForgotPasswordFormChange('securityQuestion', e.target.value)
                  }
                  data-testid="testid-forgot-password-security-question-input"
                />

                <Input placeholder="Security Answer"
                  value={forgotPasswordForm.securityAnswer}
                  onChange={(e) =>
                    handleForgotPasswordFormChange('securityAnswer', e.target.value)
                  }
                  data-testid="testid-forgot-password-security-answer-input"
                />

                <Button type="submit" data-testid="testid-reset-button">Reset</Button>
              </form>
            </DialogContent>
          </Dialog>
<Button
variant="link"
onClick={() => {

setChangePasswordForm({
email: '',
oldPassword: '',
newPassword: '',
confirmPassword: ''
});

setShowChangePasswordDialog(true);

}}
data-testid="testid-change-password-button"
>
Change Password
</Button>
<Dialog open={showChangePasswordDialog} onOpenChange={setShowChangePasswordDialog}>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Change Password</DialogTitle>
    </DialogHeader>

    <form onSubmit={handleChangePasswordSubmit} className="space-y-3">
<Input
  placeholder="Email"
  data-testid="testid-change-password-email-input"
  value={changePasswordForm.email}
  onChange={(e) =>
    handleChangePasswordFormChange("email", e.target.value)
  }
/>

{/* OLD PASSWORD */}
<div className="relative">
<Input
  type={showOldPassword ? "text" : "password"}
  placeholder="Temporary Password"
  data-testid="testid-change-password-old-input"
  value={changePasswordForm.oldPassword}
  onChange={(e) =>
    handleChangePasswordFormChange("oldPassword", e.target.value)
  }
  className="pr-10"
/>

  <Button
    type="button"
    variant="ghost"
    className="absolute right-0 top-0"
    onClick={() => setShowOldPassword(!showOldPassword)}
  >
    {showOldPassword ? <EyeOff /> : <Eye />}
  </Button>
</div>

{/* NEW PASSWORD */}
<div className="relative">
  <Input
    type={showNewPassword ? "text" : "password"}
    placeholder="New Password"
    data-testid="testid-change-password-new-input"
    value={changePasswordForm.newPassword}
    onChange={(e) =>
      handleChangePasswordFormChange('newPassword', e.target.value)
    }
    className="pr-10"
  />

  <Button
    type="button"
    variant="ghost"
    className="absolute right-0 top-0"
    onClick={() => setShowNewPassword(!showNewPassword)}
  >
    {showNewPassword ? <EyeOff /> : <Eye />}
  </Button>
</div>

{/* CONFIRM PASSWORD */}
<div className="relative">

<Input
type={
showConfirmPassword
? "text"
: "password"
}
placeholder="Confirm Password"
data-testid="testid-change-password-confirm-input"
value={changePasswordForm.confirmPassword}
onChange={(e) =>
handleChangePasswordFormChange(
'confirmPassword',
e.target.value
)
}
className="pr-10"
/>

<Button
type="button"
variant="ghost"
className="absolute right-0 top-0"
onClick={() =>
setShowConfirmPassword(
!showConfirmPassword
)
}
>
{
showConfirmPassword
? <EyeOff />
: <Eye />
}
</Button>

</div>

<Button type="submit" data-testid="testid-change-password-submit-button">
  Submit
</Button>
    </form>
  </DialogContent>
</Dialog>


        </CardContent>
      </Card>

    </div>
  );
}