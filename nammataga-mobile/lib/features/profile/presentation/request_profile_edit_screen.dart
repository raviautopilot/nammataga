import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../domain/edit_request_model.dart';
import '../domain/member_profile.dart';
import '../providers/profile_provider.dart';
import 'widgets/profile_subpage_scaffold.dart';

/// Screens 8 & 9: Request Profile Edit
/// Screen 8: Landing overview with illustrated badge and change checklist.
/// Screen 9: Multi-step breadcrumb progress edit form (Details -> Review -> Submit).
class RequestProfileEditScreen extends ConsumerStatefulWidget {
  const RequestProfileEditScreen({super.key});

  @override
  ConsumerState<RequestProfileEditScreen> createState() =>
      _RequestProfileEditScreenState();
}

class _RequestProfileEditScreenState
    extends ConsumerState<RequestProfileEditScreen> {
  bool _isFormActive = false;
  bool _isSubmitting = false;
  bool _isSuccess = false;

  // Selected sections
  bool _selectContact = true;
  bool _selectEmployment = true;
  bool _selectAddress = true;

  final _formKey = GlobalKey<FormState>();
  late TextEditingController _mobileController;
  late TextEditingController _emailController;
  late TextEditingController _designationController;
  late TextEditingController _residentialController;
  late TextEditingController _permanentController;
  late TextEditingController _remarksController;
  String? _selectedWorkingDistrict;

  @override
  void initState() {
    super.initState();
    _mobileController = TextEditingController();
    _emailController = TextEditingController();
    _designationController = TextEditingController();
    _residentialController = TextEditingController();
    _permanentController = TextEditingController();
    _remarksController = TextEditingController();
  }

  @override
  void dispose() {
    _mobileController.dispose();
    _emailController.dispose();
    _designationController.dispose();
    _residentialController.dispose();
    _permanentController.dispose();
    _remarksController.dispose();
    super.dispose();
  }

  void _populateInitialValues(MemberProfile profile) {
    if (_mobileController.text.isEmpty && profile.mobileNumber.isNotEmpty) {
      _mobileController.text = profile.mobileNumber;
    }
    if (_emailController.text.isEmpty && profile.emailId.isNotEmpty) {
      _emailController.text = profile.emailId;
    }
    if (_designationController.text.isEmpty && profile.designation.isNotEmpty) {
      _designationController.text = profile.designation;
    }
    if (_residentialController.text.isEmpty &&
        profile.residentialAddress.isNotEmpty) {
      _residentialController.text = profile.residentialAddress;
    }
    if (_permanentController.text.isEmpty &&
        profile.permanentAddress.isNotEmpty) {
      _permanentController.text = profile.permanentAddress;
    }
    if (_selectedWorkingDistrict == null && profile.workingDistrict.isNotEmpty) {
      _selectedWorkingDistrict = profile.workingDistrict;
    }
  }

  Future<void> _submitEditRequest(MemberProfile profile) async {
    final mobile = _mobileController.text.trim();
    final mail = _emailController.text.trim();
    final designation = _designationController.text.trim();
    final district = _selectedWorkingDistrict ?? profile.workingDistrict;
    final residential = _residentialController.text.trim();
    final permanent = _permanentController.text.trim();
    final remarks = _remarksController.text.trim();

    final changed =
        mobile != profile.mobileNumber ||
        mail != profile.emailId ||
        designation != profile.designation ||
        district != profile.workingDistrict ||
        residential != profile.residentialAddress ||
        permanent != profile.permanentAddress;

    if (!changed) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text(
            'No changes detected compared to your current profile.',
          ),
          backgroundColor: Color(0xFFD97706),
        ),
      );
      return;
    }

    setState(() => _isSubmitting = true);

    try {
      final payload = EditRequestPayload(
        email: profile.emailId,
        mobileNumber: mobile,
        mailId: mail,
        designation: designation,
        workingDistrict: district,
        residentialAddress: residential,
        permanentAddress: permanent,
        remarks: remarks,
      );

      final repo = ref.read(profileRepositoryProvider);
      await repo.submitEditRequest(payload);

      if (mounted) {
        setState(() {
          _isSubmitting = false;
          _isSuccess = true;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isSubmitting = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to submit request: $e'),
            backgroundColor: AppColors.destructive,
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final profileAsync = ref.watch(memberProfileProvider);
    final districtsAsync = ref.watch(districtsListProvider);

    return ProfileSubpageScaffold(
      title: _isFormActive ? 'Edit Request' : 'Request Profile Edit',
      watermarkIcon: LucideIcons.filePenLine,
      showEditNotice: false,
      body: profileAsync.when(
        data: (profile) {
          _populateInitialValues(profile);

          if (_isSuccess) {
            return _buildSuccessCard(context);
          }

          if (!_isFormActive) {
            return _buildLandingOverview(context);
          }

          return _buildEditWorkflow(profile, districtsAsync);
        },
        loading: () => Container(
          height: 320,
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: const Color(0xFFE2EBE5)),
          ),
          child: const Center(
            child: CircularProgressIndicator(color: AppColors.primary),
          ),
        ),
        error: (error, stack) => Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Text(
              'Unable to load member information.\n$error',
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: AppColors.destructive,
                fontSize: 13,
              ),
            ),
          ),
        ),
      ),
    );
  }

  // --- SCREEN 8: LANDING OVERVIEW ---
  Widget _buildLandingOverview(BuildContext context) {
    return Column(
      children: [
        // Hero Illustrated Card matching reference
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
            boxShadow: const [
              BoxShadow(
                color: Color(0x061B4332),
                blurRadius: 16,
                offset: Offset(0, 4),
              ),
            ],
          ),
          child: Column(
            children: [
              Container(
                width: 72,
                height: 72,
                decoration: BoxDecoration(
                  color: const Color(0xFFE8F5EE),
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: const Color(0xFFCBE3D5),
                    width: 1.5,
                  ),
                ),
                child: const Icon(
                  LucideIcons.filePenLine,
                  color: AppColors.primary,
                  size: 32,
                ),
              ),
              const SizedBox(height: 16),
              const Text(
                'Request Profile Edit',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 19,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                  letterSpacing: -0.2,
                ),
              ),
              const SizedBox(height: 8),
              const Text(
                'Submit changes to your profile information. Your request will be reviewed by admin and you will be notified once it is approved.',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 12.5,
                  color: AppColors.textSecondary,
                  height: 1.45,
                ),
              ),
              const SizedBox(height: 20),

              // Info Checklist Container
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: const Color(0xFFF9FAF7),
                  borderRadius: BorderRadius.circular(14),
                  border: Border.all(color: const Color(0xFFE5ECE7)),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Row(
                      children: [
                        Icon(
                          LucideIcons.info,
                          size: 16,
                          color: AppColors.primary,
                        ),
                        SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            'You can request changes for:',
                            style: TextStyle(
                              fontSize: 13,
                              fontWeight: FontWeight.bold,
                              color: AppColors.textPrimary,
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    _buildCheckItem('Contact Details (Mobile & Email)'),
                    _buildCheckItem(
                      'Employment Details (Designation & District)',
                    ),
                    _buildCheckItem(
                      'Address Details (Residential & Permanent)',
                    ),
                    _buildCheckItem('Official Details & Verification'),
                  ],
                ),
              ),
              const SizedBox(height: 24),

              // Primary "New Edit Request" Button
              BouncyButton(
                onTap: () {
                  HapticFeedback.lightImpact();
                  setState(() {
                    _isFormActive = true;
                  });
                },
                child: Container(
                  width: double.infinity,
                  height: 52,
                  decoration: BoxDecoration(
                    color: const Color(0xFF0D2C1B),
                    borderRadius: BorderRadius.circular(14),
                    boxShadow: [
                      BoxShadow(
                        color: const Color(0xFF0D2C1B).withValues(alpha: 0.28),
                        blurRadius: 12,
                        offset: const Offset(0, 4),
                      ),
                    ],
                  ),
                  child: const Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        'New Edit Request',
                        style: TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.bold,
                          color: Colors.white,
                          letterSpacing: 0.1,
                        ),
                      ),
                      SizedBox(width: 8),
                      Icon(
                        LucideIcons.arrowRight,
                        size: 16,
                        color: Colors.white,
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 12),

              // Secondary Outline Button
              BouncyButton(
                onTap: () => _showPreviousRequestsSheet(context),
                child: Container(
                  width: double.infinity,
                  height: 48,
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(14),
                    border: Border.all(
                      color: const Color(0xFFD0E5D8),
                      width: 1.2,
                    ),
                  ),
                  child: const Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(
                        LucideIcons.history,
                        size: 16,
                        color: Color(0xFF1B4332),
                      ),
                      SizedBox(width: 8),
                      Text(
                        'View Previous Requests',
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF1B4332),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  // --- SCREEN 9: MULTI-STEP EDIT WORKFLOW ---
  Widget _buildEditWorkflow(
    MemberProfile profile,
    AsyncValue<List<String>> districtsAsync,
  ) {
    return Column(
      children: [
        // Step Indicator Breadcrumbs matching reference
        _buildStepIndicator(),
        const SizedBox(height: 18),

        // Edit Form Container
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
            boxShadow: const [
              BoxShadow(
                color: Color(0x061B4332),
                blurRadius: 16,
                offset: Offset(0, 4),
              ),
            ],
          ),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Header with icon
                Row(
                  children: [
                    Container(
                      width: 38,
                      height: 38,
                      decoration: BoxDecoration(
                        color: const Color(0xFFE8F5EE),
                        borderRadius: BorderRadius.circular(10),
                      ),
                      child: const Icon(
                        LucideIcons.filePenLine,
                        size: 18,
                        color: AppColors.primary,
                      ),
                    ),
                    const SizedBox(width: 12),
                    const Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Edit Request Details',
                            style: TextStyle(
                              fontSize: 16.5,
                              fontWeight: FontWeight.bold,
                              color: AppColors.textPrimary,
                              letterSpacing: -0.2,
                            ),
                          ),
                          SizedBox(height: 2),
                          Text(
                            'Modify the fields below for admin review.',
                            style: TextStyle(
                              fontSize: 12,
                              color: AppColors.textSecondary,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                const Divider(height: 1, thickness: 1, color: Color(0xFFEFF4F0)),
                const SizedBox(height: 16),

                // Section Selector Chips
                const Text(
                  'Sections to Update',
                  style: TextStyle(
                    fontSize: 12.5,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF5C7164),
                  ),
                ),
                const SizedBox(height: 8),
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: [
                    _buildFilterChip('Contact', _selectContact, () {
                      setState(() => _selectContact = !_selectContact);
                    }),
                    _buildFilterChip('Employment', _selectEmployment, () {
                      setState(() => _selectEmployment = !_selectEmployment);
                    }),
                    _buildFilterChip('Address', _selectAddress, () {
                      setState(() => _selectAddress = !_selectAddress);
                    }),
                  ],
                ),
                const SizedBox(height: 18),

                // Contact Details Section
                if (_selectContact) ...[
                  _buildSectionHeader('Contact Details', LucideIcons.phone),
                  const SizedBox(height: 10),
                  _buildTextField(
                    controller: _mobileController,
                    label: 'Mobile Number',
                    hintText: 'Enter 10-digit mobile number',
                    icon: LucideIcons.phone,
                    keyboardType: TextInputType.phone,
                    validator: (val) {
                      if (val == null || val.trim().isEmpty) {
                        return 'Mobile number is required';
                      }
                      return null;
                    },
                  ),
                  const SizedBox(height: 14),
                  _buildTextField(
                    controller: _emailController,
                    label: 'Email Address',
                    hintText: 'Enter email address',
                    icon: LucideIcons.mail,
                    keyboardType: TextInputType.emailAddress,
                    validator: (val) {
                      if (val == null || val.trim().isEmpty) {
                        return 'Email address is required';
                      }
                      if (!val.contains('@')) {
                        return 'Enter a valid email address';
                      }
                      return null;
                    },
                  ),
                  const SizedBox(height: 18),
                ],

                // Employment Details Section
                if (_selectEmployment) ...[
                  _buildSectionHeader(
                    'Employment Details',
                    LucideIcons.briefcase,
                  ),
                  const SizedBox(height: 10),
                  _buildTextField(
                    controller: _designationController,
                    label: 'Designation',
                    hintText: 'e.g. Assistant Director of Agriculture',
                    icon: LucideIcons.briefcase,
                  ),
                  const SizedBox(height: 14),
                  districtsAsync.when(
                    data: (districts) => _buildDistrictDropdown(districts),
                    loading: () => const LinearProgressIndicator(),
                    error: (e, s) => _buildTextField(
                      controller: TextEditingController(
                        text: _selectedWorkingDistrict ?? '',
                      ),
                      label: 'Working District',
                      hintText: 'Enter district name',
                      icon: LucideIcons.mapPin,
                    ),
                  ),
                  const SizedBox(height: 18),
                ],

                // Address Details Section
                if (_selectAddress) ...[
                  _buildSectionHeader('Address Details', LucideIcons.mapPin),
                  const SizedBox(height: 10),
                  _buildTextField(
                    controller: _residentialController,
                    label: 'Residential Address',
                    hintText: 'Enter residential address',
                    icon: LucideIcons.home,
                    maxLines: 2,
                  ),
                  const SizedBox(height: 14),
                  _buildTextField(
                    controller: _permanentController,
                    label: 'Permanent Address',
                    hintText: 'Enter permanent address',
                    icon: LucideIcons.landmark,
                    maxLines: 2,
                  ),
                  const SizedBox(height: 18),
                ],

                // Remarks
                _buildSectionHeader(
                  'Reason for Request',
                  LucideIcons.messageSquare,
                ),
                const SizedBox(height: 10),
                _buildTextField(
                  controller: _remarksController,
                  label: 'Remarks (Optional)',
                  hintText: 'Reason for requesting profile changes...',
                  icon: LucideIcons.messageSquare,
                  maxLines: 2,
                ),
                const SizedBox(height: 24),

                // Primary Submit Button matching test expectation
                BouncyButton(
                  onTap: _isSubmitting
                      ? null
                      : () {
                          if (!_formKey.currentState!.validate()) return;
                          HapticFeedback.lightImpact();
                          _submitEditRequest(profile);
                        },
                  child: Container(
                    width: double.infinity,
                    height: 52,
                    decoration: BoxDecoration(
                      color: const Color(0xFF0D2C1B),
                      borderRadius: BorderRadius.circular(14),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0xFF0D2C1B)
                              .withValues(alpha: 0.28),
                          blurRadius: 12,
                          offset: const Offset(0, 4),
                        ),
                      ],
                    ),
                    child: Center(
                      child: _isSubmitting
                          ? const SizedBox(
                              width: 22,
                              height: 22,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                                color: Colors.white,
                              ),
                            )
                          : const Row(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Text(
                                  'Submit Edit Request',
                                  style: TextStyle(
                                    fontSize: 15,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.white,
                                    letterSpacing: 0.1,
                                  ),
                                ),
                                SizedBox(width: 8),
                                Icon(
                                  LucideIcons.arrowRight,
                                  size: 16,
                                  color: Colors.white,
                                ),
                              ],
                            ),
                    ),
                  ),
                ),
                const SizedBox(height: 12),

                // Cancel Button
                BouncyButton(
                  onTap: () {
                    HapticFeedback.lightImpact();
                    setState(() => _isFormActive = false);
                  },
                  child: Container(
                    width: double.infinity,
                    height: 46,
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(14),
                      border: Border.all(color: const Color(0xFFD0E5D8)),
                    ),
                    child: const Center(
                      child: Text(
                        'Cancel',
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF5C7164),
                        ),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildStepIndicator() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFE2EBE5), width: 1),
        boxShadow: const [
          BoxShadow(
            color: Color(0x061B4332),
            blurRadius: 12,
            offset: Offset(0, 3),
          ),
        ],
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceAround,
        children: [
          _buildStepItem(stepNumber: 1, label: 'Details', isActive: true, isCompleted: false),
          _buildStepDivider(isCompleted: false),
          _buildStepItem(stepNumber: 2, label: 'Review', isActive: false, isCompleted: false),
          _buildStepDivider(isCompleted: false),
          _buildStepItem(stepNumber: 3, label: 'Submit', isActive: false, isCompleted: false),
        ],
      ),
    );
  }

  Widget _buildStepItem({
    required int stepNumber,
    required String label,
    required bool isActive,
    required bool isCompleted,
  }) {
    final circleColor = isActive ? const Color(0xFF0D2C1B) : const Color(0xFFE2EBE5);
    final textColor = isActive ? const Color(0xFF0D2C1B) : const Color(0xFF94A3B8);

    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 28,
          height: 28,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: circleColor,
          ),
          child: Center(
            child: Text(
              '$stepNumber',
              style: TextStyle(
                color: isActive ? Colors.white : const Color(0xFF64748B),
                fontSize: 12,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        ),
        const SizedBox(height: 4),
        Text(
          label,
          style: TextStyle(
            fontSize: 11,
            fontWeight: isActive ? FontWeight.bold : FontWeight.w500,
            color: textColor,
          ),
        ),
      ],
    );
  }

  Widget _buildStepDivider({required bool isCompleted}) {
    return Expanded(
      child: Container(
        height: 2,
        margin: const EdgeInsets.symmetric(horizontal: 8),
        color: isCompleted ? const Color(0xFF2D6A4F) : const Color(0xFFE2EBE5),
      ),
    );
  }

  Widget _buildFilterChip(String label, bool isSelected, VoidCallback onTap) {
    return InkWell(
      onTap: () {
        HapticFeedback.lightImpact();
        onTap();
      },
      borderRadius: BorderRadius.circular(20),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
        decoration: BoxDecoration(
          color: isSelected ? const Color(0xFFE8F5EE) : const Color(0xFFF1F5F2),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(
            color: isSelected ? const Color(0xFF52B788) : const Color(0xFFD0E5D8),
            width: 1,
          ),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (isSelected) ...[
              const Icon(LucideIcons.check, size: 13, color: AppColors.primary),
              const SizedBox(width: 4),
            ],
            Text(
              label,
              style: TextStyle(
                fontSize: 12,
                fontWeight: isSelected ? FontWeight.bold : FontWeight.w500,
                color: isSelected ? AppColors.primaryDark : const Color(0xFF64748B),
              ),
            ),
          ],
        ),
      ),
    );
  }

  // --- SUCCESS STATE ---
  Widget _buildSuccessCard(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(28),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
        boxShadow: const [
          BoxShadow(
            color: Color(0x061B4332),
            blurRadius: 16,
            offset: Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        children: [
          Container(
            width: 72,
            height: 72,
            decoration: const BoxDecoration(
              color: Color(0xFFE8F5EE),
              shape: BoxShape.circle,
            ),
            child: const Icon(
              LucideIcons.checkCheck,
              size: 36,
              color: Color(0xFF2D6A4F),
            ),
          ),
          const SizedBox(height: 18),
          const Text(
            'Edit Request Submitted',
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 19,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
              letterSpacing: -0.2,
            ),
          ),
          const SizedBox(height: 8),
          const Text(
            'Your profile change request has been submitted successfully to the TAGA administration for review. You will receive an update once it is approved.',
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 13,
              color: AppColors.textSecondary,
              height: 1.45,
            ),
          ),
          const SizedBox(height: 24),
          BouncyButton(
            onTap: () {
              HapticFeedback.lightImpact();
              context.go('/profile');
            },
            child: Container(
              width: double.infinity,
              height: 50,
              decoration: BoxDecoration(
                color: const Color(0xFF0D2C1B),
                borderRadius: BorderRadius.circular(14),
              ),
              child: const Center(
                child: Text(
                  'Back to Profile',
                  style: TextStyle(
                    fontSize: 14.5,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  // --- SUB-WIDGET HELPERS ---
  Widget _buildSectionHeader(String title, IconData icon) {
    return Row(
      children: [
        Icon(icon, size: 16, color: AppColors.primary),
        const SizedBox(width: 8),
        Text(
          title,
          style: const TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
          ),
        ),
      ],
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    required String hintText,
    required IconData icon,
    TextInputType keyboardType = TextInputType.text,
    int maxLines = 1,
    String? Function(String?)? validator,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: const TextStyle(
            fontSize: 12.5,
            fontWeight: FontWeight.w600,
            color: Color(0xFF5C7164),
          ),
        ),
        const SizedBox(height: 6),
        TextFormField(
          controller: controller,
          keyboardType: keyboardType,
          maxLines: maxLines,
          validator: validator,
          style: const TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w500,
            color: AppColors.textPrimary,
          ),
          decoration: InputDecoration(
            hintText: hintText,
            hintStyle: const TextStyle(
              fontSize: 13,
              color: Color(0xFF94A3B8),
            ),
            prefixIcon: Icon(icon, size: 18, color: const Color(0xFF64748B)),
            filled: true,
            fillColor: const Color(0xFFF9FAF7),
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 14,
              vertical: 12,
            ),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: Color(0xFFE2EBE5)),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: Color(0xFFE2EBE5)),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: AppColors.primary, width: 1.5),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildDistrictDropdown(List<String> districts) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          'Working District',
          style: TextStyle(
            fontSize: 12.5,
            fontWeight: FontWeight.w600,
            color: Color(0xFF5C7164),
          ),
        ),
        const SizedBox(height: 6),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 12),
          decoration: BoxDecoration(
            color: const Color(0xFFF9FAF7),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: const Color(0xFFE2EBE5)),
          ),
          child: DropdownButtonHideUnderline(
            child: DropdownButton<String>(
              isExpanded: true,
              value: districts.contains(_selectedWorkingDistrict)
                  ? _selectedWorkingDistrict
                  : null,
              hint: const Text(
                'Select District',
                style: TextStyle(fontSize: 13, color: Color(0xFF94A3B8)),
              ),
              items: districts.map((district) {
                return DropdownMenuItem<String>(
                  value: district,
                  child: Text(
                    district,
                    style: const TextStyle(
                      fontSize: 14,
                      color: AppColors.textPrimary,
                    ),
                  ),
                );
              }).toList(),
              onChanged: (val) {
                setState(() => _selectedWorkingDistrict = val);
              },
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildCheckItem(String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        children: [
          Container(
            width: 18,
            height: 18,
            decoration: const BoxDecoration(
              shape: BoxShape.circle,
              color: Color(0xFFE8F5EE),
            ),
            child: const Icon(
              LucideIcons.check,
              size: 11,
              color: Color(0xFF2D6A4F),
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              text,
              style: const TextStyle(
                fontSize: 12.5,
                color: AppColors.textPrimary,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
        ],
      ),
    );
  }

  void _showPreviousRequestsSheet(BuildContext context) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => Container(
        padding: const EdgeInsets.all(24),
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                color: const Color(0xFFE2E8F0),
                borderRadius: BorderRadius.circular(2),
              ),
            ),
            const SizedBox(height: 20),
            const Row(
              children: [
                Icon(LucideIcons.history, color: AppColors.primary, size: 20),
                SizedBox(width: 10),
                Text(
                  'Edit Requests History',
                  style: TextStyle(
                    fontSize: 17,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            const Text(
              'Your edit requests are submitted directly to the TAGA State Administration for official verification. Once approved or reviewed, your profile will automatically reflect the updated records.',
              style: TextStyle(
                fontSize: 13,
                color: AppColors.textSecondary,
                height: 1.45,
              ),
            ),
            const SizedBox(height: 24),
            BouncyButton(
              onTap: () => Navigator.pop(ctx),
              child: Container(
                width: double.infinity,
                height: 48,
                decoration: BoxDecoration(
                  color: const Color(0xFF0D2C1B),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Center(
                  child: Text(
                    'Close',
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
