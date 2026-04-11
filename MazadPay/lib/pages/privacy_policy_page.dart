import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class PrivacyPolicyPage extends StatelessWidget {
  const PrivacyPolicyPage({super.key});

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
            AppLocalizations.of(context)!.text_284,
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
               _buildSection(AppLocalizations.of(context)!.text_285, AppLocalizations.of(context)!.text_286),
               const SizedBox(height: 24),
               _buildSection(AppLocalizations.of(context)!.text_287, AppLocalizations.of(context)!.text_288),
               const SizedBox(height: 24),
               _buildSection(AppLocalizations.of(context)!.text_289, AppLocalizations.of(context)!.text_290),
               const SizedBox(height: 24),
               _buildSection(AppLocalizations.of(context)!.text_291, AppLocalizations.of(context)!.text_292),
               const SizedBox(height: 24),
               _buildSection(AppLocalizations.of(context)!.text_293, AppLocalizations.of(context)!.text_294),
               const SizedBox(height: 40),
               Center(
                 child: Text(
                   AppLocalizations.of(context)!.text_295,
                   style: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 12),
                 ),
               ),
               const SizedBox(height: 24),
            ],
          ),
        ),
      );
  }

  Widget _buildSection(String title, String content) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
         Text(
           title,
           style: GoogleFonts.plusJakartaSans(fontSize: 18, fontWeight: FontWeight.bold, color: const Color(0xFF0081FF)),
         ),
         const SizedBox(height: 12),
         Text(
           content,
           style: GoogleFonts.plusJakartaSans(fontSize: 14, height: 1.6, color: Colors.grey[800]),
           textAlign: TextAlign.justify,
         ),
      ],
    );
  }
}
