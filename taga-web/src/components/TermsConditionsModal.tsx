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
  FileText,
  UserCheck,
  ShieldCheck,
  CreditCard,
  Calendar,
  Building,
  Mail,
  MapPin,
  ExternalLink,
  Scale,
  AlertTriangle,
  FileCheck2,
  Lock,
  Layers,
  XCircle,
  HelpCircle,
  Gavel,
} from 'lucide-react';

interface TermsConditionsModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function TermsConditionsModal({ open, onOpenChange }: TermsConditionsModalProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="sm:max-w-3xl max-h-[90vh] flex flex-col p-0 overflow-hidden bg-background border-border shadow-2xl"
        data-testid="testid-terms-conditions-modal"
      >
        {/* Modal Header */}
        <DialogHeader className="px-6 pt-6 pb-4 border-b border-border/80 bg-muted/30">
          <div className="flex items-start gap-4">
            <div className="w-12 h-12 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center text-primary shrink-0 shadow-sm">
              <Scale className="w-6 h-6" />
            </div>
            <div className="space-y-1">
              <div className="flex items-center gap-2 flex-wrap">
                <DialogTitle className="text-xl sm:text-2xl font-bold tracking-tight text-foreground">
                  Terms &amp; Conditions
                </DialogTitle>
                <Badge variant="outline" className="text-xs bg-primary/5 text-primary border-primary/30">
                  Legal Agreement
                </Badge>
                <Badge variant="secondary" className="text-xs">
                  Official Association Terms
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

        {/* Scrollable Terms Body */}
        <div className="overflow-y-auto px-6 py-5 space-y-6 text-sm text-muted-foreground leading-relaxed flex-1">
          {/* Welcome / Preamble */}
          <div className="rounded-xl p-4 bg-emerald-500/5 border border-emerald-500/20 text-foreground space-y-2">
            <p>
              Welcome to the <strong>TAGA</strong> mobile application (&ldquo;App&rdquo;). The App is operated by the <strong>Tamil Nadu Government Agricultural Graduates Association (TAGA)</strong> (&ldquo;we,&rdquo; &ldquo;us,&rdquo; or &ldquo;our&rdquo;), an association for Tamil Nadu Government Agriculture Department Officers.
            </p>
            <p className="text-xs text-muted-foreground">
              By downloading, accessing, or using the App, you agree to be bound by these Terms and Conditions (&ldquo;Terms&rdquo;). If you do not agree to these Terms, please do not use the App.
            </p>
          </div>

          {/* 1. Eligibility and Official Use */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <UserCheck className="w-4 h-4 text-primary" />
              <h3>1. Eligibility and Official Use</h3>
            </div>
            <div className="space-y-2 text-xs sm:text-sm">
              <p>
                <strong>1.1.</strong> The App is intended exclusively for the members and office bearers of the Tamil Nadu Government Agricultural Graduates Association (TAGA).
              </p>
              <p>
                <strong>1.2.</strong> By using this App, you represent that you are a current or retired officer of the Tamil Nadu Government Agriculture Department or an authorized representative of the Association.
              </p>
              <p>
                <strong>1.3.</strong> Accounts are created and administered solely by the Association administrators. Members do not self-register. The accuracy of the profile information is maintained by the Association based on records provided by the member.
              </p>
            </div>
          </section>

          {/* 2. License to Use */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <FileCheck2 className="w-4 h-4 text-primary" />
              <h3>2. License to Use</h3>
            </div>
            <div className="space-y-2 text-xs sm:text-sm">
              <p>
                <strong>2.1.</strong> Subject to your compliance with these Terms, we grant you a limited, non-exclusive, non-transferable, revocable license to download, install, and use the App on a mobile device that you own or control, solely for your personal and professional association-related purposes.
              </p>
              <div className="rounded-lg bg-card p-3 border border-border/60 space-y-1.5">
                <span className="font-medium text-foreground block"><strong>2.2. You may not:</strong></span>
                <ul className="pl-4 list-disc marker:text-destructive/70 space-y-1 text-xs sm:text-sm">
                  <li>Modify, reverse engineer, or create derivative works of the App.</li>
                  <li>Use the App for any commercial purpose or for any purpose that is unlawful or prohibited by these Terms.</li>
                  <li>Attempt to gain unauthorized access to the App&apos;s servers or databases.</li>
                  <li>Copy, reproduce, or redistribute content published on the App without the Association&apos;s written permission.</li>
                </ul>
              </div>
            </div>
          </section>

          {/* 3. Nature of the App and User Conduct */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <Layers className="w-4 h-4 text-primary" />
              <h3>3. Nature of the App and User Conduct</h3>
            </div>
            <div className="space-y-2 text-xs sm:text-sm">
              <p>
                <strong>3.1.</strong> The App is an information and communication portal. Members can view announcements, event schedules, annual fair information, and other content published by the Association.
              </p>
              <p>
                <strong>3.2.</strong> Content publication is restricted to Association administrators. Members cannot create, post, upload, or share content through the App.
              </p>
              <p>
                <strong>3.3.</strong> Members may raise queries or requests regarding their profile or association matters through the query feature provided in the App. You agree to use this feature responsibly and not to submit false, abusive, or misleading queries.
              </p>
              <p>
                <strong>3.4.</strong> You agree not to misuse the App, interfere with its operation, or attempt to access accounts or data belonging to other members.
              </p>
            </div>
          </section>

          {/* 4. Account Management */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <Lock className="w-4 h-4 text-primary" />
              <h3>4. Account Management</h3>
            </div>
            <div className="space-y-2 text-xs sm:text-sm">
              <p>
                <strong>4.1.</strong> Accounts are created, modified, and deleted solely by the Association administrators.
              </p>
              <p>
                <strong>4.2.</strong> If your membership ceases, or if the Association determines that your account is no longer required, the Association may deactivate or delete your account without prior notice.
              </p>
              <div className="rounded-lg bg-card p-3 border border-border/60">
                <p>
                  <strong>4.3.</strong> If you wish to update your profile details or request deletion of your account, you may raise a query through the App or contact the Association at{' '}
                  <a href="mailto:nammataga@gmail.com" className="text-primary font-medium hover:underline">
                    nammataga@gmail.com
                  </a>. Such requests will be processed at the discretion of the administrators in accordance with association records and applicable requirements.
                </p>
              </div>
            </div>
          </section>

          {/* 5. Payments and Razorpay Integration */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <CreditCard className="w-4 h-4 text-primary" />
              <h3>5. Payments and Razorpay Integration</h3>
            </div>
            <div className="space-y-2.5 text-xs sm:text-sm">
              <div className="p-3 rounded-lg bg-card border border-border/60 space-y-1">
                <strong className="text-foreground block">5.1. Payment Processing</strong>
                <p>
                  The App uses Razorpay as a third-party payment gateway to process payments made by members for association events, the annual fair, membership dues, or other Association-related activities. All payment transactions are processed securely through Razorpay&apos;s systems. We do not store your full card number, UPI credentials, or netbanking passwords on our servers.
                </p>
              </div>

              <div className="p-3 rounded-lg bg-card border border-border/60 space-y-1">
                <strong className="text-foreground block">5.2. Merchant of Record</strong>
                <p>
                  The Tamil Nadu Government Agricultural Graduates Association (TAGA) is the merchant of record for all transactions processed through the App. Razorpay acts solely as a payment processor. Your payment is made to the Association, not to Razorpay.
                </p>
              </div>

              <div className="p-3 rounded-lg bg-card border border-border/60 space-y-1">
                <strong className="text-foreground block">5.3. Payment Authorization</strong>
                <p>
                  By initiating a payment, you authorize the Association and Razorpay to charge the specified amount using your chosen payment method. You confirm that you are authorized to use the payment instrument provided.
                </p>
              </div>

              <div className="p-3 rounded-lg bg-card border border-border/60 space-y-1">
                <strong className="text-foreground block">5.4. Refunds and Cancellations</strong>
                <p>
                  Refunds for event registrations, annual fair fees, or other payments are subject to the specific refund policy communicated at the time of payment. Unless otherwise stated, payments for association events may be non-refundable. If you believe you are entitled to a refund, please raise a query through the App or contact the Association. Refunds, where approved, will be processed through Razorpay to the original payment method and may take 5–7 business days to reflect.
                </p>
              </div>

              <div className="p-3 rounded-lg bg-card border border-border/60 space-y-1">
                <strong className="text-foreground block">5.5. Failed Transactions</strong>
                <p>
                  If a payment fails but funds are debited from your account, the amount is typically auto-reversed by your bank or Razorpay within 5–7 business days. If the issue persists, please contact the Association with the transaction reference number.
                </p>
              </div>

              <div className="p-3 rounded-lg bg-card border border-border/60 space-y-1">
                <strong className="text-foreground block">5.6. Chargebacks</strong>
                <p>
                  If you initiate a chargeback or dispute with your bank or card issuer, the Association reserves the right to suspend your access to paid events or services pending resolution, and to recover any chargeback fees imposed by Razorpay or the bank.
                </p>
              </div>

              <div className="p-3 rounded-lg bg-card border border-border/60 space-y-1">
                <strong className="text-foreground block">5.7. Razorpay Terms</strong>
                <p>
                  Your use of Razorpay&apos;s payment services is also subject to Razorpay&apos;s own terms and conditions and privacy policy. We encourage you to review them at{' '}
                  <a href="https://razorpay.com" target="_blank" rel="noopener noreferrer" className="text-primary hover:underline inline-flex items-center gap-0.5">
                    razorpay.com <ExternalLink className="w-3 h-3" />
                  </a>.
                </p>
              </div>
            </div>
          </section>

          {/* 6. Events and Annual Fair */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <Calendar className="w-4 h-4 text-primary" />
              <h3>6. Events and Annual Fair</h3>
            </div>
            <div className="space-y-2 text-xs sm:text-sm">
              <p>
                <strong>6.1.</strong> The App facilitates the dissemination of information relating to the Association&apos;s annual fair and other events.
              </p>
              <p>
                <strong>6.2.</strong> While we strive to keep event details accurate, we reserve the right to modify the schedule, venue, or nature of any event. We are not liable for any inconvenience caused by such changes.
              </p>
              <p>
                <strong>6.3.</strong> Any payments made in connection with events or membership dues are subject to the policies stated at the time of the transaction and Section 5 above.
              </p>
            </div>
          </section>

          {/* 7. Intellectual Property */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <ShieldCheck className="w-4 h-4 text-primary" />
              <h3>7. Intellectual Property</h3>
            </div>
            <div className="space-y-2 text-xs sm:text-sm">
              <p>
                <strong>7.1.</strong> The App, including its content, features, and functionality, is owned by the Tamil Nadu Government Agricultural Graduates Association (TAGA) and is protected by copyright, trademark, and other intellectual property laws.
              </p>
              <p>
                <strong>7.2.</strong> The Association&apos;s name, logo, and all related names, logos, product and service names, designs, and slogans are trademarks of the Association. You must not use such marks without the prior written permission of the Association.
              </p>
            </div>
          </section>

          {/* 8. Termination */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <XCircle className="w-4 h-4 text-primary" />
              <h3>8. Termination</h3>
            </div>
            <div className="space-y-2 text-xs sm:text-sm">
              <p>
                <strong>8.1.</strong> We may terminate or suspend your access to the App immediately, without prior notice or liability, for any reason whatsoever, including without limitation if you breach these Terms or cease to be a member of the Association.
              </p>
              <p>
                <strong>8.2.</strong> Upon termination, your right to use the App will cease immediately.
              </p>
            </div>
          </section>

          {/* 9. Disclaimer of Warranties */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <AlertTriangle className="w-4 h-4 text-primary" />
              <h3>9. Disclaimer of Warranties</h3>
            </div>
            <div className="space-y-2 text-xs sm:text-sm">
              <p>
                <strong>9.1.</strong> The App is provided on an &ldquo;AS IS&rdquo; and &ldquo;AS AVAILABLE&rdquo; basis. The Association makes no warranties, expressed or implied, regarding the App&apos;s operation or the information, content, or materials included therein.
              </p>
              <p>
                <strong>9.2.</strong> We do not warrant that the App will be uninterrupted, error-free, or free of viruses or other harmful components. We do not control and are not responsible for the performance or availability of Razorpay&apos;s payment systems.
              </p>
            </div>
          </section>

          {/* 10. Limitation of Liability */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <HelpCircle className="w-4 h-4 text-primary" />
              <h3>10. Limitation of Liability</h3>
            </div>
            <div className="space-y-2 text-xs sm:text-sm">
              <p>
                <strong>10.1.</strong> To the maximum extent permitted by applicable law, the Tamil Nadu Government Agricultural Graduates Association (TAGA) shall not be liable for any indirect, incidental, special, consequential, or punitive damages, including without limitation, loss of profits, data, use, goodwill, or other intangible losses, resulting from your access to or use of, or inability to access or use, the App.
              </p>
              <p>
                <strong>10.2.</strong> Without limiting the foregoing, the Association shall not be liable for any failure or delay in processing payments caused by Razorpay, your bank, or any other third-party payment system.
              </p>
            </div>
          </section>

          {/* 11. Governing Law */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <Gavel className="w-4 h-4 text-primary" />
              <h3>11. Governing Law</h3>
            </div>
            <div className="space-y-2 text-xs sm:text-sm">
              <p>
                <strong>11.1.</strong> These Terms shall be governed and construed in accordance with the laws of India, specifically the state of Tamil Nadu, without regard to its conflict of law provisions.
              </p>
              <p>
                <strong>11.2.</strong> Any disputes arising under these Terms shall be subject to the exclusive jurisdiction of the courts located in Chennai, Tamil Nadu.
              </p>
            </div>
          </section>

          {/* 12. Changes to Terms */}
          <section className="space-y-2">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <Calendar className="w-4 h-4 text-primary" />
              <h3>12. Changes to Terms</h3>
            </div>
            <p className="text-xs sm:text-sm">
              <strong>12.1.</strong> We reserve the right, at our sole discretion, to modify or replace these Terms at any time. We will provide notice of any material changes by posting the new Terms on the App. By continuing to access or use the App after those revisions become effective, you agree to be bound by the revised terms.
            </p>
          </section>

          {/* 13. Contact Us */}
          <section className="space-y-3">
            <div className="flex items-center gap-2 text-foreground font-semibold text-base border-b border-border/60 pb-2">
              <Building className="w-4 h-4 text-primary" />
              <h3>13. Contact Us</h3>
            </div>
            <p className="text-xs sm:text-sm">If you have any questions about these Terms, please contact us at:</p>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs sm:text-sm">
              <div className="p-3.5 rounded-lg bg-card border border-border/60 flex items-start gap-3">
                <Mail className="w-4 h-4 text-primary mt-0.5 shrink-0" />
                <div>
                  <strong className="text-foreground block text-xs">Official Email</strong>
                  <div className="mt-1">
                    <a href="mailto:nammataga@gmail.com" className="text-primary hover:underline font-medium">
                      nammataga@gmail.com
                    </a>
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
            Official TAGA Legal Document • Terms of Use
          </span>
          <Button
            variant="default"
            size="sm"
            onClick={() => onOpenChange(false)}
            data-testid="testid-terms-conditions-close-button"
            className="px-5 font-medium ml-auto"
          >
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
