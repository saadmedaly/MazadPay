import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:mezadpay/widgets/side_menu_drawer.dart';
import 'package:mezadpay/widgets/app_modals.dart';
import '../services/user_api.dart';
import '../models/api_response.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/auth_provider.dart';
import 'account_page.dart';
import 'notifications_page.dart';

class AccountProfilePage extends ConsumerStatefulWidget {
  const AccountProfilePage({super.key});

  @override
  ConsumerState<AccountProfilePage> createState() => _AccountProfilePageState();
}

class _AccountProfilePageState extends ConsumerState<AccountProfilePage> {
  final UserApi _userApi = UserApi();
  bool _isLoading = true;
  bool _isSaving = false;
  String _errorMessage = '';
  String _successMessage = '';
  Map<String, dynamic> _userData = {};
  // Client feedback Phase B item 6: mirrors users.notifications_enabled --
  // the backend endpoint (PUT /users/me/notification-prefs) and mobile API
  // client (UserApi.updateNotificationPrefs) already existed and worked; only
  // the Account page switch itself was dead (hardcoded value:true,
  // onChanged:(_){}). No new preference system needed here.
  bool _notificationsEnabled = true;
  bool _isTogglingNotifications = false;
  
  // Controllers pour les champs éditables
  final _fullNameController = TextEditingController();
  final _phoneController = TextEditingController();
  final _emailController = TextEditingController();
  final _cityController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _loadUserProfile();
  }
  
  @override
  void dispose() {
    _fullNameController.dispose();
    _phoneController.dispose();
    _emailController.dispose();
    _cityController.dispose();
    super.dispose();
  }

  Future<void> _loadUserProfile() async {
    try {
      final ApiResponse<Map<String, dynamic>> response = await _userApi.getProfile();
      
      if (response.success && mounted) {
        setState(() {
          // Extraction robuste des données
          final data = response.data;
          if (data != null) {
            // La réponse peut être directement l'objet user ou contenir un champ 'user'
            _userData = data['user'] ?? data;
          }
          // Initialiser les controllers avec les données
          _fullNameController.text = _userData['full_name']?.toString() ?? '';
          _phoneController.text = _userData['phone']?.toString() ?? '';
          _emailController.text = _userData['email']?.toString() ?? '';
          _cityController.text = _userData['city']?.toString() ?? '';
          _notificationsEnabled = _userData['notifications_enabled'] ?? true;
          _isLoading = false;
        });
      } else {
        setState(() {
          _errorMessage = response.message ?? AppLocalizations.of(context)!.error_loading_profile;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _errorMessage = AppLocalizations.of(context)!.error_connection;
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _saveUserProfile() async {
    setState(() {
      _isSaving = true;
      _errorMessage = '';
      _successMessage = '';
    });
    
    try {
      final ApiResponse<Map<String, dynamic>> response = await _userApi.updateProfile(
        fullName: _fullNameController.text,
        email: _emailController.text,
        city: _cityController.text,
      );
      
      if (response.success && mounted) {
        setState(() {
          _successMessage = 'تم تحديث الملف الشخصي بنجاح';
          _isSaving = false;
        });
        // Rafraîchir les données
        _loadUserProfile();
      } else {
        setState(() {
          _errorMessage = response.message ?? 'فشل تحديث الملف الشخصي';
          _isSaving = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _errorMessage = 'خطأ في الاتصال';
          _isSaving = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    const Color primaryBlue = Color(0xFF0084FF);

    // Extraction des données utilisateur avec fallbacks
    final fullName = _fullNameController.text.isNotEmpty 
        ? _fullNameController.text 
        : _userData['full_name']?.toString() ?? AppLocalizations.of(context)!.text_37;
    final phone = _phoneController.text.isNotEmpty 
        ? _phoneController.text 
        : _userData['phone']?.toString() ?? '+222 20 00 00 00';
    final avatarUrl = _userData['avatar'] ?? _userData['avatar_url'] ?? _userData['profile_pic_url'];

    // Générer les initiales pour l'avatar fallback
    String initials = 'U';
    if (fullName.isNotEmpty && fullName != AppLocalizations.of(context)!.text_37) {
      final parts = fullName.trim().split(' ').where((p) => p.isNotEmpty).toList();
      if (parts.length > 1) {
        initials = '${parts[0][0]}${parts[1][0]}'.toUpperCase();
      } else if (parts.isNotEmpty) {
        initials = parts[0][0].toUpperCase();
      }
    }

    if (_isLoading) {
      return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFF5F7FA),
        body: const Center(
          child: CircularProgressIndicator(color: primaryBlue),
        ),
      );
    }

    if (_errorMessage.isNotEmpty) {
      return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFF5F7FA),
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.error_outline, size: 48, color: Colors.red[300]),
              const SizedBox(height: 16),
              Text(_errorMessage, style: TextStyle(color: isDarkMode ? Colors.white : Colors.black)),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: _loadUserProfile,
                style: ElevatedButton.styleFrom(backgroundColor: primaryBlue),
                child: Text(AppLocalizations.of(context)!.retry),
              ),
            ],
          ),
        ),
      );
    }

    return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFF5F7FA),
        endDrawer: const SideMenuDrawer(),
        appBar: PreferredSize(
          preferredSize: const Size.fromHeight(70),
          child: SafeArea(
            child: Builder(
              builder: (ctx) => SizedBox(
                height: 70,
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    IconButton(
                      icon: Icon(Icons.arrow_forward_ios, color: isDarkMode ? Colors.white : Colors.black, size: 20),
                      onPressed: () => Navigator.pop(ctx),
                    ),
                    Text(
                      AppLocalizations.of(context)!.text_35,
                      style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18, color: isDarkMode ? Colors.white : Colors.black),
                    ),
                    IconButton(
                      icon: Icon(Icons.menu, color: isDarkMode ? Colors.white : Colors.black, size: 28),
                      onPressed: () => Scaffold.of(ctx).openEndDrawer(),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
        body: SingleChildScrollView(
          child: Column(
            children: [
              // Profile Header
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(24),
                child: Column(
                  children: [
                    Stack(
                      children: [
                        Container(
                          width: 90,
                          height: 90,
                          decoration: BoxDecoration(
                            shape: BoxShape.circle,
                            boxShadow: [BoxShadow(color: primaryBlue.withValues(alpha: 0.3), blurRadius: 12, offset: const Offset(0, 4))],
                          ),
                          child: ClipOval(
                            child: avatarUrl != null && avatarUrl.toString().isNotEmpty
                              ? Image.network(
                                  avatarUrl.toString(),
                                  fit: BoxFit.cover,
                                  errorBuilder: (c, e, s) => Container(
                                    decoration: const BoxDecoration(
                                      shape: BoxShape.circle,
                                      gradient: LinearGradient(
                                        colors: [Color(0xFF0055FF), Color(0xFF0084FF)],
                                        begin: AlignmentDirectional.topStart,
                                        end: AlignmentDirectional.bottomEnd,
                                      ),
                                    ),
                                    child: Center(child: Text(initials, style: const TextStyle(color: Colors.white, fontSize: 36, fontWeight: FontWeight.bold))),
                                  ),
                                )
                              : Container(
                                  decoration: const BoxDecoration(
                                    shape: BoxShape.circle,
                                    gradient: LinearGradient(
                                      colors: [Color(0xFF0055FF), Color(0xFF0084FF)],
                                      begin: AlignmentDirectional.topStart,
                                      end: AlignmentDirectional.bottomEnd,
                                    ),
                                  ),
                                  child: Center(child: Text(initials, style: const TextStyle(color: Colors.white, fontSize: 36, fontWeight: FontWeight.bold))),
                                ),
                          ),
                        ),
                        Positioned(
                          bottom: 0,
                          left: 0,
                          child: Container(
                            width: 28,
                            height: 28,
                            decoration: const BoxDecoration(color: primaryBlue, shape: BoxShape.circle),
                            child: const Icon(Icons.camera_alt, color: Colors.white, size: 14),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Text(fullName, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 20)),
                    const SizedBox(height: 4),
                    Text(phone, style: TextStyle(color: Colors.grey[500], fontSize: 14)),
                  ],
                ),
              ),

              // Profile Info Fields
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(AppLocalizations.of(context)!.text_38, style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                        if (_isSaving)
                          const SizedBox(
                            width: 20,
                            height: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        else if (_successMessage.isNotEmpty)
                          Icon(Icons.check_circle, color: Colors.green, size: 20)
                        else
                          TextButton.icon(
                            onPressed: _saveUserProfile,
                            icon: const Icon(Icons.save, size: 18),
                            label: const Text('حفظ'),
                            style: TextButton.styleFrom(
                              foregroundColor: primaryBlue,
                              padding: EdgeInsets.zero,
                            ),
                          ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    _buildEditableField(context, AppLocalizations.of(context)!.text_39, _fullNameController, Icons.person_outline, isDarkMode),
                    _buildEditableField(context, AppLocalizations.of(context)!.text_40, _phoneController, Icons.phone_outlined, isDarkMode, readOnly: true),
                    _buildEditableField(context, AppLocalizations.of(context)!.text_41, _emailController, Icons.email_outlined, isDarkMode),
                    _buildEditableField(context, AppLocalizations.of(context)!.text_42, _cityController, Icons.location_city_outlined, isDarkMode),

                    if (_successMessage.isNotEmpty)
                      Padding(
                        padding: const EdgeInsets.only(top: 8),
                        child: Text(
                          _successMessage,
                          style: TextStyle(color: Colors.green, fontSize: 12),
                        ),
                      ),
                    if (_errorMessage.isNotEmpty)
                      Padding(
                        padding: const EdgeInsets.only(top: 8),
                        child: Text(
                          _errorMessage,
                          style: TextStyle(color: Colors.red, fontSize: 12),
                        ),
                      ),

                    const SizedBox(height: 32),
                    Text(AppLocalizations.of(context)!.text_44, style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                    const SizedBox(height: 16),

                    _buildSettingTile(context, AppLocalizations.of(context)!.text_45, Icons.lock_outline, isDarkMode, onTap: () => _showChangePasswordModal(context)),
                    _buildSettingTile(context, AppLocalizations.of(context)!.text_46, Icons.language, isDarkMode, trailing: AppLocalizations.of(context)!.text_47, onTap: () => AppModals.showLanguageModal(context)),
                    // Staging blocker fix (item 10/16 follow-up): the row
                    // below is a PREFERENCE TOGGLE only (enable/disable
                    // future notifications) -- it never opened the actual
                    // Notification Center, and Staging testing found no
                    // other reachable entry point was obvious enough to the
                    // client. Home/Account tab bells already navigate to
                    // NotificationsPage (home_page.dart, account_shell_page.
                    // dart), but per the client's explicit ask this adds a
                    // second, clearly-labeled entry right here so the
                    // preference and the history are visibly two different
                    // things, not one overloaded control.
                    _buildSettingTile(
                      context,
                      'سجل الإشعارات',
                      Icons.notifications_active_outlined,
                      isDarkMode,
                      onTap: () => Navigator.push(
                        context,
                        MaterialPageRoute(builder: (context) => const NotificationsPage()),
                      ),
                    ),
                    _buildSettingTile(context, AppLocalizations.of(context)!.text_48, Icons.notifications_outlined, isDarkMode, hasSwitch: true, switchValue: _notificationsEnabled, onSwitchChanged: _toggleNotifications),

                    const SizedBox(height: 32),

                    // Logout Button
                    SizedBox(
                      width: double.infinity,
                      height: 54,
                      child: OutlinedButton(
                        onPressed: () async {
                          // Déconnexion propre via le provider (efface les tokens et l'état)
                          await ref.read(authNotifierProvider.notifier).logout();
                          
                          if (mounted) {
                            Navigator.pushAndRemoveUntil(
                              context,
                              MaterialPageRoute(builder: (context) => const AccountPage()),
                              (route) => false,
                            );
                          }
                        },
                        style: OutlinedButton.styleFrom(
                          side: const BorderSide(color: Colors.red, width: 1.5),
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                        ),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.logout, color: Colors.red),
                            SizedBox(width: 8),
                            Text(AppLocalizations.of(context)!.text_49, style: TextStyle(color: Colors.red, fontWeight: FontWeight.bold, fontSize: 16)),
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(height: 24),
                  ],
                ),
              ),
            ],
          ),
        ),
      );
  }

  Widget _buildEditableField(BuildContext context, String label, TextEditingController controller, IconData icon, bool isDarkMode, {bool readOnly = false}) {
    return Container(
      margin: const EdgeInsetsDirectional.only(bottom: 12),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: readOnly
            ? (isDarkMode ? const Color(0xFF2A2A2A) : const Color(0xFFF5F5F5))
            : (isDarkMode ? const Color(0xFF1D1D1D) : Colors.white),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.withValues(alpha: 0.15)),
      ),
      child: Row(
        children: [
          Icon(icon, color: readOnly ? Colors.grey : const Color(0xFF0084FF), size: 20),
          const SizedBox(width: 12),
          Expanded(
            child: TextField(
              controller: controller,
              readOnly: readOnly,
              style: TextStyle(fontWeight: FontWeight.w600, fontSize: 14, color: readOnly ? Colors.grey[500] : null),
              decoration: InputDecoration(
                labelText: label,
                labelStyle: TextStyle(color: Colors.grey[500], fontSize: 11),
                border: InputBorder.none,
                isDense: true,
                contentPadding: EdgeInsets.zero,
              ),
            ),
          ),
          if (!readOnly) Icon(Icons.edit_outlined, color: Colors.grey[400], size: 16)
          else Icon(Icons.lock_outline, color: Colors.grey[400], size: 16),
        ],
      ),
    );
  }

  // Client feedback Phase B item 6: the "Change Password" tile above had no
  // onTap at all (dead UI), even though the backend endpoint and the mobile
  // provider/API layer (authNotifierProvider.changePassword ->
  // AuthApi.changePassword -> PUT /auth/change-password) were already fully
  // implemented and working. This wires the existing tile to a bottom sheet
  // matching AppModals.showLanguageModal's visual style, rather than building
  // a whole new page or reinventing the password-reset (OTP) flow.
  Future<void> _toggleNotifications(bool newValue) async {
    if (_isTogglingNotifications) return;
    final previous = _notificationsEnabled;
    setState(() {
      _notificationsEnabled = newValue;
      _isTogglingNotifications = true;
    });

    final response = await _userApi.updateNotificationPrefs(
      preferences: {'enabled': newValue},
    );

    if (!mounted) return;
    setState(() {
      _isTogglingNotifications = false;
      if (!response.success) {
        // Revert on failure -- never leave the switch showing a state the
        // backend didn't actually persist.
        _notificationsEnabled = previous;
      }
    });
  }

  void _showChangePasswordModal(BuildContext context) {
    final isDarkMode = Theme.of(context).brightness == Brightness.dark;
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(25)),
      ),
      builder: (sheetContext) => Padding(
        padding: EdgeInsets.only(bottom: MediaQuery.of(sheetContext).viewInsets.bottom),
        child: const _ChangePasswordSheet(),
      ),
    );
  }

  Widget _buildSettingTile(BuildContext context, String title, IconData icon, bool isDarkMode, {String? trailing, bool hasSwitch = false, VoidCallback? onTap, bool switchValue = true, ValueChanged<bool>? onSwitchChanged}) {
    return Container(
      margin: const EdgeInsetsDirectional.only(bottom: 10),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.withValues(alpha: 0.15)),
      ),
      child: ListTile(
        leading: Container(
          width: 36,
          height: 36,
          decoration: BoxDecoration(
            color: const Color(0xFF0084FF).withValues(alpha: 0.1),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(icon, color: const Color(0xFF0084FF), size: 18),
        ),
        title: Text(title, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
        trailing: hasSwitch
            ? Switch(
                value: switchValue,
                onChanged: onSwitchChanged,
                activeThumbColor: const Color(0xFF0084FF),
              )
            : trailing != null
                ? Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(trailing, style: TextStyle(color: Colors.grey[500], fontSize: 12)),
                      const SizedBox(width: 4),
                      Icon(Icons.arrow_back_ios, size: 14, color: Colors.grey[400]),
                    ],
                  )
                : Icon(Icons.arrow_back_ios, size: 14, color: Colors.grey[400]),
        onTap: onTap,
      ),
    );
  }
}

/// Change Password bottom sheet for the currently logged-in user (client
/// feedback Phase B item 6). Backend contract: PUT /auth/change-password with
/// {old_pin, new_pin} — the backend independently verifies old_pin, enforces
/// its own PIN-strength/length rules (min 8 chars, see auth_handler.go
/// ChangePasswordRequest), and returns a safe validation error rather than a
/// 500/panic on any failure (invalid current password, weak new password).
/// This sheet never logs password values, only generic success/failure text.
class _ChangePasswordSheet extends ConsumerStatefulWidget {
  const _ChangePasswordSheet();

  @override
  ConsumerState<_ChangePasswordSheet> createState() => _ChangePasswordSheetState();
}

class _ChangePasswordSheetState extends ConsumerState<_ChangePasswordSheet> {
  final _formKey = GlobalKey<FormState>();
  final _currentController = TextEditingController();
  final _newController = TextEditingController();
  final _confirmController = TextEditingController();
  bool _obscureCurrent = true;
  bool _obscureNew = true;
  bool _obscureConfirm = true;
  bool _isSubmitting = false;
  String? _errorMessage;
  bool _success = false;

  @override
  void dispose() {
    _currentController.dispose();
    _newController.dispose();
    _confirmController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() {
      _isSubmitting = true;
      _errorMessage = null;
    });

    final ok = await ref.read(authNotifierProvider.notifier).changePassword(
          oldPin: _currentController.text,
          newPin: _newController.text,
        );

    if (!mounted) return;

    if (ok) {
      setState(() {
        _isSubmitting = false;
        _success = true;
      });
      await Future.delayed(const Duration(milliseconds: 900));
      if (mounted) Navigator.of(context).pop();
    } else {
      final apiError = ref.read(authNotifierProvider).error;
      setState(() {
        _isSubmitting = false;
        // Safe, generic messages only -- never echo the raw password value or
        // a backend stack trace. The backend's own message (current_password
        // incorrect / validation error) is already localized/safe, so it is
        // shown as-is when present.
        _errorMessage = apiError?.isNotEmpty == true
            ? apiError
            : 'تعذر تغيير كلمة المرور. يرجى المحاولة مرة أخرى.';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final isArabic = Localizations.localeOf(context).languageCode == 'ar';

    String t(String ar, String fr, String en) {
      if (isArabic) return ar;
      final locale = Localizations.localeOf(context).languageCode;
      return locale == 'fr' ? fr : en;
    }

    return SingleChildScrollView(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 24),
        child: Form(
          key: _formKey,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                t('تغيير كلمة المرور', 'Changer le mot de passe', 'Change Password'),
                style: TextStyle(
                  fontFamily: 'Plus Jakarta Sans',
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: isDarkMode ? Colors.white : Colors.black,
                ),
              ),
              const SizedBox(height: 20),
              _passwordField(
                controller: _currentController,
                label: t('كلمة المرور الحالية', 'Mot de passe actuel', 'Current password'),
                obscure: _obscureCurrent,
                onToggle: () => setState(() => _obscureCurrent = !_obscureCurrent),
                validator: (v) => (v == null || v.isEmpty)
                    ? t('مطلوب', 'Requis', 'Required')
                    : null,
                isDarkMode: isDarkMode,
              ),
              const SizedBox(height: 12),
              _passwordField(
                controller: _newController,
                label: t('كلمة المرور الجديدة', 'Nouveau mot de passe', 'New password'),
                obscure: _obscureNew,
                onToggle: () => setState(() => _obscureNew = !_obscureNew),
                validator: (v) {
                  if (v == null || v.isEmpty) return t('مطلوب', 'Requis', 'Required');
                  if (v.length < 8) {
                    return t('8 أحرف على الأقل', 'Au moins 8 caractères', 'At least 8 characters');
                  }
                  return null;
                },
                isDarkMode: isDarkMode,
              ),
              const SizedBox(height: 12),
              _passwordField(
                controller: _confirmController,
                label: t('تأكيد كلمة المرور الجديدة', 'Confirmer le nouveau mot de passe', 'Confirm new password'),
                obscure: _obscureConfirm,
                onToggle: () => setState(() => _obscureConfirm = !_obscureConfirm),
                validator: (v) {
                  if (v == null || v.isEmpty) return t('مطلوب', 'Requis', 'Required');
                  if (v != _newController.text) {
                    return t('كلمتا المرور غير متطابقتين', 'Les mots de passe ne correspondent pas', 'Passwords do not match');
                  }
                  return null;
                },
                isDarkMode: isDarkMode,
              ),
              if (_errorMessage != null)
                Padding(
                  padding: const EdgeInsets.only(top: 12),
                  child: Text(_errorMessage!, style: const TextStyle(color: Colors.red, fontSize: 12)),
                ),
              if (_success)
                Padding(
                  padding: const EdgeInsets.only(top: 12),
                  child: Text(
                    t('تم تغيير كلمة المرور بنجاح', 'Mot de passe changé avec succès', 'Password changed successfully'),
                    style: const TextStyle(color: Colors.green, fontSize: 12),
                  ),
                ),
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                height: 50,
                child: ElevatedButton(
                  onPressed: _isSubmitting || _success ? null : _submit,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF0084FF),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                  child: _isSubmitting
                      ? const SizedBox(
                          width: 20,
                          height: 20,
                          child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                        )
                      : Text(
                          t('حفظ', 'Enregistrer', 'Save'),
                          style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 15),
                        ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _passwordField({
    required TextEditingController controller,
    required String label,
    required bool obscure,
    required VoidCallback onToggle,
    required String? Function(String?) validator,
    required bool isDarkMode,
  }) {
    return TextFormField(
      controller: controller,
      obscureText: obscure,
      validator: validator,
      style: TextStyle(color: isDarkMode ? Colors.white : Colors.black, fontSize: 14),
      decoration: InputDecoration(
        labelText: label,
        labelStyle: TextStyle(color: Colors.grey[500], fontSize: 12),
        filled: true,
        fillColor: isDarkMode ? const Color(0xFF2A2A2A) : const Color(0xFFF5F5F5),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide.none),
        suffixIcon: IconButton(
          icon: Icon(obscure ? Icons.visibility_off_outlined : Icons.visibility_outlined, size: 18, color: Colors.grey),
          onPressed: onToggle,
        ),
      ),
    );
  }
}