import 'dart:io';

void main() async {
  print('Running dart analyze...');
  var result = await Process.run('dart', ['analyze']);
  String output = result.stdout.toString() + '\n' + result.stderr.toString();
  
  var lines = output.split('\n');
  var exp = RegExp(r"error - (lib[\\/].*?\.dart):(\d+):\d+ - (Invalid constant|The values in a const|Not a constant|A value of type)");
  
  Map<String, Set<int>> fileLines = {};
  
  for (var line in lines) {
    var match = exp.firstMatch(line);
    if (match != null) {
      String file = match.group(1)!;
      int lineNum = int.parse(match.group(2)!);
      fileLines.putIfAbsent(file, () => {}).add(lineNum);
    }
  }

  for (var file in fileLines.keys) {
     File f = File(file);
     if (f.existsSync()) {
        List<String> contentLines = f.readAsLinesSync();
        var linesToFix = fileLines[file]!.toList();
        linesToFix.sort();
        
        for (int l in linesToFix) {
           int idx = l - 1; // 0-indexed
           // search upwards for up to 5 lines for a 'const ', and remove it
           for (int j = idx; j >= 0 && j > idx - 5; j--) {
              if (contentLines[j].contains('const ')) {
                 contentLines[j] = contentLines[j].replaceFirst('const ', '');
                 break;
              }
           }
        }
        f.writeAsStringSync(contentLines.join('\n'));
        print('Fixed constants in $file');
     }
  }
}
