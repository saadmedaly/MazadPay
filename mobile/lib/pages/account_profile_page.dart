import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:mezadpay/widgets/side_menu_drawer.dart';
import 'package:mezadpay/widgets/app_modals.dart';
import '../services/user_api.dart';
import '../models/api_response.dart';
import 'account_page.dart';

class AccountProfilePage extends StatefulWidget {
  const AccountProfilePage({super.key});

  @override
  State<AccountProfilePage> createState() => _AccountProfilePageState();
}

class _AccountProfilePageState extends State<AccountProfilePage> {
  final UserApi _userApi = UserApi();
  bool _isLoading = true;
  bool _isSaving = false;
  String _errorMessage = '';
  String _successMessage = '';
  Map<String, dynamic> _userData = {};
  
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
          if (data != null && data is Map<String, dynamic>) {
            // La réponse peut être directement l'objet user ou contenir un champ 'user'
            _userData = data['user'] ?? data;
          }
          // Initialiser les controllers avec les données
          _fullNameController.text = _userData['full_name']?.toString() ?? '';
          _phoneController.text = _userData['phone']?.toString() ?? '';
          _emailController.text = _userData['email']?.toString() ?? '';
          _cityController.text = _userData['city']?.toString() ?? '';
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
        phone: _phoneController.text,
        email: _emailController.text,
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
    final email = _emailController.text.isNotEmpty 
        ? _emailController.text 
        : _userData['email']?.toString() ?? 'badal@example.com';
    final city = _cityController.text.isNotEmpty 
        ? _cityController.text 
        : _userData['city']?.toString() ?? _userData['location'] ?? AppLocalizations.of(context)!.text_43;
    final avatarUrl = _userData['avatar'] ?? _userData['avatar_url'] ?? _userData['profile_pic_url'];

    // Générer les initiales pour l'avatar fallback
    String initials = 'U';
    if (fullName.isNotEmpty && fullName != AppLocalizations.of(context)!.text_37) {
      final parts = fullName.split(' ');
      if (parts.length > 1) {
        initials = '${parts[0][0]}${parts[1][0]}'.toUpperCase();
      } else {
        initials = fullName[0].toUpperCase();
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
                            boxShadow: [BoxShadow(color: primaryBlue.withOpacity(0.3), blurRadius: 12, offset: const Offset(0, 4))],
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
                    _buildEditableField(context, AppLocalizations.of(context)!.text_40, _phoneController, Icons.phone_outlined, isDarkMode),
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

                    _buildSettingTile(context, AppLocalizations.of(context)!.text_45, Icons.lock_outline, isDarkMode),
                    _buildSettingTile(context, AppLocalizations.of(context)!.text_46, Icons.language, isDarkMode, trailing: AppLocalizations.of(context)!.text_47, onTap: () => AppModals.showLanguageModal(context)),
                    _buildSettingTile(context, AppLocalizations.of(context)!.text_48, Icons.notifications_outlined, isDarkMode, hasSwitch: true),

                    const SizedBox(height: 32),

                    // Logout Button
                    SizedBox(
                      width: double.infinity,
                      height: 54,
                      child: OutlinedButton(
                        onPressed: () {
                          Navigator.pushAndRemoveUntil(
                            context,
                            MaterialPageRoute(builder: (context) => const AccountPage()),
                            (route) => false,
                          );
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

  Widget _buildInfoField(BuildContext context, String label, String value, IconData icon, bool isDarkMode) {
    return Container(
      margin: const EdgeInsetsDirectional.only(bottom: 12),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.withOpacity(0.15)),
      ),
      child: Row(
        children: [
          Icon(icon, color: const Color(0xFF0084FF), size: 20),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(label, style: TextStyle(color: Colors.grey[500], fontSize: 11)),
                const SizedBox(height: 2),
                Text(value, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
              ],
            ),
          ),
          Icon(Icons.edit_outlined, color: Colors.grey[400], size: 16),
        ],
      ),
    );
  }

  Widget _buildEditableField(BuildContext context, String label, TextEditingController controller, IconData icon, bool isDarkMode) {
    return Container(
      margin: const EdgeInsetsDirectional.only(bottom: 12),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.withOpacity(0.15)),
      ),
      child: Row(
        children: [
          Icon(icon, color: const Color(0xFF0084FF), size: 20),
          const SizedBox(width: 12),
          Expanded(
            child: TextField(
              controller: controller,
              style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
              decoration: InputDecoration(
                labelText: label,
                labelStyle: TextStyle(color: Colors.grey[500], fontSize: 11),
                border: InputBorder.none,
                isDense: true,
                contentPadding: EdgeInsets.zero,
              ),
            ),
          ),
          Icon(Icons.edit_outlined, color: Colors.grey[400], size: 16),
        ],
      ),
    );
  }

  Widget _buildSettingTile(BuildContext context, String title, IconData icon, bool isDarkMode, {String? trailing, bool hasSwitch = false, VoidCallback? onTap}) {
    return Container(
      margin: const EdgeInsetsDirectional.only(bottom: 10),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.withOpacity(0.15)),
      ),
      child: ListTile(
        leading: Container(
          width: 36,
          height: 36,
          decoration: BoxDecoration(
            color: const Color(0xFF0084FF).withOpacity(0.1),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(icon, color: const Color(0xFF0084FF), size: 18),
        ),
        title: Text(title, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
        trailing: hasSwitch
            ? Switch(
                value: true,
                onChanged: (_) {},
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