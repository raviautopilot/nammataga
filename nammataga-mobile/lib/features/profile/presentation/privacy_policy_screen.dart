import 'package:flutter/material.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import 'widgets/profile_subpage_scaffold.dart';

/// Full-screen Privacy Policy page based on the official Association document.
class PrivacyPolicyScreen extends StatelessWidget {
  const PrivacyPolicyScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return ProfileSubpageScaffold(
      title: 'Privacy Policy',
      watermarkIcon: LucideIcons.shieldCheck,
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
                          LucideIcons.lock,
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
                              'Data Privacy & DPDP Compliance',
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
                    'The Tamil Nadu Government Agricultural Graduates Association (TAGA) ("we," "us," or "our") operates the TAGA mobile application (the "App"). This Privacy Policy explains how we collect, use, disclose, and safeguard your information when you use our App.\n\nWe are committed to protecting the privacy of our members and complying with the Digital Personal Data Protection Act, 2023 (DPDP Act) and other applicable Indian laws.',
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
              title: 'Information We Collect',
              children: [
                _buildSubheading('1.1. Personal Data (Maintained by the Association):'),
                _buildBulletPoint('Identity & Contact Data: Name, official designation, department branch, email address, phone number, and postal address.'),
                _buildBulletPoint('Professional Data: Details regarding your role in the association, years of service, and participation in association events or the annual fair.'),
                _buildBulletPoint('Query Data: The content of queries you raise through the App regarding your profile or association matters.'),
                _buildParagraph(
                  'This information is entered and maintained by Association administrators; members do not create or edit their own profile data directly.',
                ),
                const SizedBox(height: 8),
                _buildSubheading('1.2. Payment Data (Processed via Razorpay):'),
                _buildParagraph(
                  'When you make a payment through the App (e.g., for the annual fair or association events), the transaction is processed by Razorpay. We do not collect or store your full payment card number, CVV, UPI PIN, or netbanking credentials on our servers. Razorpay collects and processes your payment information (including card details, UPI ID, or bank account information) directly. We receive only a confirmation of payment status, transaction ID, and the amount paid.',
                ),
                _buildParagraph(
                  'Razorpay may also collect certain device and usage data as part of its SDK, including device identifiers and app interaction data, for fraud prevention and transaction verification, in accordance with Razorpay\'s privacy policy.',
                ),
                const SizedBox(height: 8),
                _buildSubheading('1.3. Usage Data (Collected Automatically):'),
                _buildBulletPoint('Device Information: Device type and operating system version.'),
                _buildBulletPoint('Usage Data: Interactions within the App (e.g., pages viewed, features used).'),
                _buildBulletPoint('Log Data: Server logs which may include your IP address and timestamps, used for security and troubleshooting.'),
                const SizedBox(height: 8),
                _buildSubheading('1.4. Permissions:'),
                _buildParagraph(
                  'The App does not request access to your camera, photos, or storage. The App may request permission to send notifications, solely to inform you about association events, fairs, and announcements. You may disable notifications at any time through your device settings.',
                ),
              ],
            ),

            // Section 2
            _buildSection(
              number: '2',
              title: 'How We Use Your Information',
              children: [
                _buildParagraph('We use the collected data for the following purposes:'),
                _buildBulletPoint('To Provide and Maintain the App: Ensuring the App functions correctly and displays relevant information to members.'),
                _buildBulletPoint('Association Communication: To send you notices regarding meetings, the annual fair, and other official association activities.'),
                _buildBulletPoint('Event Management: To share information and coordinate logistics for the annual fair.'),
                _buildBulletPoint('Payment Processing: To facilitate and confirm payments made through Razorpay for events, fairs, or dues.'),
                _buildBulletPoint('Query Resolution: To review and respond to queries raised by members regarding their profiles, payments, or association matters.'),
                _buildBulletPoint('Improvement: To analyze usage trends and improve the App\'s functionality and user experience.'),
                _buildBulletPoint('Legal Compliance: To comply with applicable laws and regulations.'),
              ],
            ),

            // Section 3
            _buildSection(
              number: '3',
              title: 'Data Sharing and Disclosure',
              children: [
                _buildParagraph('We do not sell your personal data. We may share your information in the following situations:'),
                _buildBulletPoint('Within the Association: Your name and official designation may be visible to other members for networking purposes, as determined by the Association.'),
                _buildBulletPoint('Razorpay (Payment Processor): When you initiate a payment, your payment details and certain contact information (such as name, email, and phone number if prefilled) are shared with Razorpay to process the transaction. Razorpay is a PCI DSS compliant payment processor and handles your data in accordance with its own privacy policy and terms. We encourage you to review Razorpay\'s privacy policy at razorpay.com/privacy.'),
                _buildBulletPoint('Service Providers: We may employ other third-party companies (e.g., cloud hosting providers, notification service providers) to facilitate our App and perform App-related services. These third parties have access to your Personal Data only to perform these tasks on our behalf and are obligated not to disclose or use it for any other purpose.'),
                _buildBulletPoint('Legal Requirements: We may disclose your Personal Data if required to do so by law or in response to valid requests by public authorities (e.g., a court or a government agency).'),
              ],
            ),

            // Section 4
            _buildSection(
              number: '4',
              title: 'Data Retention and Deletion',
              children: [
                _buildParagraph(
                  '4.1. Retention: We will retain your Personal Data only for as long as is necessary for the purposes set out in this Privacy Policy and for the duration of your membership with the Association. Payment transaction records may be retained for a longer period as required by applicable tax and accounting laws.',
                ),
                _buildParagraph(
                  '4.2. Account Deletion: Accounts are deleted by Association administrators. If you wish to request deletion of your account and associated personal data, you may raise a query through the App or contact us at nammataga@gmail.com / tagastage@gmail.com. Requests will be reviewed and actioned by the administrators in accordance with association records and applicable retention requirements. Note that payment transaction records held by Razorpay are subject to Razorpay\'s own retention policies and cannot be deleted by the Association directly.',
                ),
                _buildParagraph(
                  '4.3. Razorpay Data: To exercise data deletion rights with Razorpay, please refer to Razorpay\'s privacy policy and contact Razorpay directly.',
                ),
              ],
            ),

            // Section 5
            _buildSection(
              number: '5',
              title: 'Data Security',
              children: [
                _buildParagraph(
                  'The security of your data is important to us. We implement reasonable security safeguards, including encryption and access controls, to protect your Personal Data against unauthorized access, alteration, disclosure, or destruction. Payment transactions are processed by Razorpay using industry-standard encryption (including TLS and tokenization) and Razorpay is PCI DSS compliant. However, no method of transmission over the Internet or method of electronic storage is 100% secure.',
                ),
              ],
            ),

            // Section 6
            _buildSection(
              number: '6',
              title: 'Your Rights under the DPDP Act',
              children: [
                _buildParagraph('As a Data Principal under the DPDP Act, you have the right to:'),
                _buildBulletPoint('Access: Obtain information about the personal data we process.'),
                _buildBulletPoint('Correction: Request correction of inaccurate or incomplete data (via the query feature or by contacting the Association).'),
                _buildBulletPoint('Erasure: Request deletion of your data (subject to administrator approval and legal retention requirements).'),
                _buildBulletPoint('Grievance Redressal: Lodge a complaint regarding our data processing practices.'),
                const SizedBox(height: 8),
                _buildParagraph('To exercise these rights, please contact our Grievance Officer:'),
                const SizedBox(height: 6),
                _buildContactTile(
                  icon: LucideIcons.user,
                  title: 'Grievance Officer',
                  value: 'Joint Secretary (Legal)',
                ),
                const SizedBox(height: 8),
                _buildContactTile(
                  icon: LucideIcons.mail,
                  title: 'Email',
                  value: 'tagastate@gmail.com',
                ),
                const SizedBox(height: 8),
                _buildContactTile(
                  icon: LucideIcons.phone,
                  title: 'Phone',
                  value: '044 34919949',
                ),
                const SizedBox(height: 8),
                _buildParagraph(
                  'For rights specifically relating to payment data held by Razorpay, please contact Razorpay directly.',
                ),
              ],
            ),

            // Section 7
            _buildSection(
              number: '7',
              title: 'Children\'s Privacy',
              children: [
                _buildParagraph(
                  'The App is intended for use by adults (Government Officers). We do not knowingly collect personally identifiable information from anyone under the age of 18.',
                ),
              ],
            ),

            // Section 8
            _buildSection(
              number: '8',
              title: 'Changes to This Privacy Policy',
              children: [
                _buildParagraph(
                  'We may update our Privacy Policy from time to time. We will notify you of any changes by posting the new Privacy Policy on this page and updating the "Last Updated" date. You are advised to review this Privacy Policy periodically for any changes.',
                ),
              ],
            ),

            // Section 9
            _buildSection(
              number: '9',
              title: 'Contact Us',
              children: [
                _buildParagraph('If you have any questions about this Privacy Policy, please contact us:'),
                const SizedBox(height: 6),
                _buildContactTile(
                  icon: LucideIcons.mail,
                  title: 'Email',
                  value: 'nammataga@gmail.com / tagastage@gmail.com',
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

  Widget _buildSubheading(String text) {
    return Padding(
      padding: const EdgeInsets.only(top: 4, bottom: 6),
      child: Text(
        text,
        style: const TextStyle(
          fontFamily: 'Roboto',
          fontSize: 13.5,
          fontWeight: FontWeight.bold,
          color: AppColors.textPrimary,
        ),
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
