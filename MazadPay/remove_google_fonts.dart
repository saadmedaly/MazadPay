import 'dart:io';

void main() {
  final dir = Directory('lib');
  List<FileSystemEntity> files = dir.listSync(recursive: true);

  for (var entity in files) {
    if (entity is File && entity.path.endsWith('.dart')) {
      String content = entity.readAsStringSync();
      bool modified = false;

      // Replace GoogleFonts.plusJakartaSans(...) with TextStyle(fontFamily: 'Plus Jakarta Sans', ...)
      if (content.contains('GoogleFonts.plusJakartaSans')) {
        content = content.replaceAll(
          RegExp(r'GoogleFonts\.plusJakartaSans\('),
          "TextStyle(fontFamily: 'Plus Jakarta Sans', ",
        );
        modified = true;
      }

      // Replace GoogleFonts.poppins(...)
      if (content.contains('GoogleFonts.poppins')) {
        content = content.replaceAll(
          RegExp(r'GoogleFonts\.poppins\('),
          "TextStyle(fontFamily: 'Poppins', ",
        );
        modified = true;
      }
      
      // Replace GoogleFonts.roboto(...)
      if (content.contains('GoogleFonts.roboto')) {
        content = content.replaceAll(
          RegExp(r'GoogleFonts\.roboto\('),
          "TextStyle(fontFamily: 'Plus Jakarta Sans', ", // replace roboto with PJS
        );
        modified = true;
      }

      // Remove import 'package:google_fonts/google_fonts.dart';
      if (content.contains('import \'package:google_fonts/google_fonts.dart\';')) {
        content = content.replaceAll(
          'import \'package:google_fonts/google_fonts.dart\';',
          '',
        );
        modified = true;
      }

      if (modified) {
        entity.writeAsStringSync(content);
        print('Fixed GoogleFonts in: ${entity.path}');
      }
    }
  }
}
