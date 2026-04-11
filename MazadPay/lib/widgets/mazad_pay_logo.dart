import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:flutter/material.dart';

class MazadPayLogo extends StatelessWidget {
  final double fontSize;
  final double arabicFontSize;

  const MazadPayLogo({super.key, this.fontSize = 42, this.arabicFontSize = 24});

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    Color primaryBlack = isDarkMode ? Colors.white : Colors.black;
    Color primaryBlue = const Color(0xFF135BEC);

    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        Directionality(
          textDirection: TextDirection.ltr,
          child: Row(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.baseline,
            textBaseline: TextBaseline.alphabetic,
            children: [
              // Half Black Half Blue 'M'
              ShaderMask(
                shaderCallback: (Rect bounds) {
                  return LinearGradient(
                    colors: [primaryBlack, primaryBlue],
                    stops: const [
                      0.46,
                      0.46,
                    ], // Split point roughly in the middle
                    begin: AlignmentDirectional.centerStart,
                    end: AlignmentDirectional.centerEnd,
                  ).createShader(bounds);
                },
                blendMode: BlendMode.srcIn,
                child: Text(
                  'M',
                  style: TextStyle(
                    fontSize: fontSize * 1.35,
                    fontWeight: FontWeight.w900,
                    fontStyle: FontStyle.italic,
                    fontFamily:
                        'Roboto', // Ensures the 'M' is very straight-line sans-serif
                    height: 1.0,
                    letterSpacing: -2,
                  ),
                ),
              ),
              // "azad" part
              Text(
                'azad',
                style: TextStyle(
                  fontSize: fontSize,
                  fontWeight: FontWeight.w900,
                  color: primaryBlack,
                  letterSpacing: -1.5,
                ),
              ),
              // "Pay" part
              Text(
                'Pay',
                style: TextStyle(
                  fontSize: fontSize,
                  fontWeight: FontWeight.w900,
                  color: primaryBlue,
                  letterSpacing: -1.5,
                ),
              ),
            ],
          ),
        ),
        // The Arabic part: AppLocalizations.of(context)!.text_145
        Transform.translate(
          offset: const Offset(
            0,
            -10,
          ), // Tweak offset to align properly under the English text
          child: RichText(
            textDirection: TextDirection.rtl,
            text: TextSpan(
              children: [
                TextSpan(
                  text: AppLocalizations.of(context)!.text_375,
                  style: TextStyle(
                    fontSize: arabicFontSize,
                    fontWeight: FontWeight.w900,
                    color: primaryBlack,
                  ),
                ),
                TextSpan(
                  text: AppLocalizations.of(context)!.text_376,
                  style: TextStyle(
                    fontSize: arabicFontSize,
                    fontWeight: FontWeight.w900,
                    color: primaryBlue,
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }
}
