import 'package:flutter/material.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/services/auth_service.dart';
import 'package:mezadpay/services/user_api.dart';
import 'package:mezadpay/widgets/app_modals.dart';
import 'login_page.dart';

class SettingsPage extends StatefulWidget {
  const SettingsPage({super.key});

  @override
  State<SettingsPage> createState() => _SettingsPageState();
}

class _SettingsPageState extends State<SettingsPage> {
  final UserApi _userApi = UserApi();
  bool _isLoading = true;
  bool _isDeletingAccount = false;
  Map<String, dynamic> _settings = {
    'push_notifications': true,
    'email_notifications': true,
    'sms_notifications': false,
    'public_profile': true,
    'show_bid_history': true,
  };

  @override
  void initState() {
    super.initState();
    _loadSettings();
  }

  Future<void> _loadSettings() async {
    try {
      final response = await _userApi.getUserSettings();
      if (mounted && response.success && response.data != null) {
        setState(() {
          _settings = Map<String, dynamic>.from(response.data!['settings'] ?? _settings);
          _isLoading = false;
        });
      } else {
        if (mounted) setState(() => _isLoading = false);
      }
    } catch (e) {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _updateSetting(String key, bool value) async {
    setState(() {
      _settings[key] = value;
    });

    try {
      await _userApi.updateUserSettings({key: value});
    } catch (e) {
      // Revert if failed
      setState(() {
        _settings[key] = !value;
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Erreur lors de la mise à jour des paramètres')),
        );
      }
    }
  }

  // Release Phase 1 — conformité App Store (guideline 5.1.1(v)) : le compte doit
  // pouvoir être supprimé/désactivé depuis l'app elle-même. Confirmation forte
  // requise avant tout appel réseau — aucune suppression en un seul tap.
  Future<void> _confirmDeleteAccount() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('حذف الحساب'),
        content: const Text(
          'سيتم حذف حسابك وإزالة بياناتك الشخصية التي لا نحتاج للاحتفاظ بها لأسباب '
          'قانونية أو مالية. قد نحتفظ ببعض السجلات المرتبطة بالمعاملات حسب المتطلبات '
          'القانونية.\n\nلا يمكن التراجع عن هذا الإجراء. هل أنت متأكد من رغبتك في المتابعة؟',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(dialogContext).pop(false),
            child: const Text('إلغاء'),
          ),
          TextButton(
            onPressed: () => Navigator.of(dialogContext).pop(true),
            child: const Text('حذف الحساب', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      await _deleteAccount();
    }
  }

  Future<void> _deleteAccount() async {
    setState(() => _isDeletingAccount = true);
    try {
      final response = await _userApi.deleteAccount();
      if (!mounted) return;

      if (response.success) {
        await AuthService().logout();
        if (!mounted) return;
        Navigator.pushAndRemoveUntil(
          context,
          MaterialPageRoute(builder: (context) => const LoginPage()),
          (route) => false,
        );
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('تم حذف حسابك بنجاح'), backgroundColor: Colors.green),
        );
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('تعذر حذف الحساب، يرجى المحاولة لاحقاً'), backgroundColor: Colors.red),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('تعذر حذف الحساب، يرجى المحاولة لاحقاً'), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _isDeletingAccount = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context)!;
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      appBar: AppBar(
        title: Text(localizations.text_44), // Settings
      ),
      body: _isLoading 
        ? const Center(child: CircularProgressIndicator())
        : ListView(
            children: [
              _buildSectionHeader('Notifications'),
              _buildSwitchTile(
                'Notifications Push', 
                'Recevoir des alertes sur votre téléphone', 
                'push_notifications',
                Icons.notifications_active_outlined
              ),
              _buildSwitchTile(
                'E-mail', 
                'Recevoir des mises à jour par e-mail', 
                'email_notifications',
                Icons.email_outlined
              ),
              _buildSwitchTile(
                'SMS', 
                'Recevoir des alertes critiques par SMS', 
                'sms_notifications',
                Icons.sms_outlined
              ),
              
              const Divider(),
              _buildSectionHeader('Confidentialité'),
              _buildSwitchTile(
                'Profil Public', 
                'Permettre aux autres de voir votre profil', 
                'public_profile',
                Icons.person_outline
              ),
              _buildSwitchTile(
                'Historique des Enchères', 
                'Afficher vos enchères passées sur votre profil', 
                'show_bid_history',
                Icons.history
              ),

              const Divider(),
              _buildSectionHeader('Application'),
              ListTile(
                leading: const Icon(Icons.language, color: Color(0xFF0081FF)),
                title: const Text('Langue'),
                subtitle: Text(Localizations.localeOf(context).languageCode.toUpperCase()),
                trailing: const Icon(Icons.arrow_forward_ios, size: 16),
                onTap: () => AppModals.showLanguageModal(context),
              ),
              ListTile(
                leading: const Icon(Icons.dark_mode_outlined, color: Color(0xFF0081FF)),
                title: const Text('Thème sombre'),
                trailing: Switch(
                  value: isDarkMode,
                  onChanged: (val) {
                    // This normally would involve a ThemeProvider change
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('Le changement de thème est géré par les paramètres système')),
                    );
                  },
                ),
              ),

              const Divider(),
              _buildSectionHeader('الحساب'),
              ListTile(
                leading: _isDeletingAccount
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(strokeWidth: 2, color: Colors.red),
                      )
                    : const Icon(Icons.delete_outline, color: Colors.red),
                title: const Text('حذف الحساب', style: TextStyle(color: Colors.red, fontWeight: FontWeight.w500)),
                subtitle: const Text('حذف حسابك بشكل نهائي', style: TextStyle(fontSize: 12)),
                onTap: _isDeletingAccount ? null : _confirmDeleteAccount,
              ),

              const SizedBox(height: 32),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Text(
                  'Version 2.0.1 (Stable)',
                  style: TextStyle(color: Colors.grey[400], fontSize: 12),
                  textAlign: TextAlign.center,
                ),
              ),
            ],
          ),
    );
  }

  Widget _buildSectionHeader(String title) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 24, 16, 8),
      child: Text(
        title,
        style: const TextStyle(
          fontSize: 14,
          fontWeight: FontWeight.bold,
          color: Color(0xFF0081FF),
          letterSpacing: 1.2,
        ),
      ),
    );
  }

  Widget _buildSwitchTile(String title, String subtitle, String key, IconData icon) {
    return SwitchListTile(
      secondary: Icon(icon, color: Colors.grey),
      title: Text(title, style: const TextStyle(fontWeight: FontWeight.w500)),
      subtitle: Text(subtitle, style: TextStyle(color: Colors.grey[600], fontSize: 12)),
      value: _settings[key] ?? false,
      onChanged: (val) => _updateSetting(key, val),
      activeThumbColor: const Color(0xFF0081FF),
    );
  }
}
