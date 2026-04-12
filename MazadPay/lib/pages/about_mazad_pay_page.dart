import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class AboutMazadPayPage extends StatelessWidget {
  const AboutMazadPayPage({super.key});

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    
    return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : Colors.white,
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          leading: IconButton(
            icon: Icon(
              Icons.arrow_back_ios,
              color: isDarkMode ? Colors.white : Colors.black,
            ),
            onPressed: () => Navigator.of(context).pop(),
          ),
          title: Text(
            AppLocalizations.of(context)!.text_4,
            style: TextStyle(
              fontFamily: 'Plus Jakarta Sans',
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: isDarkMode ? Colors.white : Colors.black,
            ),


          ),
          centerTitle: true,
        ),
        body: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _buildSectionTitle(AppLocalizations.of(context)!.text_4, isDarkMode),
              const SizedBox(height: 12),
              _buildParagraph(
                AppLocalizations.of(context)!.text_5,
                isDarkMode,
              ),
              const SizedBox(height: 24),
              
              _buildSectionTitle(AppLocalizations.of(context)!.text_6, isDarkMode),
              const SizedBox(height: 12),
              _buildParagraph(
                AppLocalizations.of(context)!.text_7,
                isDarkMode,
              ),
              const SizedBox(height: 12),
              _buildBulletPoint(
                AppLocalizations.of(context)!.text_8,
                isDarkMode,
              ),
              _buildBulletPoint(
                AppLocalizations.of(context)!.text_9,
                isDarkMode,
              ),
              _buildBulletPoint(
                AppLocalizations.of(context)!.text_10,
                isDarkMode,
              ),
              _buildBulletPoint(
                AppLocalizations.of(context)!.text_11,
                isDarkMode,
              ),
              _buildBulletPoint(
                AppLocalizations.of(context)!.text_12,
                isDarkMode,
              ),
              _buildBulletPoint(
                AppLocalizations.of(context)!.text_13,
                isDarkMode,
              ),
              _buildBulletPoint(
                AppLocalizations.of(context)!.text_14,
                isDarkMode,
              ),
              
              const SizedBox(height: 32),
              _buildSectionTitle(AppLocalizations.of(context)!.text_15, isDarkMode),
              const SizedBox(height: 12),
              _buildParagraph(
                AppLocalizations.of(context)!.text_16,
                isDarkMode,
              ),
              
              const SizedBox(height: 32),
              _buildSectionTitle(AppLocalizations.of(context)!.text_17, isDarkMode),
              const SizedBox(height: 12),
              _buildParagraph(
                AppLocalizations.of(context)!.text_18,
                isDarkMode,
              ),
              const SizedBox(height: 40),
            ],
          ),
        ),
      );
  }

  Widget _buildSectionTitle(String title, bool isDarkMode) {
    return Text(
      title,
      style: TextStyle(
        fontFamily: 'Plus Jakarta Sans',
        fontSize: 20,
        fontWeight: FontWeight.w800,
        color: isDarkMode ? Colors.white : Colors.black,
      ),


    );
  }

  Widget _buildParagraph(String text, bool isDarkMode) {
    return Text(
      text,
      textAlign: TextAlign.justify,
      style: TextStyle(
        fontFamily: 'Plus Jakarta Sans',
        fontSize: 15,
        height: 1.6,
        color: isDarkMode ? Colors.white70 : Colors.black87,
      ),

    );
  }

  Widget _buildBulletPoint(String text, bool isDarkMode) {
    return Padding(
      padding: const EdgeInsetsDirectional.only(bottom: 12.0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            margin: const EdgeInsetsDirectional.only(top: 8, start: 12),
            width: 6,
            height: 6,
            decoration: BoxDecoration(
              color: isDarkMode ? Colors.white38 : Colors.black38,
              shape: BoxShape.circle,
            ),
          ),
          Expanded(
            child: Text(
              text,
              style: TextStyle(
                fontFamily: 'Plus Jakarta Sans',
                fontSize: 15,
                height: 1.5,
                color: isDarkMode ? Colors.white70 : Colors.black87,
              ),


            ),
          ),
        ],
      ),
    );
  }
}
