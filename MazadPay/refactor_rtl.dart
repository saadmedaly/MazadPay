import 'dart:io';

void main() {
  final libDir = Directory('lib');
  if (!libDir.existsSync()) {
    print('lib directory not found');
    return;
  }

  int modifiedFilesCount = 0;

  for (final entity in libDir.listSync(recursive: true)) {
    if (entity is File && entity.path.endsWith('.dart')) {
      String content = entity.readAsStringSync();
      bool modified = false;

      // 1. Remove manual Directionality
      int idx = content.indexOf('return Directionality(');
      while (idx != -1) {
        // Confirm it has textDirection: TextDirection.rtl and child:
        String substring = content.substring(idx, idx + 150);
        if (substring.contains('TextDirection.rtl') && substring.contains('child:')) {
          // Find the matching close parenthesis for Directionality
          int openParen = content.indexOf('(', idx);
          int count = 1;
          int closeParen = openParen + 1;
          while (count > 0 && closeParen < content.length) {
            if (content[closeParen] == '(') count++;
            if (content[closeParen] == ')') count--;
            closeParen++;
          }
          
          closeParen--; // The actual index of ')'

          // Extract the child
          int childIdx = content.indexOf('child:', idx);
          int childStart = childIdx + 6;
          while (content[childStart] == ' ' || content[childStart] == '\n' || content[childStart] == '\r') {
            childStart++;
          }

          String childContent = content.substring(childStart, closeParen).trim();
          
          // There might be a trailing comma
          if (childContent.endsWith(',')) {
            childContent = childContent.substring(0, childContent.length - 1);
          }

          // Replace
          String before = content.substring(0, idx);
          String after = content.substring(closeParen + 1);
          
          // Maintain semicolon if it was there
          bool hasSemicolon = after.startsWith(';') || after.trimLeft().startsWith(';');
          content = before + 'return ' + childContent + (hasSemicolon ? ';' : '') + after.replaceFirst(RegExp(r'^\s*;'), '');
          modified = true;
          idx = content.indexOf('return Directionality(');
        } else {
          idx = content.indexOf('return Directionality(', idx + 1);
        }
      }
      
      // Also catch non-return Directionality if it's the root of a dialog
      // Directionality(...) without return
      idx = content.indexOf('Directionality(');
      while (idx != -1) {
         if (idx > 0 && content.substring(idx - 7, idx) == 'return ') {
            idx = content.indexOf('Directionality(', idx + 1);
            continue;
         }
         String substring = content.substring(idx, idx + 150 < content.length ? idx + 150 : content.length);
         if (substring.contains('TextDirection.rtl') && substring.contains('child:')) {
            int openParen = content.indexOf('(', idx);
            int count = 1;
            int closeParen = openParen + 1;
            while (count > 0 && closeParen < content.length) {
              if (content[closeParen] == '(') count++;
              if (content[closeParen] == ')') count--;
              closeParen++;
            }
            closeParen--;
            int childIdx = content.indexOf('child:', idx);
            int childStart = childIdx + 6;
            while (content[childStart] == ' ' || content[childStart] == '\n' || content[childStart] == '\r') {
              childStart++;
            }
            String childContent = content.substring(childStart, closeParen).trim();
            if (childContent.endsWith(',')) childContent = childContent.substring(0, childContent.length - 1);
            String before = content.substring(0, idx);
            String after = content.substring(closeParen + 1);
            content = before + childContent + after;
            modified = true;
            idx = content.indexOf('Directionality(');
         } else {
            idx = content.indexOf('Directionality(', idx + 1);
         }
      }

      // 2. Fix EdgeInsets.only(left: x, right: y) to EdgeInsetsDirectional.only(start: x, end: y)
      content = content.replaceAll('EdgeInsets.only(', 'EdgeInsetsDirectional.only(');
      content = content.replaceAll(RegExp(r'\bleft\s*:'), 'start:');
      content = content.replaceAll(RegExp(r'\bright\s*:'), 'end:');
      
      // 3. Fix Positioned(left: x, right: y)
      content = content.replaceAll(RegExp(r'\bPositioned\((\s*)start:'), r'Positioned.directional(textDirection: Directionality.of(context),$1start:');
      
      // 4. Alignment
      content = content.replaceAll('Alignment.centerLeft', 'AlignmentDirectional.centerStart');
      content = content.replaceAll('Alignment.centerRight', 'AlignmentDirectional.centerEnd');
      content = content.replaceAll('Alignment.bottomLeft', 'AlignmentDirectional.bottomStart');
      content = content.replaceAll('Alignment.bottomRight', 'AlignmentDirectional.bottomEnd');
      content = content.replaceAll('Alignment.topLeft', 'AlignmentDirectional.topStart');
      content = content.replaceAll('Alignment.topRight', 'AlignmentDirectional.topEnd');

      // 5. BorderRadius.only(topLeft/bottomLeft -> topStart/bottomStart)
      content = content.replaceAll('BorderRadius.only(', 'BorderRadiusDirectional.only(');
      content = content.replaceAll(RegExp(r'\btopLeft\s*:'), 'topStart:');
      content = content.replaceAll(RegExp(r'\btopRight\s*:'), 'topEnd:');
      content = content.replaceAll(RegExp(r'\bbottomLeft\s*:'), 'bottomStart:');
      content = content.replaceAll(RegExp(r'\bbottomRight\s*:'), 'bottomEnd:');

      // Check if actually modified by regex
      String original = entity.readAsStringSync();
      if (original != content) {
        entity.writeAsStringSync(content);
        modifiedFilesCount++;
        print('Modified ${entity.path}');
      }
    }
  }

  print('Finished modifying $modifiedFilesCount files.');
}
