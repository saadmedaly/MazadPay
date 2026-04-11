import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';

class AppModals {
  static void showLanguageModal(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    
    showModalBottomSheet(
      context: context,
      backgroundColor: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(25)),
      ),
      builder: (context) {
        return SingleChildScrollView(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 24),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    AppLocalizations.of(context)!.text_210,
                    style: TextStyle(
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 24),
                  _buildLanguageOption(
                    context,
                    title: AppLocalizations.of(context)!.text_47,
                    flag: '🇸🇦',
                    isSelected: true,
                    onTap: () => Navigator.pop(context),
                    isDarkMode: isDarkMode,
                  ),
                  _buildLanguageOption(
                    context,
                    title: 'Français',
                    flag: '🇫🇷',
                    onTap: () => Navigator.pop(context),
                    isDarkMode: isDarkMode,
                  ),
                  _buildLanguageOption(
                    context,
                    title: 'English',
                    flag: '🇺🇸',
                    onTap: () => Navigator.pop(context),
                    isDarkMode: isDarkMode,
                  ),
                  const SizedBox(height: 8),
                ],
              ),
            ),
          );
      },
    );
  }

  static Widget _buildLanguageOption(
    BuildContext context, {
    required String title,
    required String flag,
    bool isSelected = false,
    required VoidCallback onTap,
    required bool isDarkMode,
  }) {
    return Padding(
      padding: const EdgeInsetsDirectional.only(bottom: 12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
          decoration: BoxDecoration(
            color: isSelected 
              ? const Color(0xFF0084FF).withOpacity(0.1) 
              : (isDarkMode ? Colors.white.withOpacity(0.05) : Colors.grey[100]),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: isSelected ? const Color(0xFF0084FF) : Colors.transparent,
              width: 1.5,
            ),
          ),
          child: Row(
            children: [
              Text(flag, style: const TextStyle(fontSize: 24)),
              const SizedBox(width: 16),
              Text(
                title,
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                  color: isSelected ? const Color(0xFF0084FF) : (isDarkMode ? Colors.white : Colors.black),
                ),
              ),
              const Spacer(),
              if (isSelected)
                const Icon(Icons.check_circle, color: Color(0xFF0084FF)),
            ],
          ),
        ),
      ),
    );
  }

  static void showRatingModal(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    int rating = 0;

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(25)),
      ),
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setState) {
            return SingleChildScrollView(
                child: Padding(
                  padding: EdgeInsetsDirectional.only(
                    start: 24,
                    end: 24,
                    top: 32,
                    bottom: MediaQuery.of(context).viewInsets.bottom + 32,
                  ),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Container(
                        width: 80,
                        height: 80,
                        decoration: BoxDecoration(
                          color: Colors.amber.withOpacity(0.1),
                          shape: BoxShape.circle,
                        ),
                        child: const Icon(Icons.star, color: Colors.amber, size: 40),
                      ),
                      const SizedBox(height: 24),
                      Text(
                        AppLocalizations.of(context)!.text_360,
                        style: TextStyle(
                          fontSize: 22,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(height: 12),
                      Text(
                        AppLocalizations.of(context)!.text_361,
                        textAlign: TextAlign.center,
                        style: TextStyle(color: Colors.grey),
                      ),
                      const SizedBox(height: 32),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: List.generate(5, (index) {
                          return IconButton(
                            onPressed: () => setState(() => rating = index + 1),
                            icon: Icon(
                              index < rating ? Icons.star : Icons.star_border,
                              color: Colors.amber,
                              size: 40,
                            ),
                          );
                        }),
                      ),
                      const SizedBox(height: 32),
                      TextField(
                        maxLines: 3,
                        textAlign: TextAlign.right,
                        decoration: InputDecoration(
                          hintText: AppLocalizations.of(context)!.text_362,
                          hintStyle: const TextStyle(fontSize: 14),
                          filled: true,
                          fillColor: isDarkMode ? Colors.white.withOpacity(0.05) : Colors.grey[500]!.withOpacity(0.1),
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(12),
                            borderSide: BorderSide.none,
                          ),
                        ),
                      ),
                      const SizedBox(height: 32),
                      SizedBox(
                        width: double.infinity,
                        height: 54,
                        child: ElevatedButton(
                          onPressed: () => Navigator.pop(context),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: const Color(0xFF0084FF),
                            foregroundColor: Colors.white,
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(12),
                            ),
                          ),
                          child: Text(
                            AppLocalizations.of(context)!.text_363,
                            style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              );
          },
        );
      },
    );
  }

  static void showContactModal(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    showModalBottomSheet(
      context: context,
      backgroundColor: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(25)),
      ),
      builder: (context) {
        return Padding(
            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Drag handle
                Container(
                  width: 40,
                  height: 4,
                  decoration: BoxDecoration(
                    color: Colors.grey[300],
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
                const SizedBox(height: 24),
                
                Text(
                  AppLocalizations.of(context)!.text_224,
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 32),
                
                _buildContactItem(
                  context,
                  icon: FontAwesomeIcons.whatsapp,
                  isWhatsApp: true,
                  title: AppLocalizations.of(context)!.text_225,
                  subtitle: '47601175',
                  iconColor: const Color(0xFF25D366),
                  bgColor: const Color(0xFFE8F5E9),
                  isDarkMode: isDarkMode,
                  onTap: () {
                    // Action for WhatsApp
                  },
                ),
                const SizedBox(height: 16),
                
                _buildContactItem(
                  context,
                  icon: Icons.email_outlined,
                  isWhatsApp: false,
                  title: AppLocalizations.of(context)!.text_41,
                  subtitle: 'mazadpay@gmail.com',
                  iconColor: const Color(0xFF0084FF),
                  bgColor: const Color(0xFFE3F2FD),
                  isDarkMode: isDarkMode,
                  onTap: () {
                    // Action for Email
                  },
                ),
                const SizedBox(height: 32),
              ],
            ),
          );
      },
    );
  }

  static Widget _buildContactItem(
    BuildContext context, {
    required dynamic icon, // Can be IconData
    required bool isWhatsApp,
    required String title,
    required String subtitle,
    required Color iconColor,
    required Color bgColor,
    required bool isDarkMode,
    required VoidCallback onTap,
  }) {
    // In order to not depend on font_awesome_flutter in this file if not needed, 
    // we use a generic icon or we import it. I will simply use standard icons.
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: isDarkMode ? Colors.white.withOpacity(0.05) : const Color(0xFFF9FAFB),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: isDarkMode ? Colors.transparent : Colors.grey.withOpacity(0.1),
          ),
        ),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: bgColor,
                borderRadius: BorderRadius.circular(12),
              ),
              child: isWhatsApp 
                ? FaIcon(icon as IconData, color: iconColor, size: 24)
                : Icon(icon as IconData, color: iconColor, size: 24),
            ),
            const SizedBox(width: 16),
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: const TextStyle(
                    color: Colors.grey,
                    fontSize: 12,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  subtitle,
                  style: TextStyle(
                    color: isDarkMode ? Colors.white : Colors.black,
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
