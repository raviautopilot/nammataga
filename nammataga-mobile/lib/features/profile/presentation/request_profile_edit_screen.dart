import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../domain/edit_request_model.dart';
import '../providers/profile_provider.dart';
import 'widgets/profile_subpage_scaffold.dart';

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

  void _populateInitialValues(dynamic profile) {
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
    if (_selectedWorkingDistrict == null &&
        profile.workingDistrict.isNotEmpty) {
      _selectedWorkingDistrict = profile.workingDistrict;
    }
  }

  Future<void> _submitEditRequest(dynamic profile) async {
    if (!_formKey.currentState!.validate()) return;

    final mobile = _mobileController.text.trim();
    final mail = _emailController.text.trim();
    final designation = _designationController.text.trim();
    final district = _selectedWorkingDistrict ?? profile.workingDistrict;
    final residential = _residentialController.text.trim();
    final permanent = _permanentController.text.trim();
    final remarks = _remarksController.text.trim();

    // Check if at least one field changed
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
      title: 'Request Profile Edit',
      watermarkIcon: LucideIcons.filePenLine,
      showEditNotice: false,
      body: profileAsync.when(
        data: (profile) {
          _populateInitialValues(profile);

          if (_isSuccess) {
            return _buildSuccessCard(context);
          }

          if (!_isFormActive) {
            return _buildLandingCard(context);
          }

          return _buildEditForm(profile, districtsAsync);
        },
        loading: () => const Center(
          child: Padding(
            padding: EdgeInsets.all(40),
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

  Widget _buildLandingCard(BuildContext context) {
    return Column(
      children: [
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
                blurRadius: 14,
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
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                  letterSpacing: -0.2,
                ),
              ),
              const SizedBox(height: 6),
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
                          size: 15,
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
                  ],
                ),
              ),
              const SizedBox(height: 24),
              BouncyButton(
                onTap: () {
                  setState(() => _isFormActive = true);
                },
                child: Container(
                  width: double.infinity,
                  height: 52,
                  decoration: BoxDecoration(
                    color: AppColors.primaryDark,
                    borderRadius: BorderRadius.circular(14),
                    boxShadow: [
                      BoxShadow(
                        color: AppColors.primaryDark.withValues(alpha: 0.25),
                        blurRadius: 10,
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
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildEditForm(
    dynamic profile,
    AsyncValue<List<String>> districtsAsync,
  ) {
    return Form(
      key: _formKey,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
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
                  blurRadius: 14,
                  offset: Offset(0, 4),
                ),
              ],
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Header
                const Text(
                  'Edit Request Details',
                  style: TextStyle(
                    fontSize: 16.5,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
                const SizedBox(height: 4),
                const Text(
                  'Update the applicable fields below. Changes require admin approval.',
                  style: TextStyle(
                    fontSize: 12,
                    color: AppColors.textSecondary,
                  ),
                ),
                const SizedBox(height: 18),
                const Divider(height: 1, color: Color(0xFFEFF4F0)),
                const SizedBox(height: 16),

                // 1. Mobile Number
                _buildFieldLabel('Mobile Number'),
                TextFormField(
                  controller: _mobileController,
                  keyboardType: TextInputType.phone,
                  decoration: const InputDecoration(
                    prefixIcon: Icon(LucideIcons.phone, size: 16),
                    hintText: 'e.g. 9876543210',
                  ),
                  validator: (val) {
                    if (val == null || val.trim().isEmpty) {
                      return 'Mobile number cannot be empty';
                    }
                    return null;
                  },
                ),
                const SizedBox(height: 14),

                // 2. Email Address
                _buildFieldLabel('Email Address'),
                TextFormField(
                  controller: _emailController,
                  keyboardType: TextInputType.emailAddress,
                  decoration: const InputDecoration(
                    prefixIcon: Icon(LucideIcons.mail, size: 16),
                    hintText: 'e.g. member@nammataga.com',
                  ),
                  validator: (val) {
                    if (val == null || val.trim().isEmpty) {
                      return 'Email address cannot be empty';
                    }
                    if (!val.contains('@')) {
                      return 'Enter a valid email address';
                    }
                    return null;
                  },
                ),
                const SizedBox(height: 14),

                // 3. Designation
                _buildFieldLabel('Designation'),
                TextFormField(
                  controller: _designationController,
                  decoration: const InputDecoration(
                    prefixIcon: Icon(LucideIcons.briefcase, size: 16),
                    hintText: 'e.g. Agricultural Officer',
                  ),
                ),
                const SizedBox(height: 14),

                // 4. Working District
                _buildFieldLabel('Working District'),
                districtsAsync.when(
                  data: (districts) {
                    final currentVal =
                        districts.contains(_selectedWorkingDistrict)
                        ? _selectedWorkingDistrict
                        : (districts.isNotEmpty ? districts.first : null);

                    return DropdownButtonFormField<String>(
                      initialValue: currentVal,
                      decoration: const InputDecoration(
                        prefixIcon: Icon(LucideIcons.mapPin, size: 16),
                      ),
                      items: districts.map((d) {
                        return DropdownMenuItem(value: d, child: Text(d));
                      }).toList(),
                      onChanged: (val) {
                        setState(() => _selectedWorkingDistrict = val);
                      },
                    );
                  },
                  loading: () =>
                      const LinearProgressIndicator(color: AppColors.primary),
                  error: (e, s) => Text('Error loading districts: $e'),
                ),
                const SizedBox(height: 14),

                // 5. Residential Address
                _buildFieldLabel('Residential Address'),
                TextFormField(
                  controller: _residentialController,
                  maxLines: 2,
                  decoration: const InputDecoration(
                    prefixIcon: Icon(LucideIcons.home, size: 16),
                    hintText: 'Current residential address',
                  ),
                ),
                const SizedBox(height: 14),

                // 6. Permanent Address
                _buildFieldLabel('Permanent Address'),
                TextFormField(
                  controller: _permanentController,
                  maxLines: 2,
                  decoration: const InputDecoration(
                    prefixIcon: Icon(LucideIcons.home, size: 16),
                    hintText: 'Permanent native address',
                  ),
                ),
                const SizedBox(height: 14),

                // 7. Remarks
                _buildFieldLabel('Remarks / Reason (Optional)'),
                TextFormField(
                  controller: _remarksController,
                  maxLines: 2,
                  decoration: const InputDecoration(
                    prefixIcon: Icon(LucideIcons.messageSquare, size: 16),
                    hintText: 'Reason for requesting this update...',
                  ),
                ),
                const SizedBox(height: 24),

                // Submit Button
                BouncyButton(
                  onTap: _isSubmitting
                      ? () {}
                      : () => _submitEditRequest(profile),
                  child: Container(
                    width: double.infinity,
                    height: 52,
                    decoration: BoxDecoration(
                      color: AppColors.primaryDark,
                      borderRadius: BorderRadius.circular(14),
                    ),
                    child: Center(
                      child: _isSubmitting
                          ? const SizedBox(
                              width: 20,
                              height: 20,
                              child: CircularProgressIndicator(
                                color: Colors.white,
                                strokeWidth: 2,
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
                                  ),
                                ),
                                SizedBox(width: 8),
                                Icon(
                                  LucideIcons.send,
                                  size: 16,
                                  color: Colors.white,
                                ),
                              ],
                            ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSuccessCard(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
      ),
      child: Column(
        children: [
          Container(
            width: 64,
            height: 64,
            decoration: const BoxDecoration(
              color: Color(0xFFE8F5EE),
              shape: BoxShape.circle,
            ),
            child: const Icon(
              LucideIcons.circleCheck,
              color: AppColors.primary,
              size: 32,
            ),
          ),
          const SizedBox(height: 16),
          const Text(
            'Request Submitted!',
            style: TextStyle(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 6),
          const Text(
            'Your profile edit request has been successfully recorded. Association administrators will review the changes and update your official record.',
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 13,
              color: AppColors.textSecondary,
              height: 1.45,
            ),
          ),
          const SizedBox(height: 22),
          BouncyButton(
            onTap: () {
              context.go('/profile');
            },
            child: Container(
              width: double.infinity,
              height: 48,
              decoration: BoxDecoration(
                color: AppColors.primaryDark,
                borderRadius: BorderRadius.circular(12),
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

  Widget _buildFieldLabel(String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 5),
      child: Text(
        text,
        style: const TextStyle(
          fontSize: 12.5,
          fontWeight: FontWeight.w600,
          color: Color(0xFF374151),
        ),
      ),
    );
  }

  Widget _buildCheckItem(String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(2.5),
            decoration: const BoxDecoration(
              color: Color(0xFFE8F5EE),
              shape: BoxShape.circle,
            ),
            child: const Icon(
              LucideIcons.check,
              size: 12,
              color: AppColors.primary,
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              text,
              style: const TextStyle(
                fontSize: 12.5,
                color: AppColors.textSecondary,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
