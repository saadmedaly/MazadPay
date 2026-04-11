import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class SupportPage extends StatelessWidget {
  const SupportPage({super.key});

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          centerTitle: true,
          title: Text(
            AppLocalizations.of(context)!.text_309,
            style: GoogleFonts.plusJakartaSans(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: isDarkMode ? Colors.white : Colors.black,
            ),
          ),
          leading: IconButton(
            icon: Icon(Icons.arrow_back_ios, color: isDarkMode ? Colors.white : Colors.black, size: 20),
            onPressed: () => Navigator.of(context).pop(),
          ),
        ),
        body: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                AppLocalizations.of(context)!.text_310,
                style: GoogleFonts.plusJakartaSans(fontSize: 24, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),
              Text(
                AppLocalizations.of(context)!.text_311,
                style: GoogleFonts.plusJakartaSans(color: Colors.grey[600]),
              ),
              const SizedBox(height: 32),
              _buildContactTile(
                icon: Icons.chat_bubble_outline,
                title: AppLocalizations.of(context)!.text_312,
                subtitle: AppLocalizations.of(context)!.text_313,
                color: const Color(0xFF25D366),
                isDarkMode: isDarkMode,
              ),
              const SizedBox(height: 16),
              _buildContactTile(
                icon: Icons.phone_in_talk_outlined,
                title: AppLocalizations.of(context)!.text_314,
                subtitle: '+222 36 60 11 75',
                color: const Color(0xFF0081FF),
                isDarkMode: isDarkMode,
              ),
              const SizedBox(height: 16),
              _buildContactTile(
                icon: Icons.alternate_email,
                title: AppLocalizations.of(context)!.text_41,
                subtitle: 'support@mazadpay.mr',
                color: Colors.orange,
                isDarkMode: isDarkMode,
              ),
              const SizedBox(height: 40),
              Text(
                AppLocalizations.of(context)!.text_207,
                style: GoogleFonts.plusJakartaSans(fontSize: 18, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 16),
              _buildFaqItem(AppLocalizations.of(context)!.text_315, isDarkMode),
              _buildFaqItem(AppLocalizations.of(context)!.text_316, isDarkMode),
              _buildFaqItem(AppLocalizations.of(context)!.text_317, isDarkMode),
            ],
          ),
        ),
      );
  }

  Widget _buildContactTile({required IconData icon, required String title, required String subtitle, required Color color, required bool isDarkMode}) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.grey.withOpacity(0.1)),
        boxShadow: [
          BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 10, offset: const Offset(0, 4)),
        ],
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(color: color.withOpacity(0.1), shape: BoxShape.circle),
            child: Icon(icon, color: color, size: 24),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(title, style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.bold, fontSize: 16)),
                const SizedBox(height: 4),
                Text(subtitle, style: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 12)),
              ],
            ),
          ),
          const Icon(Icons.arrow_forward_ios, size: 16, color: Colors.grey),
        ],
      ),
    );
  }

  Widget _buildFaqItem(String question, bool isDarkMode) {
    return Container(
      margin: const EdgeInsetsDirectional.only(bottom: 12),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(12),
      ),
      child: ListTile(
        title: Text(question, style: GoogleFonts.plusJakartaSans(fontSize: 14, fontWeight: FontWeight.w600)),
        trailing: const Icon(Icons.add, size: 20, color: Color(0xFF0081FF)),
        onTap: () {},
      ),
    );
  }
}
