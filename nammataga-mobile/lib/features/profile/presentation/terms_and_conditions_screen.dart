import 'package:flutter/material.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import 'widgets/profile_subpage_scaffold.dart';

/// Full-screen Terms and Conditions page based on the official Association document.
class TermsAndConditionsScreen extends StatelessWidget {
  const TermsAndConditionsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return ProfileSubpageScaffold(
      title: 'Terms & Conditions',
      watermarkIcon: LucideIcons.fileCheck2,
      showEditNotice: false,
      body: SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(16, 16, 16, 32),
        physics: const BouncingScrollPhysics(),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Metadata header card
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(18),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(18),
                border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
                boxShadow: const [
                  BoxShadow(
                    color: Color(0x061B4332),
                    blurRadius: 10,
                    offset: Offset(0, 3),
                  ),
                ],
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(8),
                        decoration: BoxDecoration(
                          color: const Color(0xFFE8F5EE),
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: const Icon(
                          LucideIcons.shieldAlert,
                          color: AppColors.primary,
                          size: 20,
                        ),
                      ),
                      const SizedBox(width: 12),
                      const Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Legal Terms of Use',
                              style: TextStyle(
                                fontFamily: 'Roboto',
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                                color: AppColors.textPrimary,
                              ),
                            ),
                            SizedBox(height: 2),
                            Text(
                              'Last Updated: 14 November 2026',
                              style: TextStyle(
                                fontFamily: 'Roboto',
                                fontSize: 12,
                                color: AppColors.textSecondary,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 14),
                  const Text(
                    'Welcome to the TAGA mobile application ("App"). The App is operated by the Tamil Nadu Government Agricultural Graduates Association (TAGA) ("we," "us," or "our"), an association for Tamil Nadu Government Agriculture Department Officers.\n\nBy downloading, accessing, or using the App, you agree to be bound by these Terms and Conditions ("Terms"). If you do not agree to these Terms, please do not use the App.',
                    style: TextStyle(
                      fontFamily: 'Roboto',
                      fontSize: 13,
                      color: Color(0xFF334155),
                      height: 1.55,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 16),

            // Section 1
            _buildSection(
              number: '1',
              title: 'Eligibility and Official Use',
              children: [
                _buildParagraph(
                  '1.1. The App is intended exclusively for the members and office bearers of the Tamil Nadu Government Agricultural Graduates Association (TAGA).',
                ),
                _buildParagraph(
                  '1.2. By using this App, you represent that you are a current or retired officer of the Tamil Nadu Government Agriculture Department or an authorized representative of the Association.',
                ),
                _buildParagraph(
                  '1.3. Accounts are created and administered solely by the Association administrators. Members do not self-register. The accuracy of the profile information is maintained by the Association based on records provided by the member.',
                ),
              ],
            ),

            // Section 2
            _buildSection(
              number: '2',
              title: 'License to Use',
              children: [
                _buildParagraph(
                  '2.1. Subject to your compliance with these Terms, we grant you a limited, non-exclusive, non-transferable, revocable license to download, install, and use the App on a mobile device that you own or control, solely for your personal and professional association-related purposes.',
                ),
                _buildParagraph(
                  '2.2. You may not:',
                ),
                _buildBulletPoint('Modify, reverse engineer, or create derivative works of the App.'),
                _buildBulletPoint('Use the App for any commercial purpose or for any purpose that is unlawful or prohibited by these Terms.'),
                _buildBulletPoint('Attempt to gain unauthorized access to the App\'s servers or databases.'),
                _buildBulletPoint('Copy, reproduce, or redistribute content published on the App without the Association\'s written permission.'),
              ],
            ),

            // Section 3
            _buildSection(
              number: '3',
              title: 'Nature of the App and User Conduct',
              children: [
                _buildParagraph(
                  '3.1. The App is an information and communication portal. Members can view announcements, event schedules, annual fair information, and other content published by the Association.',
                ),
                _buildParagraph(
                  '3.2. Content publication is restricted to Association administrators. Members cannot create, post, upload, or share content through the App.',
                ),
                _buildParagraph(
                  '3.3. Members may raise queries or requests regarding their profile or association matters through the query feature provided in the App. You agree to use this feature responsibly and not to submit false, abusive, or misleading queries.',
                ),
                _buildParagraph(
                  '3.4. You agree not to misuse the App, interfere with its operation, or attempt to access accounts or data belonging to other members.',
                ),
              ],
            ),

            // Section 4
            _buildSection(
              number: '4',
              title: 'Account Management',
              children: [
                _buildParagraph(
                  '4.1. Accounts are created, modified, and deleted solely by the Association administrators.',
                ),
                _buildParagraph(
                  '4.2. If your membership ceases, or if the Association determines that your account is no longer required, the Association may deactivate or delete your account without prior notice.',
                ),
                _buildParagraph(
                  '4.3. If you wish to update your profile details or request deletion of your account, you may raise a query through the App or contact the Association at nammataga@gmail.com. Such requests will be processed at the discretion of the administrators in accordance with association records and applicable requirements.',
                ),
              ],
            ),

            // Section 5
            _buildSection(
              number: '5',
              title: 'Payments and Razorpay Integration',
              children: [
                _buildParagraph(
                  '5.1. Payment Processing. The App uses Razorpay as a third-party payment gateway to process payments made by members for association events, the annual fair, membership dues, or other Association-related activities. All payment transactions are processed securely through Razorpay\'s systems. We do not store your full card number, UPI credentials, or netbanking passwords on our servers.',
                ),
                _buildParagraph(
                  '5.2. Merchant of Record. The Tamil Nadu Government Agricultural Graduates Association (TAGA) is the merchant of record for all transactions processed through the App. Razorpay acts solely as a payment processor. Your payment is made to the Association, not to Razorpay.',
                ),
                _buildParagraph(
                  '5.3. Payment Authorization. By initiating a payment, you authorize the Association and Razorpay to charge the specified amount using your chosen payment method. You confirm that you are authorized to use the payment instrument provided.',
                ),
                _buildParagraph(
                  '5.4. Refunds and Cancellations. Refunds for event registrations, annual fair fees, or other payments are subject to the specific refund policy communicated at the time of payment. Unless otherwise stated, payments for association events may be non-refundable. If you believe you are entitled to a refund, please raise a query through the App or contact the Association. Refunds, where approved, will be processed through Razorpay to the original payment method and may take 5–7 business days to reflect.',
                ),
                _buildParagraph(
                  '5.5. Failed Transactions. If a payment fails but funds are debited from your account, the amount is typically auto-reversed by your bank or Razorpay within 5–7 business days. If the issue persists, please contact the Association with the transaction reference number.',
                ),
                _buildParagraph(
                  '5.6. Chargebacks. If you initiate a chargeback or dispute with your bank or card issuer, the Association reserves the right to suspend your access to paid events or services pending resolution, and to recover any chargeback fees imposed by Razorpay or the bank.',
                ),
                _buildParagraph(
                  '5.7. Razorpay Terms. Your use of Razorpay\'s payment services is also subject to Razorpay\'s own terms and conditions and privacy policy. We encourage you to review them at razorpay.com.',
                ),
              ],
            ),

            // Section 6
            _buildSection(
              number: '6',
              title: 'Events and Annual Fair',
              children: [
                _buildParagraph(
                  '6.1. The App facilitates the dissemination of information relating to the Association\'s annual fair and other events.',
                ),
                _buildParagraph(
                  '6.2. While we strive to keep event details accurate, we reserve the right to modify the schedule, venue, or nature of any event. We are not liable for any inconvenience caused by such changes.',
                ),
                _buildParagraph(
                  '6.3. Any payments made in connection with events or membership dues are subject to the policies stated at the time of the transaction and Section 5 above.',
                ),
              ],
            ),

            // Section 7
            _buildSection(
              number: '7',
              title: 'Intellectual Property',
              children: [
                _buildParagraph(
                  '7.1. The App, including its content, features, and functionality, is owned by the Tamil Nadu Government Agricultural Graduates Association (TAGA) and is protected by copyright, trademark, and other intellectual property laws.',
                ),
                _buildParagraph(
                  '7.2. The Association\'s name, logo, and all related names, logos, product and service names, designs, and slogans are trademarks of the Association. You must not use such marks without the prior written permission of the Association.',
                ),
              ],
            ),

            // Section 8
            _buildSection(
              number: '8',
              title: 'Termination',
              children: [
                _buildParagraph(
                  '8.1. We may terminate or suspend your access to the App immediately, without prior notice or liability, for any reason whatsoever, including without limitation if you breach these Terms or cease to be a member of the Association.',
                ),
                _buildParagraph(
                  '8.2. Upon termination, your right to use the App will cease immediately.',
                ),
              ],
            ),

            // Section 9
            _buildSection(
              number: '9',
              title: 'Disclaimer of Warranties',
              children: [
                _buildParagraph(
                  '9.1. The App is provided on an "AS IS" and "AS AVAILABLE" basis. The Association makes no warranties, expressed or implied, regarding the App\'s operation or the information, content, or materials included therein.',
                ),
                _buildParagraph(
                  '9.2. We do not warrant that the App will be uninterrupted, error-free, or free of viruses or other harmful components. We do not control and are not responsible for the performance or availability of Razorpay\'s payment systems.',
                ),
              ],
            ),

            // Section 10
            _buildSection(
              number: '10',
              title: 'Limitation of Liability',
              children: [
                _buildParagraph(
                  '10.1. To the maximum extent permitted by applicable law, the Tamil Nadu Government Agricultural Graduates Association (TAGA) shall not be liable for any indirect, incidental, special, consequential, or punitive damages, including without limitation, loss of profits, data, use, goodwill, or other intangible losses, resulting from your access to or use of, or inability to access or use, the App.',
                ),
                _buildParagraph(
                  '10.2. Without limiting the foregoing, the Association shall not be liable for any failure or delay in processing payments caused by Razorpay, your bank, or any other third-party payment system.',
                ),
              ],
            ),

            // Section 11
            _buildSection(
              number: '11',
              title: 'Governing Law',
              children: [
                _buildParagraph(
                  '11.1. These Terms shall be governed and construed in accordance with the laws of India, specifically the state of Tamil Nadu, without regard to its conflict of law provisions.',
                ),
                _buildParagraph(
                  '11.2. Any disputes arising under these Terms shall be subject to the exclusive jurisdiction of the courts located in Chennai, Tamil Nadu.',
                ),
              ],
            ),

            // Section 12
            _buildSection(
              number: '12',
              title: 'Changes to Terms',
              children: [
                _buildParagraph(
                  '12.1. We reserve the right, at our sole discretion, to modify or replace these Terms at any time. We will provide notice of any material changes by posting the new Terms on the App. By continuing to access or use the App after those revisions become effective, you agree to be bound by the revised terms.',
                ),
              ],
            ),

            // Section 13: Contact Us
            _buildSection(
              number: '13',
              title: 'Contact Us',
              children: [
                _buildParagraph(
                  'If you have any questions about these Terms, please contact us at:',
                ),
                const SizedBox(height: 6),
                _buildContactTile(
                  icon: LucideIcons.mail,
                  title: 'Email',
                  value: 'nammataga@gmail.com',
                ),
                const SizedBox(height: 8),
                _buildContactTile(
                  icon: LucideIcons.mapPin,
                  title: 'Address',
                  value: 'Velanmai Illam, 45, Sait Colony, II Street, Egmore, Chennai - 600008, Tamil Nadu, India',
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSection({
    required String number,
    required String title,
    required List<Widget> children,
  }) {
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.only(bottom: 14),
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
        boxShadow: const [
          BoxShadow(
            color: Color(0x061B4332),
            blurRadius: 8,
            offset: Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                width: 26,
                height: 26,
                decoration: const BoxDecoration(
                  shape: BoxShape.circle,
                  color: Color(0xFFE8F5EE),
                ),
                alignment: Alignment.center,
                child: Text(
                  number,
                  style: const TextStyle(
                    fontFamily: 'Roboto',
                    fontSize: 12.5,
                    fontWeight: FontWeight.bold,
                    color: AppColors.primary,
                  ),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  title,
                  style: const TextStyle(
                    fontFamily: 'Roboto',
                    fontSize: 15,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          ...children,
        ],
      ),
    );
  }

  Widget _buildParagraph(String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Text(
        text,
        style: const TextStyle(
          fontFamily: 'Roboto',
          fontSize: 13,
          color: Color(0xFF475569),
          height: 1.52,
        ),
      ),
    );
  }

  Widget _buildBulletPoint(String text) {
    return Padding(
      padding: const EdgeInsets.only(left: 8, bottom: 6),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Padding(
            padding: EdgeInsets.only(top: 6),
            child: Icon(
              Icons.circle,
              size: 5,
              color: AppColors.primary,
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              text,
              style: const TextStyle(
                fontFamily: 'Roboto',
                fontSize: 13,
                color: Color(0xFF475569),
                height: 1.48,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildContactTile({
    required IconData icon,
    required String title,
    required String value,
  }) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFF7FAF8),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFFE5EFE9)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, size: 16, color: AppColors.primary),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: const TextStyle(
                    fontFamily: 'Roboto',
                    fontSize: 11.5,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textSecondary,
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  value,
                  style: const TextStyle(
                    fontFamily: 'Roboto',
                    fontSize: 13,
                    color: AppColors.textPrimary,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
