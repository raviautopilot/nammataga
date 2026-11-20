import React from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from './ui/dialog';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import {
  ShieldCheck,
  Lock,
  CreditCard,
  Server,
  UserCheck,
  Bell,
  Scale,
  Calendar,
  Mail,
  Phone,
  MapPin,
  ExternalLink,
  CheckCircle2,
  FileText,
  AlertCircle,
  Building,
} from 'lucide-react';

interface PrivacyPolicyModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function PrivacyPolicyModal({ open, onOpenChange }: PrivacyPolicyModalProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="sm:max-w-3xl max-h-[90vh] flex flex-col p-0 overflow-hidden bg-background border-border shadow-2xl"
        data-testid="testid-privacy-policy-modal"
      >
        {/* Modal Header */}
        <DialogHeader className="px-6 pt-6 pb-4 border-b border-border/80 bg-muted/30">
          <div className="flex items-start gap-4">
            <div className="w-12 h-12 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center text-primary shrink-0 shadow-sm">
              <ShieldCheck className="w-6 h-6" />
            </div>
            <div className="space-y-1">
              <div className="flex items-center gap-2 flex-wrap">
                <DialogTitle className="text-xl sm:text-2xl font-bold tracking-tight text-foreground">
                  Privacy Policy
                </DialogTitle>
                <Badge variant="outline" className="text-xs bg-primary/5 text-primary border-primary/30">
                  DPDP Act, 2023 Compliant
                </Badge>
                <Badge variant="secondary" className="text-xs">
                  Google Play
                </Badge>
              </div>
              <DialogDescription className="text-xs sm:text-sm text-muted-foreground flex items-center gap-2 flex-wrap">
                <span>Tamil Nadu Government Agricultural Graduates Association (TAGA)</span>
                <span>•</span>
                <span className="inline-flex items-center gap-1 font-medium text-foreground/80">
                  <Calendar className="w-3.5 h-3.5" />
                  Last Updated: 14 November 2026
                </span>
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        {/* Scrollable Policy Body */}
        <div className="overflow-y-auto px-6 py-5 space-y-6 text-sm text-muted-foreground leading-relaxed flex-1">
          {/* Official Preamble & Scope */}
          <div className="rounded-xl p-4 bg-emerald-500/5 border border-emerald-500/20 text-foreground space-y-2">
            <p>
              The <strong>Tamil Nadu Government Agricultural Graduates Association (TAGA)</strong> (&ldquo;we,&rdquo; &ldquo;us,&rdquo; or &ldquo;our&rdquo;) operates the TAGA mobile application (the &ldquo;App&rdquo;). This Privacy Policy informs you of our policies regarding the collection, use, and disclosure of personal data when you use our App and the choices you have associated with that data.
            </p>
            <p className="text-xs text-muted-foreground">
              We are committed to protecting the privacy of our members and complying with the <strong>Digital Personal Data Protection Act, 2023 (DPDP Act)</strong> of India and the requirements of Google Play.
            </p>
          </div>

          {/* 1. Information We Collect */}
          <section className="space-y-4">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <FileText className="w-4 h-4 text-primary" />
              <h3>1. Information We Collect</h3>
            </div>

            {/* 1.1 Personal Data */}
            <div className="space-y-2 rounded-lg bg-card p-3.5 border border-border/60">
              <h4 className="font-medium text-foreground flex items-center gap-2 text-sm">
                <span className="w-1.5 h-1.5 rounded-full bg-primary" />
                1.1. Personal Data (Maintained by the Association)
              </h4>
              <ul className="space-y-1.5 pl-4 list-disc marker:text-primary/70 text-xs sm:text-sm">
                <li>
                  <strong className="text-foreground">Identity &amp; Contact Data:</strong> Name, official designation, department branch, email address, phone number, and postal address.
                </li>
                <li>
                  <strong className="text-foreground">Professional Data:</strong> Details regarding your role in the association, years of service, and participation in association events.
                </li>
                <li>
                  <strong className="text-foreground">Query Data:</strong> The content of queries you raise through the App regarding your profile or association matters.
                </li>
              </ul>
              <div className="mt-2 text-xs bg-muted/60 text-muted-foreground rounded p-2 italic flex items-center gap-1.5">
                <UserCheck className="w-3.5 h-3.5 text-primary shrink-0" />
                <span>This information is entered and maintained by Association administrators; members do not create or edit their own profiles directly.</span>
              </div>
            </div>

            {/* 1.2 Payment Data */}
            <div className="space-y-2 rounded-lg bg-card p-3.5 border border-border/60">
              <h4 className="font-medium text-foreground flex items-center gap-2 text-sm">
                <CreditCard className="w-4 h-4 text-primary" />
                1.2. Payment Data (Processed via Razorpay)
              </h4>
              <p className="text-xs sm:text-sm">
                When you make a payment through the App (e.g., for the annual fair or association events), the transaction is processed by <strong>Razorpay</strong>. We do not collect or store your full payment card number, CVV, UPI PIN, or netbanking credentials on our servers. Razorpay collects and processes your payment information (including card details, UPI ID, or bank account information) directly. We receive only a confirmation of payment status, transaction ID, and the amount paid.
              </p>
              <p className="text-xs sm:text-sm">
                Razorpay may also collect certain device and usage data as part of its SDK, including device identifiers and app interaction data on screens rendered by Razorpay, as described in Razorpay&apos;s own privacy policy.
              </p>
              <div className="pt-1">
                <a
                  href="https://razorpay.com/privacy/"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1 text-xs text-primary hover:underline font-medium"
                >
                  View Razorpay Privacy Policy <ExternalLink className="w-3 h-3" />
                </a>
              </div>
            </div>

            {/* 1.3 Usage Data */}
            <div className="space-y-2 rounded-lg bg-card p-3.5 border border-border/60">
              <h4 className="font-medium text-foreground flex items-center gap-2 text-sm">
                <Server className="w-4 h-4 text-primary" />
                1.3. Usage Data (Collected Automatically)
              </h4>
              <ul className="space-y-1.5 pl-4 list-disc marker:text-primary/70 text-xs sm:text-sm">
                <li>
                  <strong className="text-foreground">Device Information:</strong> Device type and operating system version.
                </li>
                <li>
                  <strong className="text-foreground">Usage Data:</strong> Interactions within the App (e.g., pages viewed, features used).
                </li>
                <li>
                  <strong className="text-foreground">Log Data:</strong> Server logs which may include your IP address and timestamps, used for security and troubleshooting.
                </li>
              </ul>
            </div>

            {/* 1.4 Permissions */}
            <div className="space-y-2 rounded-lg bg-card p-3.5 border border-border/60">
              <h4 className="font-medium text-foreground flex items-center gap-2 text-sm">
                <Bell className="w-4 h-4 text-primary" />
                1.4. Permissions
              </h4>
              <p className="text-xs sm:text-sm">
                The App <strong>does not</strong> request access to your camera, photos, or storage. The App may request permission to send notifications, solely to inform you about association events, fairs, and announcements. You may disable notifications at any time through your device settings.
              </p>
            </div>
          </section>

          {/* 2. How We Use Your Information */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <CheckCircle2 className="w-4 h-4 text-primary" />
              <h3>2. How We Use Your Information</h3>
            </div>
            <p className="text-xs sm:text-sm">We use the collected data for the following purposes:</p>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
              <div className="p-3 rounded-lg bg-muted/40 border border-border/50 text-xs sm:text-sm">
                <strong className="text-foreground block mb-0.5">To Provide and Maintain the App</strong>
                Ensuring the App functions correctly and displays relevant information to members.
              </div>
              <div className="p-3 rounded-lg bg-muted/40 border border-border/50 text-xs sm:text-sm">
                <strong className="text-foreground block mb-0.5">Association Communication</strong>
                To send you notices regarding meetings, the annual fair, and other official association activities.
              </div>
              <div className="p-3 rounded-lg bg-muted/40 border border-border/50 text-xs sm:text-sm">
                <strong className="text-foreground block mb-0.5">Event Management</strong>
                To share information and coordinate logistics for the annual fair.
              </div>
              <div className="p-3 rounded-lg bg-muted/40 border border-border/50 text-xs sm:text-sm">
                <strong className="text-foreground block mb-0.5">Payment Processing</strong>
                To facilitate and confirm payments made through Razorpay for events, fairs, or dues.
              </div>
              <div className="p-3 rounded-lg bg-muted/40 border border-border/50 text-xs sm:text-sm">
                <strong className="text-foreground block mb-0.5">Query Resolution</strong>
                To review and respond to queries raised by members regarding their profiles, payments, or association matters.
              </div>
              <div className="p-3 rounded-lg bg-muted/40 border border-border/50 text-xs sm:text-sm">
                <strong className="text-foreground block mb-0.5">Improvement &amp; Compliance</strong>
                To analyze usage trends and improve experience, and comply with applicable laws and regulations.
              </div>
            </div>
          </section>

          {/* 3. Data Sharing and Disclosure */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <Scale className="w-4 h-4 text-primary" />
              <h3>3. Data Sharing and Disclosure</h3>
            </div>
            <div className="rounded-lg bg-primary/5 border border-primary/20 p-2.5 text-xs text-foreground font-medium flex items-center gap-2">
              <ShieldCheck className="w-4 h-4 text-primary shrink-0" />
              <span>We do not sell your personal data. We may share your information only in the specific situations outlined below:</span>
            </div>
            <ul className="space-y-2 pl-4 list-disc marker:text-primary/70 text-xs sm:text-sm">
              <li>
                <strong className="text-foreground">Within the Association:</strong> Your name and official designation may be visible to other members for networking purposes, as determined by the Association.
              </li>
              <li>
                <strong className="text-foreground">Razorpay (Payment Processor):</strong> When you initiate a payment, your payment details and certain contact information (such as name, email, and phone number if prefilled) are shared with Razorpay to process the transaction. Razorpay is a PCI DSS compliant payment processor and handles your data in accordance with its own privacy policy and terms. We encourage you to review Razorpay&apos;s privacy policy at{' '}
                <a href="https://razorpay.com/privacy" target="_blank" rel="noopener noreferrer" className="text-primary hover:underline">
                  razorpay.com/privacy
                </a>.
              </li>
              <li>
                <strong className="text-foreground">Service Providers:</strong> We may employ other third-party companies (e.g., cloud hosting providers, notification service providers) to facilitate our App and perform App-related services. These third parties have access to your Personal Data only to perform these tasks on our behalf and are obligated not to disclose or use it for any other purpose.
              </li>
              <li>
                <strong className="text-foreground">Legal Requirements:</strong> We may disclose your Personal Data if required to do so by law or in response to valid requests by public authorities (e.g., a court or a government agency).
              </li>
            </ul>
          </section>

          {/* 4. Data Retention and Deletion */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <Lock className="w-4 h-4 text-primary" />
              <h3>4. Data Retention and Deletion</h3>
            </div>
            <div className="space-y-2 text-xs sm:text-sm">
              <p>
                <strong>4.1. Retention:</strong> We will retain your Personal Data only for as long as is necessary for the purposes set out in this Privacy Policy and for the duration of your membership with the Association. Payment transaction records may be retained for a longer period as required by applicable tax and accounting laws.
              </p>
              <p>
                <strong>4.2. Account Deletion:</strong> Accounts are deleted by Association administrators. If you wish to request deletion of your account and associated personal data, you may raise a query through the App or contact us at{' '}
                <a href="mailto:nammataga@gmail.com" className="text-primary font-medium hover:underline">
                  nammataga@gmail.com
                </a>{' '}
                /{' '}
                <a href="mailto:tagastage@gmail.com" className="text-primary font-medium hover:underline">
                  tagastage@gmail.com
                </a>. Requests will be reviewed and actioned by the administrators in accordance with association records and applicable retention requirements. Note that payment transaction records held by Razorpay are subject to Razorpay&apos;s own retention policies and cannot be deleted by the Association directly.
              </p>
              <p>
                <strong>4.3. Razorpay Data:</strong> To exercise data deletion rights with Razorpay, please refer to Razorpay&apos;s privacy policy and contact Razorpay directly.
              </p>
            </div>
          </section>

          {/* 5. Data Security */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <ShieldCheck className="w-4 h-4 text-primary" />
              <h3>5. Data Security</h3>
            </div>
            <p className="text-xs sm:text-sm">
              The security of your data is important to us. We implement reasonable security safeguards, including encryption and access controls, to protect your Personal Data against unauthorized access, alteration, disclosure, or destruction. Payment transactions are processed by Razorpay using industry-standard encryption (including TLS and tokenization) and Razorpay is PCI DSS compliant. However, no method of transmission over the Internet or method of electronic storage is 100% secure.
            </p>
          </section>

          {/* 6. Your Rights under the DPDP Act */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <UserCheck className="w-4 h-4 text-primary" />
              <h3>6. Your Rights under the DPDP Act</h3>
            </div>
            <p className="text-xs sm:text-sm">As a Data Principal under the DPDP Act, you have the right to:</p>
            <ul className="space-y-1.5 pl-4 list-disc marker:text-primary/70 text-xs sm:text-sm">
              <li><strong className="text-foreground">Access:</strong> Obtain information about the personal data we process.</li>
              <li><strong className="text-foreground">Correction:</strong> Request correction of inaccurate or incomplete data (via the query feature or by contacting the Association).</li>
              <li><strong className="text-foreground">Erasure:</strong> Request deletion of your data (subject to administrator approval and legal retention requirements).</li>
              <li><strong className="text-foreground">Grievance Redressal:</strong> Lodge a complaint regarding our data processing practices.</li>
            </ul>

            {/* Grievance Officer Box */}
            <div className="rounded-xl border border-primary/20 bg-primary/5 p-4 space-y-2.5">
              <div className="flex items-center gap-2 text-foreground font-medium text-xs sm:text-sm">
                <Scale className="w-4 h-4 text-primary" />
                <span>To exercise these rights, please contact our Grievance Officer:</span>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 pt-1 text-xs">
                <div className="bg-background/80 p-2.5 rounded-lg border border-border/50">
                  <span className="text-muted-foreground block text-[11px]">Officer Name</span>
                  <span className="font-semibold text-foreground">Joint Secretary (Legal)</span>
                </div>
                <div className="bg-background/80 p-2.5 rounded-lg border border-border/50">
                  <span className="text-muted-foreground block text-[11px]">Email</span>
                  <a href="mailto:tagastate@gmail.com" className="font-semibold text-primary hover:underline">
                    tagastate@gmail.com
                  </a>
                </div>
                <div className="bg-background/80 p-2.5 rounded-lg border border-border/50">
                  <span className="text-muted-foreground block text-[11px]">Phone</span>
                  <a href="tel:04434919949" className="font-semibold text-foreground hover:text-primary">
                    044 34919949
                  </a>
                </div>
              </div>
              <p className="text-[11px] text-muted-foreground italic">
                For rights specifically relating to payment data held by Razorpay, please contact Razorpay directly.
              </p>
            </div>
          </section>

          {/* 7. Children's Privacy */}
          <section className="space-y-2">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <AlertCircle className="w-4 h-4 text-primary" />
              <h3>7. Children&apos;s Privacy</h3>
            </div>
            <p className="text-xs sm:text-sm">
              The App is intended for use by adults (Government Officers). We do not knowingly collect personally identifiable information from anyone under the age of 18.
            </p>
          </section>

          {/* 8. Changes to This Privacy Policy */}
          <section className="space-y-2">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <Calendar className="w-4 h-4 text-primary" />
              <h3>8. Changes to This Privacy Policy</h3>
            </div>
            <p className="text-xs sm:text-sm">
              We may update our Privacy Policy from time to time. We will notify you of any changes by posting the new Privacy Policy on this page and updating the &ldquo;Last Updated&rdquo; date. You are advised to review this Privacy Policy periodically for any changes.
            </p>
          </section>

          {/* 9. Contact Us */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <Building className="w-4 h-4 text-primary" />
              <h3>9. Contact Us</h3>
            </div>
            <p className="text-xs sm:text-sm">If you have any questions about this Privacy Policy, please contact us:</p>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs sm:text-sm">
              <div className="p-3.5 rounded-lg bg-card border border-border/60 flex items-start gap-3">
                <Mail className="w-4 h-4 text-primary mt-0.5 shrink-0" />
                <div>
                  <strong className="text-foreground block text-xs">Official Email Addresses</strong>
                  <div className="mt-1 space-y-0.5">
                    <div>
                      <a href="mailto:nammataga@gmail.com" className="text-primary hover:underline">
                        nammataga@gmail.com
                      </a>
                    </div>
                    <div>
                      <a href="mailto:tagastage@gmail.com" className="text-primary hover:underline">
                        tagastage@gmail.com
                      </a>
                    </div>
                  </div>
                </div>
              </div>

              <div className="p-3.5 rounded-lg bg-card border border-border/60 flex items-start gap-3">
                <MapPin className="w-4 h-4 text-primary mt-0.5 shrink-0" />
                <div>
                  <strong className="text-foreground block text-xs">Headquarters Address</strong>
                  <p className="mt-1 text-muted-foreground text-xs leading-relaxed">
                    Velanmai Illam, 45, Sait Colony, II Street, Egmore, Chennai - 600008, Tamil Nadu, India
                  </p>
                </div>
              </div>
            </div>
          </section>
        </div>

        {/* Modal Footer */}
        <DialogFooter className="px-6 py-3.5 border-t border-border/80 bg-muted/20 flex flex-row items-center justify-between sm:justify-between">
          <span className="text-[11px] text-muted-foreground hidden sm:inline">
            Official TAGA Legal Document • DPDP Act 2023
          </span>
          <Button
            variant="default"
            size="sm"
            onClick={() => onOpenChange(false)}
            data-testid="testid-privacy-policy-close-button"
            className="px-5 font-medium ml-auto"
          >
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
