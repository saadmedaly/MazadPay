import 'dart:io';
import 'dart:convert';

void main() {
  final filesToRevert = [
    'lib/pages/delivery_details_page.dart',
    'lib/pages/login_controller.dart',
    'lib/pages/notifications_page.dart',
    'lib/providers/auction_provider.dart'
  ];
  
  final arbContent = File('lib/l10n/app_ar.arb').readAsStringSync();
  final arbMap = jsonDecode(arbContent) as Map;

  for (String filePath in filesToRevert) {
    File file = File(filePath);
    if (!file.existsSync()) continue;
    
    String content = file.readAsStringSync();
    
    // Replace AppLocalizations.of(context)!.text_X back to its value
    final exp = RegExp(r"AppLocalizations\.of\(context\)!\.(text_\d+)");
    content = content.replaceAllMapped(exp, (match) {
       String key = match.group(1)!;
       String arabicText = arbMap[key] ?? '';
       // wrap it back in quotes
       return "'$arabicText'";
    });

    file.writeAsStringSync(content);
    print('Fixed context issue in: $filePath');
  }
}
