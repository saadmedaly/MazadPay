import 'dart:io';
import 'dart:convert';

void main() {
  final libDir = Directory('lib');
  if (!libDir.existsSync()) return;

  final regExp = RegExp(r"'([^\n']*?[\u0600-\u06FF]+[^\n']*?)'");
  final regExpDoubleQuote = RegExp(r'"([^\n"]*?[\u0600-\u06FF]+[^\n"]*?)"');

  Map<String, String> translationMap = {};
  int keyCounter = 1;

  for (final entity in libDir.listSync(recursive: true)) {
    if (entity is File && entity.path.endsWith('.dart')) {
      String content = entity.readAsStringSync();
      bool modified = false;

      // Extract single quotes
      content = content.replaceAllMapped(regExp, (match) {
        String arabicText = match.group(1)!;
        
        // Skip strings with interpolation for now to avoid breaking logic
        if (arabicText.contains(r'$')) {
          return match.group(0)!;
        }

        String mapKey = '';
        var entry = translationMap.entries.where((e) => e.value == arabicText).toList();
        if (entry.isNotEmpty) {
           mapKey = entry.first.key;
        } else {
           mapKey = "text_$keyCounter";
           translationMap[mapKey] = arabicText;
           keyCounter++;
        }
        
        modified = true;
        return "AppLocalizations.of(context)!.$mapKey";
      });

      // Extract double quotes
      content = content.replaceAllMapped(regExpDoubleQuote, (match) {
        String arabicText = match.group(1)!;
        
        if (arabicText.contains(r'$')) {
          return match.group(0)!;
        }

        String mapKey = '';
        var entry = translationMap.entries.where((e) => e.value == arabicText).toList();
        if (entry.isNotEmpty) {
           mapKey = entry.first.key;
        } else {
           mapKey = "text_$keyCounter";
           translationMap[mapKey] = arabicText;
           keyCounter++;
        }
        
        modified = true;
        return "AppLocalizations.of(context)!.$mapKey";
      });

      if (modified) {
        // We also need to add import 'package:flutter_gen/gen_l10n/app_localizations.dart'; at the top of the file
        if (!content.contains('app_localizations.dart')) {
          content = "import 'package:flutter_gen/gen_l10n/app_localizations.dart';\n" + content;
        }
        entity.writeAsStringSync(content);
      }
    }
  }

  // Generate ARB map
  Map<String, String> finalArb = {
     "@@locale": "ar",
  };
  finalArb.addAll(translationMap);

  File('lib/l10n/app_ar.arb').writeAsStringSync(JsonEncoder.withIndent('  ').convert(finalArb));

  print('Extracted ${translationMap.length} strings.');
}
